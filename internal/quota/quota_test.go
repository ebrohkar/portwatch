package quota

import (
	"testing"
	"time"
)

var (
	now   = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	fixed = func() time.Time { return now }
)

func TestNew_PanicsOnZeroMax(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for max=0")
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
	New(1, 0)
}

func TestAllow_PermitsUpToMax(t *testing.T) {
	q := newWithClock(3, time.Minute, fixed)
	for i := 0; i < 3; i++ {
		if !q.Allow(80) {
			t.Fatalf("call %d should be allowed", i+1)
		}
	}
	if q.Allow(80) {
		t.Fatal("4th call should be denied")
	}
}

func TestAllow_DifferentPortsAreIndependent(t *testing.T) {
	q := newWithClock(1, time.Minute, fixed)
	if !q.Allow(80) {
		t.Fatal("port 80 first call should be allowed")
	}
	if !q.Allow(443) {
		t.Fatal("port 443 first call should be allowed")
	}
	if q.Allow(80) {
		t.Fatal("port 80 second call should be denied")
	}
}

func TestAllow_WindowExpiryResetsQuota(t *testing.T) {
	var current = now
	clock := func() time.Time { return current }

	q := newWithClock(2, time.Minute, clock)
	q.Allow(8080)
	q.Allow(8080)
	if q.Allow(8080) {
		t.Fatal("should be denied before window expiry")
	}

	current = now.Add(2 * time.Minute)
	if !q.Allow(8080) {
		t.Fatal("should be allowed after window expiry")
	}
}

func TestRemaining_DecreasesWithUse(t *testing.T) {
	q := newWithClock(3, time.Minute, fixed)
	if got := q.Remaining(9090); got != 3 {
		t.Fatalf("want 3, got %d", got)
	}
	q.Allow(9090)
	if got := q.Remaining(9090); got != 2 {
		t.Fatalf("want 2, got %d", got)
	}
}

func TestReset_ClearsPortState(t *testing.T) {
	q := newWithClock(1, time.Minute, fixed)
	q.Allow(22)
	if q.Allow(22) {
		t.Fatal("should be denied before reset")
	}
	q.Reset(22)
	if !q.Allow(22) {
		t.Fatal("should be allowed after reset")
	}
}

func TestString_ContainsConfig(t *testing.T) {
	q := New(5, time.Minute)
	s := q.String()
	if s == "" {
		t.Fatal("String should not be empty")
	}
}
