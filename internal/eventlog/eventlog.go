// Package eventlog provides a structured, append-only log of port-change
// events suitable for audit trails and post-incident review.
package eventlog

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Entry represents a single recorded port event.
type Entry struct {
	Timestamp time.Time `json:"timestamp"`
	Port      int       `json:"port"`
	Event     string    `json:"event"`
	Severity  string    `json:"severity"`
	Message   string    `json:"message"`
}

// EventLog writes structured JSON event entries to an io.Writer.
type EventLog struct {
	mu  sync.Mutex
	out io.Writer
	now func() time.Time
}

// New returns an EventLog that writes to w.
// If w is nil, os.Stdout is used.
func New(w io.Writer) *EventLog {
	if w == nil {
		w = os.Stdout
	}
	return &EventLog{
		out: w,
		now: time.Now,
	}
}

// Record writes a single Entry as a newline-delimited JSON object.
// The Timestamp field is set to the current time if zero.
func (el *EventLog) Record(e Entry) error {
	if e.Timestamp.IsZero() {
		e.Timestamp = el.now().UTC()
	}
	if e.Port < 1 || e.Port > 65535 {
		return fmt.Errorf("eventlog: invalid port %d", e.Port)
	}
	if e.Event == "" {
		return fmt.Errorf("eventlog: event must not be empty")
	}

	data, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("eventlog: marshal: %w", err)
	}

	el.mu.Lock()
	defer el.mu.Unlock()

	_, err = fmt.Fprintf(el.out, "%s\n", data)
	return err
}
