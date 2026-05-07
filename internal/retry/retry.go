// Package retry provides a simple retry mechanism with backoff for
// transient failures in portwatch subsystems (e.g. notifier sends).
package retry

import (
	"context"
	"errors"
	"time"
)

// Policy defines how retries are attempted.
type Policy struct {
	// MaxAttempts is the total number of attempts (including the first).
	MaxAttempts int
	// InitialDelay is the wait time before the second attempt.
	InitialDelay time.Duration
	// Multiplier scales the delay after each failure.
	Multiplier float64
	// MaxDelay caps the inter-attempt wait regardless of multiplier growth.
	MaxDelay time.Duration
	// clock is used for sleeping; defaults to time.Sleep.
	clock func(time.Duration)
}

// DefaultPolicy returns a Policy suitable for most portwatch use-cases.
func DefaultPolicy() Policy {
	return Policy{
		MaxAttempts:  3,
		InitialDelay: 200 * time.Millisecond,
		Multiplier:   2.0,
		MaxDelay:     5 * time.Second,
		clock:        time.Sleep,
	}
}

// ErrExhausted is returned when all attempts have been consumed.
var ErrExhausted = errors.New("retry: all attempts exhausted")

// Do calls fn up to p.MaxAttempts times, backing off between failures.
// It stops early if ctx is cancelled or fn returns nil.
func (p Policy) Do(ctx context.Context, fn func() error) error {
	if p.MaxAttempts <= 0 {
		p.MaxAttempts = 1
	}
	sleep := p.clock
	if sleep == nil {
		sleep = time.Sleep
	}

	delay := p.InitialDelay
	var lastErr error

	for attempt := 0; attempt < p.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		lastErr = fn()
		if lastErr == nil {
			return nil
		}
		if attempt < p.MaxAttempts-1 {
			wait := delay
			if wait > p.MaxDelay && p.MaxDelay > 0 {
				wait = p.MaxDelay
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-after(wait, sleep):
			}
			delay = time.Duration(float64(delay) * p.Multiplier)
		}
	}
	return errors.Join(ErrExhausted, lastErr)
}

func after(d time.Duration, sleep func(time.Duration)) <-chan struct{} {
	ch := make(chan struct{}, 1)
	go func() {
		sleep(d)
		ch <- struct{}{}
	}()
	return ch
}
