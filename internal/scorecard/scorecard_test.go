package scorecard

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) Clock {
	return func() time.Time { return t }
}

func TestNew_PanicsOnZeroDecay(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero decayPer")
		}
	}()
	New(0)
}

func TestGet_UnknownPort_ReturnsZero(t *testing.T) {
	sc := newWithClock(time.Minute, fixedClock(time.Now()))
	if got := sc.Get(9999); got != 0 {
		t.Fatalf("expected 0, got %f", got)
	}
}

func TestRecord_IncreasesScore(t *testing.T) {
	now := time.Now()
	sc := newWithClock(time.Hour, fixedClock(now))
	sc.Record(80, 1.0)
	if got := sc.Get(80); got <= 0 {
		t.Fatalf("expected positive score, got %f", got)
	}
}

func TestRecord_AccumulatesMultipleDeltas(t *testing.T) {
	now := time.Now()
	sc := newWithClock(time.Hour, fixedClock(now))
	sc.Record(443, 1.0)
	sc.Record(443, 1.0)
	score := sc.Get(443)
	if score < 1.5 {
		t.Fatalf("expected score >= 1.5, got %f", score)
	}
}

func TestGet_ScoreDecaysAfterHalfLife(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	half := time.Minute

	var now time.Time
	clock := func() time.Time { return now }

	sc := newWithClock(half, clock)
	now = base
	sc.Record(22, 2.0)

	// Advance by exactly one half-life.
	now = base.Add(half)
	got := sc.Get(22)
	// After one half-life the score should be roughly 1.0 (within 20%).
	if got < 0.8 || got > 1.2 {
		t.Fatalf("expected ~1.0 after half-life, got %f", got)
	}
}

func TestReset_ClearsScore(t *testing.T) {
	now := time.Now()
	sc := newWithClock(time.Hour, fixedClock(now))
	sc.Record(8080, 5.0)
	sc.Reset(8080)
	if got := sc.Get(8080); got != 0 {
		t.Fatalf("expected 0 after reset, got %f", got)
	}
}

func TestRecord_DifferentPortsAreIndependent(t *testing.T) {
	now := time.Now()
	sc := newWithClock(time.Hour, fixedClock(now))
	sc.Record(80, 3.0)
	sc.Record(443, 1.0)
	if sc.Get(80) <= sc.Get(443) {
		t.Fatal("expected port 80 to have higher score than port 443")
	}
}

func TestReset_NonExistentPort_IsNoOp(t *testing.T) {
	sc := newWithClock(time.Minute, fixedClock(time.Now()))
	sc.Reset(1234) // should not panic
}
