package escalation

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) clock { return func() time.Time { return t } }

func TestNew_PanicsOnZeroThreshold(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero threshold")
		}
	}()
	New(0, time.Minute)
}

func TestNew_PanicsOnZeroWindow(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero window")
		}
	}()
	New(2, 0)
}

func TestEvaluate_BelowThreshold_ReturnsSameSeverity(t *testing.T) {
	e := New(3, time.Minute)
	got := e.Evaluate(8080, "opened", SeverityInfo)
	if got != SeverityInfo {
		t.Fatalf("expected %q, got %q", SeverityInfo, got)
	}
}

func TestEvaluate_AtThreshold_PromotesSeverity(t *testing.T) {
	e := New(2, time.Minute)
	e.Evaluate(8080, "opened", SeverityInfo) // hit 1
	got := e.Evaluate(8080, "opened", SeverityInfo) // hit 2 — should escalate
	if got != SeverityWarning {
		t.Fatalf("expected %q, got %q", SeverityWarning, got)
	}
}

func TestEvaluate_WarningPromotesToCritical(t *testing.T) {
	e := New(1, time.Minute)
	got := e.Evaluate(443, "closed", SeverityWarning)
	if got != SeverityCritical {
		t.Fatalf("expected %q, got %q", SeverityCritical, got)
	}
}

func TestEvaluate_CriticalStaysCritical(t *testing.T) {
	e := New(1, time.Minute)
	got := e.Evaluate(22, "opened", SeverityCritical)
	if got != SeverityCritical {
		t.Fatalf("expected %q, got %q", SeverityCritical, got)
	}
}

func TestEvaluate_DifferentPortsAreIndependent(t *testing.T) {
	e := New(2, time.Minute)
	e.Evaluate(80, "opened", SeverityInfo)
	// port 443 has only one hit — should not escalate
	got := e.Evaluate(443, "opened", SeverityInfo)
	if got != SeverityInfo {
		t.Fatalf("expected %q, got %q", SeverityInfo, got)
	}
}

func TestEvaluate_WindowExpiry_ResetsCount(t *testing.T) {
	now := time.Now()
	e := New(2, time.Second)
	e.now = fixedClock(now)
	e.Evaluate(8080, "opened", SeverityInfo) // hit 1 in first window

	// Advance past the window so state expires.
	e.now = fixedClock(now.Add(2 * time.Second))
	got := e.Evaluate(8080, "opened", SeverityInfo) // hit 1 in new window
	if got != SeverityInfo {
		t.Fatalf("expected %q after window expiry, got %q", SeverityInfo, got)
	}
}

func TestReset_ClearsState(t *testing.T) {
	e := New(2, time.Minute)
	e.Evaluate(8080, "opened", SeverityInfo) // hit 1
	e.Reset(8080, "opened")
	// After reset hit count is 1 again — below threshold.
	got := e.Evaluate(8080, "opened", SeverityInfo)
	if got != SeverityInfo {
		t.Fatalf("expected %q after reset, got %q", SeverityInfo, got)
	}
}
