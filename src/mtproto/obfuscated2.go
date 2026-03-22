package mtproto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"syscall"
	"time"

	"github.com/daniellavrushin/b4/log"
	"golang.org/x/sys/unix"
)

const (
	obfuscatedFrameLen = 64
	connectionTagPadded = 0xdddddddd
)

type ObfuscatedConn struct {
	net.Conn
	reader cipher.Stream
	writer cipher.Stream
}

func (c *ObfuscatedConn) Read(p []byte) (int, error) {
	n, err := c.Conn.Read(p)
	if n > 0 {
		c.reader.XORKeyStream(p[:n], p[:n])
	}
	return n, err
}

func (c *ObfuscatedConn) Write(p []byte) (int, error) {
	buf := make([]byte, len(p))
	c.writer.XORKeyStream(buf, p)
	return c.Conn.Write(buf)
}

type ClientHandshakeResult struct {
	DC   int
	Conn *ObfuscatedConn
}

func AcceptObfuscated(conn net.Conn, secret *Secret) (*ClientHandshakeResult, error) {
	frame := make([]byte, obfuscatedFrameLen)
	if _, err := io.ReadFull(conn, frame); err != nil {
		return nil, fmt.Errorf("read handshake: %w", err)
	}

	decKey := deriveKey(frame[8:40], secret.Key[:])
	decIV := make([]byte, 16)
	copy(decIV, frame[40:56])
	decStream, err := newAESCTR(decKey, decIV)
	if err != nil {
		return nil, fmt.Errorf("init decrypt: %w", err)
	}

	reversed := make([]byte, 48)
	for i := 0; i < 48; i++ {
		reversed[i] = frame[55-i]
	}
	encKey := deriveKey(reversed[:32], secret.Key[:])
	encIV := make([]byte, 16)
	copy(encIV, reversed[32:48])
	encStream, err := newAESCTR(encKey, encIV)
	if err != nil {
		return nil, fmt.Errorf("init encrypt: %w", err)
	}

	decrypted := make([]byte, obfuscatedFrameLen)
	copy(decrypted, frame)
	decStream.XORKeyStream(decrypted, decrypted)

	tag := binary.LittleEndian.Uint32(decrypted[56:60])
	if tag != connectionTagPadded {
		return nil, fmt.Errorf("invalid connection tag: 0x%08x", tag)
	}

	dc := int(int16(binary.LittleEndian.Uint16(decrypted[60:62])))

	return &ClientHandshakeResult{
		DC: dc,
		Conn: &ObfuscatedConn{
			Conn:   conn,
			reader: decStream,
			writer: encStream,
		},
	}, nil
}

func DialObfuscatedDC(addr string, dc int, mark uint) (*ObfuscatedConn, error) {
	log.Debugf("MTProto dialing DC %d at %s (mark=0x%x)", dc, addr, mark)
	dialer := net.Dialer{Timeout: 30 * secondDuration}
	if mark > 0 {
		dialer.Control = func(network, address string, c syscall.RawConn) error {
			var sErr error
			if err := c.Control(func(fd uintptr) {
				sErr = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, unix.SO_MARK, int(mark))
			}); err != nil {
				return err
			}
			return sErr
		}
	}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("dial DC %d at %s: %w", dc, addr, err)
	}
	log.Debugf("MTProto TCP connected to DC %d at %s", dc, addr)

	frame := generateFrame(dc)

	encKey := frame[8:40]
	encIV := make([]byte, 16)
	copy(encIV, frame[40:56])
	encStream, err := newAESCTR(encKey, encIV)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("init encrypt: %w", err)
	}

	reversed := make([]byte, 48)
	for i := 0; i < 48; i++ {
		reversed[i] = frame[55-i]
	}
	decKey := reversed[:32]
	decIV := make([]byte, 16)
	copy(decIV, reversed[32:48])
	decStream, err := newAESCTR(decKey, decIV)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("init decrypt: %w", err)
	}

	encrypted := make([]byte, obfuscatedFrameLen)
	copy(encrypted, frame)
	encStream.XORKeyStream(encrypted, encrypted)
	copy(encrypted[8:56], frame[8:56])

	verifyKey := make([]byte, 32)
	copy(verifyKey, encrypted[8:40])
	verifyIV := make([]byte, 16)
	copy(verifyIV, encrypted[40:56])
	verifyStream, _ := newAESCTR(verifyKey, verifyIV)
	verifyBuf := make([]byte, obfuscatedFrameLen)
	copy(verifyBuf, encrypted)
	verifyStream.XORKeyStream(verifyBuf, verifyBuf)
	verifyTag := binary.LittleEndian.Uint32(verifyBuf[56:60])
	verifyDC := int16(binary.LittleEndian.Uint16(verifyBuf[60:62]))
	log.Debugf("MTProto handshake verify: tag=0x%08x dc=%d (expect tag=0x%08x dc=%d)",
		verifyTag, verifyDC, uint32(connectionTagPadded), dc)

	if tc, ok := conn.(*net.TCPConn); ok {
		tc.SetNoDelay(true)
	}
	if _, err := conn.Write(encrypted[:32]); err != nil {
		conn.Close()
		return nil, fmt.Errorf("send handshake part1: %w", err)
	}
	time.Sleep(50 * time.Millisecond)
	if _, err := conn.Write(encrypted[32:]); err != nil {
		conn.Close()
		return nil, fmt.Errorf("send handshake part2: %w", err)
	}

	return &ObfuscatedConn{
		Conn:   conn,
		reader: decStream,
		writer: encStream,
	}, nil
}

func generateFrame(dc int) []byte {
	frame := make([]byte, obfuscatedFrameLen)
	for {
		if _, err := rand.Read(frame); err != nil {
			continue
		}

		if frame[0] == 0xef {
			continue
		}
		first4 := binary.BigEndian.Uint32(frame[0:4])
		if first4 == 0x44414548 || first4 == 0x54534f50 ||
			first4 == 0x20544547 || first4 == 0x4954504f ||
			first4 == 0x02010316 || first4 == 0xdddddddd ||
			first4 == 0xeeeeeeee {
			continue
		}
		if binary.BigEndian.Uint32(frame[4:8]) == 0 {
			continue
		}
		break
	}

	binary.LittleEndian.PutUint32(frame[56:60], connectionTagPadded)
	binary.LittleEndian.PutUint16(frame[60:62], uint16(int16(dc)))
	return frame
}

func deriveKey(rawKey []byte, secret []byte) []byte {
	h := sha256.New()
	h.Write(rawKey)
	h.Write(secret)
	return h.Sum(nil)
}

func newAESCTR(key, iv []byte) (cipher.Stream, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewCTR(block, iv), nil
}
