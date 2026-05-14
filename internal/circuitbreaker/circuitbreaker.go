// Package circuitbreaker implements a simple circuit breaker that opens after
// a configurable number of consecutive failures and resets after a cooldown.
package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// ErrOpen is returned when the circuit is open and calls are rejected.
var ErrOpen = errors.New("circuit breaker is open")

// State represents the current circuit breaker state.
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

// Clock abstracts time for testing.
type Clock func() time.Time

// Breaker tracks consecutive failures and opens the circuit when the threshold
// is exceeded, allowing recovery after a configurable reset timeout.
type Breaker struct {
	mu           sync.Mutex
	threshold    int
	resetTimeout time.Duration
	clock        Clock

	failures  int
	state     State
	openedAt  time.Time
}

// New creates a Breaker with the given failure threshold and reset timeout.
// Panics if threshold < 1 or resetTimeout <= 0.
func New(threshold int, resetTimeout time.Duration) *Breaker {
	return WithClock(threshold, resetTimeout, time.Now)
}

// WithClock creates a Breaker with a custom clock (useful for testing).
func WithClock(threshold int, resetTimeout time.Duration, clock Clock) *Breaker {
	if threshold < 1 {
		panic("circuitbreaker: threshold must be >= 1")
	}
	if resetTimeout <= 0 {
		panic("circuitbreaker: resetTimeout must be positive")
	}
	if clock == nil {
		panic("circuitbreaker: clock must not be nil")
	}
	return &Breaker{
		threshold:    threshold,
		resetTimeout: resetTimeout,
		clock:        clock,
	}
}

// Allow returns nil if the call is permitted, or ErrOpen if the circuit is open.
func (b *Breaker) Allow() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case StateOpen:
		if b.clock().Sub(b.openedAt) >= b.resetTimeout {
			b.state = StateHalfOpen
			return nil
		}
		return ErrOpen
	default:
		return nil
	}
}

// RecordSuccess resets the breaker to closed state.
func (b *Breaker) RecordSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures = 0
	b.state = StateClosed
}

// RecordFailure increments the failure counter and opens the circuit if the
// threshold is reached.
func (b *Breaker) RecordFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures++
	if b.failures >= b.threshold {
		b.state = StateOpen
		b.openedAt = b.clock()
	}
}

// State returns the current circuit state.
func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}

// Failures returns the current consecutive failure count.
func (b *Breaker) Failures() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.failures
}
