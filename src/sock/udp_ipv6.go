package sock

import (
	"encoding/binary"
)

// udpChecksumIPv6 calculates and sets the UDP checksum for IPv6 packets
func udpChecksumIPv6(pkt []byte) {
	if len(pkt) < 48 { // 40 (IPv6 header) + 8 (UDP header)
		return
	}

	ipv6HdrLen := 40
	udpOffset := ipv6HdrLen
	udpLen := int(binary.BigEndian.Uint16(pkt[udpOffset+4 : udpOffset+6]))

	// Build IPv6 pseudo-header
	pseudo := make([]byte, 40)
	copy(pseudo[0:16], pkt[8:24])   // Source address
	copy(pseudo[16:32], pkt[24:40]) // Destination address
	binary.BigEndian.PutUint32(pseudo[32:36], uint32(udpLen))
	pseudo[39] = 17 // Next header = UDP

	// Clear existing checksum
	pkt[udpOffset+6], pkt[udpOffset+7] = 0, 0

	var sum uint32

	// Sum pseudo-header
	for i := 0; i < len(pseudo); i += 2 {
		sum += uint32(binary.BigEndian.Uint16(pseudo[i : i+2]))
	}

	// Sum UDP segment
	udp := pkt[udpOffset : udpOffset+udpLen]
	for i := 0; i+1 < len(udp); i += 2 {
		sum += uint32(binary.BigEndian.Uint16(udp[i : i+2]))
	}
	if len(udp)%2 == 1 {
		sum += uint32(udp[len(udp)-1]) << 8
	}

	// Fold carries
	for sum > 0xffff {
		sum = (sum >> 16) + (sum & 0xffff)
	}

	checksum := ^uint16(sum)
	if checksum == 0 {
		checksum = 0xffff // UDP checksum of 0 means no checksum
	}
	binary.BigEndian.PutUint16(pkt[udpOffset+6:udpOffset+8], checksum)
}

func BuildFakeUDPFromOriginalV6(orig []byte, fakeLen int, hopLimit uint8, payload []byte) ([]byte, bool) {
	if len(orig) < 48 || orig[0]>>4 != 6 {
		return nil, false
	}

	ipv6HdrLen := 40
	if len(orig) < ipv6HdrLen+8 {
		return nil, false
	}

	out := make([]byte, ipv6HdrLen+8+fakeLen)

	copy(out, orig[:ipv6HdrLen])

	out[7] = hopLimit

	binary.BigEndian.PutUint16(out[4:6], uint16(8+fakeLen))

	copy(out[ipv6HdrLen:], orig[ipv6HdrLen:ipv6HdrLen+8])

	binary.BigEndian.PutUint16(out[ipv6HdrLen+4:ipv6HdrLen+6], uint16(8+fakeLen))

	if n := len(payload); n > 0 {
		copyLen := n
		if copyLen > fakeLen {
			copyLen = fakeLen
		}
		copy(out[ipv6HdrLen+8:ipv6HdrLen+8+copyLen], payload[:copyLen])
		for i := copyLen; i < fakeLen; i++ {
			out[ipv6HdrLen+8+i] = 0
		}
	} else {
		for i := 0; i < fakeLen; i++ {
			out[ipv6HdrLen+8+i] = 0
		}
	}

	udpChecksumIPv6(out)

	return out, true
}

// IPv6FragmentUDP fragments an IPv6 UDP packet
// Note: IPv6 fragmentation is handled differently than IPv4
// Fragment headers are extension headers in IPv6
func IPv6FragmentUDP(orig []byte, split int) ([][]byte, bool) {
	if len(orig) < 48 || orig[0]>>4 != 6 {
		return nil, false
	}

	ipv6HdrLen := 40
	if len(orig) < ipv6HdrLen+8 {
		return nil, false
	}

	payloadLen := int(binary.BigEndian.Uint16(orig[4:6]))
	if payloadLen < 8 || ipv6HdrLen+payloadLen > len(orig) {
		return nil, false
	}

	udp := orig[ipv6HdrLen : ipv6HdrLen+payloadLen]
	if len(udp) < 8 {
		return nil, false
	}

	payload := udp[8:]
	if split < 1 || split >= len(payload) {
		split = 8
	}

	// Align to 8-byte boundary for IPv6 fragmentation
	firstData := 8 + split
	firstDataAligned := firstData - (firstData % 8)
	if firstDataAligned < 8 {
		firstDataAligned = 8
	}
	if firstDataAligned >= len(udp) {
		return nil, false
	}

	// IPv6 uses a fragment extension header (8 bytes)
	// Format: Next Header (1) + Reserved (1) + Fragment Offset (2) + Res (1) + M flag (1) + Identification (4)
	fragHdrLen := 8

	// Generate a unique identification for this fragmented packet
	var identification uint32 = generateFragmentID()

	// First fragment: IPv6 header + Fragment header + first part of UDP
	frag1Len := ipv6HdrLen + fragHdrLen + firstDataAligned
	frag1 := make([]byte, frag1Len)

	// Copy base IPv6 header
	copy(frag1, orig[:ipv6HdrLen])
	// Change next header to Fragment (44)
	frag1[6] = 44
	// Update payload length
	binary.BigEndian.PutUint16(frag1[4:6], uint16(fragHdrLen+firstDataAligned))

	// Build fragment header
	fragHdr1 := frag1[ipv6HdrLen : ipv6HdrLen+fragHdrLen]
	fragHdr1[0] = 17                                    // Next header = UDP
	fragHdr1[1] = 0                                     // Reserved
	binary.BigEndian.PutUint16(fragHdr1[2:4], 0|0x0001) // Offset 0, M flag set
	binary.BigEndian.PutUint32(fragHdr1[4:8], identification)

	// Copy UDP data
	copy(frag1[ipv6HdrLen+fragHdrLen:], udp[:firstDataAligned])

	// Second fragment: IPv6 header + Fragment header + rest of UDP
	remainingData := udp[firstDataAligned:]
	frag2Len := ipv6HdrLen + fragHdrLen + len(remainingData)
	frag2 := make([]byte, frag2Len)

	// Copy base IPv6 header
	copy(frag2, orig[:ipv6HdrLen])
	// Change next header to Fragment (44)
	frag2[6] = 44
	// Update payload length
	binary.BigEndian.PutUint16(frag2[4:6], uint16(fragHdrLen+len(remainingData)))

	// Build fragment header
	fragHdr2 := frag2[ipv6HdrLen : ipv6HdrLen+fragHdrLen]
	fragHdr2[0] = 17                                              // Next header = UDP
	fragHdr2[1] = 0                                               // Reserved
	offsetUnits := uint16(firstDataAligned / 8)                   // Fragment offset in 8-byte units
	binary.BigEndian.PutUint16(fragHdr2[2:4], (offsetUnits << 3)) // Offset, M flag not set (last fragment)
	binary.BigEndian.PutUint32(fragHdr2[4:8], identification)

	// Copy remaining UDP data
	copy(frag2[ipv6HdrLen+fragHdrLen:], remainingData)

	// Return fragments in reverse order (common DPI evasion technique)
	return [][]byte{frag2, frag1}, true
}
