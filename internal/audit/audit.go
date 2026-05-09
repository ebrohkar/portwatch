// Package audit provides a structured audit log for portwatch events,
// recording scan results and alert decisions with timestamps.
package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Entry represents a single audit log record.
type Entry struct {
	Timestamp time.Time `json:"timestamp"`
	Event     string    `json:"event"`
	Port      int       `json:"port,omitempty"`
	Severity  string    `json:"severity,omitempty"`
	Message   string    `json:"message"`
}

// Logger writes audit entries as newline-delimited JSON.
type Logger struct {
	mu  sync.Mutex
	out io.Writer
	now func() time.Time
}

// New returns a Logger writing to w. If w is nil, os.Stdout is used.
func New(w io.Writer) *Logger {
	if w == nil {
		w = os.Stdout
	}
	return &Logger{
		out: w,
		now: time.Now,
	}
}

// Log writes a single audit entry to the underlying writer.
// It is safe for concurrent use.
func (l *Logger) Log(event, message string, port int, severity string) error {
	e := Entry{
		Timestamp: l.now().UTC(),
		Event:     event,
		Port:      port,
		Severity:  severity,
		Message:   message,
	}
	b, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("audit: marshal entry: %w", err)
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	_, err = fmt.Fprintf(l.out, "%s\n", b)
	if err != nil {
		return fmt.Errorf("audit: write entry: %w", err)
	}
	return nil
}
