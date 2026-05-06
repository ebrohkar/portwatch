package reporter_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/user/portwatch/internal/notifier"
	"github.com/user/portwatch/internal/reporter"
	"github.com/user/portwatch/internal/rules"
	"github.com/user/portwatch/internal/scanner"
	"github.com/user/portwatch/internal/state"
)

func buildReporter(t *testing.T, buf *bytes.Buffer, ruleBytes []byte) *reporter.Reporter {
	t.Helper()

	sc, err := scanner.NewScanner(scanner.Config{Host: "127.0.0.1", StartPort: 1, EndPort: 1024})
	if err != nil {
		t.Fatalf("scanner: %v", err)
	}

	rs, err := rules.NewRuleSet(ruleBytes)
	if err != nil {
		t.Fatalf("ruleset: %v", err)
	}

	st := state.NewStore("")

	var n *notifier.Notifier
	if buf != nil {
		n = notifier.New(notifier.WithWriter(buf))
	} else {
		n = notifier.New()
	}

	return reporter.New(sc, rs, st, n)
}

func TestNew_ReturnsReporter(t *testing.T) {
	r := buildReporter(t, nil, []byte(`rules: []`))
	if r == nil {
		t.Fatal("expected non-nil Reporter")
	}
}

func TestRun_CancelledContext(t *testing.T) {
	r := buildReporter(t, nil, []byte(`rules: []`))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	// A cancelled context should propagate cleanly; scan may return an error
	// or empty results — either is acceptable.
	_ = r.Run(ctx)
}

func TestAlertCount_NoRules(t *testing.T) {
	r := buildReporter(t, nil, []byte(`rules: []`))
	count := r.AlertCount([]int{80, 443}, []int{8080})
	if count != 0 {
		t.Errorf("expected 0 alerts with no rules, got %d", count)
	}
}

func TestAlertCount_WithAlertRule(t *testing.T) {
	rulesYAML := []byte(`
rules:
  - port: 9999
    event: opened
    action: alert
    severity: high
`)
	r := buildReporter(t, nil, rulesYAML)

	count := r.AlertCount([]int{9999}, []int{})
	if count != 1 {
		t.Errorf("expected 1 alert, got %d", count)
	}
}

func TestAlertCount_ClosedEvent(t *testing.T) {
	rulesYAML := []byte(`
rules:
  - port: 22
    event: closed
    action: alert
    severity: medium
`)
	r := buildReporter(t, nil, rulesYAML)

	count := r.AlertCount([]int{}, []int{22})
	if count != 1 {
		t.Errorf("expected 1 alert for closed port, got %d", count)
	}
}
