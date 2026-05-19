package pause

import (
	"testing"
	"time"
)

// fixedClock returns a Clock that always returns t.
func fixedClock(t time.Time) Clock {
	return func() time.Time { return t }
}

var epoch = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

func TestNew_EmptyStore(t *testing.T) {
	s := New()
	if s.Len() != 0 {
		t.Fatalf("expected 0 entries, got %d", s.Len())
	}
}

func TestPause_And_IsPaused(t *testing.T) {
	s := withClock(fixedClock(epoch))
	if err := s.Pause(8080, "open", time.Minute); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !s.IsPaused(8080, "open") {
		t.Fatal("expected port to be paused")
	}
}

func TestIsPaused_WrongPort(t *testing.T) {
	s := withClock(fixedClock(epoch))
	_ = s.Pause(8080, "open", time.Minute)
	if s.IsPaused(9090, "open") {
		t.Fatal("different port should not be paused")
	}
}

func TestIsPaused_WrongEvent(t *testing.T) {
	s := withClock(fixedClock(epoch))
	_ = s.Pause(8080, "open", time.Minute)
	if s.IsPaused(8080, "close") {
		t.Fatal("different event should not be paused")
	}
}

func TestIsPaused_ExpiredEntry(t *testing.T) {
	now := epoch
	s := withClock(func() time.Time { return now })
	_ = s.Pause(8080, "open", time.Minute)
	now = epoch.Add(2 * time.Minute) // advance past expiry
	if s.IsPaused(8080, "open") {
		t.Fatal("expired entry should not be paused")
	}
}

func TestResume_LiftsPauseEarly(t *testing.T) {
	s := withClock(fixedClock(epoch))
	_ = s.Pause(8080, "open", time.Hour)
	s.Resume(8080, "open")
	if s.IsPaused(8080, "open") {
		t.Fatal("expected pause to be lifted after Resume")
	}
}

func TestPause_ExtendsExpiry(t *testing.T) {
	now := epoch
	s := withClock(func() time.Time { return now })
	_ = s.Pause(8080, "open", time.Minute)
	_ = s.Pause(8080, "open", 2*time.Hour) // extend
	now = epoch.Add(90 * time.Minute)       // past original, within extended
	if !s.IsPaused(8080, "open") {
		t.Fatal("expected extended pause to still be active")
	}
}

func TestPause_InvalidPort(t *testing.T) {
	s := New()
	if err := s.Pause(0, "open", time.Minute); err == nil {
		t.Fatal("expected error for port 0")
	}
	if err := s.Pause(70000, "open", time.Minute); err == nil {
		t.Fatal("expected error for port 70000")
	}
}

func TestPause_EmptyEvent(t *testing.T) {
	s := New()
	if err := s.Pause(8080, "", time.Minute); err == nil {
		t.Fatal("expected error for empty event")
	}
}

func TestPause_NegativeDuration(t *testing.T) {
	s := New()
	if err := s.Pause(8080, "open", -time.Second); err == nil {
		t.Fatal("expected error for negative duration")
	}
}

func TestLen_CountsActiveEntries(t *testing.T) {
	now := epoch
	s := withClock(func() time.Time { return now })
	_ = s.Pause(80, "open", time.Minute)
	_ = s.Pause(443, "open", time.Minute)
	_ = s.Pause(22, "close", 5*time.Second)
	now = epoch.Add(10 * time.Second) // expire port 22
	if got := s.Len(); got != 2 {
		t.Fatalf("expected 2 active entries, got %d", got)
	}
}
