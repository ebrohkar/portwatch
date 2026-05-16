package dedup

import (
	"testing"
	"time"

	"portwatch/internal/alert"
)

func fixedClock(t time.Time) clock {
	return func() time.Time { return t }
}

func makeAlert(port int, event, severity string) alert.Alert {
	return alert.Alert{Port: port, Event: event, Severity: severity}
}

func TestNew_PanicsOnZeroTTL(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on zero TTL")
		}
	}()
	New(0)
}

func TestNew_PanicsOnNilClock(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on nil clock")
		}
	}()
	WithClock(time.Second, nil)
}

func TestAllow_FirstCallPermitted(t *testing.T) {
	now := time.Now()
	d := WithClock(time.Minute, fixedClock(now))
	a := makeAlert(8080, "open", "warn")
	if !d.Allow(a) {
		t.Fatal("expected first call to be allowed")
	}
}

func TestAllow_DuplicateWithinTTLBlocked(t *testing.T) {
	now := time.Now()
	d := WithClock(time.Minute, fixedClock(now))
	a := makeAlert(8080, "open", "warn")
	d.Allow(a)
	if d.Allow(a) {
		t.Fatal("expected duplicate within TTL to be blocked")
	}
}

func TestAllow_DuplicateAfterTTLPermitted(t *testing.T) {
	now := time.Now()
	var current time.Time = now
	clock := func() time.Time { return current }

	d := WithClock(time.Minute, clock)
	a := makeAlert(8080, "open", "warn")
	d.Allow(a)

	current = now.Add(2 * time.Minute)
	if !d.Allow(a) {
		t.Fatal("expected alert after TTL expiry to be allowed")
	}
}

func TestAllow_DifferentPortsAreIndependent(t *testing.T) {
	now := time.Now()
	d := WithClock(time.Minute, fixedClock(now))
	a1 := makeAlert(8080, "open", "warn")
	a2 := makeAlert(9090, "open", "warn")
	d.Allow(a1)
	if !d.Allow(a2) {
		t.Fatal("expected different port to be allowed")
	}
}

func TestAllow_DifferentEventsAreIndependent(t *testing.T) {
	now := time.Now()
	d := WithClock(time.Minute, fixedClock(now))
	a1 := makeAlert(8080, "open", "warn")
	a2 := makeAlert(8080, "close", "warn")
	d.Allow(a1)
	if !d.Allow(a2) {
		t.Fatal("expected different event to be allowed")
	}
}

func TestFlush_RemovesExpiredEntries(t *testing.T) {
	now := time.Now()
	var current time.Time = now
	clock := func() time.Time { return current }

	d := WithClock(time.Minute, clock)
	a := makeAlert(8080, "open", "warn")
	d.Allow(a)

	current = now.Add(2 * time.Minute)
	d.Flush()

	if len(d.entries) != 0 {
		t.Fatalf("expected entries to be flushed, got %d", len(d.entries))
	}
}

func TestFlush_KeepsActiveEntries(t *testing.T) {
	now := time.Now()
	d := WithClock(time.Hour, fixedClock(now))
	a := makeAlert(8080, "open", "warn")
	d.Allow(a)
	d.Flush()

	if len(d.entries) != 1 {
		t.Fatalf("expected active entry to remain, got %d", len(d.entries))
	}
}
