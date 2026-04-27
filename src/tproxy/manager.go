package tproxy

import (
	"context"
	"sync"
	"time"

	"github.com/daniellavrushin/b4/config"
	"github.com/daniellavrushin/b4/log"
	"github.com/daniellavrushin/b4/socks5"
)

type Manager struct {
	mu        sync.Mutex
	listeners map[string]*Listener
	resolver  DomainResolver
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewManager(resolver DomainResolver) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		listeners: make(map[string]*Listener),
		resolver:  resolver,
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (m *Manager) SetResolver(r DomainResolver) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.resolver = r
	for _, l := range m.listeners {
		l.Resolver = r
	}
}

func (m *Manager) SyncConfig(cfg *config.Config) {
	if cfg == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	desired := make(map[string]*config.SetConfig, len(cfg.Sets))
	for _, set := range cfg.Sets {
		if set == nil || !set.Enabled || !set.Routing.Enabled {
			continue
		}
		if set.Routing.Mode != config.RoutingModeProxy {
			continue
		}
		desired[set.Id] = set
	}

	for id, l := range m.listeners {
		set, keep := desired[id]
		if !keep {
			log.Infof("tproxy: stopping listener for removed set %q", l.SetName)
			_ = l.Stop()
			delete(m.listeners, id)
			continue
		}
		mark := effectiveMark(set)
		port := PortFor(mark)
		if l.Port != port ||
			l.Upstream.Host != set.Routing.Upstream.Host ||
			l.Upstream.Port != set.Routing.Upstream.Port ||
			l.Upstream.Username != set.Routing.Upstream.Username ||
			l.Upstream.Password != set.Routing.Upstream.Password ||
			l.UseDomain != set.Routing.Upstream.UseDomain ||
			l.FailOpen != set.Routing.Upstream.FailOpen {
			log.Infof("tproxy: restarting listener for set %q (config changed)", set.Name)
			_ = l.Stop()
			delete(m.listeners, id)
		}
	}

	for id, set := range desired {
		if _, ok := m.listeners[id]; ok {
			continue
		}
		mark := effectiveMark(set)
		port := PortFor(mark)
		l := &Listener{
			SetID:    set.Id,
			SetName:  set.Name,
			Port:     port,
			Upstream: socks5.ClientConfig{
				Host:     set.Routing.Upstream.Host,
				Port:     set.Routing.Upstream.Port,
				Username: set.Routing.Upstream.Username,
				Password: set.Routing.Upstream.Password,
				Timeout:  10 * time.Second,
			},
			UseDomain: set.Routing.Upstream.UseDomain,
			FailOpen:  set.Routing.Upstream.FailOpen,
			Resolver:  m.resolver,
		}
		if err := l.Start(m.ctx); err != nil {
			log.Errorf("tproxy: failed to start listener for set %q: %v", set.Name, err)
			continue
		}
		m.listeners[id] = l
	}
}

func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, l := range m.listeners {
		_ = l.Stop()
		delete(m.listeners, id)
	}
	if m.cancel != nil {
		m.cancel()
	}
}

func (m *Manager) PortForSet(set *config.SetConfig) int {
	if set == nil {
		return 0
	}
	return PortFor(effectiveMark(set))
}

func effectiveMark(set *config.SetConfig) uint32 {
	if set == nil {
		return 0
	}
	return MarkForSet(set.Id, set.Routing.FWMark)
}
