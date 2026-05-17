// Package labeler attaches human-readable labels to alerts based on port-to-service mappings.
package labeler

import (
	"fmt"

	"github.com/example/portwatch/internal/alert"
)

// Labeler maps port numbers to service labels and annotates alerts.
type Labeler struct {
	mappings map[int]string
}

// New creates a Labeler from the provided port-to-label map.
// Returns an error if any port is out of range or any label is blank.
func New(mappings map[int]string) (*Labeler, error) {
	if len(mappings) == 0 {
		return &Labeler{mappings: map[int]string{}}, nil
	}
	copy := make(map[int]string, len(mappings))
	for port, label := range mappings {
		if port < 1 || port > 65535 {
			return nil, fmt.Errorf("labeler: port %d out of range (1-65535)", port)
		}
		if label == "" {
			return nil, fmt.Errorf("labeler: empty label for port %d", port)
		}
		copy[port] = label
	}
	return &Labeler{mappings: copy}, nil
}

// Label returns the service label for the given port, and whether one was found.
func (l *Labeler) Label(port int) (string, bool) {
	label, ok := l.mappings[port]
	return label, ok
}

// Annotate returns a copy of the alert with the Service field set to the
// mapped label when one exists. If no mapping is found the alert is returned
// unchanged.
func (l *Labeler) Annotate(a alert.Alert) alert.Alert {
	if label, ok := l.mappings[a.Port]; ok {
		a.Service = label
	}
	return a
}

// Len returns the number of port-to-label mappings.
func (l *Labeler) Len() int {
	return len(l.mappings)
}
