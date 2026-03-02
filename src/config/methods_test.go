package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveToFile_And_LoadFromFile(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("save and load roundtrip", func(t *testing.T) {
		path := filepath.Join(tmpDir, "config.json")
		cfg := NewConfig()
		cfg.Queue.StartNum = 999
		cfg.Queue.Threads = 8

		if err := cfg.SaveToFile(path); err != nil {
			t.Fatalf("SaveToFile failed: %v", err)
		}

		loaded := NewConfig()
		if err := loaded.LoadFromFile(path); err != nil {
			t.Fatalf("LoadFromFile failed: %v", err)
		}

		if loaded.Queue.StartNum != 999 {
			t.Errorf("expected StartNum=999, got %d", loaded.Queue.StartNum)
		}
		if loaded.Queue.Threads != 8 {
			t.Errorf("expected Threads=8, got %d", loaded.Queue.Threads)
		}
	})

	t.Run("empty path does nothing", func(t *testing.T) {
		cfg := NewConfig()
		if err := cfg.SaveToFile(""); err != nil {
			t.Errorf("expected nil error for empty path, got %v", err)
		}
		if err := cfg.LoadFromFile(""); err != nil {
			t.Errorf("expected nil error for empty path, got %v", err)
		}
	})

	t.Run("load from nonexistent file", func(t *testing.T) {
		cfg := NewConfig()
		err := cfg.LoadFromFile(filepath.Join(tmpDir, "nonexistent.json"))
		if err == nil {
			t.Error("expected error for nonexistent file")
		}
	})

	t.Run("load from directory fails", func(t *testing.T) {
		cfg := NewConfig()
		err := cfg.LoadFromFile(tmpDir)
		if err == nil {
			t.Error("expected error when path is directory")
		}
	})

	t.Run("load malformed json", func(t *testing.T) {
		path := filepath.Join(tmpDir, "bad.json")
		os.WriteFile(path, []byte("{invalid json"), 0644)

		cfg := NewConfig()
		err := cfg.LoadFromFile(path)
		if err == nil {
			t.Error("expected error for malformed json")
		}
	})

	t.Run("save creates default set if empty", func(t *testing.T) {
		path := filepath.Join(tmpDir, "empty_sets.json")
		cfg := NewConfig()
		cfg.Sets = []*SetConfig{}

		if err := cfg.SaveToFile(path); err != nil {
			t.Fatalf("SaveToFile failed: %v", err)
		}

		loaded := NewConfig()
		loaded.LoadFromFile(path)
		if len(loaded.Sets) == 0 {
			t.Error("expected Sets to have default set after save")
		}
	})
}

func TestValidate(t *testing.T) {
	t.Run("default config is valid", func(t *testing.T) {
		cfg := NewConfig()
		if err := cfg.Validate(); err != nil {
			t.Errorf("default config should be valid: %v", err)
		}
	})

	t.Run("threads < 1 fails", func(t *testing.T) {
		cfg := NewConfig()
		cfg.Queue.Threads = 0
		if err := cfg.Validate(); err == nil {
			t.Error("expected error for threads=0")
		}
	})

	t.Run("queue num out of range", func(t *testing.T) {
		cfg := NewConfig()
		cfg.Queue.StartNum = -1
		if err := cfg.Validate(); err == nil {
			t.Error("expected error for negative queue num")
		}

		cfg.Queue.StartNum = 70000
		if err := cfg.Validate(); err == nil {
			t.Error("expected error for queue num > 65535")
		}
	})

	t.Run("geosite categories without path", func(t *testing.T) {
		cfg := NewConfig()
		mainSet := NewSetConfig()
		mainSet.Id = MAIN_SET_ID
		mainSet.Targets.GeoSiteCategories = []string{"youtube"}
		cfg.Sets = []*SetConfig{&mainSet}
		cfg.System.Geo.GeoSitePath = ""

		if err := cfg.Validate(); err == nil {
			t.Error("expected error when geosite categories set without path")
		}
	})

	t.Run("geoip categories without path", func(t *testing.T) {
		cfg := NewConfig()
		mainSet := NewSetConfig()
		mainSet.Id = MAIN_SET_ID
		mainSet.Targets.GeoIpCategories = []string{"ru"}
		cfg.Sets = []*SetConfig{&mainSet}
		cfg.System.Geo.GeoIpPath = ""

		if err := cfg.Validate(); err == nil {
			t.Error("expected error when geoip categories set without path")
		}
	})

	t.Run("MainSet nil gets initialized from Sets", func(t *testing.T) {
		cfg := NewConfig()
		mainSet := NewSetConfig()
		mainSet.Id = MAIN_SET_ID
		mainSet.TCP.ConnBytesLimit = 42
		mainSet.Targets = TargetsConfig{}
		cfg.Sets = []*SetConfig{&mainSet}
		cfg.MainSet = nil

		if err := cfg.Validate(); err != nil {
			t.Fatalf("validation failed: %v", err)
		}
		if cfg.MainSet == nil {
			t.Error("MainSet should be initialized")
		}
		if cfg.MainSet.TCP.ConnBytesLimit != 42 {
			t.Error("MainSet should be found from Sets")
		}
	})

	t.Run("MainSet nil and not in Sets gets default", func(t *testing.T) {
		cfg := NewConfig()
		cfg.Sets = []*SetConfig{}
		cfg.MainSet = nil

		if err := cfg.Validate(); err != nil {
			t.Fatalf("validation failed: %v", err)
		}
		if cfg.MainSet == nil {
			t.Error("MainSet should be initialized to default")
		}
	})

	t.Run("set TCP ConnBytesLimit > main gets capped", func(t *testing.T) {
		cfg := NewConfig()
		cfg.Validate()

		secondSet := NewSetConfig()
		secondSet.Id = "second"
		secondSet.TCP.ConnBytesLimit = cfg.MainSet.TCP.ConnBytesLimit + 10
		cfg.Sets = append(cfg.Sets, &secondSet)

		if err := cfg.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if secondSet.TCP.ConnBytesLimit != cfg.MainSet.TCP.ConnBytesLimit {
			t.Errorf("expected TCP ConnBytesLimit to be capped to %d, got %d",
				cfg.MainSet.TCP.ConnBytesLimit, secondSet.TCP.ConnBytesLimit)
		}
	})

	t.Run("set UDP ConnBytesLimit > main gets capped", func(t *testing.T) {
		cfg := NewConfig()
		cfg.Validate()

		secondSet := NewSetConfig()
		secondSet.Id = "second"
		secondSet.UDP.ConnBytesLimit = cfg.MainSet.UDP.ConnBytesLimit + 10
		cfg.Sets = append(cfg.Sets, &secondSet)

		if err := cfg.Validate(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if secondSet.UDP.ConnBytesLimit != cfg.MainSet.UDP.ConnBytesLimit {
			t.Errorf("expected UDP ConnBytesLimit to be capped to %d, got %d",
				cfg.MainSet.UDP.ConnBytesLimit, secondSet.UDP.ConnBytesLimit)
		}
	})
	t.Run("set without id fails", func(t *testing.T) {
		cfg := NewConfig()
		cfg.Validate()

		secondSet := NewSetConfig()
		secondSet.Id = ""
		cfg.Sets = append(cfg.Sets, &secondSet)

		if err := cfg.Validate(); err == nil {
			t.Error("expected error for set without id")
		}
	})

	t.Run("web server port enables/disables", func(t *testing.T) {
		cfg := NewConfig()
		cfg.System.WebServer.Port = 0
		cfg.Validate()
		if cfg.System.WebServer.IsEnabled {
			t.Error("port 0 should disable web server")
		}

		cfg.System.WebServer.Port = 8080
		cfg.Validate()
		if !cfg.System.WebServer.IsEnabled {
			t.Error("valid port should enable web server")
		}

		cfg.System.WebServer.Port = 70000
		cfg.Validate()
		if cfg.System.WebServer.IsEnabled {
			t.Error("port > 65535 should disable web server")
		}
	})
}

func TestAppendIP(t *testing.T) {
	t.Run("appends new IPs", func(t *testing.T) {
		targets := &TargetsConfig{}
		targets.AppendIP([]string{"1.1.1.1", "8.8.8.8"})

		if len(targets.IPs) != 2 {
			t.Errorf("expected 2 IPs, got %d", len(targets.IPs))
		}
		if len(targets.IpsToMatch) != 2 {
			t.Errorf("expected 2 IpsToMatch, got %d", len(targets.IpsToMatch))
		}
	})

	t.Run("deduplicates IPs", func(t *testing.T) {
		targets := &TargetsConfig{
			IPs:        []string{"1.1.1.1"},
			IpsToMatch: []string{"1.1.1.1"},
		}
		targets.AppendIP([]string{"1.1.1.1", "8.8.8.8"})

		if len(targets.IPs) != 2 {
			t.Errorf("expected 2 IPs after dedup, got %d", len(targets.IPs))
		}
	})
}

func TestAppendSNI(t *testing.T) {
	t.Run("appends new SNI", func(t *testing.T) {
		targets := &TargetsConfig{}
		if err := targets.AppendSNI("example.com"); err != nil {
			t.Errorf("AppendSNI failed: %v", err)
		}

		if len(targets.SNIDomains) != 1 || targets.SNIDomains[0] != "example.com" {
			t.Error("SNI not appended to SNIDomains")
		}
		if len(targets.DomainsToMatch) != 1 {
			t.Error("SNI not appended to DomainsToMatch")
		}
	})

	t.Run("rejects duplicate in SNIDomains", func(t *testing.T) {
		targets := &TargetsConfig{
			SNIDomains: []string{"example.com"},
		}
		if err := targets.AppendSNI("example.com"); err == nil {
			t.Error("expected error for duplicate SNI")
		}
	})

	t.Run("rejects duplicate in DomainsToMatch", func(t *testing.T) {
		targets := &TargetsConfig{
			DomainsToMatch: []string{"example.com"},
		}
		if err := targets.AppendSNI("example.com"); err == nil {
			t.Error("expected error for duplicate in DomainsToMatch")
		}
	})
}

func TestGetSetById(t *testing.T) {
	cfg := NewConfig()
	set1 := NewSetConfig()
	set1.Id = "set-1"
	set2 := NewSetConfig()
	set2.Id = "set-2"
	cfg.Sets = []*SetConfig{&set1, &set2}

	t.Run("finds existing set", func(t *testing.T) {
		found := cfg.GetSetById("set-2")
		if found == nil || found.Id != "set-2" {
			t.Error("should find set-2")
		}
	})

	t.Run("returns nil for unknown id", func(t *testing.T) {
		if cfg.GetSetById("unknown") != nil {
			t.Error("should return nil for unknown id")
		}
	})
}

func TestGetTargetsForSet(t *testing.T) {
	t.Run("combines manual domains", func(t *testing.T) {
		cfg := NewConfig()
		set := NewSetConfig()
		set.Targets.SNIDomains = []string{"a.com", "b.com"}

		domains, ips, err := cfg.GetTargetsForSet(&set)
		if err != nil {
			t.Fatalf("GetTargetsForSet failed: %v", err)
		}

		if len(domains) != 2 {
			t.Errorf("expected 2 domains, got %d", len(domains))
		}
		if len(ips) != 0 {
			t.Errorf("expected 0 ips, got %d", len(ips))
		}
	})

	t.Run("combines manual IPs", func(t *testing.T) {
		cfg := NewConfig()
		set := NewSetConfig()
		set.Targets.IPs = []string{"1.1.1.1", "8.8.8.8"}

		domains, ips, err := cfg.GetTargetsForSet(&set)
		if err != nil {
			t.Fatalf("GetTargetsForSet failed: %v", err)
		}

		if len(domains) != 0 {
			t.Errorf("expected 0 domains, got %d", len(domains))
		}
		if len(ips) != 2 {
			t.Errorf("expected 2 ips, got %d", len(ips))
		}
	})
}

func TestLoadTargets(t *testing.T) {
	t.Run("skips disabled sets", func(t *testing.T) {
		cfg := NewConfig()

		enabled := NewSetConfig()
		enabled.Id = "enabled"
		enabled.Enabled = true
		enabled.Targets.SNIDomains = []string{"a.com"}

		disabled := NewSetConfig()
		disabled.Id = "disabled"
		disabled.Enabled = false
		disabled.Targets.SNIDomains = []string{"b.com"}

		cfg.Sets = []*SetConfig{&enabled, &disabled}

		sets, domainCount, _, err := cfg.LoadTargets()
		if err != nil {
			t.Fatalf("LoadTargets failed: %v", err)
		}

		if len(sets) != 1 {
			t.Errorf("expected 1 enabled set, got %d", len(sets))
		}
		if domainCount != 1 {
			t.Errorf("expected 1 domain from enabled set, got %d", domainCount)
		}
	})

	t.Run("aggregates counts from multiple sets", func(t *testing.T) {
		cfg := NewConfig()

		set1 := NewSetConfig()
		set1.Id = "set1"
		set1.Enabled = true
		set1.Targets.SNIDomains = []string{"a.com", "b.com"}
		set1.Targets.IPs = []string{"1.1.1.1"}

		set2 := NewSetConfig()
		set2.Id = "set2"
		set2.Enabled = true
		set2.Targets.SNIDomains = []string{"c.com"}
		set2.Targets.IPs = []string{"8.8.8.8", "8.8.4.4"}

		cfg.Sets = []*SetConfig{&set1, &set2}

		_, domainCount, ipCount, err := cfg.LoadTargets()
		if err != nil {
			t.Fatalf("LoadTargets failed: %v", err)
		}

		if domainCount != 3 {
			t.Errorf("expected 3 total domains, got %d", domainCount)
		}
		if ipCount != 3 {
			t.Errorf("expected 3 total ips, got %d", ipCount)
		}
	})
}

func TestApplyLogLevel(t *testing.T) {
	levels := []string{"debug", "trace", "info", "error", "silent", "unknown"}
	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			cfg := NewConfig()
			cfg.ApplyLogLevel(level) // just verify no panic
		})
	}
}

func TestHasGlobalMSSClamp(t *testing.T) {
	t.Run("disabled returns false", func(t *testing.T) {
		cfg := NewConfig()
		cfg.Validate()
		ok, _ := cfg.HasGlobalMSSClamp()
		if ok {
			t.Error("expected false when MSS clamp is disabled")
		}
	})

	t.Run("enabled on main set without targets returns true", func(t *testing.T) {
		cfg := NewConfig()
		cfg.Validate()
		cfg.MainSet.TCP.MSSClamp.Enabled = true
		cfg.MainSet.TCP.MSSClamp.Size = 88
		cfg.MainSet.Targets.IPs = nil
		cfg.MainSet.Targets.GeoIpCategories = nil

		ok, size := cfg.HasGlobalMSSClamp()
		if !ok {
			t.Error("expected true for main set with MSS clamp and no targets")
		}
		if size != 88 {
			t.Errorf("expected size 88, got %d", size)
		}
	})

	t.Run("enabled on main set with IP targets returns false", func(t *testing.T) {
		cfg := NewConfig()
		cfg.Validate()
		cfg.MainSet.TCP.MSSClamp.Enabled = true
		cfg.MainSet.TCP.MSSClamp.Size = 88
		cfg.MainSet.Targets.IPs = []string{"1.1.1.1"}

		ok, _ := cfg.HasGlobalMSSClamp()
		if ok {
			t.Error("expected false when main set has IP targets")
		}
	})

	t.Run("enabled on main set with geoip categories returns false", func(t *testing.T) {
		cfg := NewConfig()
		cfg.Validate()
		cfg.MainSet.TCP.MSSClamp.Enabled = true
		cfg.MainSet.TCP.MSSClamp.Size = 88
		cfg.MainSet.Targets.GeoIpCategories = []string{"ru"}

		ok, _ := cfg.HasGlobalMSSClamp()
		if ok {
			t.Error("expected false when main set has geoip categories")
		}
	})

	t.Run("size zero returns false", func(t *testing.T) {
		cfg := NewConfig()
		cfg.Validate()
		cfg.MainSet.TCP.MSSClamp.Enabled = true
		cfg.MainSet.TCP.MSSClamp.Size = 0

		ok, _ := cfg.HasGlobalMSSClamp()
		if ok {
			t.Error("expected false when size is 0")
		}
	})
}

func TestCollectMSSClampIPs(t *testing.T) {
	t.Run("no sets enabled returns empty maps", func(t *testing.T) {
		cfg := NewConfig()
		cfg.Validate()
		ipv4, ipv6 := cfg.CollectMSSClampIPs()
		if len(ipv4) != 0 || len(ipv6) != 0 {
			t.Error("expected empty maps when no MSS clamp enabled")
		}
	})

	t.Run("collects IPv4 and IPv6 grouped by size", func(t *testing.T) {
		cfg := NewConfig()
		cfg.Validate()

		set1 := NewSetConfig()
		set1.Id = "set1"
		set1.Enabled = true
		set1.TCP.MSSClamp.Enabled = true
		set1.TCP.MSSClamp.Size = 88
		set1.Targets.IpsToMatch = []string{"1.1.1.1", "2.2.2.2", "2001:db8::1"}

		set2 := NewSetConfig()
		set2.Id = "set2"
		set2.Enabled = true
		set2.TCP.MSSClamp.Enabled = true
		set2.TCP.MSSClamp.Size = 200
		set2.Targets.IpsToMatch = []string{"3.3.3.3"}

		cfg.Sets = append(cfg.Sets, &set1, &set2)

		ipv4, ipv6 := cfg.CollectMSSClampIPs()

		if len(ipv4[88]) != 2 {
			t.Errorf("expected 2 IPv4 for size 88, got %d", len(ipv4[88]))
		}
		if len(ipv4[200]) != 1 {
			t.Errorf("expected 1 IPv4 for size 200, got %d", len(ipv4[200]))
		}
		if len(ipv6[88]) != 1 {
			t.Errorf("expected 1 IPv6 for size 88, got %d", len(ipv6[88]))
		}
	})

	t.Run("skips disabled sets", func(t *testing.T) {
		cfg := NewConfig()
		cfg.Validate()

		set := NewSetConfig()
		set.Id = "disabled"
		set.Enabled = false
		set.TCP.MSSClamp.Enabled = true
		set.TCP.MSSClamp.Size = 88
		set.Targets.IpsToMatch = []string{"1.1.1.1"}
		cfg.Sets = append(cfg.Sets, &set)

		ipv4, ipv6 := cfg.CollectMSSClampIPs()
		if len(ipv4) != 0 || len(ipv6) != 0 {
			t.Error("expected empty maps for disabled set")
		}
	})

	t.Run("skips sets with MSS clamp disabled", func(t *testing.T) {
		cfg := NewConfig()
		cfg.Validate()

		set := NewSetConfig()
		set.Id = "no-mss"
		set.Enabled = true
		set.TCP.MSSClamp.Enabled = false
		set.Targets.IpsToMatch = []string{"1.1.1.1"}
		cfg.Sets = append(cfg.Sets, &set)

		ipv4, ipv6 := cfg.CollectMSSClampIPs()
		if len(ipv4) != 0 || len(ipv6) != 0 {
			t.Error("expected empty maps when MSS clamp disabled on set")
		}
	})
}

func TestMSSClampFingerprint(t *testing.T) {
	t.Run("stable ordering", func(t *testing.T) {
		cfg := NewConfig()
		cfg.Validate()

		set := NewSetConfig()
		set.Id = "set1"
		set.Enabled = true
		set.TCP.MSSClamp.Enabled = true
		set.TCP.MSSClamp.Size = 88
		set.Targets.IpsToMatch = []string{"2.2.2.2", "1.1.1.1"}
		cfg.Sets = append(cfg.Sets, &set)

		fp1 := cfg.MSSClampFingerprint()
		fp2 := cfg.MSSClampFingerprint()
		if fp1 != fp2 {
			t.Errorf("fingerprint not stable: %q vs %q", fp1, fp2)
		}
	})

	t.Run("changes when config changes", func(t *testing.T) {
		cfg := NewConfig()
		cfg.Validate()

		set := NewSetConfig()
		set.Id = "set1"
		set.Enabled = true
		set.TCP.MSSClamp.Enabled = true
		set.TCP.MSSClamp.Size = 88
		set.Targets.IpsToMatch = []string{"1.1.1.1"}
		cfg.Sets = append(cfg.Sets, &set)

		fp1 := cfg.MSSClampFingerprint()

		set.TCP.MSSClamp.Size = 100
		fp2 := cfg.MSSClampFingerprint()

		if fp1 == fp2 {
			t.Error("fingerprint should change when size changes")
		}
	})

	t.Run("includes global when main set has no targets", func(t *testing.T) {
		cfg := NewConfig()
		cfg.Validate()
		cfg.MainSet.TCP.MSSClamp.Enabled = true
		cfg.MainSet.TCP.MSSClamp.Size = 88
		cfg.MainSet.Targets.IPs = nil
		cfg.MainSet.Targets.GeoIpCategories = nil

		fp := cfg.MSSClampFingerprint()
		if fp == "" {
			t.Error("fingerprint should not be empty for global MSS clamp")
		}
		if !contains(fp, "global:88") {
			t.Errorf("fingerprint should contain 'global:88', got %q", fp)
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestResetToDefaults(t *testing.T) {
	set := NewSetConfig()
	set.Id = "custom"
	set.Name = "my-set"
	set.Targets.SNIDomains = []string{"keep.me"}
	set.Fragmentation.SNIPosition = 99

	set.ResetToDefaults()

	if set.Id != "custom" {
		t.Error("Id should be preserved")
	}
	if set.Name != "my-set" {
		t.Error("Name should be preserved")
	}
	if len(set.Targets.SNIDomains) != 1 || set.Targets.SNIDomains[0] != "keep.me" {
		t.Error("Targets should be preserved")
	}
	if set.Fragmentation.SNIPosition != DefaultSetConfig.Fragmentation.SNIPosition {
		t.Errorf("SNIPosition should reset to default %d, got %d",
			DefaultSetConfig.Fragmentation.SNIPosition, set.Fragmentation.SNIPosition)
	}
}
