package backoff_test

import (
	"testing"
	"time"

	"github.com/example/portwatch/internal/backoff"
)

func TestNew_PanicsOnZeroBase(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero base")
		}
	}()
	backoff.New(0, time.Second)
}

func TestNew_PanicsOnZeroMax(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero max")
		}
	}()
	backoff.New(time.Second, 0)
}

func TestNew_PanicsWhenBaseExceedsMax(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when base > max")
		}
	}()
	backoff.New(2*time.Second, time.Second)
}

func TestWithFactor_PanicsOnFactorBelowOne(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for factor < 1")
		}
	}()
	backoff.New(time.Millisecond, time.Second, backoff.WithFactor(0.5))
}

func TestNext_ExponentialGrowth(t *testing.T) {
	b := backoff.New(100*time.Millisecond, 10*time.Second)

	d0 := b.Next()
	d1 := b.Next()
	d2 := b.Next()

	if d0 != 100*time.Millisecond {
		t.Fatalf("attempt 0: got %v, want 100ms", d0)
	}
	if d1 != 200*time.Millisecond {
		t.Fatalf("attempt 1: got %v, want 200ms", d1)
	}
	if d2 != 400*time.Millisecond {
		t.Fatalf("attempt 2: got %v, want 400ms", d2)
	}
}

func TestNext_CapsAtMax(t *testing.T) {
	b := backoff.New(time.Second, 3*time.Second)

	var last time.Duration
	for i := 0; i < 10; i++ {
		last = b.Next()
	}
	if last > 3*time.Second {
		t.Fatalf("delay exceeded max: got %v", last)
	}
}

func TestReset_ResetsAttemptCounter(t *testing.T) {
	b := backoff.New(100*time.Millisecond, time.Second)
	b.Next()
	b.Next()
	if b.Attempt() != 2 {
		t.Fatalf("expected attempt=2, got %d", b.Attempt())
	}
	b.Reset()
	if b.Attempt() != 0 {
		t.Fatalf("expected attempt=0 after reset, got %d", b.Attempt())
	}
	if b.Next() != 100*time.Millisecond {
		t.Fatal("expected base delay after reset")
	}
}

func TestWithJitter_DelayWithinBounds(t *testing.T) {
	b := backoff.New(100*time.Millisecond, 2*time.Second, backoff.WithJitter())

	for i := 0; i < 20; i++ {
		d := b.Next()
		if d <= 0 {
			t.Fatalf("jitter produced non-positive delay: %v", d)
		}
		if d > 2*time.Second {
			t.Fatalf("jitter exceeded max: %v", d)
		}
	}
}

func TestWithFactor_CustomFactor(t *testing.T) {
	b := backoff.New(100*time.Millisecond, time.Minute, backoff.WithFactor(3.0))

	d0 := b.Next()
	d1 := b.Next()

	if d0 != 100*time.Millisecond {
		t.Fatalf("attempt 0: got %v, want 100ms", d0)
	}
	if d1 != 300*time.Millisecond {
		t.Fatalf("attempt 1: got %v, want 300ms", d1)
	}
}
