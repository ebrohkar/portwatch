package pipeline_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/notifier"
	"github.com/user/portwatch/internal/pipeline"
)

func makeAlert(port int, event string) alert.Alert {
	return alert.Alert{
		Port:      port,
		Event:     event,
		Severity:  "warn",
		Timestamp: time.Now().UTC(),
	}
}

func buildNotifier(t *testing.T, w *bytes.Buffer) *notifier.Notifier {
	t.Helper()
	n, err := notifier.New(notifier.WithWriter(w))
	if err != nil {
		t.Fatalf("notifier.New: %v", err)
	}
	return n
}

func TestNew_PanicsOnNilNotifier(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil notifier")
		}
	}()
	pipeline.New(nil)
}

func TestRun_AllStagesPass_AlertSent(t *testing.T) {
	var buf bytes.Buffer
	n := buildNotifier(t, &buf)
	p := pipeline.New(n,
		func(alert.Alert) bool { return true },
		func(alert.Alert) bool { return true },
	)

	if err := p.Run(context.Background(), makeAlert(8080, "open")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected notifier output, got none")
	}
}

func TestRun_StageDrops_AlertNotSent(t *testing.T) {
	var buf bytes.Buffer
	n := buildNotifier(t, &buf)
	p := pipeline.New(n,
		func(alert.Alert) bool { return true },
		func(alert.Alert) bool { return false }, // drop
		func(alert.Alert) bool { return true },
	)

	if err := p.Run(context.Background(), makeAlert(8080, "open")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Fatalf("expected no output, got %d bytes", buf.Len())
	}
}

func TestRun_NoStages_AlertSent(t *testing.T) {
	var buf bytes.Buffer
	n := buildNotifier(t, &buf)
	p := pipeline.New(n)

	if err := p.Run(context.Background(), makeAlert(443, "open")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected output with no stages")
	}
}

func TestRun_FirstStageDrops_ShortCircuits(t *testing.T) {
	called := false
	var buf bytes.Buffer
	n := buildNotifier(t, &buf)
	p := pipeline.New(n,
		func(alert.Alert) bool { return false },
		func(alert.Alert) bool { called = true; return true },
	)

	_ = p.Run(context.Background(), makeAlert(22, "open"))
	if called {
		t.Fatal("second stage should not have been called after first dropped")
	}
}
