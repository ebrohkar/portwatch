package circuitbreaker_test

import (
	"testing"
	"time"

	"github.com/user/portwatch/internal/circuitbreaker"
)

var (
	now     = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	fixedAt = now
)

func fixedClock(t time.Time) circuitbreaker.Clock {
	return func() time.Time { return t }
}

func TestNew_PanicsOnZeroThreshold(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero threshold")
		}
	}()
	circuitbreaker.New(0, time.Second)
}

func TestNew_PanicsOnZeroTimeout(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero reset timeout")
		}
	}()
	circuitbreaker.New(1, 0)
}

func TestAllow_ClosedByDefault(t *testing.T) {
	b := circuitbreaker.New(3, time.Second)
	if err := b.Allow(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestRecordFailure_BelowThreshold_StaysClosed(t *testing.T) {
	b := circuitbreaker.New(3, time.Second)
	b.RecordFailure()
	b.RecordFailure()
	if b.State() != circuitbreaker.StateClosed {
		t.Fatal("expected circuit to remain closed below threshold")
	}
	if b.Failures() != 2 {
		t.Fatalf("expected 2 failures, got %d", b.Failures())
	}
}

func TestRecordFailure_AtThreshold_Opens(t *testing.T) {
	b := circuitbreaker.New(3, time.Second)
	b.RecordFailure()
	b.RecordFailure()
	b.RecordFailure()
	if b.State() != circuitbreaker.StateOpen {
		t.Fatal("expected circuit to open at threshold")
	}
	if err := b.Allow(); err != circuitbreaker.ErrOpen {
		t.Fatalf("expected ErrOpen, got %v", err)
	}
}

func TestAllow_HalfOpenAfterTimeout(t *testing.T) {
	var current = fixedAt
	clock := func() time.Time { return current }

	b := circuitbreaker.WithClock(2, 5*time.Second, clock)
	b.RecordFailure()
	b.RecordFailure()

	// advance past reset timeout
	current = fixedAt.Add(6 * time.Second)

	if err := b.Allow(); err != nil {
		t.Fatalf("expected half-open to permit call, got %v", err)
	}
	if b.State() != circuitbreaker.StateHalfOpen {
		t.Fatalf("expected StateHalfOpen, got %v", b.State())
	}
}

func TestRecordSuccess_ResetsClosed(t *testing.T) {
	b := circuitbreaker.New(2, time.Second)
	b.RecordFailure()
	b.RecordFailure()
	b.RecordSuccess()
	if b.State() != circuitbreaker.StateClosed {
		t.Fatal("expected closed after success")
	}
	if b.Failures() != 0 {
		t.Fatalf("expected 0 failures after reset, got %d", b.Failures())
	}
}

func TestAllow_StillOpenBeforeTimeout(t *testing.T) {
	var current = fixedAt
	clock := func() time.Time { return current }

	b := circuitbreaker.WithClock(1, 10*time.Second, clock)
	b.RecordFailure()

	current = fixedAt.Add(3 * time.Second) // before timeout

	if err := b.Allow(); err != circuitbreaker.ErrOpen {
		t.Fatalf("expected ErrOpen before timeout, got %v", err)
	}
}
