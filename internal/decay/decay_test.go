package decay

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) Clock {
	return func() time.Time { return t }
}

func TestNew_PanicsOnZeroHalfLife(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on zero halfLife")
		}
	}()
	New(0)
}

func TestNew_PanicsOnNegativeHalfLife(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on negative halfLife")
		}
	}()
	New(-time.Second)
}

func TestGet_UnknownPort_ReturnsZero(t *testing.T) {
	d := newWithClock(time.Minute, fixedClock(time.Now()))
	if got := d.Get(8080); got != 0 {
		t.Fatalf("expected 0, got %f", got)
	}
}

func TestAdd_IncreasesScore(t *testing.T) {
	now := time.Now()
	d := newWithClock(time.Minute, fixedClock(now))
	score := d.Add(8080, 10)
	if score != 10 {
		t.Fatalf("expected 10, got %f", score)
	}
}

func TestAdd_AccumulatesWithoutDecay(t *testing.T) {
	now := time.Now()
	d := newWithClock(time.Minute, fixedClock(now))
	d.Add(8080, 5)
	score := d.Add(8080, 5)
	if score != 10 {
		t.Fatalf("expected 10, got %f", score)
	}
}

func TestGet_AfterOneHalfLife_HalvesScore(t *testing.T) {
	now := time.Now()
	halfLife := time.Minute
	d := newWithClock(halfLife, fixedClock(now))
	d.Add(9090, 100)

	// Advance clock by exactly one half-life.
	d.clock = fixedClock(now.Add(halfLife))
	got := d.Get(9090)
	if got != 50 {
		t.Fatalf("expected 50 after one half-life, got %f", got)
	}
}

func TestGet_AfterTwoHalfLives_QuartersScore(t *testing.T) {
	now := time.Now()
	halfLife := time.Minute
	d := newWithClock(halfLife, fixedClock(now))
	d.Add(443, 100)

	d.clock = fixedClock(now.Add(2 * halfLife))
	got := d.Get(443)
	if got != 25 {
		t.Fatalf("expected 25 after two half-lives, got %f", got)
	}
}

func TestReset_ClearsScore(t *testing.T) {
	now := time.Now()
	d := newWithClock(time.Minute, fixedClock(now))
	d.Add(22, 50)
	d.Reset(22)
	if got := d.Get(22); got != 0 {
		t.Fatalf("expected 0 after reset, got %f", got)
	}
}

func TestAdd_DifferentPortsAreIndependent(t *testing.T) {
	now := time.Now()
	d := newWithClock(time.Minute, fixedClock(now))
	d.Add(80, 10)
	d.Add(443, 20)

	if got := d.Get(80); got != 10 {
		t.Fatalf("port 80: expected 10, got %f", got)
	}
	if got := d.Get(443); got != 20 {
		t.Fatalf("port 443: expected 20, got %f", got)
	}
}
