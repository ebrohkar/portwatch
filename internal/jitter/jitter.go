// Package jitter adds randomised delay to periodic operations to avoid
// thundering-herd problems when many portwatch instances run together.
package jitter

import (
	"math/rand"
	"sync"
	"time"
)

// Source is a function that returns a pseudo-random float64 in [0.0, 1.0).
type Source func() float64

// Jitter applies a bounded random offset to a base duration.
type Jitter struct {
	mu     sync.Mutex
	base   time.Duration
	factor float64 // fraction of base, e.g. 0.25 → ±25 %
	src    Source
}

// New returns a Jitter that offsets d by up to factor*d.
// factor must be in (0, 1]; it panics otherwise.
func New(d time.Duration, factor float64) *Jitter {
	if factor <= 0 || factor > 1 {
		panic("jitter: factor must be in (0, 1]")
	}
	if d <= 0 {
		panic("jitter: base duration must be positive")
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
	return &Jitter{
		base:   d,
		factor: factor,
		src:    r.Float64,
	}
}

// withSource replaces the random source; used in tests.
func withSource(d time.Duration, factor float64, src Source) *Jitter {
	j := New(d, factor)
	j.src = src
	return j
}

// Duration returns the base duration offset by a random amount in
// [-factor*base, +factor*base].
func (j *Jitter) Duration() time.Duration {
	j.mu.Lock()
	v := j.src()
	j.mu.Unlock()

	// map [0,1) → [-1,1)
	offset := (v*2 - 1) * j.factor * float64(j.base)
	return j.base + time.Duration(offset)
}

// Sleep blocks for Duration().
func (j *Jitter) Sleep() {
	time.Sleep(j.Duration())
}
