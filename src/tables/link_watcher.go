package tables

import (
	"encoding/binary"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/daniellavrushin/b4/config"
	"github.com/daniellavrushin/b4/log"
	"github.com/mdlayher/netlink"
	"golang.org/x/sys/unix"
)

const (
	ifInfoMsgSize       = 16
	linkWatcherDebounce = 500 * time.Millisecond
)

// linkWatcher subscribes to kernel RTNETLINK link events so routing can
// react immediately when an interface referenced by a set's routing
// config (EgressInterface or SourceInterfaces) is created or torn down.
// This closes a gap where the periodic monitor only detects IP-address
// changes — common with tun2socks / sing-box which destroy and recreate
// their TUN interface with identical addressing.
type linkWatcher struct {
	cfgPtr *atomic.Pointer[config.Config]
	conn   *netlink.Conn
	stop   chan struct{}
	wg     sync.WaitGroup

	debounceMu    sync.Mutex
	debounceTimer *time.Timer
}

func newLinkWatcher(cfgPtr *atomic.Pointer[config.Config]) *linkWatcher {
	return &linkWatcher{
		cfgPtr: cfgPtr,
		stop:   make(chan struct{}),
	}
}

func (w *linkWatcher) Start() error {
	conn, err := netlink.Dial(unix.NETLINK_ROUTE, &netlink.Config{
		Groups: unix.RTMGRP_LINK,
	})
	if err != nil {
		return err
	}
	w.conn = conn
	w.wg.Add(1)
	go w.loop()
	return nil
}

func (w *linkWatcher) Stop() {
	select {
	case <-w.stop:
		return
	default:
	}
	close(w.stop)
	if w.conn != nil {
		_ = w.conn.Close()
	}
	w.wg.Wait()

	w.debounceMu.Lock()
	if w.debounceTimer != nil {
		w.debounceTimer.Stop()
		w.debounceTimer = nil
	}
	w.debounceMu.Unlock()
}

func (w *linkWatcher) loop() {
	defer w.wg.Done()
	for {
		msgs, err := w.conn.Receive()
		if err != nil {
			select {
			case <-w.stop:
				return
			default:
			}
			log.Tracef("Link watcher: receive error: %v", err)
			time.Sleep(time.Second)
			continue
		}
		for _, m := range msgs {
			switch m.Header.Type {
			case unix.RTM_NEWLINK:
				if name, up := parseIfInfoMsg(m.Data); name != "" {
					w.handleEvent(name, true, up)
				}
			case unix.RTM_DELLINK:
				if name, _ := parseIfInfoMsg(m.Data); name != "" {
					w.handleEvent(name, false, false)
				}
			}
		}
	}
}

// parseIfInfoMsg decodes the ifinfomsg header + netlink attributes payload
// of an RTM_NEWLINK / RTM_DELLINK message and returns the interface name
// and whether IFF_UP is set.
func parseIfInfoMsg(b []byte) (name string, up bool) {
	if len(b) < ifInfoMsgSize {
		return "", false
	}
	flags := binary.LittleEndian.Uint32(b[8:12])
	up = flags&unix.IFF_UP != 0
	ad, err := netlink.NewAttributeDecoder(b[ifInfoMsgSize:])
	if err != nil {
		return "", up
	}
	for ad.Next() {
		if ad.Type() == unix.IFLA_IFNAME {
			name = strings.TrimRight(ad.String(), "\x00")
		}
	}
	return name, up
}

func (w *linkWatcher) handleEvent(ifname string, isNew bool, up bool) {
	cfg := w.cfgPtr.Load()
	if cfg == nil {
		return
	}
	if !isWatchedIface(cfg, ifname) {
		return
	}
	if isNew && up {
		log.Infof("Link watcher: interface %s is up, scheduling routing reinstall", ifname)
		w.scheduleReinstall(ifname)
		return
	}
	if !isNew {
		log.Infof("Link watcher: interface %s went away; routing will be reinstalled when it returns", ifname)
	}
}

func isWatchedIface(cfg *config.Config, ifname string) bool {
	for _, set := range cfg.Sets {
		if set == nil || !set.Enabled || !set.Routing.Enabled {
			continue
		}
		if set.Routing.EgressInterface == ifname {
			return true
		}
		for _, src := range set.Routing.SourceInterfaces {
			if strings.TrimSpace(src) == ifname {
				return true
			}
		}
	}
	return false
}

// scheduleReinstall coalesces bursts of RTM_NEWLINK events (a single link
// transition typically emits several) into one reinstall.
func (w *linkWatcher) scheduleReinstall(ifname string) {
	w.debounceMu.Lock()
	defer w.debounceMu.Unlock()
	if w.debounceTimer != nil {
		w.debounceTimer.Stop()
	}
	w.debounceTimer = time.AfterFunc(linkWatcherDebounce, func() {
		cfg := w.cfgPtr.Load()
		if cfg == nil {
			return
		}
		RoutingReinstallForInterface(cfg, ifname)
	})
}
