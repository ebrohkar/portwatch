package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func instantClock(_ time.Duration) {}

func policyNoSleep(maxAttempts int) Policy {
	return Policy{
		MaxAttempts:  maxAttempts,
		InitialDelay: time.Millisecond,
		Multiplier:   2.0,
		MaxDelay:     time.Second,
		clock:        instantClock,
	}
}

func TestDo_SucceedsFirstAttempt(t *testing.T) {
	p := policyNoSleep(3)
	calls := 0
	err := p.Do(context.Background(), func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestDo_RetriesOnFailure(t *testing.T) {
	p := policyNoSleep(3)
	calls := 0
	sentinel := errors.New("boom")
	err := p.Do(context.Background(), func() error {
		calls++
		if calls < 3 {
			return sentinel
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil after eventual success, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestDo_ExhaustsAttempts(t *testing.T) {
	p := policyNoSleep(3)
	calls := 0
	sentinel := errors.New("persistent")
	err := p.Do(context.Background(), func() error {
		calls++
		return sentinel
	})
	if !errors.Is(err, ErrExhausted) {
		t.Fatalf("expected ErrExhausted, got %v", err)
	}
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected wrapped sentinel error, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestDo_CancelledContext(t *testing.T) {
	p := policyNoSleep(5)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := p.Do(ctx, func() error { return nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestDo_MaxDelayCapApplied(t *testing.T) {
	var slept []time.Duration
	p := Policy{
		MaxAttempts:  4,
		InitialDelay: 100 * time.Millisecond,
		Multiplier:   10.0,
		MaxDelay:     150 * time.Millisecond,
		clock:        func(d time.Duration) { slept = append(slept, d) },
	}
	sentinel := errors.New("err")
	_ = p.Do(context.Background(), func() error { return sentinel })
	for _, d := range slept {
		if d > 150*time.Millisecond {
			t.Fatalf("sleep %v exceeded MaxDelay", d)
		}
	}
}

func TestDefaultPolicy_Fields(t *testing.T) {
	p := DefaultPolicy()
	if p.MaxAttempts != 3 {
		t.Errorf("expected MaxAttempts=3, got %d", p.MaxAttempts)
	}
	if p.Multiplier != 2.0 {
		t.Errorf("expected Multiplier=2.0, got %f", p.Multiplier)
	}
}
