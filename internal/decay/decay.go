// Package decay provides a time-based score decay mechanism that reduces
// alert severity scores over time when no new events are observed.
package decay

import (
	"sync"
	"time"
)

// Clock is a function that returns the current time.
type Clock func() time.Time

// Entry holds the current score and last update time for a port.
type Entry struct {
	Score     float64
	UpdatedAt time.Time
}

// Decayer applies exponential decay to per-port scores.
type Decayer struct {
	mu       sync.Mutex
	entries  map[int]*Entry
	halfLife time.Duration
	clock    Clock
}

// New creates a Decayer with the given half-life duration.
// Panics if halfLife is zero or negative.
func New(halfLife time.Duration) *Decayer {
	return newWithClock(halfLife, time.Now)
}

func newWithClock(halfLife time.Duration, clock Clock) *Decayer {
	if halfLife <= 0 {
		panic("decay: halfLife must be positive")
	}
	return &Decayer{
		entries:  make(map[int]*Entry),
		halfLife: halfLife,
		clock:    clock,
	}
}

// Add increases the score for the given port by delta after applying
// any pending decay since the last update. Returns the new score.
func (d *Decayer) Add(port int, delta float64) float64 {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := d.clock()
	e, ok := d.entries[port]
	if !ok {
		e = &Entry{Score: 0, UpdatedAt: now}
		d.entries[port] = e
	}

	elapsed := now.Sub(e.UpdatedAt)
	e.Score = d.applyDecay(e.Score, elapsed) + delta
	e.UpdatedAt = now
	return e.Score
}

// Get returns the current decayed score for the given port.
func (d *Decayer) Get(port int) float64 {
	d.mu.Lock()
	defer d.mu.Unlock()

	e, ok := d.entries[port]
	if !ok {
		return 0
	}
	elapsed := d.clock().Sub(e.UpdatedAt)
	return d.applyDecay(e.Score, elapsed)
}

// Reset clears the score for the given port.
func (d *Decayer) Reset(port int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.entries, port)
}

// applyDecay computes score * 0.5^(elapsed/halfLife).
func (d *Decayer) applyDecay(score float64, elapsed time.Duration) float64 {
	if elapsed <= 0 {
		return score
	}
	exponent := float64(elapsed) / float64(d.halfLife)
	// 0.5^exponent = e^(-exponent * ln2)
	// Computed iteratively to avoid math import.
	result := score
	for exponent >= 1 {
		result *= 0.5
		exponent--
	}
	result *= 1.0 - exponent*0.5 // linear approx for fractional part
	return result
}
