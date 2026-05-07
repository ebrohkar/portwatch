// Package metrics tracks runtime counters for portwatch: scans performed,
// alerts emitted, and suppressed events. Values are safe for concurrent use.
package metrics

import (
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"text/tabwriter"
	"time"
)

// Counters holds atomic runtime metrics for a single daemon lifecycle.
type Counters struct {
	ScansTotal    atomic.Int64
	AlertsTotal   atomic.Int64
	SuppressTotal atomic.Int64
	ErrorsTotal   atomic.Int64
	startedAt     time.Time
}

// New returns a Counters instance with the start time set to now.
func New() *Counters {
	return &Counters{startedAt: time.Now()}
}

// IncScans increments the scan counter by 1.
func (c *Counters) IncScans() { c.ScansTotal.Add(1) }

// IncAlerts increments the alert counter by 1.
func (c *Counters) IncAlerts() { c.AlertsTotal.Add(1) }

// IncSuppressed increments the suppressed-event counter by 1.
func (c *Counters) IncSuppressed() { c.SuppressTotal.Add(1) }

// IncErrors increments the error counter by 1.
func (c *Counters) IncErrors() { c.ErrorsTotal.Add(1) }

// Uptime returns the duration since the Counters were created.
func (c *Counters) Uptime() time.Duration {
	return time.Since(c.startedAt)
}

// WriteTo writes a human-readable summary of all counters to w.
// It returns the number of bytes written and any error.
func (c *Counters) WriteTo(w io.Writer) (int64, error) {
	if w == nil {
		w = os.Stdout
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	lines := []string{
		fmt.Sprintf("uptime\t%s", c.Uptime().Round(time.Second)),
		fmt.Sprintf("scans_total\t%d", c.ScansTotal.Load()),
		fmt.Sprintf("alerts_total\t%d", c.AlertsTotal.Load()),
		fmt.Sprintf("suppressed_total\t%d", c.SuppressTotal.Load()),
		fmt.Sprintf("errors_total\t%d", c.ErrorsTotal.Load()),
	}
	var total int64
	for _, l := range lines {
		n, err := fmt.Fprintln(tw, l)
		total += int64(n)
		if err != nil {
			return total, err
		}
	}
	return total, tw.Flush()
}
