// Package escalation provides severity escalation logic for repeated alerts.
// When the same port+event combination triggers more than a configured threshold
// of alerts within a sliding window, the severity is promoted to a higher level.
package escalation

import (
	"fmt"
	"sync"
	"time"
)

// Severity levels in ascending order.
const (
	SeverityInfo     = "info"
	SeverityWarning  = "warning"
	SeverityCritical = "critical"
)

// clock abstracts time for testing.
type clock func() time.Time

// entry tracks hit count and window start for a single key.
type entry struct {
	count     int
	windowEnd time.Time
}

// Escalator promotes alert severity when a port+event pair exceeds a threshold
// within a rolling time window.
type Escalator struct {
	mu        sync.Mutex
	threshold int
	window    time.Duration
	promote   map[string]string
	state     map[string]*entry
	now       clock
}

// New returns an Escalator. threshold is the minimum hit count that triggers
// promotion; window is the rolling duration over which hits are counted.
// Panics if threshold < 1 or window <= 0.
func New(threshold int, window time.Duration) *Escalator {
	if threshold < 1 {
		panic("escalation: threshold must be >= 1")
	}
	if window <= 0 {
		panic("escalation: window must be > 0")
	}
	return &Escalator{
		threshold: threshold,
		window:    window,
		promote: map[string]string{
			SeverityInfo:    SeverityWarning,
			SeverityWarning: SeverityCritical,
		},
		state: make(map[string]*entry),
		now:   time.Now,
	}
}

// Evaluate records a hit for the given port+event pair and returns the
// (possibly escalated) severity. If the hit count within the current window
// meets or exceeds the threshold the severity is promoted one level.
func (e *Escalator) Evaluate(port int, event, severity string) string {
	e.mu.Lock()
	defer e.mu.Unlock()

	k := fmt.Sprintf("%d:%s", port, event)
	now := e.now()

	en, ok := e.state[k]
	if !ok || now.After(en.windowEnd) {
		en = &entry{windowEnd: now.Add(e.window)}
		e.state[k] = en
	}
	en.count++

	if en.count >= e.threshold {
		if promoted, ok := e.promote[severity]; ok {
			return promoted
		}
	}
	return severity
}

// Reset clears all tracked state for the given port+event pair.
func (e *Escalator) Reset(port int, event string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.state, fmt.Sprintf("%d:%s", port, event))
}
