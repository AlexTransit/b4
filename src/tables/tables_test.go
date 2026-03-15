package tables

import (
	"testing"

	"github.com/daniellavrushin/b4/config"
)

func TestIPTablesManager_BuildNFQSpec(t *testing.T) {
	cfg := config.NewConfig()
	manager := NewIPTablesManager(&cfg)

	t.Run("single thread", func(t *testing.T) {
		spec := manager.buildNFQSpec(100, 1)

		expected := []string{"-j", "NFQUEUE", "--queue-num", "100", "--queue-bypass"}
		if len(spec) != len(expected) {
			t.Fatalf("expected %d elements, got %d", len(expected), len(spec))
		}
		for i, v := range expected {
			if spec[i] != v {
				t.Errorf("spec[%d] = %q, want %q", i, spec[i], v)
			}
		}
	})

	t.Run("multiple threads", func(t *testing.T) {
		spec := manager.buildNFQSpec(100, 4)

		expected := []string{"-j", "NFQUEUE", "--queue-balance", "100:103", "--queue-bypass"}
		if len(spec) != len(expected) {
			t.Fatalf("expected %d elements, got %d", len(expected), len(spec))
		}
		for i, v := range expected {
			if spec[i] != v {
				t.Errorf("spec[%d] = %q, want %q", i, spec[i], v)
			}
		}
	})

	t.Run("queue balance range calculation", func(t *testing.T) {
		spec := manager.buildNFQSpec(537, 8)

		// Should be 537:544 (537 + 8 - 1 = 544)
		if spec[3] != "537:544" {
			t.Errorf("expected queue-balance 537:544, got %s", spec[3])
		}
	})
}

func TestNFTablesManager_BuildNFQueueAction(t *testing.T) {
	t.Run("single thread", func(t *testing.T) {
		cfg := config.NewConfig()
		cfg.Queue.StartNum = 100
		cfg.Queue.Threads = 1
		manager := NewNFTablesManager(&cfg)

		action := manager.buildNFQueueAction()
		expected := "queue num 100 bypass"
		if action != expected {
			t.Errorf("got %q, want %q", action, expected)
		}
	})

	t.Run("multiple threads", func(t *testing.T) {
		cfg := config.NewConfig()
		cfg.Queue.StartNum = 100
		cfg.Queue.Threads = 4
		manager := NewNFTablesManager(&cfg)

		action := manager.buildNFQueueAction()
		expected := "queue num 100-103 bypass"
		if action != expected {
			t.Errorf("got %q, want %q", action, expected)
		}
	})
}

func TestNewIPTablesManager(t *testing.T) {
	cfg := config.NewConfig()
	manager := NewIPTablesManager(&cfg)

	if manager == nil {
		t.Fatal("expected non-nil manager")
	}
	if manager.cfg != &cfg {
		t.Error("manager.cfg not set correctly")
	}
}

func TestNewNFTablesManager(t *testing.T) {
	cfg := config.NewConfig()
	manager := NewNFTablesManager(&cfg)

	if manager == nil {
		t.Fatal("expected non-nil manager")
	}
	if manager.cfg != &cfg {
		t.Error("manager.cfg not set correctly")
	}
}

func TestNewMonitor(t *testing.T) {
	t.Run("default interval", func(t *testing.T) {
		cfg := config.NewConfig()
		cfg.System.Tables.MonitorInterval = 0 // Will use fallback

		monitor := NewMonitor(&cfg)

		if monitor == nil {
			t.Fatal("expected non-nil monitor")
		}
		if monitor.interval < 1e9 { // 1 second in nanoseconds
			t.Error("interval should be at least 1 second")
		}
	})

	t.Run("custom interval", func(t *testing.T) {
		cfg := config.NewConfig()
		cfg.System.Tables.MonitorInterval = 30

		monitor := NewMonitor(&cfg)

		if monitor.interval.Seconds() != 30 {
			t.Errorf("expected 30s interval, got %v", monitor.interval)
		}
	})
}

func TestManifest_Apply_Empty(t *testing.T) {
	m := Manifest{}
	err := m.Apply()
	if err != nil {
		t.Errorf("empty manifest should apply without error: %v", err)
	}
}

func TestSysctlSetting(t *testing.T) {
	// Just test struct creation - actual apply/revert requires root
	s := SysctlSetting{
		Name:    "net.test.setting",
		Desired: "1",
		Revert:  "0",
	}

	if s.Name != "net.test.setting" {
		t.Error("Name not set")
	}
	if s.Desired != "1" {
		t.Error("Desired not set")
	}
	if s.Revert != "0" {
		t.Error("Revert not set")
	}
}

func TestRule_Struct(t *testing.T) {

	r := Rule{
		IPT:   "iptables",
		Table: "mangle",
		Chain: "B4",
		Spec:  []string{"-p", "tcp", "--dport", "443"},
	}

	if r.IPT != "iptables" {
		t.Error("IPT not set")
	}
	if r.Table != "mangle" {
		t.Error("Table not set")
	}
	if r.Chain != "B4" {
		t.Error("Chain not set")
	}
	if len(r.Spec) != 4 {
		t.Error("Spec not set correctly")
	}
}

func TestChain_Struct(t *testing.T) {

	c := Chain{
		IPT:   "iptables",
		Table: "mangle",
		Name:  "B4",
	}

	if c.IPT != "iptables" {
		t.Error("IPT not set")
	}
	if c.Table != "mangle" {
		t.Error("Table not set")
	}
	if c.Name != "B4" {
		t.Error("Name not set")
	}
}

func TestAddRules_SkipSetup(t *testing.T) {
	cfg := config.NewConfig()
	cfg.System.Tables.SkipSetup = true

	err := AddRules(&cfg)
	if err != nil {
		t.Errorf("AddRules with SkipSetup should return nil: %v", err)
	}
}

func TestClearRules_SkipSetup(t *testing.T) {
	cfg := config.NewConfig()
	cfg.System.Tables.SkipSetup = true

	err := ClearRules(&cfg)
	if err != nil {
		t.Errorf("ClearRules with SkipSetup should return nil: %v", err)
	}
}

func TestMonitor_StartStop_Disabled(t *testing.T) {
	cfg := config.NewConfig()
	cfg.System.Tables.SkipSetup = true

	monitor := NewMonitor(&cfg)

	// Should not panic or block
	monitor.Start()
	monitor.Stop()
}

func TestMonitor_StartStop_IntervalZero(t *testing.T) {
	cfg := config.NewConfig()
	cfg.System.Tables.MonitorInterval = 0

	monitor := NewMonitor(&cfg)

	// interval <= 0 disables monitor
	monitor.Start()
	monitor.Stop()
}

func TestHasBinary(t *testing.T) {
	// "sh" should exist on any unix system
	if !hasBinary("sh") {
		t.Error("sh should be found")
	}

	// Non-existent binary
	if hasBinary("nonexistent_binary_xyz123") {
		t.Error("nonexistent binary should not be found")
	}
}

func TestNFTablesConstants(t *testing.T) {
	if nftTableName != "b4_mangle" {
		t.Errorf("nftTableName = %q, want b4_mangle", nftTableName)
	}
	if nftChainName != "b4_chain" {
		t.Errorf("nftChainName = %q, want b4_chain", nftChainName)
	}
}

func TestIPTablesManager_BuildManifest_NoIPTables(t *testing.T) {
	cfg := config.NewConfig()
	cfg.Queue.IPv4Enabled = false
	cfg.Queue.IPv6Enabled = false

	manager := NewIPTablesManager(&cfg)
	_, err := manager.buildManifest()

	if err == nil {
		t.Error("expected error when no iptables binaries enabled")
	}
}

func TestLoadSysctlSnapshot_NoFile(t *testing.T) {
	// Temporarily change path to non-existent file
	origPath := sysctlSnapPath
	sysctlSnapPath = "/tmp/nonexistent_test_snapshot.json"
	defer func() { sysctlSnapPath = origPath }()

	snap := loadSysctlSnapshot()
	if snap == nil {
		t.Error("should return empty map, not nil")
	}
	if len(snap) != 0 {
		t.Error("should return empty map for non-existent file")
	}
}
