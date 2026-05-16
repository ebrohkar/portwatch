package grouper_test

import (
	"testing"
	"time"

	"github.com/example/portwatch/internal/alert"
	"github.com/example/portwatch/internal/grouper"
)

func makeAlert(port int) alert.Alert {
	return alert.Alert{
		Port:      port,
		Event:     "open",
		Severity:  "info",
		Timestamp: time.Now(),
	}
}

func portsOf(ports ...int) map[int]struct{} {
	m := make(map[int]struct{}, len(ports))
	for _, p := range ports {
		m[p] = struct{}{}
	}
	return m
}

func TestNew_ValidRules(t *testing.T) {
	_, err := grouper.New([]grouper.Rule{
		{Name: "web", Ports: portsOf(80, 443)},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNew_EmptyRules_ReturnsError(t *testing.T) {
	_, err := grouper.New(nil)
	if err == nil {
		t.Fatal("expected error for empty rules")
	}
}

func TestNew_EmptyName_ReturnsError(t *testing.T) {
	_, err := grouper.New([]grouper.Rule{
		{Name: "", Ports: portsOf(80)},
	})
	if err == nil {
		t.Fatal("expected error for empty rule name")
	}
}

func TestNew_EmptyPorts_ReturnsError(t *testing.T) {
	_, err := grouper.New([]grouper.Rule{
		{Name: "web", Ports: map[int]struct{}{}},
	})
	if err == nil {
		t.Fatal("expected error for rule with no ports")
	}
}

func TestAdd_And_Len(t *testing.T) {
	g, _ := grouper.New([]grouper.Rule{
		{Name: "web", Ports: portsOf(80, 443)},
	})
	g.Add(makeAlert(80))
	g.Add(makeAlert(443))
	if got := g.Len(); got != 2 {
		t.Fatalf("expected 2, got %d", got)
	}
}

func TestAdd_UnmatchedPort_Ignored(t *testing.T) {
	g, _ := grouper.New([]grouper.Rule{
		{Name: "web", Ports: portsOf(80)},
	})
	g.Add(makeAlert(9999))
	if got := g.Len(); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestFlush_ReturnsAndResets(t *testing.T) {
	g, _ := grouper.New([]grouper.Rule{
		{Name: "web", Ports: portsOf(80)},
	})
	g.Add(makeAlert(80))
	groups := g.Flush()
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if len(groups[0].Alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(groups[0].Alerts))
	}
	if g.Len() != 0 {
		t.Fatal("expected grouper to be empty after flush")
	}
}

func TestFlush_EmptyGroups_Omitted(t *testing.T) {
	g, _ := grouper.New([]grouper.Rule{
		{Name: "web", Ports: portsOf(80)},
		{Name: "db", Ports: portsOf(5432)},
	})
	g.Add(makeAlert(80))
	groups := g.Flush()
	if len(groups) != 1 {
		t.Fatalf("expected 1 non-empty group, got %d", len(groups))
	}
	if groups[0].Name != "web" {
		t.Fatalf("expected group 'web', got %q", groups[0].Name)
	}
}

func TestAdd_MultipleRules_SamePort(t *testing.T) {
	g, _ := grouper.New([]grouper.Rule{
		{Name: "web", Ports: portsOf(443)},
		{Name: "tls", Ports: portsOf(443)},
	})
	g.Add(makeAlert(443))
	if got := g.Len(); got != 2 {
		t.Fatalf("expected alert in both groups (len=2), got %d", got)
	}
}
