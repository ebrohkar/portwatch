// Package trend tracks alert frequency over time and detects rising or falling
// patterns across a sliding window of intervals.
package trend

import (
	"fmt"
	"sync"
	"time"
)

// Direction indicates whether alert volume is rising, falling, or stable.
type Direction int

const (
	Stable  Direction = iota
	Rising            // count increased relative to previous interval
	Falling           // count decreased relative to previous interval
)

func (d Direction) String() string {
	switch d {
	case Rising:
		return "rising"
	case Falling:
		return "falling"
	default:
		return "stable"
	}
}

// Tracker records per-port alert counts bucketed by interval and reports
// the trend direction when queried.
type Tracker struct {
	mu       sync.Mutex
	interval time.Duration
	buckets  map[int][]int64 // port -> ordered bucket counts
	stamps   map[int][]time.Time
	maxBuckets int
	clock    func() time.Time
}

// New returns a Tracker that buckets counts into intervals of the given
// duration and retains up to maxBuckets historical buckets per port.
func New(interval time.Duration, maxBuckets int) *Tracker {
	if interval <= 0 {
		panic("trend: interval must be positive")
	}
	if maxBuckets < 2 {
		panic("trend: maxBuckets must be at least 2")
	}
	return &Tracker{
		interval:   interval,
		maxBuckets: maxBuckets,
		buckets:    make(map[int][]int64),
		stamps:     make(map[int][]time.Time),
		clock:      time.Now,
	}
}

func withClock(interval time.Duration, maxBuckets int, clk func() time.Time) *Tracker {
	t := New(interval, maxBuckets)
	t.clock = clk
	return t
}

// Record increments the count for port in the current interval bucket.
func (t *Tracker) Record(port int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := t.clock()
	bkts := t.buckets[port]
	sts := t.stamps[port]
	if len(bkts) == 0 || now.Sub(sts[len(sts)-1]) >= t.interval {
		bkts = append(bkts, 0)
		sts = append(sts, now)
		if len(bkts) > t.maxBuckets {
			bkts = bkts[len(bkts)-t.maxBuckets:]
			sts = sts[len(sts)-t.maxBuckets:]
		}
	}
	bkts[len(bkts)-1]++
	t.buckets[port] = bkts
	t.stamps[port] = sts
}

// Trend returns the Direction for port based on the last two completed buckets.
// If fewer than two buckets exist, Stable is returned.
func (t *Tracker) Trend(port int) Direction {
	t.mu.Lock()
	defer t.mu.Unlock()
	bkts := t.buckets[port]
	if len(bkts) < 2 {
		return Stable
	}
	prev := bkts[len(bkts)-2]
	curr := bkts[len(bkts)-1]
	switch {
	case curr > prev:
		return Rising
	case curr < prev:
		return Falling
	default:
		return Stable
	}
}

// Summary returns a human-readable summary for port.
func (t *Tracker) Summary(port int) string {
	d := t.Trend(port)
	t.mu.Lock()
	bkts := t.buckets[port]
	t.mu.Unlock()
	var last int64
	if len(bkts) > 0 {
		last = bkts[len(bkts)-1]
	}
	return fmt.Sprintf("port %d: %s (current bucket count: %d)", port, d, last)
}
