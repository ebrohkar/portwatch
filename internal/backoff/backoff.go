// Package backoff provides exponential backoff with jitter for retry delays.
package backoff

import (
	"math"
	"math/rand"
	"time"
)

// Clock abstracts time for testing.
type Clock func() time.Time

// Backoff computes delay durations using exponential backoff with optional jitter.
type Backoff struct {
	base    time.Duration
	max     time.Duration
	factor  float64
	jitter  bool
	attempt int
	clock   Clock
}

// Option configures a Backoff.
type Option func(*Backoff)

// WithJitter enables random jitter up to 25% of the computed delay.
func WithJitter() Option {
	return func(b *Backoff) { b.jitter = true }
}

// WithFactor sets the exponential growth factor (default 2.0).
func WithFactor(f float64) Option {
	return func(b *Backoff) {
		if f < 1 {
			panic("backoff: factor must be >= 1")
		}
		b.factor = f
	}
}

// New creates a Backoff with the given base and max durations.
// Panics if base or max is zero, or if base > max.
func New(base, max time.Duration, opts ...Option) *Backoff {
	if base <= 0 {
		panic("backoff: base must be > 0")
	}
	if max <= 0 {
		panic("backoff: max must be > 0")
	}
	if base > max {
		panic("backoff: base must be <= max")
	}
	b := &Backoff{
		base:   base,
		max:    max,
		factor: 2.0,
		clock:  time.Now,
	}
	for _, o := range opts {
		o(b)
	}
	return b
}

// Next returns the delay for the current attempt and advances the internal counter.
func (b *Backoff) Next() time.Duration {
	delay := float64(b.base) * math.Pow(b.factor, float64(b.attempt))
	if delay > float64(b.max) {
		delay = float64(b.max)
	}
	if b.jitter {
		delay += delay * 0.25 * rand.Float64() //nolint:gosec
		if delay > float64(b.max) {
			delay = float64(b.max)
		}
	}
	b.attempt++
	return time.Duration(delay)
}

// Reset resets the attempt counter to zero.
func (b *Backoff) Reset() {
	b.attempt = 0
}

// Attempt returns the current attempt index (zero-based).
func (b *Backoff) Attempt() int {
	return b.attempt
}
