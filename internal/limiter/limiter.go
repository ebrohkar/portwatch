// Package limiter provides a concurrent-alert limiter that caps the number
// of alerts dispatched within a sliding time window across all ports.
package limiter

import (
	"fmt"
	"sync"
	"time"
)

// Clock is a time source used for testing.
type Clock func() time.Time

// Limiter enforces a global cap on alerts per time window.
type Limiter struct {
	mu      sync.Mutex
	max     int
	window  time.Duration
	clock   Clock
	buckets []time.Time
}

// New returns a Limiter that allows at most max alerts per window.
// Panics if max < 1 or window is zero.
func New(max int, window time.Duration) *Limiter {
	if max < 1 {
		panic("limiter: max must be >= 1")
	}
	if window <= 0 {
		panic("limiter: window must be > 0")
	}
	return &Limiter{
		max:    max,
		window: window,
		clock:  time.Now,
	}
}

// withClock returns a Limiter with a custom clock (for testing).
func withClock(max int, window time.Duration, clk Clock) *Limiter {
	l := New(max, window)
	l.clock = clk
	return l
}

// Allow returns true and records the alert timestamp if the global rate
// limit has not been exceeded. Returns false if the cap is reached.
func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.clock()
	cutoff := now.Add(-l.window)

	// Evict expired entries.
	valid := l.buckets[:0]
	for _, t := range l.buckets {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	l.buckets = valid

	if len(l.buckets) >= l.max {
		return false
	}
	l.buckets = append(l.buckets, now)
	return true
}

// Remaining returns the number of additional alerts permitted in the
// current window without advancing the clock.
func (l *Limiter) Remaining() int {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.clock()
	cutoff := now.Add(-l.window)
	count := 0
	for _, t := range l.buckets {
		if t.After(cutoff) {
			count++
		}
	}
	r := l.max - count
	if r < 0 {
		return 0
	}
	return r
}

// String returns a human-readable description of the limiter configuration.
func (l *Limiter) String() string {
	return fmt.Sprintf("Limiter(max=%d, window=%s)", l.max, l.window)
}
