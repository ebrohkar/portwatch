package alert_test

import (
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
)

func TestNew_SetsFields(t *testing.T) {
	before := time.Now().UTC()
	a := alert.New(8080, "tcp", "opened", alert.SeverityWarning, "unexpected port")
	after := time.Now().UTC()

	if a.Port != 8080 {
		t.Errorf("expected port 8080, got %d", a.Port)
	}
	if a.Protocol != "tcp" {
		t.Errorf("expected protocol tcp, got %s", a.Protocol)
	}
	if a.Event != "opened" {
		t.Errorf("expected event opened, got %s", a.Event)
	}
	if a.Severity != alert.SeverityWarning {
		t.Errorf("expected severity warning, got %s", a.Severity)
	}
	if a.Message != "unexpected port" {
		t.Errorf("unexpected message: %s", a.Message)
	}
	if a.Timestamp.Before(before) || a.Timestamp.After(after) {
		t.Errorf("timestamp out of expected range: %v", a.Timestamp)
	}
}

func TestIsValid_ValidAlert(t *testing.T) {
	a := alert.New(443, "tcp", "closed", alert.SeverityCritical, "expected port closed")
	if !a.IsValid() {
		t.Error("expected alert to be valid")
	}
}

func TestIsValid_InvalidPort(t *testing.T) {
	a := alert.New(0, "tcp", "opened", alert.SeverityInfo, "")
	if a.IsValid() {
		t.Error("expected alert with port 0 to be invalid")
	}

	a2 := alert.New(99999, "tcp", "opened", alert.SeverityInfo, "")
	if a2.IsValid() {
		t.Error("expected alert with port 99999 to be invalid")
	}
}

func TestIsValid_InvalidEvent(t *testing.T) {
	a := alert.New(80, "tcp", "changed", alert.SeverityInfo, "")
	if a.IsValid() {
		t.Error("expected alert with unknown event to be invalid")
	}
}

func TestIsValid_InvalidSeverity(t *testing.T) {
	a := alert.Alert{
		Port:     80,
		Protocol: "tcp",
		Event:    "opened",
		Severity: "unknown",
	}
	if a.IsValid() {
		t.Error("expected alert with unknown severity to be invalid")
	}
}

func TestIsValid_EmptyProtocol(t *testing.T) {
	a := alert.New(80, "", "opened", alert.SeverityInfo, "")
	if a.IsValid() {
		t.Error("expected alert with empty protocol to be invalid")
	}
}
