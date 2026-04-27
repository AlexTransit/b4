package tproxy

import "hash/fnv"

const (
	DefaultPortBase = 13000
	PortRange       = 2000
	MarkBase        = 0x10000
	MarkRange       = 0xFE00
)

func MarkForSet(setID string, pinned uint32) uint32 {
	if pinned > 0 {
		return pinned
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(setID))
	return MarkBase + (h.Sum32() % MarkRange)
}

func PortFor(mark uint32) int {
	if mark == 0 {
		return DefaultPortBase
	}
	return DefaultPortBase + int(mark%PortRange)
}

