// Package pressure tracks alert volume over a sliding window and exposes
// a normalised load value in the range [0.0, 1.0].
package pressure

import (
	"sync"
	"time"
)

// Clock allows tests to inject a fake time source.
type Clock func() time.Time

// Gauge accumulates alert counts and reports the current pressure level
// relative to a configurable capacity.
type Gauge struct {
	mu       sync.Mutex
	clock    Clock
	window   time.Duration
	capacity int
	buckets  []bucket
}

type bucket struct {
	at    time.Time
	count int
}

// New creates a Gauge that measures pressure over the given window against
// the given capacity. It panics if either value is zero.
func New(window time.Duration, capacity int) *Gauge {
	return newWithClock(window, capacity, time.Now)
}

func newWithClock(window time.Duration, capacity int, clock Clock) *Gauge {
	if window <= 0 {
		panic("pressure: window must be positive")
	}
	if capacity <= 0 {
		panic("pressure: capacity must be positive")
	}
	return &Gauge{clock: clock, window: window, capacity: capacity}
}

// Record adds n events at the current time.
func (g *Gauge) Record(n int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	now := g.clock()
	g.evict(now)
	g.buckets = append(g.buckets, bucket{at: now, count: n})
}

// Level returns the current pressure as a value in [0.0, 1.0].
// A value ≥ 1.0 means the capacity has been reached or exceeded.
func (g *Gauge) Level() float64 {
	g.mu.Lock()
	defer g.mu.Unlock()
	now := g.clock()
	g.evict(now)
	var total int
	for _, b := range g.buckets {
		total += b.count
	}
	v := float64(total) / float64(g.capacity)
	if v > 1.0 {
		return 1.0
	}
	return v
}

// Reset clears all accumulated state.
func (g *Gauge) Reset() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.buckets = g.buckets[:0]
}

func (g *Gauge) evict(now time.Time) {
	cutoff := now.Add(-g.window)
	i := 0
	for i < len(g.buckets) && g.buckets[i].at.Before(cutoff) {
		i++
	}
	g.buckets = g.buckets[i:]
}
