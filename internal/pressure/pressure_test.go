package pressure

import (
	"testing"
	"time"
)

var epoch = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

func fixedClock(t time.Time) Clock {
	return func() time.Time { return t }
}

func TestNew_PanicsOnZeroWindow(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	New(0, 10)
}

func TestNew_PanicsOnZeroCapacity(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	New(time.Minute, 0)
}

func TestLevel_NoEvents_ReturnsZero(t *testing.T) {
	g := newWithClock(time.Minute, 100, fixedClock(epoch))
	if got := g.Level(); got != 0.0 {
		t.Fatalf("expected 0.0, got %f", got)
	}
}

func TestLevel_BelowCapacity(t *testing.T) {
	g := newWithClock(time.Minute, 100, fixedClock(epoch))
	g.Record(25)
	if got := g.Level(); got != 0.25 {
		t.Fatalf("expected 0.25, got %f", got)
	}
}

func TestLevel_AtCapacity_ReturnsOne(t *testing.T) {
	g := newWithClock(time.Minute, 10, fixedClock(epoch))
	g.Record(10)
	if got := g.Level(); got != 1.0 {
		t.Fatalf("expected 1.0, got %f", got)
	}
}

func TestLevel_ExceedsCapacity_ClampedToOne(t *testing.T) {
	g := newWithClock(time.Minute, 10, fixedClock(epoch))
	g.Record(50)
	if got := g.Level(); got != 1.0 {
		t.Fatalf("expected 1.0, got %f", got)
	}
}

func TestLevel_OldEventsEvicted(t *testing.T) {
	now := epoch
	clock := func() time.Time { return now }
	g := newWithClock(time.Minute, 100, clock)
	g.Record(50)
	now = epoch.Add(2 * time.Minute) // advance past window
	if got := g.Level(); got != 0.0 {
		t.Fatalf("expected 0.0 after eviction, got %f", got)
	}
}

func TestReset_ClearsState(t *testing.T) {
	g := newWithClock(time.Minute, 100, fixedClock(epoch))
	g.Record(80)
	g.Reset()
	if got := g.Level(); got != 0.0 {
		t.Fatalf("expected 0.0 after reset, got %f", got)
	}
}

func TestRecord_AccumulatesMultipleCalls(t *testing.T) {
	g := newWithClock(time.Minute, 100, fixedClock(epoch))
	g.Record(10)
	g.Record(20)
	g.Record(30)
	if got := g.Level(); got != 0.60 {
		t.Fatalf("expected 0.60, got %f", got)
	}
}
