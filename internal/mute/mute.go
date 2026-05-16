// Package mute provides a time-bounded muting mechanism that suppresses
// alerts for specific ports during a defined window (e.g. maintenance).
package mute

import (
	"sync"
	"time"
)

// Clock is a function that returns the current time.
type Clock func() time.Time

// Entry represents a single mute window for a port.
type Entry struct {
	Port    int
	Until   time.Time
	Reason  string
}

// Store holds active mute windows.
type Store struct {
	mu      sync.RWMutex
	entries []Entry
	clock   Clock
}

// New returns a Store using the real wall clock.
func New() *Store {
	return withClock(time.Now)
}

func withClock(c Clock) *Store {
	return &Store{clock: c}
}

// Add registers a mute window for port lasting for the given duration.
// An empty reason is accepted.
func (s *Store) Add(port int, duration time.Duration, reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = append(s.entries, Entry{
		Port:   port,
		Until:  s.clock().Add(duration),
		Reason: reason,
	})
}

// IsMuted reports whether port is currently muted.
// Expired entries are lazily pruned on each call.
func (s *Store) IsMuted(port int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.clock()
	active := s.entries[:0]
	muted := false
	for _, e := range s.entries {
		if e.Until.After(now) {
			active = append(active, e)
			if e.Port == port {
				muted = true
			}
		}
	}
	s.entries = active
	return muted
}

// Active returns a snapshot of all currently active mute entries.
func (s *Store) Active() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := s.clock()
	out := make([]Entry, 0, len(s.entries))
	for _, e := range s.entries {
		if e.Until.After(now) {
			out = append(out, e)
		}
	}
	return out
}

// Clear removes all mute entries for port.
func (s *Store) Clear(port int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	filtered := s.entries[:0]
	for _, e := range s.entries {
		if e.Port != port {
			filtered = append(filtered, e)
		}
	}
	s.entries = filtered
}
