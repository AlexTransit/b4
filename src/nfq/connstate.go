package nfq

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/daniellavrushin/b4/config"
)

type connInfo struct {
	bytesIn       uint64
	threshold     uint64
	set           *config.SetConfig
	lastSeen      time.Time
	serverTTL     uint8
	ttlRecorded   bool
	responseSeen  bool
	rstCount      int
	serverHasOpts bool
}

type tlsInfo struct {
	host       string
	tlsVersion uint16
	lastSeen   time.Time
}

type tlsInfoCache struct {
	mu    sync.RWMutex
	conns map[string]*tlsInfo
}

const maxTLSCacheEntries = 20000

func (c *tlsInfoCache) Store(connKey string, host string, tlsVersion uint16) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.conns) >= maxTLSCacheEntries {
		now := time.Now()
		var oldestKey string
		var oldestTime time.Time
		for k, v := range c.conns {
			if now.Sub(v.lastSeen) > 120*time.Second {
				delete(c.conns, k)
			} else if oldestTime.IsZero() || v.lastSeen.Before(oldestTime) {
				oldestKey = k
				oldestTime = v.lastSeen
			}
		}
		if len(c.conns) >= maxTLSCacheEntries && oldestKey != "" {
			delete(c.conns, oldestKey)
		}
	}

	c.conns[connKey] = &tlsInfo{
		host:       host,
		tlsVersion: tlsVersion,
		lastSeen:   time.Now(),
	}
}

func (c *tlsInfoCache) Lookup(connKey string) (string, uint16, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	info, exists := c.conns[connKey]
	if !exists {
		return "", 0, false
	}
	return info.host, info.tlsVersion, true
}

func (c *tlsInfoCache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	for k, v := range c.conns {
		if now.Sub(v.lastSeen) > 120*time.Second {
			delete(c.conns, k)
		}
	}
}

type connStateTracker struct {
	mu    sync.RWMutex
	conns map[string]*connInfo
}

const maxConnStateEntries = 10000

type ipBlockEntry struct {
	firstSeen   time.Time
	retransmits int
	rstSent     bool
	host        string
}

type IPBlockCache interface {
	IsBlocked(dstIPPort string) bool
	AddBlocked(dstIPPort string)
}

type ipBlockTracker struct {
	mu          sync.RWMutex
	conns       map[string]*ipBlockEntry
	blocked     map[string]time.Time
	escalations map[string]*escalationEntry
	rstKills    map[string]*rstKillEntry
}

type escalationEntry struct {
	setId  string
	setAt  time.Time
	hops   int
}

type rstKillEntry struct {
	count   int
	firstAt time.Time
}

const maxIPBlockEntries = 10000
const maxIPBlockCacheEntries = 5000
const maxEscalationCacheEntries = 5000
const maxRSTKillEntries = 5000

// RSTKillThreshold is the number of suspicious RSTs to a destination
// within RSTKillWindow before escalation triggers. Independent from
// IPBlockDetect because the timing pattern is different (each RST kill
// is its own connection attempt by the client).
const RSTKillThreshold = 3
const RSTKillWindow = 30 * time.Second

// EscalationTTL is how long a per-host escalation decision is remembered.
// After expiry, b4 retries the original (non-escalated) set for that host.
const EscalationTTL = time.Hour

// MaxEscalationHops bounds how many times a host can be escalated to
// guard against accidental cycles in the EscalateTo chain.
const MaxEscalationHops = 8

func (t *ipBlockTracker) RecordClientHello(connKey, host string) (int, time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(t.conns) >= maxIPBlockEntries {
		now := time.Now()
		for k, v := range t.conns {
			if now.Sub(v.firstSeen) > 120*time.Second {
				delete(t.conns, k)
			}
		}
	}

	entry, exists := t.conns[connKey]
	if !exists {
		entry = &ipBlockEntry{
			firstSeen: time.Now(),
			host:      host,
		}
		t.conns[connKey] = entry
	}
	entry.retransmits++
	return entry.retransmits, entry.firstSeen
}

func (t *ipBlockTracker) HasRSTSent(connKey string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if entry, ok := t.conns[connKey]; ok {
		return entry.rstSent
	}
	return false
}

func (t *ipBlockTracker) MarkRSTSent(connKey string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if entry, ok := t.conns[connKey]; ok {
		entry.rstSent = true
	}
}

func (t *ipBlockTracker) IsBlocked(dstIPPort string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	_, ok := t.blocked[dstIPPort]
	if ok {
		t.blocked[dstIPPort] = time.Now()
	}
	return ok
}

func (t *ipBlockTracker) AddBlocked(dstIPPort string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.blocked) >= maxIPBlockCacheEntries {
		now := time.Now()
		for k, v := range t.blocked {
			if now.Sub(v) > 300*time.Second {
				delete(t.blocked, k)
			}
		}
	}
	t.blocked[dstIPPort] = time.Now()
}

func (t *ipBlockTracker) Cleanup(cacheTTL time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	for k, v := range t.conns {
		if now.Sub(v.firstSeen) > 120*time.Second {
			delete(t.conns, k)
		}
	}
	if cacheTTL > 0 {
		for k, v := range t.blocked {
			if now.Sub(v) > cacheTTL {
				delete(t.blocked, k)
			}
		}
	}
	for k, v := range t.escalations {
		if now.Sub(v.setAt) > EscalationTTL {
			delete(t.escalations, k)
		}
	}
	for k, v := range t.rstKills {
		if now.Sub(v.firstAt) > RSTKillWindow*2 {
			delete(t.rstKills, k)
		}
	}
}

// GetEscalation returns the escalated set id for a destination, if a
// non-expired escalation entry exists.
func (t *ipBlockTracker) GetEscalation(dstIPPort string) (setId string, hops int, ok bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	e, exists := t.escalations[dstIPPort]
	if !exists {
		return "", 0, false
	}
	if time.Since(e.setAt) > EscalationTTL {
		return "", 0, false
	}
	return e.setId, e.hops, true
}

// SetEscalation records that future connections to dstIPPort should use setId.
// Returns false if the per-host escalation chain has reached MaxEscalationHops.
func (t *ipBlockTracker) SetEscalation(dstIPPort, setId string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.escalations) >= maxEscalationCacheEntries {
		now := time.Now()
		for k, v := range t.escalations {
			if now.Sub(v.setAt) > EscalationTTL {
				delete(t.escalations, k)
			}
		}
	}
	prev := t.escalations[dstIPPort]
	hops := 1
	if prev != nil {
		hops = prev.hops + 1
	}
	if hops > MaxEscalationHops {
		return false
	}
	t.escalations[dstIPPort] = &escalationEntry{
		setId: setId,
		setAt: time.Now(),
		hops:  hops,
	}
	return true
}

// ClearEscalation removes any escalation entry for dstIPPort.
func (t *ipBlockTracker) ClearEscalation(dstIPPort string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.escalations, dstIPPort)
}

// ResetEscalations drops all escalation state. Called on config reload to
// guarantee renamed/removed sets cannot be referenced by stale entries.
func (t *ipBlockTracker) ResetEscalations() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.escalations = make(map[string]*escalationEntry)
	t.rstKills = make(map[string]*rstKillEntry)
}

// RecordRSTKill bumps the RST-kill counter for dstIPPort and reports
// whether the threshold has been reached within the rolling window.
// Returns true at most once per window: when the counter trips, the
// entry is reset so callers don't re-trigger on every subsequent RST.
func (t *ipBlockTracker) RecordRSTKill(dstIPPort string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(t.rstKills) >= maxRSTKillEntries {
		now := time.Now()
		for k, v := range t.rstKills {
			if now.Sub(v.firstAt) > RSTKillWindow*2 {
				delete(t.rstKills, k)
			}
		}
	}

	now := time.Now()
	entry, exists := t.rstKills[dstIPPort]
	if !exists || now.Sub(entry.firstAt) > RSTKillWindow {
		t.rstKills[dstIPPort] = &rstKillEntry{count: 1, firstAt: now}
		return false
	}
	entry.count++
	if entry.count >= RSTKillThreshold {
		delete(t.rstKills, dstIPPort)
		return true
	}
	return false
}

type runtimeState struct {
	tlsCache  *tlsInfoCache
	connState *connStateTracker
	ipBlocker *ipBlockTracker
}

func newRuntimeState() *runtimeState {
	return &runtimeState{
		tlsCache: &tlsInfoCache{
			conns: make(map[string]*tlsInfo),
		},
		connState: &connStateTracker{
			conns: make(map[string]*connInfo),
		},
		ipBlocker: &ipBlockTracker{
			conns:       make(map[string]*ipBlockEntry),
			blocked:     make(map[string]time.Time),
			escalations: make(map[string]*escalationEntry),
			rstKills:    make(map[string]*rstKillEntry),
		},
	}
}

func (t *connStateTracker) RegisterOutgoing(connKey string, set *config.SetConfig) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// If at capacity, evict oldest entries before adding
	if len(t.conns) >= maxConnStateEntries {
		now := time.Now()
		var oldestKey string
		var oldestTime time.Time
		for k, v := range t.conns {
			if now.Sub(v.lastSeen) > 120*time.Second {
				delete(t.conns, k)
			} else if oldestTime.IsZero() || v.lastSeen.Before(oldestTime) {
				oldestKey = k
				oldestTime = v.lastSeen
			}
		}
		// If still at capacity after removing stale entries, evict the oldest
		if len(t.conns) >= maxConnStateEntries && oldestKey != "" {
			delete(t.conns, oldestKey)
		}
	}

	if existing, ok := t.conns[connKey]; ok {
		existing.set = set
		existing.lastSeen = time.Now()
	} else {
		t.conns[connKey] = &connInfo{
			set:      set,
			lastSeen: time.Now(),
		}
	}
}

func (t *connStateTracker) RecordServerResponse(clientIP string, clientPort uint16, serverIP string, serverPort uint16, ttl uint8, hasOpts bool) {
	outKey := fmt.Sprintf("%s:%d->%s:%d", clientIP, clientPort, serverIP, serverPort)
	t.mu.Lock()
	defer t.mu.Unlock()
	info, exists := t.conns[outKey]
	if !exists {
		return
	}
	info.responseSeen = true
	if !info.ttlRecorded {
		info.serverTTL = ttl
		info.ttlRecorded = true
	}
	if hasOpts {
		info.serverHasOpts = true
	}
}

func (t *connStateTracker) CheckRST(clientIP string, clientPort uint16, serverIP string, serverPort uint16, rstTTL uint8, rstHasOpts bool, rstHasACK bool, tolerance int) (drop bool, reason string) {
	outKey := fmt.Sprintf("%s:%d->%s:%d", clientIP, clientPort, serverIP, serverPort)
	t.mu.Lock()
	defer t.mu.Unlock()
	info, exists := t.conns[outKey]
	if !exists {
		return false, ""
	}

	info.rstCount++

	if info.rstCount > 1 {
		return true, fmt.Sprintf("multiple RSTs (count=%d)", info.rstCount)
	}

	if !info.responseSeen {
		return true, "RST before any server response"
	}

	if info.serverHasOpts && !rstHasOpts {
		return true, "TCP options stripped (server uses options, RST does not)"
	}

	if info.serverHasOpts && !rstHasACK {
		return true, "bare RST on established connection"
	}

	if info.ttlRecorded {
		delta := int(rstTTL) - int(info.serverTTL)
		if delta < 0 {
			delta = -delta
		}
		if delta > tolerance {
			return true, fmt.Sprintf("TTL mismatch (RST=%d, server=%d, delta=%d)", rstTTL, info.serverTTL, delta)
		}
	}

	return false, ""
}

func (t *connStateTracker) GetSetForIncoming(clientIP string, clientPort uint16, serverIP string, serverPort uint16) *config.SetConfig {
	outKey := fmt.Sprintf("%s:%d->%s:%d", clientIP, clientPort, serverIP, serverPort)

	t.mu.Lock()
	defer t.mu.Unlock()

	info, exists := t.conns[outKey]
	if !exists || info.set == nil {
		return nil
	}

	info.lastSeen = time.Now()
	return info.set
}

func (t *connStateTracker) TrackIncomingBytes(clientIP string, clientPort uint16, serverIP string, serverPort uint16, bytes uint64, inc *config.IncomingConfig) bool {
	outKey := fmt.Sprintf("%s:%d->%s:%d", clientIP, clientPort, serverIP, serverPort)

	t.mu.Lock()
	defer t.mu.Unlock()

	info, exists := t.conns[outKey]
	if !exists {
		return false
	}

	if info.threshold == 0 {
		minKB := inc.Min
		maxKB := inc.Max
		if maxKB == 0 || maxKB < minKB {
			maxKB = minKB
		}
		if minKB <= 0 {
			minKB = 14
			maxKB = 14
		}

		if minKB == maxKB {
			info.threshold = uint64(minKB * 1024)
		} else {
			info.threshold = uint64((minKB + rand.Intn(maxKB-minKB+1)) * 1024)
		}
	}

	prevBytes := info.bytesIn
	info.bytesIn += bytes
	info.lastSeen = time.Now()

	if prevBytes < info.threshold && info.bytesIn >= info.threshold {
		info.bytesIn = 0
		info.threshold = 0
		return true
	}

	return false
}

func (t *connStateTracker) Cleanup() {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	for k, v := range t.conns {
		if now.Sub(v.lastSeen) > 120*time.Second {
			delete(t.conns, k)
		}
	}
}
