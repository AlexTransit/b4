package sock

import (
	"encoding/binary"
	"testing"
)

func buildMinimalIPv4UDPPacket(payloadSize int) []byte {
	ipHdrLen := 20
	udpHdrLen := 8
	totalLen := ipHdrLen + udpHdrLen + payloadSize

	pkt := make([]byte, totalLen)

	pkt[0] = 0x45
	binary.BigEndian.PutUint16(pkt[2:4], uint16(totalLen))
	pkt[8] = 64
	pkt[9] = 17 // UDP
	copy(pkt[12:16], []byte{192, 168, 1, 1})
	copy(pkt[16:20], []byte{10, 0, 0, 1})

	binary.BigEndian.PutUint16(pkt[ipHdrLen:], 12345)
	binary.BigEndian.PutUint16(pkt[ipHdrLen+2:], 53)
	binary.BigEndian.PutUint16(pkt[ipHdrLen+4:], uint16(udpHdrLen+payloadSize))

	for i := 0; i < payloadSize; i++ {
		pkt[ipHdrLen+udpHdrLen+i] = byte(i % 256)
	}

	FixIPv4Checksum(pkt[:ipHdrLen])

	return pkt
}

func TestBuildFakeUDPFromOriginalV4_TooShort(t *testing.T) {
	_, ok := BuildFakeUDPFromOriginalV4(make([]byte, 10), 100, 3, nil)
	if ok {
		t.Error("expected false for short packet")
	}
}

func TestBuildFakeUDPFromOriginalV4_NotIPv4(t *testing.T) {
	pkt := make([]byte, 50)
	pkt[0] = 0x60
	_, ok := BuildFakeUDPFromOriginalV4(pkt, 100, 3, nil)
	if ok {
		t.Error("expected false for non-IPv4")
	}
}

func TestBuildFakeUDPFromOriginalV4_Valid(t *testing.T) {
	pkt := buildMinimalIPv4UDPPacket(20)
	result, ok := BuildFakeUDPFromOriginalV4(pkt, 50, 3, nil)
	if !ok {
		t.Fatal("expected success")
	}

	// Check TTL
	if result[8] != 3 {
		t.Errorf("TTL not set: expected 3, got %d", result[8])
	}

	// Check length
	expectedLen := 20 + 8 + 50
	if len(result) != expectedLen {
		t.Errorf("expected len %d, got %d", expectedLen, len(result))
	}
}

func TestIPv4FragmentUDP_TooShort(t *testing.T) {
	_, ok := IPv4FragmentUDP(make([]byte, 20), 8)
	if ok {
		t.Error("expected false")
	}
}

func TestIPv4FragmentUDP_Valid(t *testing.T) {
	pkt := buildMinimalIPv4UDPPacket(100)
	frags, ok := IPv4FragmentUDP(pkt, 20)
	if !ok {
		t.Fatal("expected success")
	}
	if len(frags) != 2 {
		t.Errorf("expected 2 fragments, got %d", len(frags))
	}
}
