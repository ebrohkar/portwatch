// Package rollup groups repeated alerts into a single summary alert
// after a configurable count threshold is reached within a time window.
package rollup

import (
	"fmt"
	"sync"
	"time"

	"portwatch/internal/alert"
)

// clock allows injecting a fake time source in tests.
type clock func() time.Time

// bucket tracks repeated occurrences of the same (port, event) pair.
type bucket struct {
	count     int
	firstSeen time.Time
	lastAlert alert.Alert
}

// Rollup collapses repeated alerts into a summary once a threshold is met.
type Rollup struct {
	mu        sync.Mutex
	threshold int
	window    time.Duration
	now       clock
	buckets   map[string]*bucket
}

// New returns a Rollup that fires a summary alert after threshold occurrences
// of the same (port, event) pair within window. Panics if threshold < 2 or
// window is zero.
func New(threshold int, window time.Duration) *Rollup {
	if threshold < 2 {
		panic("rollup: threshold must be >= 2")
	}
	if window <= 0 {
		panic("rollup: window must be > 0")
	}
	return &Rollup{
		threshold: threshold,
		window:    window,
		now:       time.Now,
		buckets:   make(map[string]*bucket),
	}
}

func withClock(threshold int, window time.Duration, c clock) *Rollup {
	r := New(threshold, window)
	r.now = c
	return r
}

// Add records the alert and returns (summary, true) when the threshold is
// reached within the window; otherwise it returns (zero, false).
func (r *Rollup) Add(a alert.Alert) (alert.Alert, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := r.now()
	k := key(a)

	b, ok := r.buckets[k]
	if !ok || now.Sub(b.firstSeen) > r.window {
		r.buckets[k] = &bucket{count: 1, firstSeen: now, lastAlert: a}
		return alert.Alert{}, false
	}

	b.count++
	b.lastAlert = a

	if b.count >= r.threshold {
		summary := b.lastAlert
		summary.Message = fmt.Sprintf(
			"[rollup] port %d %s triggered %d times in %s",
			a.Port, a.Event, b.count, r.window,
		)
		delete(r.buckets, k)
		return summary, true
	}

	return alert.Alert{}, false
}

// Reset clears the state for a specific (port, event) pair.
func (r *Rollup) Reset(port int, event string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.buckets, fmt.Sprintf("%d:%s", port, event))
}

func key(a alert.Alert) string {
	return fmt.Sprintf("%d:%s", a.Port, a.Event)
}
