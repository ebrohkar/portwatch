package correlator

import (
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func makeAlert(event string, port int) alert.Alert {
	return alert.Alert{
		Port:     port,
		Event:    event,
		Severity: "info",
	}
}

func TestNew_PanicsOnZeroWindow(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero window")
		}
	}()
	New(0)
}

func TestAdd_CreatesNewGroup(t *testing.T) {
	now := time.Now()
	c := withClock(5*time.Second, fixedClock(now))

	id := c.Add(makeAlert("open", 8080))
	if id == "" {
		t.Fatal("expected non-empty group ID")
	}
	if c.Len() != 1 {
		t.Fatalf("expected 1 group, got %d", c.Len())
	}
}

func TestAdd_SameEventWithinWindow_MergesGroup(t *testing.T) {
	now := time.Now()
	c := withClock(5*time.Second, fixedClock(now))

	id1 := c.Add(makeAlert("open", 8080))
	id2 := c.Add(makeAlert("open", 9090))

	if id1 != id2 {
		t.Fatalf("expected same group ID, got %q and %q", id1, id2)
	}
	if c.Len() != 1 {
		t.Fatalf("expected 1 group, got %d", c.Len())
	}
}

func TestAdd_SameEventOutsideWindow_StartsNewGroup(t *testing.T) {
	now := time.Now()
	c := withClock(1*time.Second, fixedClock(now))

	id1 := c.Add(makeAlert("open", 8080))

	// advance clock beyond window
	c.clock = fixedClock(now.Add(2 * time.Second))
	id2 := c.Add(makeAlert("open", 9090))

	if id1 == id2 {
		t.Fatal("expected different group IDs after window expiry")
	}
	if c.Len() != 2 {
		t.Fatalf("expected 2 groups, got %d", c.Len())
	}
}

func TestAdd_DifferentEvents_CreateSeparateGroups(t *testing.T) {
	now := time.Now()
	c := withClock(10*time.Second, fixedClock(now))

	c.Add(makeAlert("open", 8080))
	c.Add(makeAlert("close", 8080))

	if c.Len() != 2 {
		t.Fatalf("expected 2 groups for different events, got %d", c.Len())
	}
}

func TestFlush_ReturnsGroupsAndResets(t *testing.T) {
	now := time.Now()
	c := withClock(10*time.Second, fixedClock(now))

	c.Add(makeAlert("open", 8080))
	c.Add(makeAlert("close", 443))

	groups := c.Flush()
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups from Flush, got %d", len(groups))
	}
	if c.Len() != 0 {
		t.Fatalf("expected 0 groups after Flush, got %d", c.Len())
	}
}

func TestFlush_GroupContainsAllAlerts(t *testing.T) {
	now := time.Now()
	c := withClock(10*time.Second, fixedClock(now))

	c.Add(makeAlert("open", 8080))
	c.Add(makeAlert("open", 8081))
	c.Add(makeAlert("open", 8082))

	groups := c.Flush()
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if len(groups[0].Alerts) != 3 {
		t.Fatalf("expected 3 alerts in group, got %d", len(groups[0].Alerts))
	}
}
