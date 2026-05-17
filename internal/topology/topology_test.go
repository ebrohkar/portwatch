package topology_test

import (
	"testing"

	"github.com/example/portwatch/internal/topology"
)

func TestNew_ValidEntries(t *testing.T) {
	entries := []topology.Entry{
		{Port: 80, Service: "nginx", Tier: topology.TierFrontend},
		{Port: 5432, Service: "postgres", Tier: topology.TierDatabase},
	}
	topo, err := topology.New(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if topo.Len() != 2 {
		t.Fatalf("expected 2 entries, got %d", topo.Len())
	}
}

func TestNew_EmptyEntries_ReturnsError(t *testing.T) {
	_, err := topology.New(nil)
	if err == nil {
		t.Fatal("expected error for empty entries")
	}
}

func TestNew_InvalidPort_ReturnsError(t *testing.T) {
	entries := []topology.Entry{
		{Port: 0, Service: "bad", Tier: topology.TierInternal},
	}
	_, err := topology.New(entries)
	if err == nil {
		t.Fatal("expected error for port 0")
	}
}

func TestNew_EmptyService_ReturnsError(t *testing.T) {
	entries := []topology.Entry{
		{Port: 8080, Service: "  ", Tier: topology.TierBackend},
	}
	_, err := topology.New(entries)
	if err == nil {
		t.Fatal("expected error for blank service name")
	}
}

func TestNew_InvalidTier_ReturnsError(t *testing.T) {
	entries := []topology.Entry{
		{Port: 9090, Service: "metrics", Tier: topology.Tier("unknown")},
	}
	_, err := topology.New(entries)
	if err == nil {
		t.Fatal("expected error for unknown tier")
	}
}

func TestLookup_Found(t *testing.T) {
	entries := []topology.Entry{
		{Port: 443, Service: "tls-proxy", Tier: topology.TierFrontend},
	}
	topo, _ := topology.New(entries)
	e, ok := topo.Lookup(443)
	if !ok {
		t.Fatal("expected entry to be found")
	}
	if e.Service != "tls-proxy" {
		t.Errorf("expected service %q, got %q", "tls-proxy", e.Service)
	}
	if e.Tier != topology.TierFrontend {
		t.Errorf("expected tier %q, got %q", topology.TierFrontend, e.Tier)
	}
}

func TestLookup_NotFound(t *testing.T) {
	entries := []topology.Entry{
		{Port: 8080, Service: "api", Tier: topology.TierBackend},
	}
	topo, _ := topology.New(entries)
	_, ok := topo.Lookup(9999)
	if ok {
		t.Fatal("expected Lookup to return false for unregistered port")
	}
}

func TestLen_ReflectsEntryCount(t *testing.T) {
	entries := []topology.Entry{
		{Port: 3000, Service: "app", Tier: topology.TierBackend},
		{Port: 3306, Service: "mysql", Tier: topology.TierDatabase},
		{Port: 6379, Service: "redis", Tier: topology.TierInternal},
	}
	topo, _ := topology.New(entries)
	if topo.Len() != 3 {
		t.Errorf("expected Len 3, got %d", topo.Len())
	}
}
