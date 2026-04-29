package tproxy

import (
	"testing"

	"github.com/google/uuid"
)

const bypassBit uint32 = 0x8000

func TestMarkForSet_NeverCollidesWithBypassBit(t *testing.T) {
	const iterations = 100_000
	for i := 0; i < iterations; i++ {
		id := uuid.NewString()
		m := MarkForSet(id, 0)
		if m&bypassBit == bypassBit {
			t.Fatalf("MarkForSet(%q) = %#x collides with bypass bit %#x", id, m, bypassBit)
		}
	}
}

func TestMarkForSet_StaysWithinDeclaredRange(t *testing.T) {
	const iterations = 10_000
	min := uint32(MarkBase)
	max := uint32(MarkBase + MarkRange)
	for i := 0; i < iterations; i++ {
		id := uuid.NewString()
		m := MarkForSet(id, 0)
		if m < min || m >= max {
			t.Fatalf("MarkForSet(%q) = %#x out of range [%#x, %#x)", id, m, min, max)
		}
	}
}

func TestMarkForSet_PinnedReturnsAsIs(t *testing.T) {
	cases := []uint32{1, 0x100, 0x12345, 0xDEADBEEF}
	for _, want := range cases {
		if got := MarkForSet("any-id", want); got != want {
			t.Fatalf("pinned mark: got %#x, want %#x", got, want)
		}
	}
}

func TestMarkForSet_DeterministicForSameID(t *testing.T) {
	id := "718e0020-ee8d-4055-851c-b99deeeb5abf"
	first := MarkForSet(id, 0)
	for i := 0; i < 1000; i++ {
		if got := MarkForSet(id, 0); got != first {
			t.Fatalf("MarkForSet not deterministic: got %#x, want %#x", got, first)
		}
	}
}

func TestMarkForSet_RegressionOriginalFailingUUID(t *testing.T) {
	id := "718e0020-ee8d-4055-851c-b99deeeb5abf"
	m := MarkForSet(id, 0)
	if m&bypassBit != 0 {
		t.Fatalf("regression: mark %#x for the original failing UUID still hits bypass bit %#x", m, bypassBit)
	}
}
