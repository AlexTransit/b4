package nfq

import (
	"encoding/binary"
	"net"
	"time"

	"github.com/daniellavrushin/b4/log"
	"github.com/daniellavrushin/b4/sock"
)

func (w *Worker) sendRSTToClientV4(raw []byte, ihl int, srcIP, dstIP net.IP) {
	tcp := raw[ihl:]
	if len(tcp) < 20 {
		return
	}

	clientPort := binary.BigEndian.Uint16(tcp[0:2])
	serverPort := binary.BigEndian.Uint16(tcp[2:4])
	clientAck := binary.BigEndian.Uint32(tcp[8:12])

	rst := make([]byte, 40)

	rst[0] = 0x45
	binary.BigEndian.PutUint16(rst[2:4], 40)
	binary.BigEndian.PutUint16(rst[4:6], uint16(time.Now().UnixNano()&0xFFFF))
	rst[8] = 64
	rst[9] = 6
	copy(rst[12:16], dstIP.To4())
	copy(rst[16:20], srcIP.To4())

	binary.BigEndian.PutUint16(rst[20:22], serverPort)
	binary.BigEndian.PutUint16(rst[22:24], clientPort)
	binary.BigEndian.PutUint32(rst[24:28], clientAck)
	rst[32] = 0x50
	rst[33] = 0x04

	sock.FixIPv4Checksum(rst[:20])
	sock.FixTCPChecksum(rst)

	if err := w.sock.SendIPv4(rst, srcIP); err != nil {
		log.Tracef("ip-block: failed to send RST to client %s:%d: %v", srcIP, clientPort, err)
	}
}

func (w *Worker) sendRSTToClientV6(raw []byte, srcIP, dstIP net.IP) {
	ipv6HdrLen := 40
	tcp := raw[ipv6HdrLen:]
	if len(tcp) < 20 {
		return
	}

	clientPort := binary.BigEndian.Uint16(tcp[0:2])
	serverPort := binary.BigEndian.Uint16(tcp[2:4])
	clientAck := binary.BigEndian.Uint32(tcp[8:12])

	rst := make([]byte, 60)

	rst[0] = 0x60
	binary.BigEndian.PutUint16(rst[4:6], 20)
	rst[6] = 6
	rst[7] = 64
	copy(rst[8:24], dstIP.To16())
	copy(rst[24:40], srcIP.To16())

	binary.BigEndian.PutUint16(rst[40:42], serverPort)
	binary.BigEndian.PutUint16(rst[42:44], clientPort)
	binary.BigEndian.PutUint32(rst[44:48], clientAck)
	rst[52] = 0x50
	rst[53] = 0x14

	sock.FixTCPChecksumV6(rst)

	if err := w.sock.SendIPv6(rst, srcIP); err != nil {
		log.Tracef("ip-block: failed to send RST to client %s:%d: %v", srcIP, clientPort, err)
	}
}
