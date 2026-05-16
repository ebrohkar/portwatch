package window

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) Clock {
	return func() time.Time { return t }
}

func TestNew_PanicsOnZeroSize(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero size")
		}
	}()
	New(0)
}

func TestAdd_ReturnsCount(t *testing.T) {
	now := time.Now()
	w := New(time.Minute).WithClock(fixedClock(now))

	if c := w.Add(80); c != 1 {
		t.Fatalf("expected 1, got %d", c)
	}
	if c := w.Add(80); c != 2 {
		t.Fatalf("expected 2, got %d", c)
	}
}

func TestCount_ReflectsWindow(t *testing.T) {
	base := time.Now()
	w := New(time.Minute)

	// Add two events at base time.
	w.WithClock(fixedClock(base))
	w.Add(443)
	w.Add(443)

	// Advance clock beyond window; events should be evicted.
	w.WithClock(fixedClock(base.Add(2 * time.Minute)))
	if c := w.Count(443); c != 0 {
		t.Fatalf("expected 0 after window expiry, got %d", c)
	}
}

func TestAdd_DifferentKeysAreIndependent(t *testing.T) {
	now := time.Now()
	w := New(time.Minute).WithClock(fixedClock(now))

	w.Add(80)
	w.Add(80)
	w.Add(443)

	if c := w.Count(80); c != 2 {
		t.Fatalf("port 80: expected 2, got %d", c)
	}
	if c := w.Count(443); c != 1 {
		t.Fatalf("port 443: expected 1, got %d", c)
	}
}

func TestReset_ClearsKey(t *testing.T) {
	now := time.Now()
	w := New(time.Minute).WithClock(fixedClock(now))

	w.Add(8080)
	w.Add(8080)
	w.Reset(8080)

	if c := w.Count(8080); c != 0 {
		t.Fatalf("expected 0 after reset, got %d", c)
	}
}

func TestAdd_EvictsExpiredEntries(t *testing.T) {
	base := time.Now()
	w := New(30 * time.Second)

	w.WithClock(fixedClock(base))
	w.Add(9090)
	w.Add(9090)

	// Advance by 40 s — old entries should be gone, new one added.
	w.WithClock(fixedClock(base.Add(40 * time.Second)))
	count := w.Add(9090)
	if count != 1 {
		t.Fatalf("expected 1 after eviction + new add, got %d", count)
	}
}
