package throttle_test

import (
	"sync"
	"testing"
	"time"

	"github.com/yourorg/portwatch/internal/throttle"
)

func fixedClock(t time.Time) throttle.Clock {
	return func() time.Time { return t }
}

func TestNew_PanicsOnZeroMax(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for max=0")
		}
	}()
	throttle.New(0, time.Second)
}

func TestNew_PanicsOnZeroWindow(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for window=0")
		}
	}()
	throttle.New(1, 0)
}

func TestAllow_PermitsUpToMax(t *testing.T) {
	now := time.Now()
	th := throttle.New(3, time.Minute, throttle.WithClock(fixedClock(now)))

	for i := 0; i < 3; i++ {
		if !th.Allow() {
			t.Fatalf("expected Allow()=true on call %d", i+1)
		}
	}
	if th.Allow() {
		t.Fatal("expected Allow()=false after budget exhausted")
	}
}

func TestAllow_WindowExpiryResetsQuota(t *testing.T) {
	base := time.Now()
	clock := func() time.Time { return base }

	th := throttle.New(2, time.Minute, throttle.WithClock(func() time.Time { return clock() }))
	th.Allow()
	th.Allow()

	if th.Allow() {
		t.Fatal("budget should be exhausted")
	}

	// advance past the window
	base = base.Add(61 * time.Second)

	if !th.Allow() {
		t.Fatal("expected Allow()=true after window expired")
	}
}

func TestRemaining_DecreasesOnAllow(t *testing.T) {
	now := time.Now()
	th := throttle.New(5, time.Minute, throttle.WithClock(fixedClock(now)))

	if got := th.Remaining(); got != 5 {
		t.Fatalf("want 5, got %d", got)
	}
	th.Allow()
	if got := th.Remaining(); got != 4 {
		t.Fatalf("want 4, got %d", got)
	}
}

func TestReset_ClearsState(t *testing.T) {
	now := time.Now()
	th := throttle.New(2, time.Minute, throttle.WithClock(fixedClock(now)))
	th.Allow()
	th.Allow()
	th.Reset()

	if got := th.Remaining(); got != 2 {
		t.Fatalf("want 2 after reset, got %d", got)
	}
}

func TestAllow_ConcurrentSafe(t *testing.T) {
	th := throttle.New(50, time.Minute)
	var wg sync.WaitGroup
	allowed := make(chan bool, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			allowed <- th.Allow()
		}()
	}
	wg.Wait()
	close(allowed)

	count := 0
	for v := range allowed {
		if v {
			count++
		}
	}
	if count != 50 {
		t.Fatalf("want exactly 50 allowed, got %d", count)
	}
}
