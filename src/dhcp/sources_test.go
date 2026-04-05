package dhcp

import (
	"os"
	"testing"
)

func TestParseDnsmasqHostnames(t *testing.T) {
	content := `1712345678 aa:bb:cc:dd:ee:ff 192.168.1.10 my-phone 01:aa:bb:cc:dd:ee:ff
1712345679 11:22:33:44:55:66 192.168.1.20 * 01:11:22:33:44:55:66
1712345680 aa:00:bb:11:cc:22 192.168.1.30 laptop 01:aa:00:bb:11:cc:22
`
	f, err := os.CreateTemp("", "dnsmasq-leases-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString(content)
	f.Close()

	result := parseDnsmasqHostnames(f.Name())

	if len(result) != 2 {
		t.Fatalf("expected 2 hostnames, got %d: %v", len(result), result)
	}
	if result["AA:BB:CC:DD:EE:FF"] != "my-phone" {
		t.Errorf("expected my-phone, got %q", result["AA:BB:CC:DD:EE:FF"])
	}
	if result["AA:00:BB:11:CC:22"] != "laptop" {
		t.Errorf("expected laptop, got %q", result["AA:00:BB:11:CC:22"])
	}
}

func TestParseDnsmasqHostnames_NotFound(t *testing.T) {
	result := parseDnsmasqHostnames("/nonexistent/path")
	if result != nil {
		t.Errorf("expected nil for missing file, got %v", result)
	}
}

func TestParseISCHostnames(t *testing.T) {
	content := `lease 192.168.1.10 {
  hardware ethernet aa:bb:cc:dd:ee:ff;
  client-hostname "my-phone";
}
lease 192.168.1.20 {
  hardware ethernet 11:22:33:44:55:66;
}
lease 192.168.1.30 {
  hardware ethernet aa:00:bb:11:cc:22;
  client-hostname "laptop";
}
`
	f, err := os.CreateTemp("", "isc-leases-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString(content)
	f.Close()

	result := parseISCHostnames(f.Name())

	if len(result) != 2 {
		t.Fatalf("expected 2 hostnames, got %d: %v", len(result), result)
	}
	if result["AA:BB:CC:DD:EE:FF"] != "my-phone" {
		t.Errorf("expected my-phone, got %q", result["AA:BB:CC:DD:EE:FF"])
	}
	if result["AA:00:BB:11:CC:22"] != "laptop" {
		t.Errorf("expected laptop, got %q", result["AA:00:BB:11:CC:22"])
	}
}

func TestParseISCHostnames_NotFound(t *testing.T) {
	result := parseISCHostnames("/nonexistent/path")
	if result != nil {
		t.Errorf("expected nil for missing file, got %v", result)
	}
}
