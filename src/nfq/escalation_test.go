package nfq

import (
	"testing"
	"time"
)

func newIPBlocker() *ipBlockTracker {
	return &ipBlockTracker{
		conns:       make(map[string]*ipBlockEntry),
		blocked:     make(map[string]time.Time),
		escalations: make(map[string]*escalationEntry),
		rstKills:    make(map[string]*rstKillEntry),
	}
}

func TestEscalation_GetMissReturnsFalse(t *testing.T) {
	b := newIPBlocker()
	if id, hops, ok := b.GetEscalation("1.2.3.4:443"); ok || id != "" || hops != 0 {
		t.Fatalf("expected miss, got id=%q hops=%d ok=%v", id, hops, ok)
	}
}

func TestEscalation_SetThenGet(t *testing.T) {
	b := newIPBlocker()
	if !b.SetEscalation("1.2.3.4:443", "set-b") {
		t.Fatal("SetEscalation should succeed on first hop")
	}
	id, hops, ok := b.GetEscalation("1.2.3.4:443")
	if !ok || id != "set-b" || hops != 1 {
		t.Fatalf("expected set-b hop=1, got id=%q hops=%d ok=%v", id, hops, ok)
	}
}

func TestEscalation_ChainIncrementsHops(t *testing.T) {
	b := newIPBlocker()
	b.SetEscalation("1.2.3.4:443", "set-b")
	b.SetEscalation("1.2.3.4:443", "set-c")
	id, hops, ok := b.GetEscalation("1.2.3.4:443")
	if !ok || id != "set-c" || hops != 2 {
		t.Fatalf("expected set-c hop=2, got id=%q hops=%d ok=%v", id, hops, ok)
	}
}

func TestEscalation_StopsAtMaxHops(t *testing.T) {
	b := newIPBlocker()
	for i := 0; i < MaxEscalationHops; i++ {
		if !b.SetEscalation("1.2.3.4:443", "set-x") {
			t.Fatalf("hop %d should still be allowed", i)
		}
	}
	if b.SetEscalation("1.2.3.4:443", "set-y") {
		t.Fatal("escalation past MaxEscalationHops must be rejected")
	}
}

func TestEscalation_Clear(t *testing.T) {
	b := newIPBlocker()
	b.SetEscalation("1.2.3.4:443", "set-b")
	b.ClearEscalation("1.2.3.4:443")
	if _, _, ok := b.GetEscalation("1.2.3.4:443"); ok {
		t.Fatal("ClearEscalation should drop the entry")
	}
}

func TestEscalation_Reset(t *testing.T) {
	b := newIPBlocker()
	b.SetEscalation("1.2.3.4:443", "set-b")
	b.SetEscalation("5.6.7.8:443", "set-c")
	b.ResetEscalations()
	if _, _, ok := b.GetEscalation("1.2.3.4:443"); ok {
		t.Fatal("ResetEscalations should drop all entries")
	}
	if _, _, ok := b.GetEscalation("5.6.7.8:443"); ok {
		t.Fatal("ResetEscalations should drop all entries")
	}
}

func TestEscalation_ExpiresAfterTTL(t *testing.T) {
	b := newIPBlocker()
	b.SetEscalation("1.2.3.4:443", "set-b")
	// Manually backdate the entry past the TTL.
	b.mu.Lock()
	b.escalations["1.2.3.4:443"].setAt = time.Now().Add(-EscalationTTL - time.Minute)
	b.mu.Unlock()

	if _, _, ok := b.GetEscalation("1.2.3.4:443"); ok {
		t.Fatal("expired escalation must not be returned")
	}
}

func TestEscalation_CleanupRemovesExpired(t *testing.T) {
	b := newIPBlocker()
	b.SetEscalation("a:1", "x")
	b.SetEscalation("b:2", "y")
	b.mu.Lock()
	b.escalations["a:1"].setAt = time.Now().Add(-EscalationTTL - time.Minute)
	b.mu.Unlock()

	b.Cleanup(0)

	b.mu.RLock()
	_, hasA := b.escalations["a:1"]
	_, hasB := b.escalations["b:2"]
	b.mu.RUnlock()
	if hasA {
		t.Fatal("Cleanup should drop expired entry a:1")
	}
	if !hasB {
		t.Fatal("Cleanup should keep fresh entry b:2")
	}
}

func TestEscalation_DoesNotInterfereWithBlockedCache(t *testing.T) {
	b := newIPBlocker()
	b.AddBlocked("9.9.9.9:443")
	b.SetEscalation("1.2.3.4:443", "set-b")

	if !b.IsBlocked("9.9.9.9:443") {
		t.Fatal("blocked IP should still be reported as blocked")
	}
	if b.IsBlocked("1.2.3.4:443") {
		t.Fatal("escalated IP must not be reported as blocked")
	}
}

func TestRSTKill_BelowThresholdReturnsFalse(t *testing.T) {
	b := newIPBlocker()
	for i := 0; i < RSTKillThreshold-1; i++ {
		if b.RecordRSTKill("1.2.3.4:443") {
			t.Fatalf("hit %d should not trip threshold (= %d)", i+1, RSTKillThreshold)
		}
	}
}

func TestRSTKill_TripsAtThreshold(t *testing.T) {
	b := newIPBlocker()
	var tripped bool
	for i := 0; i < RSTKillThreshold; i++ {
		tripped = b.RecordRSTKill("1.2.3.4:443")
	}
	if !tripped {
		t.Fatalf("threshold (%d) should have tripped", RSTKillThreshold)
	}
}

func TestRSTKill_ResetsAfterTrip(t *testing.T) {
	b := newIPBlocker()
	for i := 0; i < RSTKillThreshold; i++ {
		b.RecordRSTKill("1.2.3.4:443")
	}
	if b.RecordRSTKill("1.2.3.4:443") {
		t.Fatal("immediate next kill after trip must NOT re-fire (would spam escalations)")
	}
}

func TestRSTKill_WindowExpiry(t *testing.T) {
	b := newIPBlocker()
	b.RecordRSTKill("1.2.3.4:443")
	// Backdate the entry past the rolling window.
	b.mu.Lock()
	b.rstKills["1.2.3.4:443"].firstAt = time.Now().Add(-RSTKillWindow - time.Second)
	b.mu.Unlock()
	// Next kill is treated as a fresh start, not as count=2.
	if b.RecordRSTKill("1.2.3.4:443") {
		t.Fatal("kill after window expiry must restart counting, not trip")
	}
	b.mu.RLock()
	count := b.rstKills["1.2.3.4:443"].count
	b.mu.RUnlock()
	if count != 1 {
		t.Fatalf("expected counter reset to 1 after window expiry, got %d", count)
	}
}

func TestRSTKill_DistinctDestinationsTrackedSeparately(t *testing.T) {
	b := newIPBlocker()
	for i := 0; i < RSTKillThreshold-1; i++ {
		b.RecordRSTKill("1.2.3.4:443")
	}
	if b.RecordRSTKill("5.6.7.8:443") {
		t.Fatal("first kill on a different destination must not trip")
	}
}

func TestRSTKill_ResetEscalationsClearsKills(t *testing.T) {
	b := newIPBlocker()
	b.RecordRSTKill("1.2.3.4:443")
	b.ResetEscalations()
	b.mu.RLock()
	_, has := b.rstKills["1.2.3.4:443"]
	b.mu.RUnlock()
	if has {
		t.Fatal("ResetEscalations should also drop RST-kill state")
	}
}
