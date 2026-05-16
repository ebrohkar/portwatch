// Package quota enforces per-port alert quotas over a rolling time window.
// Once a port exhausts its quota, further alerts are dropped until the
// window resets. This prevents alert storms from a single noisy port.
package quota

import (
	"fmt"
	"sync"
	"time"
)

// Clock allows injecting a fake time source in tests.
type Clock func() time.Time

// entry tracks the hit count and window start for a single port.
type entry struct {
	count     int
	windowEnd time.Time
}

// Quota enforces a maximum number of alerts per port within a window.
type Quota struct {
	mu     sync.Mutex
	max    int
	window time.Duration
	clock  Clock
	ports  map[int]*entry
}

// New creates a Quota with the given maximum hits and rolling window.
// Panics if max < 1 or window is zero.
func New(max int, window time.Duration) *Quota {
	return newWithClock(max, window, time.Now)
}

func newWithClock(max int, window time.Duration, clock Clock) *Quota {
	if max < 1 {
		panic("quota: max must be >= 1")
	}
	if window <= 0 {
		panic("quota: window must be > 0")
	}
	return &Quota{
		max:    max,
		window: window,
		clock:  clock,
		ports:  make(map[int]*entry),
	}
}

// Allow returns true and increments the counter if the port is within quota.
// Returns false if the quota has been exhausted for the current window.
func (q *Quota) Allow(port int) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := q.clock()
	e, ok := q.ports[port]
	if !ok || now.After(e.windowEnd) {
		q.ports[port] = &entry{count: 1, windowEnd: now.Add(q.window)}
		return true
	}
	if e.count >= q.max {
		return false
	}
	e.count++
	return true
}

// Remaining returns how many more alerts the port may emit in the current window.
func (q *Quota) Remaining(port int) int {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := q.clock()
	e, ok := q.ports[port]
	if !ok || now.After(e.windowEnd) {
		return q.max
	}
	rem := q.max - e.count
	if rem < 0 {
		return 0
	}
	return rem
}

// Reset clears the quota state for the given port.
func (q *Quota) Reset(port int) {
	q.mu.Lock()
	defer q.mu.Unlock()
	delete(q.ports, port)
}

// String returns a human-readable description of the quota configuration.
func (q *Quota) String() string {
	return fmt.Sprintf("Quota(max=%d, window=%s)", q.max, q.window)
}
