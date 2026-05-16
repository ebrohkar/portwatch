package digest

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/example/portwatch/internal/alert"
)

func makeAlert(port int, event string) alert.Alert {
	return alert.Alert{Port: port, Event: event, Severity: "info"}
}

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestNew_PanicsOnZeroWindow(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	New(nil, 0)
}

func TestNew_DefaultsToStdout(t *testing.T) {
	d := New(nil, time.Minute)
	if d.w == nil {
		t.Fatal("expected non-nil writer")
	}
}

func TestAdd_AccumulatesCounts(t *testing.T) {
	var buf bytes.Buffer
	d := New(&buf, time.Minute)
	d.Add(makeAlert(8080, "opened"))
	d.Add(makeAlert(8080, "opened"))
	d.Add(makeAlert(8080, "closed"))
	entries := d.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	e := entries[0]
	if e.Port != 8080 || e.Opened != 2 || e.Closed != 1 || e.Total != 3 {
		t.Errorf("unexpected entry: %+v", e)
	}
}

func TestAdd_MultiplePorts(t *testing.T) {
	var buf bytes.Buffer
	d := New(&buf, time.Minute)
	d.Add(makeAlert(80, "opened"))
	d.Add(makeAlert(443, "closed"))
	entries := d.Entries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Port != 80 || entries[1].Port != 443 {
		t.Error("entries not sorted by port")
	}
}

func TestFlush_BeforeWindow_ReturnsFalse(t *testing.T) {
	now := time.Now()
	var buf bytes.Buffer
	d := withClock(&buf, time.Minute, fixedClock(now))
	d.Add(makeAlert(8080, "opened"))
	if d.Flush() {
		t.Fatal("expected Flush to return false before window elapses")
	}
	if buf.Len() != 0 {
		t.Error("expected no output before window")
	}
}

func TestFlush_AfterWindow_WritesAndResets(t *testing.T) {
	now := time.Now()
	var buf bytes.Buffer
	d := withClock(&buf, time.Minute, fixedClock(now))
	d.Add(makeAlert(8080, "opened"))
	// advance clock past window
	d.clock = fixedClock(now.Add(2 * time.Minute))
	if !d.Flush() {
		t.Fatal("expected Flush to return true after window")
	}
	out := buf.String()
	if !strings.Contains(out, "port 8080") {
		t.Errorf("expected port 8080 in output, got: %s", out)
	}
	// counts should be reset
	if len(d.Entries()) != 0 {
		t.Error("expected counts to be reset after flush")
	}
}

func TestFlush_NoActivity_WritesNoActivity(t *testing.T) {
	now := time.Now()
	var buf bytes.Buffer
	d := withClock(&buf, time.Minute, fixedClock(now))
	d.clock = fixedClock(now.Add(2 * time.Minute))
	d.Flush()
	if !strings.Contains(buf.String(), "no activity") {
		t.Errorf("expected '(no activity)' in output, got: %s", buf.String())
	}
}
