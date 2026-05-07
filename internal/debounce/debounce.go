// Package debounce provides a mechanism to suppress repeated alerts
// for the same port event within a configurable cooldown window.
package debounce

import (
	"sync"
	"time"
)

// key uniquely identifies a port+event pair.
type key struct {
	port  int
	event string
}

// entry holds the last time an alert was emitted for a given key.
type entry struct {
	lastSeen time.Time
}

// Debouncer tracks recent alerts and decides whether a new alert
// for the same port+event should be forwarded or dropped.
type Debouncer struct {
	mu       sync.Mutex
	cooldown time.Duration
	entries  map[key]entry
	now      func() time.Time
}

// DefaultCooldown is used when no cooldown is specified.
const DefaultCooldown = 30 * time.Second

// New creates a Debouncer with the given cooldown duration.
// If cooldown is zero, DefaultCooldown is used.
func New(cooldown time.Duration) *Debouncer {
	if cooldown <= 0 {
		cooldown = DefaultCooldown
	}
	return &Debouncer{
		cooldown: cooldown,
		entries:  make(map[key]entry),
		now:      time.Now,
	}
}

// Allow returns true if the alert for the given port and event should
// be forwarded. It returns false if an alert was already emitted within
// the cooldown window.
func (d *Debouncer) Allow(port int, event string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	k := key{port: port, event: event}
	now := d.now()

	if e, ok := d.entries[k]; ok {
		if now.Sub(e.lastSeen) < d.cooldown {
			return false
		}
	}

	d.entries[k] = entry{lastSeen: now}
	return true
}

// Reset clears the debounce state for a specific port and event.
func (d *Debouncer) Reset(port int, event string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.entries, key{port: port, event: event})
}

// Purge removes all entries whose cooldown window has expired.
func (d *Debouncer) Purge() {
	d.mu.Lock()
	defer d.mu.Unlock()
	now := d.now()
	for k, e := range d.entries {
		if now.Sub(e.lastSeen) >= d.cooldown {
			delete(d.entries, k)
		}
	}
}
