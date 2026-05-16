// Package dedup provides alert deduplication based on a composite key
// of port, event type, and severity. Duplicate alerts within a configurable
// TTL window are silently dropped.
package dedup

import (
	"fmt"
	"sync"
	"time"

	"portwatch/internal/alert"
)

// clock allows tests to inject a fake time source.
type clock func() time.Time

// entry tracks when a key was last seen.
type entry struct {
	seenAt time.Time
}

// Deduplicator drops alerts whose composite key was already seen within TTL.
type Deduplicator struct {
	mu      sync.Mutex
	ttl     time.Duration
	clock   clock
	entries map[string]entry
}

// New returns a Deduplicator with the given TTL.
// Panics if ttl is zero or negative.
func New(ttl time.Duration) *Deduplicator {
	return newWithClock(ttl, time.Now)
}

// WithClock returns a Deduplicator using a custom clock (for testing).
func WithClock(ttl time.Duration, c clock) *Deduplicator {
	return newWithClock(ttl, c)
}

func newWithClock(ttl time.Duration, c clock) *Deduplicator {
	if ttl <= 0 {
		panic("dedup: ttl must be positive")
	}
	if c == nil {
		panic("dedup: clock must not be nil")
	}
	return &Deduplicator{
		ttl:     ttl,
		clock:   c,
		entries: make(map[string]entry),
	}
}

// Allow returns true if the alert is new (not a duplicate within TTL).
// It records the alert so subsequent identical alerts are suppressed.
func (d *Deduplicator) Allow(a alert.Alert) bool {
	key := compositeKey(a)
	now := d.clock()

	d.mu.Lock()
	defer d.mu.Unlock()

	if e, ok := d.entries[key]; ok {
		if now.Sub(e.seenAt) < d.ttl {
			return false
		}
	}
	d.entries[key] = entry{seenAt: now}
	return true
}

// Flush removes all expired entries, freeing memory.
func (d *Deduplicator) Flush() {
	now := d.clock()
	d.mu.Lock()
	defer d.mu.Unlock()
	for k, e := range d.entries {
		if now.Sub(e.seenAt) >= d.ttl {
			delete(d.entries, k)
		}
	}
}

func compositeKey(a alert.Alert) string {
	return fmt.Sprintf("%d|%s|%s", a.Port, a.Event, a.Severity)
}
