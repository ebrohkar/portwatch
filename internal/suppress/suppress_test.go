package suppress

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestNew_EmptyList(t *testing.T) {
	l := New()
	if l.Len() != 0 {
		t.Fatalf("expected 0 entries, got %d", l.Len())
	}
}

func TestAdd_And_IsSuppressed(t *testing.T) {
	base := time.Now()
	l := &List{now: fixedClock(base)}
	l.Add(8080, "open", 5*time.Minute)

	if !l.IsSuppressed(8080, "open") {
		t.Fatal("expected port 8080/open to be suppressed")
	}
}

func TestIsSuppressed_WrongPort(t *testing.T) {
	base := time.Now()
	l := &List{now: fixedClock(base)}
	l.Add(8080, "open", 5*time.Minute)

	if l.IsSuppressed(9090, "open") {
		t.Fatal("port 9090 should not be suppressed")
	}
}

func TestIsSuppressed_WrongEvent(t *testing.T) {
	base := time.Now()
	l := &List{now: fixedClock(base)}
	l.Add(8080, "open", 5*time.Minute)

	if l.IsSuppressed(8080, "close") {
		t.Fatal("event 'close' on port 8080 should not be suppressed")
	}
}

func TestIsSuppressed_EmptyEventMatchesAll(t *testing.T) {
	base := time.Now()
	l := &List{now: fixedClock(base)}
	l.Add(8080, "", 5*time.Minute)

	for _, ev := range []string{"open", "close", "anything"} {
		if !l.IsSuppressed(8080, ev) {
			t.Fatalf("expected event %q on port 8080 to be suppressed by wildcard", ev)
		}
	}
}

func TestIsSuppressed_ExpiredEntry(t *testing.T) {
	base := time.Now()
	l := &List{now: fixedClock(base)}
	l.Add(8080, "open", 1*time.Second)

	// advance clock past expiry
	l.now = fixedClock(base.Add(2 * time.Second))

	if l.IsSuppressed(8080, "open") {
		t.Fatal("expired suppression should not block alert")
	}
}

func TestLen_PrunesExpired(t *testing.T) {
	base := time.Now()
	l := &List{now: fixedClock(base)}
	l.Add(80, "open", 1*time.Second)
	l.Add(443, "open", 10*time.Minute)

	l.now = fixedClock(base.Add(2 * time.Second))

	if got := l.Len(); got != 1 {
		t.Fatalf("expected 1 active suppression after expiry, got %d", got)
	}
}
