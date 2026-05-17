package jitter

import (
	"testing"
	"time"
)

func TestNew_PanicsOnZeroDuration(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero duration")
		}
	}()
	New(0, 0.1)
}

func TestNew_PanicsOnZeroFactor(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero factor")
		}
	}()
	New(time.Second, 0)
}

func TestNew_PanicsOnFactorAboveOne(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for factor > 1")
		}
	}()
	New(time.Second, 1.1)
}

func TestDuration_WithMaxOffset_ReturnsBase(t *testing.T) {
	// src always returns 0.5 → offset = 0
	j := withSource(10*time.Second, 0.25, func() float64 { return 0.5 })
	got := j.Duration()
	if got != 10*time.Second {
		t.Fatalf("expected 10s, got %v", got)
	}
}

func TestDuration_WithZeroRand_ReturnsBelowBase(t *testing.T) {
	// src returns 0.0 → offset = -factor*base
	j := withSource(10*time.Second, 0.25, func() float64 { return 0.0 })
	got := j.Duration()
	want := 10*time.Second - 2500*time.Millisecond
	if got != want {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestDuration_WithOneRand_ReturnsAboveBase(t *testing.T) {
	// src returns value close to 1.0 → offset ≈ +factor*base
	j := withSource(10*time.Second, 0.25, func() float64 { return 0.9999 })
	got := j.Duration()
	if got <= 10*time.Second {
		t.Fatalf("expected duration > 10s, got %v", got)
	}
}

func TestDuration_StaysWithinBounds(t *testing.T) {
	base := 20 * time.Second
	factor := 0.3
	lo := time.Duration(float64(base) * (1 - factor))
	hi := time.Duration(float64(base) * (1 + factor))

	j := New(base, factor)
	for i := 0; i < 500; i++ {
		d := j.Duration()
		if d < lo || d > hi {
			t.Fatalf("duration %v out of bounds [%v, %v]", d, lo, hi)
		}
	}
}

func TestDuration_FactorOne_DoublesOrZeroesBase(t *testing.T) {
	j := withSource(10*time.Second, 1.0, func() float64 { return 1.0 })
	got := j.Duration()
	// offset = (1*2-1)*1.0*10s = 10s → total = 20s
	if got != 20*time.Second {
		t.Fatalf("expected 20s, got %v", got)
	}
}
