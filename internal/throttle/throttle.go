// Package throttle provides a token-bucket style throttle that limits
// the total number of alerts dispatched per time window across all ports.
package throttle

import (
	"sync"
	"time"
)

// Clock allows injecting a fake time source in tests.
type Clock func() time.Time

// Throttle tracks a global alert budget per rolling window.
type Throttle struct {
	mu       sync.Mutex
	max      int
	window   time.Duration
	clock    Clock
	buckets  []time.Time // timestamps of recent allowed events
}

// Option is a functional option for Throttle.
type Option func(*Throttle)

// WithClock overrides the time source (useful for testing).
func WithClock(c Clock) Option {
	return func(t *Throttle) { t.clock = c }
}

// New creates a Throttle that permits at most max events per window.
// max must be >= 1 and window must be > 0, otherwise New panics.
func New(max int, window time.Duration, opts ...Option) *Throttle {
	if max < 1 {
		panic("throttle: max must be >= 1")
	}
	if window <= 0 {
		panic("throttle: window must be > 0")
	}
	t := &Throttle{
		max:    max,
		window: window,
		clock:  time.Now,
		buckets: make([]time.Time, 0, max),
	}
	for _, o := range opts {
		o(t)
	}
	return t
}

// Allow returns true if the event is within budget, recording it if so.
func (t *Throttle) Allow() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.clock()
	cutoff := now.Add(-t.window)

	// evict expired timestamps
	valid := t.buckets[:0]
	for _, ts := range t.buckets {
		if ts.After(cutoff) {
			valid = append(valid, ts)
		}
	}
	t.buckets = valid

	if len(t.buckets) >= t.max {
		return false
	}
	t.buckets = append(t.buckets, now)
	return true
}

// Remaining returns how many more events are permitted in the current window.
func (t *Throttle) Remaining() int {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.clock()
	cutoff := now.Add(-t.window)
	count := 0
	for _, ts := range t.buckets {
		if ts.After(cutoff) {
			count++
		}
	}
	r := t.max - count
	if r < 0 {
		return 0
	}
	return r
}

// Reset clears all recorded events.
func (t *Throttle) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.buckets = t.buckets[:0]
}
