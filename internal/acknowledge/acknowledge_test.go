package acknowledge

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) clock {
	return func() time.Time { return t }
}

var epoch = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

func TestNew_EmptyStore(t *testing.T) {
	s := New()
	if s.Len() != 0 {
		t.Fatalf("expected 0, got %d", s.Len())
	}
}

func TestAcknowledge_And_IsAcknowledged(t *testing.T) {
	s := withClock(fixedClock(epoch))
	s.Acknowledge(8080, "open", time.Minute)
	if !s.IsAcknowledged(8080, "open") {
		t.Fatal("expected port 8080/open to be acknowledged")
	}
}

func TestIsAcknowledged_WrongPort(t *testing.T) {
	s := withClock(fixedClock(epoch))
	s.Acknowledge(8080, "open", time.Minute)
	if s.IsAcknowledged(9090, "open") {
		t.Fatal("expected port 9090/open NOT to be acknowledged")
	}
}

func TestIsAcknowledged_WrongEvent(t *testing.T) {
	s := withClock(fixedClock(epoch))
	s.Acknowledge(8080, "open", time.Minute)
	if s.IsAcknowledged(8080, "closed") {
		t.Fatal("expected port 8080/closed NOT to be acknowledged")
	}
}

func TestIsAcknowledged_Expired(t *testing.T) {
	var now = epoch
	s := withClock(func() time.Time { return now })
	s.Acknowledge(8080, "open", time.Minute)

	now = epoch.Add(2 * time.Minute) // advance past TTL
	if s.IsAcknowledged(8080, "open") {
		t.Fatal("expected acknowledgement to have expired")
	}
}

func TestRevoke_RemovesEntry(t *testing.T) {
	s := withClock(fixedClock(epoch))
	s.Acknowledge(8080, "open", time.Hour)
	s.Revoke(8080, "open")
	if s.IsAcknowledged(8080, "open") {
		t.Fatal("expected acknowledgement to be revoked")
	}
}

func TestRevoke_NonExistentIsNoop(t *testing.T) {
	s := withClock(fixedClock(epoch))
	s.Revoke(1234, "open") // should not panic
}

func TestLen_CountsOnlyActive(t *testing.T) {
	var now = epoch
	s := withClock(func() time.Time { return now })

	s.Acknowledge(80, "open", time.Minute)
	s.Acknowledge(443, "open", time.Minute)
	s.Acknowledge(22, "open", 10*time.Second)

	if s.Len() != 3 {
		t.Fatalf("expected 3, got %d", s.Len())
	}

	now = epoch.Add(30 * time.Second) // port 22 entry has expired
	if got := s.Len(); got != 2 {
		t.Fatalf("expected 2 after partial expiry, got %d", got)
	}
}

func TestAcknowledge_ZeroTTL_IsNoop(t *testing.T) {
	s := withClock(fixedClock(epoch))
	s.Acknowledge(8080, "open", 0)
	if s.IsAcknowledged(8080, "open") {
		t.Fatal("zero-TTL acknowledge should be a no-op")
	}
}
