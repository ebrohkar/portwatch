// Package suppress provides a suppression list for alerts, allowing
// specific ports or port/event combinations to be silenced for a
// configurable duration.
package suppress

import (
	"sync"
	"time"
)

// Entry represents a single suppression rule.
type Entry struct {
	Port      int
	Event     string // empty string matches any event
	ExpiresAt time.Time
}

// List holds active suppressions and provides thread-safe access.
type List struct {
	mu      sync.Mutex
	entries []Entry
	now     func() time.Time
}

// New returns a new, empty suppression List.
func New() *List {
	return &List{now: time.Now}
}

// Add adds a suppression for the given port and event for the specified
// duration. An empty event string suppresses all events on that port.
func (l *List) Add(port int, event string, duration time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, Entry{
		Port:      port,
		Event:     event,
		ExpiresAt: l.now().Add(duration),
	})
}

// IsSuppressed reports whether the given port and event are currently
// suppressed. Expired entries are pruned on each call.
func (l *List) IsSuppressed(port int, event string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prune()
	for _, e := range l.entries {
		if e.Port != port {
			continue
		}
		if e.Event == "" || e.Event == event {
			return true
		}
	}
	return false
}

// Len returns the number of active (non-expired) suppressions.
func (l *List) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prune()
	return len(l.entries)
}

// prune removes expired entries. Caller must hold l.mu.
func (l *List) prune() {
	now := l.now()
	active := l.entries[:0]
	for _, e := range l.entries {
		if e.ExpiresAt.After(now) {
			active = append(active, e)
		}
	}
	l.entries = active
}
