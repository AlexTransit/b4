package sock

import (
	"encoding/binary"
	"testing"
)

func buildMinimalIPv6UDPPacket(payloadSize int) []byte {
	ipv6HdrLen := 40
	udpHdrLen := 8
	totalLen := ipv6HdrLen + udpHdrLen + payloadSize

	pkt := make([]byte, totalLen)

	pkt[0] = 0x60
	binary.BigEndian.PutUint16(pkt[4:6], uint16(udpHdrLen+payloadSize))
	pkt[6] = 17 // UDP
	pkt[7] = 64

	pkt[23] = 1
	pkt[39] = 1

	binary.BigEndian.PutUint16(pkt[ipv6HdrLen:], 12345)
	binary.BigEndian.PutUint16(pkt[ipv6HdrLen+2:], 53)
	binary.BigEndian.PutUint16(pkt[ipv6HdrLen+4:], uint16(udpHdrLen+payloadSize))

	for i := 0; i < payloadSize; i++ {
		pkt[ipv6HdrLen+udpHdrLen+i] = byte(i % 256)
	}

	udpChecksumIPv6(pkt)

	return pkt
}

func TestUdpChecksumIPv6_TooShort(t *testing.T) {
	pkt := make([]byte, 40)
	udpChecksumIPv6(pkt) // Should not panic
}

func TestBuildFakeUDPFromOriginalV6_TooShort(t *testing.T) {
	_, ok := BuildFakeUDPFromOriginalV6(make([]byte, 40), 100, 3, nil)
	if ok {
		t.Error("expected false")
	}
}

func TestBuildFakeUDPFromOriginalV6_NotIPv6(t *testing.T) {
	pkt := make([]byte, 60)
	pkt[0] = 0x45
	_, ok := BuildFakeUDPFromOriginalV6(pkt, 100, 3, nil)
	if ok {
		t.Error("expected false")
	}
}

func TestBuildFakeUDPFromOriginalV6_Valid(t *testing.T) {
	pkt := buildMinimalIPv6UDPPacket(20)
	result, ok := BuildFakeUDPFromOriginalV6(pkt, 50, 5, nil)
	if !ok {
		t.Fatal("expected success")
	}

	if result[7] != 5 {
		t.Errorf("hop limit not set: expected 5, got %d", result[7])
	}

	expectedLen := 40 + 8 + 50
	if len(result) != expectedLen {
		t.Errorf("expected len %d, got %d", expectedLen, len(result))
	}
}

func TestIPv6FragmentUDP_TooShort(t *testing.T) {
	_, ok := IPv6FragmentUDP(make([]byte, 40), 8)
	if ok {
		t.Error("expected false")
	}
}

func TestIPv6FragmentUDP_Valid(t *testing.T) {
	pkt := buildMinimalIPv6UDPPacket(100)
	frags, ok := IPv6FragmentUDP(pkt, 20)
	if !ok {
		t.Fatal("expected success")
	}
	if len(frags) != 2 {
		t.Errorf("expected 2 fragments, got %d", len(frags))
	}

	// Both fragments should have next header = 44 (Fragment)
	for i, frag := range frags {
		if frag[6] != 44 {
			t.Errorf("fragment %d: expected next header 44, got %d", i, frag[6])
		}
	}
}
