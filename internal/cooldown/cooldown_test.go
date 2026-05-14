package cooldown_test

import (
	"testing"
	"time"

	"github.com/user/portwatch/internal/cooldown"
)

func fixedClock(t time.Time) cooldown.Clock {
	return func() time.Time { return t }
}

func TestNew_PanicsOnZeroDuration(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero duration")
		}
	}()
	cooldown.New(0)
}

func TestNew_PanicsOnNilClock(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil clock")
		}
	}()
	cooldown.WithClock(time.Second, nil)
}

func TestAllow_FirstCallPermitted(t *testing.T) {
	now := time.Now()
	s := cooldown.WithClock(5*time.Second, fixedClock(now))
	if !s.Allow(8080) {
		t.Fatal("expected first call to be allowed")
	}
}

func TestAllow_SecondCallWithinCooldownBlocked(t *testing.T) {
	now := time.Now()
	s := cooldown.WithClock(5*time.Second, fixedClock(now))
	s.Allow(8080)
	if s.Allow(8080) {
		t.Fatal("expected second call within cooldown to be blocked")
	}
}

func TestAllow_CallAfterCooldownPermitted(t *testing.T) {
	base := time.Now()
	var current time.Time
	clock := func() time.Time { return current }

	s := cooldown.WithClock(5*time.Second, clock)
	current = base
	s.Allow(8080)

	current = base.Add(6 * time.Second)
	if !s.Allow(8080) {
		t.Fatal("expected call after cooldown to be allowed")
	}
}

func TestAllow_DifferentPortsAreIndependent(t *testing.T) {
	now := time.Now()
	s := cooldown.WithClock(5*time.Second, fixedClock(now))
	s.Allow(8080)
	if !s.Allow(9090) {
		t.Fatal("expected different port to be allowed independently")
	}
}

func TestReset_AllowsImmediateNext(t *testing.T) {
	now := time.Now()
	s := cooldown.WithClock(5*time.Second, fixedClock(now))
	s.Allow(8080)
	s.Reset(8080)
	if !s.Allow(8080) {
		t.Fatal("expected allow after reset")
	}
}

func TestLen_TracksEntries(t *testing.T) {
	now := time.Now()
	s := cooldown.WithClock(5*time.Second, fixedClock(now))
	if s.Len() != 0 {
		t.Fatalf("expected 0, got %d", s.Len())
	}
	s.Allow(8080)
	s.Allow(9090)
	if s.Len() != 2 {
		t.Fatalf("expected 2, got %d", s.Len())
	}
	s.Reset(8080)
	if s.Len() != 1 {
		t.Fatalf("expected 1, got %d", s.Len())
	}
}
