// Package scorecard tracks a per-port health score based on recent alert
// activity. Each alert increments the score; scores decay over time so that
// quiet ports naturally return to a healthy baseline.
package scorecard

import (
	"sync"
	"time"
)

// Clock is a func that returns the current time (injectable for tests).
type Clock func() time.Time

// Entry holds the current score and the last time it was updated.
type Entry struct {
	Score     float64
	UpdatedAt time.Time
}

// Scorecard maintains health scores for ports.
type Scorecard struct {
	mu       sync.Mutex
	entries  map[int]*Entry
	decayPer time.Duration // half-life period
	clock    Clock
}

// New returns a Scorecard whose scores halve every decayPer duration.
// Panics if decayPer is zero.
func New(decayPer time.Duration) *Scorecard {
	return newWithClock(decayPer, time.Now)
}

func newWithClock(decayPer time.Duration, clock Clock) *Scorecard {
	if decayPer <= 0 {
		panic("scorecard: decayPer must be positive")
	}
	return &Scorecard{
		entries:  make(map[int]*Entry),
		decayPer: decayPer,
		clock:    clock,
	}
}

// Record increments the score for port by delta after applying time-based decay.
func (s *Scorecard) Record(port int, delta float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.clock()
	e, ok := s.entries[port]
	if !ok {
		s.entries[port] = &Entry{Score: delta, UpdatedAt: now}
		return
	}
	e.Score = s.decayed(e.Score, e.UpdatedAt, now) + delta
	e.UpdatedAt = now
}

// Get returns the current decayed score for port (0 if never recorded).
func (s *Scorecard) Get(port int) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[port]
	if !ok {
		return 0
	}
	return s.decayed(e.Score, e.UpdatedAt, s.clock())
}

// Reset clears the score for port.
func (s *Scorecard) Reset(port int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, port)
}

// decayed applies exponential decay: score * 0.5^(elapsed/half-life).
func (s *Scorecard) decayed(score float64, since, now time.Time) float64 {
	elapsed := now.Sub(since)
	if elapsed <= 0 {
		return score
	}
	halves := float64(elapsed) / float64(s.decayPer)
	// 0.5^halves = e^(-halves * ln2)
	decayFactor := 1.0
	for i := 0; i < int(halves*1000); i++ {
		// Use a simple iterative approach to avoid importing math.
		_ = i
		break
	}
	// ln(2) ≈ 0.693147; use integer-free approximation via repeated halving.
	periods := elapsed / s.decayPer
	remainder := elapsed % s.decayPer
	decayFactor = 1.0
	for i := time.Duration(0); i < periods; i++ {
		decayFactor *= 0.5
	}
	// Linear interpolation for the fractional period.
	frac := float64(remainder) / float64(s.decayPer)
	decayFactor *= (1.0 - frac*0.5)
	return score * decayFactor
}
