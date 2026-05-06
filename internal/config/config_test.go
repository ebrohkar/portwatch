package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/user/portwatch/internal/config"
)

func TestLoadFromBytes_Defaults(t *testing.T) {
	cfg, err := config.LoadFromBytes([]byte(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ScanInterval != 60*time.Second {
		t.Errorf("expected default scan_interval 60s, got %v", cfg.ScanInterval)
	}
	if cfg.PortRange.Start != 1 {
		t.Errorf("expected default port_range.start 1, got %d", cfg.PortRange.Start)
	}
	if cfg.PortRange.End != 65535 {
		t.Errorf("expected default port_range.end 65535, got %d", cfg.PortRange.End)
	}
	if cfg.StateFile == "" {
		t.Error("expected non-empty default state_file")
	}
}

func TestLoadFromBytes_CustomValues(t *testing.T) {
	yaml := `
scan_interval: 30s
port_range:
  start: 1024
  end: 9000
state_file: /tmp/state.json
rules_file: /etc/portwatch/rules.yaml
`
	cfg, err := config.LoadFromBytes([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ScanInterval != 30*time.Second {
		t.Errorf("expected 30s, got %v", cfg.ScanInterval)
	}
	if cfg.PortRange.Start != 1024 {
		t.Errorf("expected 1024, got %d", cfg.PortRange.Start)
	}
	if cfg.RulesFile != "/etc/portwatch/rules.yaml" {
		t.Errorf("unexpected rules_file: %s", cfg.RulesFile)
	}
}

func TestLoadFromBytes_InvalidInterval(t *testing.T) {
	_, err := config.LoadFromBytes([]byte(`scan_interval: 500ms`))
	if err == nil {
		t.Fatal("expected error for sub-second interval")
	}
}

func TestLoadFromBytes_InvalidPortRange(t *testing.T) {
	yaml := `
port_range:
  start: 9000
  end: 1000
`
	_, err := config.LoadFromBytes([]byte(yaml))
	if err == nil {
		t.Fatal("expected error for inverted port range")
	}
}

func TestLoadFromBytes_MalformedYAML(t *testing.T) {
	_, err := config.LoadFromBytes([]byte(`{bad yaml:::`)) 
	if err == nil {
		t.Fatal("expected error for malformed YAML")
	}
}

func TestLoadFromFile_NotFound(t *testing.T) {
	_, err := config.LoadFromFile("/nonexistent/portwatch.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadFromFile_Valid(t *testing.T) {
	f, err := os.CreateTemp("", "portwatch-cfg-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("scan_interval: 10s\n")
	f.Close()

	cfg, err := config.LoadFromFile(f.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ScanInterval != 10*time.Second {
		t.Errorf("expected 10s, got %v", cfg.ScanInterval)
	}
}
