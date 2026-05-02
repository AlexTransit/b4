package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/daniellavrushin/b4/geodat"
	"github.com/daniellavrushin/b4/log"
	"github.com/daniellavrushin/b4/tlsgen"
	"github.com/daniellavrushin/b4/utils"
)

func (c *Config) SaveToFile(path string) error {
	if path == "" {
		log.Tracef("config path is not defined")
		return nil
	}

	c.Version = CurrentConfigVersion

	data, err := MarshalSparse(c)
	if err != nil {
		return log.Errorf("failed to marshal config: %v", err)
	}

	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return log.Errorf("failed to create config directory: %v", err)
		}
	}

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return log.Errorf("failed to create config file: %v", err)
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return log.Errorf("failed to write config file: %v", err)
	}
	return nil
}

func (c *Config) LoadFromFile(path string) error {
	if path == "" {
		log.Tracef("config path is not defined")
		return nil
	}

	info, err := os.Stat(path)
	if err != nil {
		return log.Errorf("failed to stat config file: %v", err)
	}

	if info.IsDir() {
		return log.Errorf("config path is a directory, not a file: %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return log.Errorf("failed to read config file: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return log.Errorf("failed to parse config file: %v", err)
	}

	rawSets := raw["sets"]
	delete(raw, "sets")

	withoutSets, _ := json.Marshal(raw)
	if err := json.Unmarshal(withoutSets, c); err != nil {
		return log.Errorf("failed to parse config file: %v", err)
	}

	if rawSets != nil {
		var setArray []json.RawMessage
		if err := json.Unmarshal(rawSets, &setArray); err != nil {
			return log.Errorf("failed to parse sets: %v", err)
		}
		c.Sets = make([]*SetConfig, 0, len(setArray))
		for _, rawSet := range setArray {
			set := NewSetConfig()
			if err := json.Unmarshal(rawSet, &set); err != nil {
				return log.Errorf("failed to parse set: %v", err)
			}
			c.Sets = append(c.Sets, &set)
		}
	}

	return nil
}

func (cfg *Config) ApplyLogLevel(level string) {
	switch level {
	case "debug":
		cfg.System.Logging.Level = log.LevelDebug
	case "trace":
		cfg.System.Logging.Level = log.LevelTrace
	case "info":
		cfg.System.Logging.Level = log.LevelInfo
	case "error":
		cfg.System.Logging.Level = log.LevelError
	case "silent":
		cfg.System.Logging.Level = -1
	default:
		cfg.System.Logging.Level = log.LevelInfo
	}
}

func (c *Config) MainInjectedMark() uint {
	if c.Queue.Mark == 0 {
		return DefaultConfig.Queue.Mark
	}
	return c.Queue.Mark
}

func (c *Config) DiscoveryFlowMark() uint {
	if c.System.Checker.DiscoveryFlowMark != 0 {
		return c.System.Checker.DiscoveryFlowMark
	}
	return c.MainInjectedMark() + 1
}

func (c *Config) DiscoveryInjectedMark() uint {
	if c.System.Checker.DiscoveryInjectedMark != 0 {
		return c.System.Checker.DiscoveryInjectedMark
	}
	return c.MainInjectedMark() + 2
}

func (c *Config) Validate() error {
	c.System.WebServer.IsEnabled = c.System.WebServer.Port > 0 && c.System.WebServer.Port <= 65535

	hasCert := c.System.WebServer.TLSCert != ""
	hasKey := c.System.WebServer.TLSKey != ""
	if hasCert != hasKey {
		return fmt.Errorf("both tls_cert and tls_key must be specified together")
	}
	if hasCert {
		if _, err := os.Stat(c.System.WebServer.TLSCert); err != nil {
			return fmt.Errorf("TLS certificate file not found: %s", c.System.WebServer.TLSCert)
		}
		if _, err := os.Stat(c.System.WebServer.TLSKey); err != nil {
			return fmt.Errorf("TLS key file not found: %s", c.System.WebServer.TLSKey)
		}
	}

	if c.Queue.TCPConnBytesLimit < DefaultConfig.Queue.TCPConnBytesLimit {
		c.Queue.TCPConnBytesLimit = DefaultConfig.Queue.TCPConnBytesLimit
	} else if c.Queue.TCPConnBytesLimit > 100 {
		c.Queue.TCPConnBytesLimit = 100
	}
	if c.Queue.UDPConnBytesLimit < DefaultConfig.Queue.UDPConnBytesLimit {
		c.Queue.UDPConnBytesLimit = DefaultConfig.Queue.UDPConnBytesLimit
	} else if c.Queue.UDPConnBytesLimit > 30 {
		c.Queue.UDPConnBytesLimit = 30
	}

	for _, set := range c.Sets {
		if set.Routing.Table < 0 {
			set.Routing.Table = 0
		}
		if set.Routing.IPTTLSeconds <= 0 {
			set.Routing.IPTTLSeconds = DefaultSetConfig.Routing.IPTTLSeconds
		}
		switch set.Routing.Mode {
		case "":
			set.Routing.Mode = RoutingModeInterface
		case RoutingModeProxy, RoutingModeInterface:
		default:
			return fmt.Errorf("set %q: unknown routing mode %q", set.Name, set.Routing.Mode)
		}
		set.Routing.EgressInterface = sanitizeIfaceName(set.Routing.EgressInterface)
		for i, src := range set.Routing.SourceInterfaces {
			set.Routing.SourceInterfaces[i] = sanitizeIfaceName(src)
		}

		if set.Routing.Enabled && set.Routing.Mode == RoutingModeProxy {
			if set.Routing.Upstream.Port < 1 || set.Routing.Upstream.Port > 65535 {
				return fmt.Errorf("set %q: upstream proxy port must be 1-65535", set.Name)
			}
			h := strings.ToLower(strings.TrimSpace(set.Routing.Upstream.Host))
			if h == "" {
				h = "127.0.0.1"
			}
			if c.System.Socks5.Enabled && set.Routing.Upstream.Port == c.System.Socks5.Port {
				if h == "127.0.0.1" || h == "::1" || h == "localhost" || h == "0.0.0.0" {
					return fmt.Errorf("set %q: upstream proxy points to b4's own SOCKS5 server (loop)", set.Name)
				}
			}
		}

		if len(set.Fragmentation.SeqOverlapPattern) > 0 {
			set.Fragmentation.SeqOverlapBytes = make([]byte, len(set.Fragmentation.SeqOverlapPattern))
			for i, s := range set.Fragmentation.SeqOverlapPattern {
				s = strings.TrimPrefix(s, "0x")
				b, _ := strconv.ParseUint(s, 16, 8)
				set.Fragmentation.SeqOverlapBytes[i] = byte(b)
			}
		}

		if set.TCP.Duplicate.Enabled {
			if set.TCP.Duplicate.Count < 1 {
				set.TCP.Duplicate.Count = 1
			}
			if set.TCP.Duplicate.Count > 10 {
				set.TCP.Duplicate.Count = 10
			}
			if len(set.Targets.IPs) == 0 && len(set.Targets.GeoIpCategories) == 0 {
				log.Warnf("Set '%s' has duplication enabled but no IP targets configured", set.Name)
			}
		}

		if set.TCP.ConnBytesLimit > c.Queue.TCPConnBytesLimit {
			set.TCP.ConnBytesLimit = c.Queue.TCPConnBytesLimit
		}
		if set.UDP.ConnBytesLimit > c.Queue.UDPConnBytesLimit {
			set.UDP.ConnBytesLimit = c.Queue.UDPConnBytesLimit
		}

		if len(set.Targets.GeoSiteCategories) > 0 && c.System.Geo.GeoSitePath == "" {
			return fmt.Errorf("--geosite must be specified when using --geo-categories")
		}

		if len(set.Targets.GeoIpCategories) > 0 && c.System.Geo.GeoIpPath == "" {
			return fmt.Errorf("--geoip must be specified when using --geoip-categories")
		}
	}

	c.sanitizeEscalation()

	// Validate global MSS clamp
	if c.Queue.MSSClamp.Enabled {
		if c.Queue.MSSClamp.Size < 10 {
			c.Queue.MSSClamp.Size = 10
		}
		if c.Queue.MSSClamp.Size > 1460 {
			c.Queue.MSSClamp.Size = 1460
		}
	}

	for i := range c.Queue.Devices.Devices {
		d := &c.Queue.Devices.Devices[i]
		d.MAC = strings.ToUpper(strings.TrimSpace(d.MAC))
		if d.MSSClamp > 0 {
			if d.MSSClamp < 10 {
				d.MSSClamp = 10
			}
			if d.MSSClamp > 1460 {
				d.MSSClamp = 1460
			}
		}
	}

	if c.Queue.Threads < 1 {
		return fmt.Errorf("threads must be at least 1")
	}

	if c.Queue.StartNum < 0 || c.Queue.StartNum > 65535 {
		return fmt.Errorf("queue-num must be between 0 and 65535")
	}

	c.Queue.Mark = c.MainInjectedMark()

	maxMark := uint(^uint32(0))
	if c.Queue.Mark > maxMark {
		return fmt.Errorf("mark value 0x%x exceeds uint32 max", c.Queue.Mark)
	}
	if c.Queue.Mark > maxMark-2 && c.System.Checker.DiscoveryFlowMark == 0 {
		return fmt.Errorf("mark value 0x%x is too high for auto-derived discovery marks", c.Queue.Mark)
	}

	const perSetReachableBits uint32 = 0x17FFF
	if c.Queue.Mark != 0 && uint32(c.Queue.Mark)&^perSetReachableBits == 0 {
		return fmt.Errorf("mark value 0x%x conflicts with per-set mark bits {0-14, 16}; bypass rule would catch TPROXY-redirected traffic. Use a value with at least one bit in {15, 17-31} (default 0x8000 has bit 15)", c.Queue.Mark)
	}

	c.System.Checker.DiscoveryFlowMark = c.DiscoveryFlowMark()
	c.System.Checker.DiscoveryInjectedMark = c.DiscoveryInjectedMark()

	if c.System.Checker.DiscoveryFlowMark > maxMark || c.System.Checker.DiscoveryInjectedMark > maxMark {
		return fmt.Errorf("discovery mark values exceed uint32 max")
	}
	if c.Queue.Mark == c.System.Checker.DiscoveryFlowMark ||
		c.Queue.Mark == c.System.Checker.DiscoveryInjectedMark ||
		c.System.Checker.DiscoveryFlowMark == c.System.Checker.DiscoveryInjectedMark {
		return fmt.Errorf("queue marks must be unique: mark, discovery_flow_mark, discovery_injected_mark")
	}

	for _, set := range c.Sets {
		if set.Id == "" {
			return fmt.Errorf("each set must have a unique non-empty ID")
		}

		set.UDP.DPortFilter = utils.ValidatePorts(set.UDP.DPortFilter)
		set.TCP.DPortFilter = utils.ValidatePorts(set.TCP.DPortFilter)
	}

	c.LoadCapturePayloads()
	c.BuildTCPPortMap()
	c.BuildSetPortRanges()

	return nil
}

func (c *Config) LogString() string {
	return ""
}

// LoadTargets returns all targets (domains and IPs) from all sets grouped by set name
func (c *Config) LoadTargets() ([]*SetConfig, int, int, error) {
	result := make([]*SetConfig, 0, len(c.Sets))
	totalDomains := 0
	totalIps := 0

	// Process all sets
	for _, set := range c.Sets {

		if !set.Enabled {
			continue
		}

		domains, ips, err := c.GetTargetsForSet(set)
		if err != nil {
			return nil, -1, -1, fmt.Errorf("failed to load domains for set '%s': %w", set.Name, err)
		}
		if len(domains) > 0 {
			totalDomains += len(domains)
		}
		if len(ips) > 0 {
			totalIps += len(ips)
		}
		result = append(result, set)
	}

	return result, totalDomains, totalIps, nil
}

func (c *Config) GetTargetsForSet(set *SetConfig) ([]string, []string, error) {
	return c.GetTargetsForSetWithCache(set, nil, nil)
}

func (c *Config) GetTargetsForSetWithCache(set *SetConfig, geositeDomains, geoipIPs map[string][]string) ([]string, []string, error) {
	domains := []string{}
	ips := []string{}

	if len(set.Targets.GeoSiteCategories) > 0 && c.System.Geo.GeoSitePath != "" {
		if geositeDomains != nil {
			// Use cached data
			for _, cat := range set.Targets.GeoSiteCategories {
				if cached, ok := geositeDomains[cat]; ok {
					domains = append(domains, cached...)
				}
			}
		} else {
			// Fallback to disk (slow path)
			geoDomains, err := geodat.LoadDomainsFromCategories(
				c.System.Geo.GeoSitePath,
				set.Targets.GeoSiteCategories,
			)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to load geosite domains for set '%s': %w", set.Name, err)
			}
			domains = append(domains, geoDomains...)
		}
	}

	if len(set.Targets.SNIDomains) > 0 {
		domains = append(domains, set.Targets.SNIDomains...)
	}
	set.Targets.DomainsToMatch = domains

	if len(set.Targets.GeoIpCategories) > 0 && c.System.Geo.GeoIpPath != "" {
		if geoipIPs != nil {
			for _, cat := range set.Targets.GeoIpCategories {
				if cached, ok := geoipIPs[cat]; ok {
					ips = append(ips, cached...)
				}
			}
		} else {
			geoIps, err := geodat.LoadIpsFromCategories(
				c.System.Geo.GeoIpPath,
				set.Targets.GeoIpCategories,
			)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to load geoip for set '%s': %w", set.Name, err)
			}
			ips = append(ips, geoIps...)
		}
	}

	if len(set.Targets.IPs) > 0 {
		ips = append(ips, set.Targets.IPs...)
	}

	set.Targets.IpsToMatch = ips
	return domains, ips, nil
}

func (c *Config) GetSetById(id string) *SetConfig {
	for _, set := range c.Sets {
		if set.Id == id {
			return set
		}
	}
	return nil
}

func (set *SetConfig) ResetToDefaults() {
	defaultSet := DefaultSetConfig

	id := set.Id
	name := set.Name
	targets := set.Targets

	*set = defaultSet

	set.Id = id
	set.Name = name
	set.Targets = targets

	set.TCP.Win.Values = make([]int, len(defaultSet.TCP.Win.Values))
	copy(set.TCP.Win.Values, defaultSet.TCP.Win.Values)

	set.Faking.SNIMutation.FakeSNIs = make([]string, len(defaultSet.Faking.SNIMutation.FakeSNIs))
	copy(set.Faking.SNIMutation.FakeSNIs, defaultSet.Faking.SNIMutation.FakeSNIs)

	set.Fragmentation.SeqOverlapPattern = make([]string, len(defaultSet.Fragmentation.SeqOverlapPattern))
	copy(set.Fragmentation.SeqOverlapPattern, defaultSet.Fragmentation.SeqOverlapPattern)

	set.Faking.TLSMod = make([]string, len(defaultSet.Faking.TLSMod))
	copy(set.Faking.TLSMod, defaultSet.Faking.TLSMod)

	set.Routing.SourceInterfaces = make([]string, len(defaultSet.Routing.SourceInterfaces))
	copy(set.Routing.SourceInterfaces, defaultSet.Routing.SourceInterfaces)

}

func (t *TargetsConfig) AppendIP(ip []string) error {
	for _, newIP := range ip {
		exists := false
		for _, existingIP := range t.IPs {
			if existingIP == newIP {
				exists = true
				break
			}
		}
		if !exists {
			t.IPs = append(t.IPs, newIP)
		}
	}

	for _, newIP := range ip {
		exists := false
		for _, existingIP := range t.IpsToMatch {
			if existingIP == newIP {
				exists = true
				break
			}
		}
		if !exists {
			t.IpsToMatch = append(t.IpsToMatch, newIP)
		}
	}

	return nil
}
func (t *TargetsConfig) AppendSNI(sni string) error {

	for _, existingDomain := range t.SNIDomains {
		if existingDomain == sni {
			return log.Errorf("SNI '%s' already exists in the set", sni)
		}
	}
	t.SNIDomains = append(t.SNIDomains, sni)

	for _, existingDomain := range t.DomainsToMatch {
		if existingDomain == sni {
			return log.Errorf("SNI '%s' already exists in the set", sni)
		}
	}
	t.DomainsToMatch = append(t.DomainsToMatch, sni)
	return nil
}

func (cfg *Config) CollectUDPPorts() []string {
	portSet := make(map[string]bool)
	portSet["443"] = true

	for _, set := range cfg.Sets {
		if !set.Enabled || set.UDP.DPortFilter == "" {
			continue
		}
		for _, p := range strings.Split(set.UDP.DPortFilter, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				portSet[p] = true
			}
		}
	}

	ports := make([]string, 0, len(portSet))
	for p := range portSet {
		ports = append(ports, p)
	}
	sort.Strings(ports)
	ports = mergeAndNormalizePorts(ports)
	return ports
}

func (cfg *Config) CollectTCPPorts() []string {
	portSet := make(map[string]bool)
	portSet["443"] = true

	for _, set := range cfg.Sets {
		if !set.Enabled || set.TCP.DPortFilter == "" {
			continue
		}
		for _, p := range strings.Split(set.TCP.DPortFilter, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				portSet[p] = true
			}
		}
	}

	ports := make([]string, 0, len(portSet))
	for p := range portSet {
		ports = append(ports, p)
	}
	sort.Strings(ports)
	ports = mergeAndNormalizePorts(ports)
	return ports
}

// BuildTCPPortMap builds a fast lookup map of all configured TCP ports.
// Call this after CollectTCPPorts to pre-compute the set for packet handler use.
func (cfg *Config) BuildTCPPortMap() {
	cfg.tcpPortMap = make(map[uint16]bool)
	for _, p := range cfg.CollectTCPPorts() {
		p = strings.TrimSpace(p)
		if strings.Contains(p, "-") {
			parts := strings.Split(p, "-")
			if len(parts) == 2 {
				start, _ := strconv.Atoi(parts[0])
				end, _ := strconv.Atoi(parts[1])
				for i := start; i <= end; i++ {
					cfg.tcpPortMap[uint16(i)] = true
				}
			}
		} else {
			port, _ := strconv.Atoi(p)
			if port > 0 {
				cfg.tcpPortMap[uint16(port)] = true
			}
		}
	}
}

type PortRange struct {
	Min uint16
	Max uint16
}

// TLSVersionCode converts a TLS version filter string ("1.2", "1.3") to wire format.
// Returns 0 if the filter is empty (match any).
func TLSVersionCode(filter string) uint16 {
	switch filter {
	case "1.0":
		return 0x0301
	case "1.1":
		return 0x0302
	case "1.2":
		return 0x0303
	case "1.3":
		return 0x0304
	default:
		return 0
	}
}

// TLSVersionString converts a TLS wire version to a human-readable string.
func TLSVersionString(ver uint16) string {
	switch ver {
	case 0x0301:
		return "1.0"
	case 0x0302:
		return "1.1"
	case 0x0303:
		return "1.2"
	case 0x0304:
		return "1.3"
	default:
		return ""
	}
}

// MatchesTLSVersion checks if the client's max TLS version matches this set's filter.
// Returns true if no filter is configured or if tlsVersion is 0 (unknown).
func (set *SetConfig) MatchesTLSVersion(tlsVersion uint16) bool {
	if set.Targets.TLSVersion == "" {
		return true
	}
	if tlsVersion == 0 {
		return true
	}
	filterVer := TLSVersionCode(set.Targets.TLSVersion)
	if filterVer == 0 {
		return false // invalid filter value — don't silently match everything
	}
	return tlsVersion == filterVer
}

func (set *SetConfig) HasIPOrDomainTargets() bool {
	return len(set.Targets.IpsToMatch) > 0 || len(set.Targets.DomainsToMatch) > 0
}

func (set *SetConfig) MatchesTCPDPort(port uint16) bool {
	if len(set.TCPPortRanges) == 0 {
		return true
	}
	for _, r := range set.TCPPortRanges {
		if port >= r.Min && port <= r.Max {
			return true
		}
	}
	return false
}

func (set *SetConfig) MatchesUDPDPort(port uint16) bool {
	if len(set.UDPPortRanges) == 0 {
		return true
	}
	for _, r := range set.UDPPortRanges {
		if port >= r.Min && port <= r.Max {
			return true
		}
	}
	return false
}

func (cfg *Config) BuildSetPortRanges() {
	for _, set := range cfg.Sets {
		set.TCPPortRanges = nil
		set.UDPPortRanges = nil

		if set.TCP.DPortFilter != "" {
			for _, part := range strings.Split(set.TCP.DPortFilter, ",") {
				if pr, ok := parsePortRangeStr(part); ok {
					set.TCPPortRanges = append(set.TCPPortRanges, pr)
				}
			}
		}
		if set.UDP.DPortFilter != "" {
			for _, part := range strings.Split(set.UDP.DPortFilter, ",") {
				if pr, ok := parsePortRangeStr(part); ok {
					set.UDPPortRanges = append(set.UDPPortRanges, pr)
				}
			}
		}
	}
}

func parsePortRangeStr(part string) (PortRange, bool) {
	part = strings.TrimSpace(part)
	if part == "" {
		return PortRange{}, false
	}
	if strings.Contains(part, "-") {
		bounds := strings.SplitN(part, "-", 2)
		if len(bounds) == 2 {
			min, err1 := strconv.Atoi(bounds[0])
			max, err2 := strconv.Atoi(bounds[1])
			if err1 == nil && err2 == nil && min >= 1 && max >= 1 && min <= max && min <= 65535 && max <= 65535 {
				return PortRange{Min: uint16(min), Max: uint16(max)}, true
			}
		}
	} else {
		port, err := strconv.Atoi(part)
		if err == nil && port >= 1 && port <= 65535 {
			return PortRange{Min: uint16(port), Max: uint16(port)}, true
		}
	}
	return PortRange{}, false
}

// IsTCPPort checks if a port is in the configured TCP port set.
func (cfg *Config) IsTCPPort(port uint16) bool {
	if cfg.tcpPortMap == nil {
		return port == 443
	}
	return cfg.tcpPortMap[port]
}

func (cfg *Config) CollectDeviceMSSClamps() map[int][]string {
	result := make(map[int][]string)
	for _, d := range cfg.Queue.Devices.Devices {
		mac := strings.ToUpper(strings.TrimSpace(d.MAC))
		if mac == "" || d.MSSClamp <= 0 {
			continue
		}
		result[d.MSSClamp] = append(result[d.MSSClamp], mac)
	}
	return result
}

func (dc *DevicesConfig) SelectedMACs() []string {
	var macs []string
	for _, d := range dc.Devices {
		if d.Selected && !d.IsManual {
			mac := strings.ToUpper(strings.TrimSpace(d.MAC))
			if mac != "" {
				macs = append(macs, mac)
			}
		}
	}
	return macs
}

func (dc *DevicesConfig) FindByMAC(mac string) *Device {
	mac = strings.ToUpper(strings.TrimSpace(mac))
	for i := range dc.Devices {
		if strings.ToUpper(dc.Devices[i].MAC) == mac {
			return &dc.Devices[i]
		}
	}
	return nil
}

func (dc *DevicesConfig) ManualEntries() []Device {
	var result []Device
	for _, d := range dc.Devices {
		if d.IsManual {
			result = append(result, d)
		}
	}
	return result
}

func (cfg *Config) HasGlobalMSSClamp() (bool, int) {
	if cfg.Queue.MSSClamp.Enabled && cfg.Queue.MSSClamp.Size > 0 {
		return true, cfg.Queue.MSSClamp.Size
	}
	return false, 0
}

// MSSClampFingerprint returns a string representation of the MSS clamp configuration for comparison.
func (cfg *Config) MSSClampFingerprint() string {
	parts := []string{}

	global, globalSize := cfg.HasGlobalMSSClamp()
	if global {
		parts = append(parts, fmt.Sprintf("global:%d", globalSize))
	}

	deviceClamps := cfg.CollectDeviceMSSClamps()
	for size, macs := range deviceClamps {
		sort.Strings(macs)
		parts = append(parts, fmt.Sprintf("dev:%d:%s", size, strings.Join(macs, ",")))
	}

	sort.Strings(parts)
	return strings.Join(parts, ";")
}

// CollectDuplicateIPs returns IPv4 and IPv6 IPs/CIDRs from sets with duplication enabled.
// Used for firewall rules that queue packets without connbytes limit.
func (cfg *Config) CollectDuplicateIPs() (ipv4 []string, ipv6 []string) {
	seen4 := make(map[string]struct{})
	seen6 := make(map[string]struct{})
	for _, set := range cfg.Sets {
		if !set.Enabled || !set.TCP.Duplicate.Enabled {
			continue
		}
		for _, ipStr := range set.Targets.IpsToMatch {
			ipStr = strings.TrimSpace(ipStr)
			if ipStr == "" {
				continue
			}
			if strings.Contains(ipStr, ":") {
				if _, ok := seen6[ipStr]; !ok {
					seen6[ipStr] = struct{}{}
					ipv6 = append(ipv6, ipStr)
				}
			} else {
				if _, ok := seen4[ipStr]; !ok {
					seen4[ipStr] = struct{}{}
					ipv4 = append(ipv4, ipStr)
				}
			}
		}
	}
	return
}

func (c *Config) Clone() *Config {
	data, _ := json.Marshal(c)
	var clone Config
	_ = json.Unmarshal(data, &clone)
	clone.ConfigPath = c.ConfigPath

	for _, set := range clone.Sets {
		for _, origSet := range c.Sets {
			if set.Id == origSet.Id {
				set.Targets.DomainsToMatch = make([]string, len(origSet.Targets.DomainsToMatch))
				copy(set.Targets.DomainsToMatch, origSet.Targets.DomainsToMatch)

				set.Targets.IpsToMatch = make([]string, len(origSet.Targets.IpsToMatch))
				copy(set.Targets.IpsToMatch, origSet.Targets.IpsToMatch)
				break
			}
		}
	}

	clone.Validate()
	return &clone
}

func (c *Config) LoadCapturePayloads() {
	capturesDir := ""
	if c.ConfigPath != "" {
		capturesDir = filepath.Dir(c.ConfigPath)
	}

	for _, set := range c.Sets {
		if !set.Enabled {
			continue
		}
		switch set.Faking.SNIType {
		case FakePayloadDefault1:
			set.Faking.PayloadData = FakeSNI1
		case FakePayloadDefault2:
			set.Faking.PayloadData = FakeSNI2
		case FakePayloadCustom:
			set.Faking.PayloadData = []byte(set.Faking.CustomPayload)
		case FakePayloadCapture:
			if capturesDir == "" || set.Faking.PayloadFile == "" {
				set.Faking.PayloadData = nil
				continue
			}
			capturePath := filepath.Join(capturesDir, set.Faking.PayloadFile)
			data, err := os.ReadFile(capturePath)
			if err != nil {
				log.Errorf("Failed to load capture file %s: %v", set.Faking.PayloadFile, err)
				set.Faking.PayloadData = nil
				continue
			}
			set.Faking.PayloadData = data
			log.Tracef("Loaded capture payload %s (%d bytes)", set.Faking.PayloadFile, len(data))
		case FakePayloadDomain:
			if set.Faking.PayloadDomain == "" {
				set.Faking.PayloadData = nil
				continue
			}
			data, err := tlsgen.GenerateTLSClientHello(set.Faking.PayloadDomain)
			if err != nil {
				log.Errorf("Failed to generate domain payload for %s: %v", set.Faking.PayloadDomain, err)
				set.Faking.PayloadData = nil
				continue
			}
			set.Faking.PayloadData = data
			log.Tracef("Generated domain payload for %s (%d bytes)", set.Faking.PayloadDomain, len(data))
		default:
			set.Faking.PayloadData = nil
		}
	}
}

func mergeAndNormalizePorts(ports []string) []string {
	type portRange struct{ start, end int }
	var ranges []portRange

	for _, p := range ports {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		p = strings.ReplaceAll(p, ":", "-")
		if strings.Contains(p, "-") {
			parts := strings.Split(p, "-")
			if len(parts) == 2 {
				start, _ := strconv.Atoi(parts[0])
				end, _ := strconv.Atoi(parts[1])
				if start > 0 && end > 0 && start <= end {
					ranges = append(ranges, portRange{start, end})
				}
			}
		} else {
			port, _ := strconv.Atoi(p)
			if port > 0 {
				ranges = append(ranges, portRange{port, port})
			}
		}
	}

	if len(ranges) == 0 {
		return nil
	}

	sort.Slice(ranges, func(i, j int) bool { return ranges[i].start < ranges[j].start })

	merged := []portRange{ranges[0]}
	for i := 1; i < len(ranges); i++ {
		last := &merged[len(merged)-1]
		cur := ranges[i]
		if cur.start <= last.end+1 {
			if cur.end > last.end {
				last.end = cur.end
			}
		} else {
			merged = append(merged, cur)
		}
	}

	result := make([]string, len(merged))
	for i, r := range merged {
		if r.start == r.end {
			result[i] = strconv.Itoa(r.start)
		} else {
			result[i] = fmt.Sprintf("%d-%d", r.start, r.end)
		}
	}
	return result
}

func (c *Config) sanitizeEscalation() {
	byID := make(map[string]*SetConfig, len(c.Sets))
	for _, s := range c.Sets {
		if s.Id != "" {
			byID[s.Id] = s
		}
	}

	for _, s := range c.Sets {
		if s.EscalateTo == "" {
			continue
		}
		if s.EscalateTo == s.Id {
			log.Warnf("Set %q: escalate_to references self, clearing", s.Name)
			s.EscalateTo = ""
			continue
		}
		target, ok := byID[s.EscalateTo]
		if !ok {
			log.Warnf("Set %q: escalate_to %q not found, clearing", s.Name, s.EscalateTo)
			s.EscalateTo = ""
			continue
		}
		if !target.Enabled {
			log.Warnf("Set %q: escalate_to %q is disabled, clearing", s.Name, target.Name)
			s.EscalateTo = ""
			continue
		}
	}

	for _, s := range c.Sets {
		if s.EscalateTo == "" {
			continue
		}
		seen := map[string]bool{s.Id: true}
		cur := s
		for cur.EscalateTo != "" {
			if seen[cur.EscalateTo] {
				log.Warnf("Set %q: escalate_to chain has a cycle at %q, breaking", s.Name, cur.Name)
				cur.EscalateTo = ""
				break
			}
			seen[cur.EscalateTo] = true
			cur = byID[cur.EscalateTo]
			if cur == nil {
				break
			}
		}
	}
}

// sanitizeIfaceName strips any characters not valid in a Linux interface name.
func sanitizeIfaceName(name string) string {
	var b strings.Builder
	for _, c := range name {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.' {
			b.WriteRune(c)
		}
	}
	return b.String()
}
