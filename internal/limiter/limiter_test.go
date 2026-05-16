package limiter

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) Clock {
	return func() time.Time { return t }
}

func TestNew_PanicsOnZeroMax(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for max=0")
		}
	}()
	New(0, time.Second)
}

func TestNew_PanicsOnZeroWindow(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for window=0")
		}
	}()
	New(5, 0)
}

func TestAllow_PermitsUpToMax(t *testing.T) {
	now := time.Now()
	l := withClock(3, time.Minute, fixedClock(now))

	for i := 0; i < 3; i++ {
		if !l.Allow() {
			t.Fatalf("expected Allow()=true on call %d", i+1)
		}
	}
	if l.Allow() {
		t.Fatal("expected Allow()=false after max reached")
	}
}

func TestAllow_WindowExpiryResetsQuota(t *testing.T) {
	base := time.Now()
	clk := fixedClock(base)
	l := withClock(2, time.Second, clk)

	l.Allow()
	l.Allow()
	if l.Allow() {
		t.Fatal("expected Allow()=false after quota exhausted")
	}

	// Advance clock beyond the window.
	l.clock = fixedClock(base.Add(2 * time.Second))
	if !l.Allow() {
		t.Fatal("expected Allow()=true after window expiry")
	}
}

func TestRemaining_DecreasesWithAllows(t *testing.T) {
	now := time.Now()
	l := withClock(5, time.Minute, fixedClock(now))

	if l.Remaining() != 5 {
		t.Fatalf("expected Remaining=5, got %d", l.Remaining())
	}
	l.Allow()
	l.Allow()
	if l.Remaining() != 3 {
		t.Fatalf("expected Remaining=3, got %d", l.Remaining())
	}
}

func TestRemaining_ZeroWhenExhausted(t *testing.T) {
	now := time.Now()
	l := withClock(2, time.Minute, fixedClock(now))
	l.Allow()
	l.Allow()

	if r := l.Remaining(); r != 0 {
		t.Fatalf("expected Remaining=0, got %d", r)
	}
}

func TestString_ContainsConfig(t *testing.T) {
	l := New(10, 30*time.Second)
	s := l.String()
	if s == "" {
		t.Fatal("expected non-empty String()")
	}
	expected := "Limiter(max=10, window=30s)"
	if s != expected {
		t.Fatalf("expected %q, got %q", expected, s)
	}
}
