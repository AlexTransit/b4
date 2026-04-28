package mtproto

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/daniellavrushin/b4/log"
)

var dcAddressesV4 = map[int]string{
	1: "149.154.175.50:443",
	2: "149.154.167.51:443",
	3: "149.154.175.100:443",
	4: "149.154.167.91:443",
	5: "149.154.171.5:443",
}

var dcAddressesV6 = map[int]string{
	1: "[2001:b28:f23d:f001::a]:443",
	2: "[2001:67c:04e8:f002::a]:443",
	3: "[2001:b28:f23d:f003::a]:443",
	4: "[2001:67c:04e8:f004::a]:443",
	5: "[2001:b28:f23f:f005::a]:443",
}

var (
	dcRuntimeMu sync.RWMutex
	dcRuntime   = map[int]string{}
)

var proxyConfigURLs = []string{
	"https://core.telegram.org/getProxyConfig",
	"https://proxy.lavrush.in/telegram/getProxyConfig",
}

var (
	dcRefresherMu   sync.Mutex
	dcRefresherStop context.CancelFunc
)

func StartDCRefresher(ctx context.Context) {
	dcRefresherMu.Lock()
	if dcRefresherStop != nil {
		dcRefresherStop()
	}
	ctx, cancel := context.WithCancel(ctx)
	dcRefresherStop = cancel
	dcRefresherMu.Unlock()

	go func() {
		t := time.NewTimer(0)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				if err := RefreshDCs(); err != nil {
					log.Warnf("MTProto DC refresh failed: %v", err)
				}
				t.Reset(24 * time.Hour)
			}
		}
	}()
}

func DCSnapshot() map[int]string {
	dcRuntimeMu.RLock()
	defer dcRuntimeMu.RUnlock()
	out := make(map[int]string, len(dcRuntime))
	for k, v := range dcRuntime {
		out[k] = v
	}
	return out
}

func RefreshDCs() error {
	cli := &http.Client{Timeout: 3 * time.Second}
	var body []byte
	var lastErr error
	for _, u := range proxyConfigURLs {
		req, err := http.NewRequest("GET", u, nil)
		if err != nil {
			lastErr = err
			continue
		}
		resp, err := cli.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("%s: %w", u, err)
			continue
		}
		if resp.StatusCode != 200 {
			resp.Body.Close()
			lastErr = fmt.Errorf("%s: status %d", u, resp.StatusCode)
			continue
		}
		body, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("%s: %w", u, err)
			continue
		}
		lastErr = nil
		break
	}
	if body == nil {
		return lastErr
	}
	next := map[int]string{}
	sc := bufio.NewScanner(strings.NewReader(string(body)))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if !strings.HasPrefix(line, "proxy_for ") {
			continue
		}
		line = strings.TrimSuffix(line, ";")
		f := strings.Fields(line)
		if len(f) != 3 {
			continue
		}
		id, err := strconv.Atoi(f[1])
		if err != nil {
			continue
		}
		if id < 0 {
			id = -id
		}
		_, port, err := net.SplitHostPort(f[2])
		if err != nil || port != "443" {
			continue
		}
		if _, ok := next[id]; ok {
			continue
		}
		next[id] = f[2]
	}
	if len(next) == 0 {
		return fmt.Errorf("no :443 entries parsed")
	}
	dcRuntimeMu.Lock()
	dcRuntime = next
	dcRuntimeMu.Unlock()
	log.Infof("MTProto DC list refreshed: %d entries", len(next))
	return nil
}

func ResolveDC(dc int, preferV6 bool, relay string) (string, error) {
	absDC := dc
	if absDC < 0 {
		absDC = -absDC
	}

	if relay != "" {
		host, portStr, err := net.SplitHostPort(relay)
		if err != nil {
			return relay, nil
		}
		basePort, err := strconv.Atoi(portStr)
		if err != nil {
			return relay, nil
		}
		return net.JoinHostPort(host, strconv.Itoa(basePort+absDC-1)), nil
	}

	dc = absDC

	dcRuntimeMu.RLock()
	if addr, ok := dcRuntime[dc]; ok {
		dcRuntimeMu.RUnlock()
		return addr, nil
	}
	dcRuntimeMu.RUnlock()

	if preferV6 {
		if addr, ok := dcAddressesV6[dc]; ok {
			return addr, nil
		}
	}

	addr, ok := dcAddressesV4[dc]
	if !ok {
		return "", fmt.Errorf("unknown DC %d", dc)
	}

	return addr, nil
}
