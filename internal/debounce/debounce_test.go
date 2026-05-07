package debounce

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestNew_DefaultCooldown(t *testing.T) {
	d := New(0)
	if d.cooldown != DefaultCooldown {
		t.Fatalf("expected default cooldown %v, got %v", DefaultCooldown, d.cooldown)
	}
}

func TestNew_CustomCooldown(t *testing.T) {
	d := New(5 * time.Second)
	if d.cooldown != 5*time.Second {
		t.Fatalf("expected 5s cooldown, got %v", d.cooldown)
	}
}

func TestAllow_FirstCallPermitted(t *testing.T) {
	d := New(10 * time.Second)
	if !d.Allow(8080, "open") {
		t.Fatal("expected first call to be allowed")
	}
}

func TestAllow_SecondCallWithinCooldownBlocked(t *testing.T) {
	now := time.Now()
	d := New(10 * time.Second)
	d.now = fixedClock(now)

	d.Allow(8080, "open")
	if d.Allow(8080, "open") {
		t.Fatal("expected second call within cooldown to be blocked")
	}
}

func TestAllow_CallAfterCooldownPermitted(t *testing.T) {
	now := time.Now()
	d := New(10 * time.Second)
	d.now = fixedClock(now)
	d.Allow(8080, "open")

	// Advance clock beyond cooldown.
	d.now = fixedClock(now.Add(11 * time.Second))
	if !d.Allow(8080, "open") {
		t.Fatal("expected call after cooldown to be allowed")
	}
}

func TestAllow_DifferentEventsAreIndependent(t *testing.T) {
	d := New(10 * time.Second)
	d.Allow(8080, "open")
	if !d.Allow(8080, "close") {
		t.Fatal("expected different event on same port to be allowed")
	}
}

func TestAllow_DifferentPortsAreIndependent(t *testing.T) {
	d := New(10 * time.Second)
	d.Allow(8080, "open")
	if !d.Allow(9090, "open") {
		t.Fatal("expected same event on different port to be allowed")
	}
}

func TestReset_ClearsState(t *testing.T) {
	now := time.Now()
	d := New(10 * time.Second)
	d.now = fixedClock(now)
	d.Allow(8080, "open")

	d.Reset(8080, "open")
	if !d.Allow(8080, "open") {
		t.Fatal("expected allow after reset")
	}
}

func TestPurge_RemovesExpiredEntries(t *testing.T) {
	now := time.Now()
	d := New(10 * time.Second)
	d.now = fixedClock(now)
	d.Allow(8080, "open")
	d.Allow(9090, "open")

	// Advance clock so entries are expired.
	d.now = fixedClock(now.Add(11 * time.Second))
	d.Purge()

	if len(d.entries) != 0 {
		t.Fatalf("expected entries to be purged, got %d", len(d.entries))
	}
}
