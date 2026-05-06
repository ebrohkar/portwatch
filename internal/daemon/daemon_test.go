package daemon_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/user/portwatch/internal/daemon"
	"github.com/user/portwatch/internal/notifier"
	"github.com/user/portwatch/internal/rules"
	"github.com/user/portwatch/internal/scanner"
	"github.com/user/portwatch/internal/state"
)

func buildConfig(t *testing.T, buf *bytes.Buffer) daemon.Config {
	t.Helper()

	sc, err := scanner.NewScanner(scanner.Options{StartPort: 1, EndPort: 1024, Timeout: 50 * time.Millisecond})
	if err != nil {
		t.Fatalf("scanner: %v", err)
	}

	rs, err := rules.NewRuleSet(nil)
	if err != nil {
		t.Fatalf("ruleset: %v", err)
	}

	store := state.NewStore("")

	var writers []notifier.Writer
	if buf != nil {
		writers = append(writers, buf)
	}
	nt := notifier.New(writers...)

	return daemon.Config{
		Interval: 100 * time.Millisecond,
		RuleSet:  rs,
		Scanner:  sc,
		Store:    store,
		Notifier: nt,
	}
}

func TestNew_DefaultInterval(t *testing.T) {
	cfg := buildConfig(t, nil)
	cfg.Interval = 0
	d := daemon.New(cfg)
	if d == nil {
		t.Fatal("expected non-nil daemon")
	}
}

func TestRun_CancelImmediately(t *testing.T) {
	cfg := buildConfig(t, nil)
	d := daemon.New(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before Run

	err := d.Run(ctx)
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestRun_TicksAtLeastOnce(t *testing.T) {
	var buf bytes.Buffer
	cfg := buildConfig(t, &buf)
	cfg.Interval = 50 * time.Millisecond
	d := daemon.New(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Should return without error (context deadline exceeded is acceptable).
	err := d.Run(ctx)
	if err == nil {
		t.Fatal("expected context error, got nil")
	}
}
