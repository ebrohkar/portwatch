package normalize_test

import (
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/normalize"
)

func makeAlert(port int, event, severity, message string) alert.Alert {
	return alert.Alert{
		Port:      port,
		Event:     event,
		Severity:  severity,
		Message:   message,
		Timestamp: time.Now(),
	}
}

func TestNew_Defaults(t *testing.T) {
	n := normalize.New()
	a, err := n.Apply(makeAlert(80, "", "", ""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Event != "open" {
		t.Errorf("expected default event 'open', got %q", a.Event)
	}
	if a.Severity != "info" {
		t.Errorf("expected default severity 'info', got %q", a.Severity)
	}
}

func TestNew_CustomDefaults(t *testing.T) {
	n := normalize.New(
		normalize.WithDefaultSeverity("warn"),
		normalize.WithDefaultEvent("close"),
	)
	a, err := n.Apply(makeAlert(443, "", "", ""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Event != "close" {
		t.Errorf("expected 'close', got %q", a.Event)
	}
	if a.Severity != "warn" {
		t.Errorf("expected 'warn', got %q", a.Severity)
	}
}

func TestApply_LowerCasesFields(t *testing.T) {
	n := normalize.New()
	a, err := n.Apply(makeAlert(8080, "OPEN", "CRITICAL", "  hello  "))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Event != "open" {
		t.Errorf("expected 'open', got %q", a.Event)
	}
	if a.Severity != "critical" {
		t.Errorf("expected 'critical', got %q", a.Severity)
	}
	if a.Message != "hello" {
		t.Errorf("expected trimmed message 'hello', got %q", a.Message)
	}
}

func TestApply_InvalidPort_ReturnsError(t *testing.T) {
	n := normalize.New()
	for _, port := range []int{0, -1, 65536, 99999} {
		_, err := n.Apply(makeAlert(port, "open", "info", ""))
		if err == nil {
			t.Errorf("expected error for port %d, got nil", port)
		}
	}
}

func TestApply_ValidBoundaryPorts(t *testing.T) {
	n := normalize.New()
	for _, port := range []int{1, 65535} {
		_, err := n.Apply(makeAlert(port, "open", "info", ""))
		if err != nil {
			t.Errorf("unexpected error for port %d: %v", port, err)
		}
	}
}

func TestNewStage_PanicsOnNil(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil Normalizer")
		}
	}()
	normalize.NewStage(nil)
}

func TestStage_Allow_ValidAlert_Passes(t *testing.T) {
	s := normalize.NewStage(normalize.New())
	a := makeAlert(22, "OPEN", "", "  ssh  ")
	if !s.Allow(&a) {
		t.Fatal("expected Allow to return true")
	}
	if a.Event != "open" {
		t.Errorf("expected normalized event 'open', got %q", a.Event)
	}
	if a.Message != "ssh" {
		t.Errorf("expected trimmed message 'ssh', got %q", a.Message)
	}
}

func TestStage_Allow_InvalidPort_Drops(t *testing.T) {
	s := normalize.NewStage(normalize.New())
	a := makeAlert(0, "open", "info", "")
	if s.Allow(&a) {
		t.Error("expected Allow to return false for invalid port")
	}
}

func TestStage_String(t *testing.T) {
	s := normalize.NewStage(normalize.New())
	if s.String() != "normalize" {
		t.Errorf("unexpected stage name: %q", s.String())
	}
}
