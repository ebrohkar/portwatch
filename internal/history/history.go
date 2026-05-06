// Package history maintains a rolling log of alert events for reporting
// and audit purposes.
package history

import (
	"sync"
	"time"

	"github.com/user/portwatch/internal/alert"
)

// Entry wraps an alert with the time it was recorded.
type Entry struct {
	RecordedAt time.Time
	Alert      alert.Alert
}

// Log is a thread-safe, bounded ring buffer of alert entries.
type Log struct {
	mu      sync.RWMutex
	entries []Entry
	maxSize int
}

// New creates a Log that retains at most maxSize entries.
// If maxSize is <= 0 it defaults to 100.
func New(maxSize int) *Log {
	if maxSize <= 0 {
		maxSize = 100
	}
	return &Log{
		entries: make([]Entry, 0, maxSize),
		maxSize: maxSize,
	}
}

// Add appends an alert to the log. When the log is full the oldest entry
// is evicted to make room.
func (l *Log) Add(a alert.Alert) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := Entry{RecordedAt: time.Now(), Alert: a}
	if len(l.entries) >= l.maxSize {
		// evict oldest
		l.entries = append(l.entries[1:], entry)
	} else {
		l.entries = append(l.entries, entry)
	}
}

// All returns a snapshot of all entries in chronological order.
func (l *Log) All() []Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	snap := make([]Entry, len(l.entries))
	copy(snap, l.entries)
	return snap
}

// Since returns entries recorded at or after t.
func (l *Log) Since(t time.Time) []Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var result []Entry
	for _, e := range l.entries {
		if !e.RecordedAt.Before(t) {
			result = append(result, e)
		}
	}
	return result
}

// Len returns the current number of entries in the log.
func (l *Log) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.entries)
}
