// Package summary provides periodic scan summary reporting for portwatch.
package summary

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/history"
)

// Reporter writes periodic summaries of alert activity to a writer.
type Reporter struct {
	history *history.History
	out     io.Writer
	interval time.Duration
}

// New creates a new summary Reporter. If out is nil, os.Stdout is used.
func New(h *history.History, out io.Writer, interval time.Duration) *Reporter {
	if out == nil {
		out = os.Stdout
	}
	if interval <= 0 {
		interval = 5 * time.Minute
	}
	return &Reporter{
		history:  h,
		out:      out,
		interval: interval,
	}
}

// Interval returns the configured summary interval.
func (r *Reporter) Interval() time.Duration {
	return r.interval
}

// Write formats and writes a summary of all alerts in history to the writer.
func (r *Reporter) Write() {
	alerts := r.history.All()
	now := time.Now().UTC().Format(time.RFC3339)

	fmt.Fprintf(r.out, "--- portwatch summary [%s] ---\n", now)
	fmt.Fprintf(r.out, "Total alerts: %d\n", len(alerts))

	if len(alerts) == 0 {
		fmt.Fprintln(r.out, "No alerts recorded.")
		return
	}

	counts := countBySeverity(alerts)
	for _, sev := range []string{"critical", "warning", "info"} {
		if n, ok := counts[sev]; ok {
			fmt.Fprintf(r.out, "  %-10s %d\n", strings.Title(sev)+":", n)
		}
	}

	fmt.Fprintln(r.out, "Recent alerts:")
	start := len(alerts) - 5
	if start < 0 {
		start = 0
	}
	for _, a := range alerts[start:] {
		fmt.Fprintf(r.out, "  [%s] port=%d event=%s severity=%s\n",
			a.Timestamp.UTC().Format(time.RFC3339), a.Port, a.Event, a.Severity)
	}
	fmt.Fprintln(r.out, "---")
}

func countBySeverity(alerts []alert.Alert) map[string]int {
	m := make(map[string]int)
	for _, a := range alerts {
		m[a.Severity]++
	}
	return m
}
