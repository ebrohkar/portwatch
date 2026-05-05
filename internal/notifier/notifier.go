package notifier

import (
	"fmt"
	"io"
	"os"
	"time"
)

// Level represents the severity of a notification.
type Level string

const (
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelAlert Level = "ALERT"
)

// Event describes a port state change that triggered a rule.
type Event struct {
	Timestamp time.Time
	Level     Level
	Port      int
	Action    string
	Message   string
}

// Notifier sends events to one or more outputs.
type Notifier struct {
	writers []io.Writer
}

// New creates a Notifier that writes to the provided writers.
// If no writers are supplied it defaults to os.Stdout.
func New(writers ...io.Writer) *Notifier {
	if len(writers) == 0 {
		writers = []io.Writer{os.Stdout}
	}
	return &Notifier{writers: writers}
}

// Send formats and dispatches an Event to all configured writers.
func (n *Notifier) Send(e Event) error {
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now()
	}
	line := fmt.Sprintf("%s [%s] port=%d action=%s msg=%q\n",
		e.Timestamp.UTC().Format(time.RFC3339),
		e.Level,
		e.Port,
		e.Action,
		e.Message,
	)
	var lastErr error
	for _, w := range n.writers {
		if _, err := fmt.Fprint(w, line); err != nil {
			lastErr = err
		}
	}
	return lastErr
}
