// Package metrics tracks runtime counters for the portwatch daemon.
package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

// Snapshot is a point-in-time copy of all counters.
type Snapshot struct {
	Uptime     time.Duration
	Scans      int64
	Alerts     int64
	Suppressed int64
	Errors     int64
}

// Metrics holds atomic counters and a start timestamp.
type Metrics struct {
	startTime  time.Time
	scans      atomic.Int64
	alerts     atomic.Int64
	suppressed atomic.Int64
	errors     atomic.Int64
	mu         sync.Mutex // guards startTime resets in tests
}

// New creates a Metrics instance with the current time as start.
func New() *Metrics {
	return &Metrics{startTime: time.Now()}
}

// IncScans increments the scan counter by one.
func (m *Metrics) IncScans() { m.scans.Add(1) }

// IncAlerts increments the alert counter by one.
func (m *Metrics) IncAlerts() { m.alerts.Add(1) }

// IncSuppressed increments the suppressed-alert counter by one.
func (m *Metrics) IncSuppressed() { m.suppressed.Add(1) }

// IncErrors increments the error counter by one.
func (m *Metrics) IncErrors() { m.errors.Add(1) }

// Snapshot returns a consistent point-in-time view of all counters.
func (m *Metrics) Snapshot() Snapshot {
	m.mu.Lock()
	start := m.startTime
	m.mu.Unlock()
	return Snapshot{
		Uptime:     time.Since(start),
		Scans:      m.scans.Load(),
		Alerts:     m.alerts.Load(),
		Suppressed: m.suppressed.Load(),
		Errors:     m.errors.Load(),
	}
}

// Reset zeroes all counters and resets the start time. Intended for tests.
func (m *Metrics) Reset() {
	m.scans.Store(0)
	m.alerts.Store(0)
	m.suppressed.Store(0)
	m.errors.Store(0)
	m.mu.Lock()
	m.startTime = time.Now()
	m.mu.Unlock()
}
