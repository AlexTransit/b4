package mtproto

import (
	"fmt"
	"net"
	"strconv"
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
