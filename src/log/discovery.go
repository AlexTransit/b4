package log

import (
	"fmt"
	"sync"
)

var (
	discoveryHub     *DiscoveryLogHub
	discoveryHubOnce sync.Once
)

type DiscoveryLogHub struct {
	mu        sync.RWMutex
	listeners []chan string
}

func GetDiscoveryHub() *DiscoveryLogHub {
	discoveryHubOnce.Do(func() {
		discoveryHub = &DiscoveryLogHub{
			listeners: make([]chan string, 0),
		}
	})
	return discoveryHub
}

func (h *DiscoveryLogHub) Subscribe() chan string {
	h.mu.Lock()
	defer h.mu.Unlock()
	ch := make(chan string, 256)
	h.listeners = append(h.listeners, ch)
	return ch
}

func (h *DiscoveryLogHub) Unsubscribe(ch chan string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for i, l := range h.listeners {
		if l == ch {
			h.listeners = append(h.listeners[:i], h.listeners[i+1:]...)
			close(ch)
			return
		}
	}
}

func (h *DiscoveryLogHub) Broadcast(msg string) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, ch := range h.listeners {
		select {
		case ch <- msg:
		default:
		}
	}
}

func DiscoveryLogf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	GetDiscoveryHub().Broadcast(msg)
	Infof("[DISCOVERY] %s", msg)
}
