package tproxy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/daniellavrushin/b4/log"
	"github.com/daniellavrushin/b4/socks5"
	"golang.org/x/sys/unix"
)

type DomainResolver interface {
	DomainFor(ip net.IP) string
}

type Listener struct {
	SetID    string
	SetName  string
	BindAddr string
	Port     int
	Upstream socks5.ClientConfig
	UseDomain bool
	FailOpen bool
	Resolver DomainResolver

	ctx    context.Context
	cancel context.CancelFunc
	ln     net.Listener

	activeConns atomic.Int64
}

func (l *Listener) Start(parent context.Context) error {
	if l.Port < 1 || l.Port > 65535 {
		return fmt.Errorf("invalid tproxy port: %d", l.Port)
	}
	bind := l.BindAddr
	if bind == "" {
		bind = "0.0.0.0"
	}
	addr := net.JoinHostPort(bind, fmt.Sprintf("%d", l.Port))

	lc := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			var ctlErr error
			err := c.Control(func(fd uintptr) {
				if e := unix.SetsockoptInt(int(fd), unix.SOL_IP, unix.IP_TRANSPARENT, 1); e != nil {
					ctlErr = fmt.Errorf("set IP_TRANSPARENT: %w", e)
					return
				}
				if e := unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1); e != nil {
					ctlErr = fmt.Errorf("set SO_REUSEADDR: %w", e)
					return
				}
			})
			if err != nil {
				return err
			}
			return ctlErr
		},
	}

	l.ctx, l.cancel = context.WithCancel(parent)
	ln, err := lc.Listen(l.ctx, "tcp", addr)
	if err != nil {
		l.cancel()
		return fmt.Errorf("tproxy listen %s: %w", addr, err)
	}
	l.ln = ln

	go l.acceptLoop()
	log.Infof("tproxy: listening on %s for set %q -> %s:%d", addr, l.SetName, l.Upstream.Host, l.Upstream.Port)
	return nil
}

func (l *Listener) Stop() error {
	if l.cancel != nil {
		l.cancel()
	}
	if l.ln != nil {
		return l.ln.Close()
	}
	return nil
}

func (l *Listener) Active() int64 {
	return l.activeConns.Load()
}

func (l *Listener) acceptLoop() {
	for {
		conn, err := l.ln.Accept()
		if err != nil {
			if l.ctx.Err() != nil {
				return
			}
			if errors.Is(err, net.ErrClosed) {
				return
			}
			log.Tracef("tproxy: accept error on set %q: %v", l.SetName, err)
			time.Sleep(50 * time.Millisecond)
			continue
		}
		go l.handle(conn)
	}
}

func (l *Listener) handle(client net.Conn) {
	l.activeConns.Add(1)
	defer l.activeConns.Add(-1)
	defer client.Close()

	tcpAddr, ok := client.LocalAddr().(*net.TCPAddr)
	if !ok || tcpAddr == nil || tcpAddr.IP == nil {
		log.Tracef("tproxy: missing original dst on set %q", l.SetName)
		return
	}
	origIP := tcpAddr.IP
	origPort := tcpAddr.Port

	targetHost := origIP.String()
	if l.UseDomain && l.Resolver != nil {
		if d := l.Resolver.DomainFor(origIP); d != "" {
			targetHost = d
		}
	}

	dialCtx, cancel := context.WithTimeout(l.ctx, 15*time.Second)
	upstream, err := socks5.DialUpstream(dialCtx, l.Upstream, targetHost, origPort)
	cancel()
	if err != nil {
		log.Tracef("tproxy: upstream dial failed for %s:%d on set %q: %v", targetHost, origPort, l.SetName, err)
		if !l.FailOpen {
			return
		}
		direct, derr := net.DialTimeout("tcp", net.JoinHostPort(origIP.String(), fmt.Sprintf("%d", origPort)), 10*time.Second)
		if derr != nil {
			log.Tracef("tproxy: fail-open direct dial failed: %v", derr)
			return
		}
		upstream = direct
	}
	defer upstream.Close()

	pipe(client, upstream)
}

func pipe(a, b net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, _ = io.Copy(a, b)
		if c, ok := a.(*net.TCPConn); ok {
			_ = c.CloseWrite()
		}
	}()
	go func() {
		defer wg.Done()
		_, _ = io.Copy(b, a)
		if c, ok := b.(*net.TCPConn); ok {
			_ = c.CloseWrite()
		}
	}()
	wg.Wait()
}
