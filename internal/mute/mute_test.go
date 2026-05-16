package mute

import (
	"testing"
	"time"
)

var epoch = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

func fixedClock(t time.Time) Clock {
	return func() time.Time { return t }
}

func TestNew_EmptyStore(t *testing.T) {
	s := New()
	if got := s.Active(); len(got) != 0 {
		t.Fatalf("expected empty active list, got %d entries", len(got))
	}
}

func TestAdd_And_IsMuted(t *testing.T) {
	s := withClock(fixedClock(epoch))
	s.Add(8080, time.Hour, "maintenance")
	if !s.IsMuted(8080) {
		t.Fatal("expected port 8080 to be muted")
	}
}

func TestIsMuted_WrongPort(t *testing.T) {
	s := withClock(fixedClock(epoch))
	s.Add(8080, time.Hour, "")
	if s.IsMuted(9090) {
		t.Fatal("expected port 9090 to not be muted")
	}
}

func TestIsMuted_ExpiredEntry(t *testing.T) {
	now := epoch
	s := withClock(func() time.Time { return now })
	s.Add(8080, time.Minute, "")
	now = epoch.Add(2 * time.Minute) // advance past expiry
	if s.IsMuted(8080) {
		t.Fatal("expected expired entry to not be muted")
	}
}

func TestActive_ReturnsOnlyLive(t *testing.T) {
	now := epoch
	s := withClock(func() time.Time { return now })
	s.Add(8080, time.Hour, "long")
	s.Add(9090, time.Second, "short")
	now = epoch.Add(2 * time.Second)
	active := s.Active()
	if len(active) != 1 {
		t.Fatalf("expected 1 active entry, got %d", len(active))
	}
	if active[0].Port != 8080 {
		t.Fatalf("expected port 8080, got %d", active[0].Port)
	}
}

func TestClear_RemovesPort(t *testing.T) {
	s := withClock(fixedClock(epoch))
	s.Add(8080, time.Hour, "")
	s.Add(9090, time.Hour, "")
	s.Clear(8080)
	if s.IsMuted(8080) {
		t.Fatal("expected 8080 to be cleared")
	}
	if !s.IsMuted(9090) {
		t.Fatal("expected 9090 to still be muted")
	}
}

func TestAdd_ReasonStored(t *testing.T) {
	s := withClock(fixedClock(epoch))
	s.Add(443, time.Hour, "patch window")
	active := s.Active()
	if len(active) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(active))
	}
	if active[0].Reason != "patch window" {
		t.Fatalf("unexpected reason: %q", active[0].Reason)
	}
}
