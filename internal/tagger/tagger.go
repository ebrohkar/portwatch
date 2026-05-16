// Package tagger assigns severity and category tags to alerts based on
// configurable port-to-tag mappings. Tags enrich alert metadata for
// downstream consumers such as the notifier and audit log.
package tagger

import (
	"fmt"
	"sync"

	"github.com/example/portwatch/internal/alert"
)

// Tag holds the label and severity override for a port mapping.
type Tag struct {
	Label    string
	Severity string
}

// Tagger maps port numbers to descriptive tags.
type Tagger struct {
	mu      sync.RWMutex
	mapping map[int]Tag
}

// New returns a Tagger initialised with the provided port-to-tag mapping.
// An error is returned if any port number is out of the valid 1–65535 range
// or if a tag label is empty.
func New(mapping map[int]Tag) (*Tagger, error) {
	for port, tag := range mapping {
		if port < 1 || port > 65535 {
			return nil, fmt.Errorf("tagger: invalid port %d", port)
		}
		if tag.Label == "" {
			return nil, fmt.Errorf("tagger: empty label for port %d", port)
		}
	}
	cp := make(map[int]Tag, len(mapping))
	for k, v := range mapping {
		cp[k] = v
	}
	return &Tagger{mapping: cp}, nil
}

// Lookup returns the Tag registered for port, and a boolean indicating
// whether a mapping exists.
func (t *Tagger) Lookup(port int) (Tag, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	tag, ok := t.mapping[port]
	return tag, ok
}

// Annotate enriches a copy of the provided alert with the label stored for
// its port. If no mapping exists the alert is returned unchanged. If the tag
// carries a non-empty Severity it overrides the alert's severity field.
func (t *Tagger) Annotate(a alert.Alert) alert.Alert {
	tag, ok := t.Lookup(a.Port)
	if !ok {
		return a
	}
	a.Message = fmt.Sprintf("[%s] %s", tag.Label, a.Message)
	if tag.Severity != "" {
		a.Severity = tag.Severity
	}
	return a
}

// Set adds or replaces the Tag for the given port. It returns an error for
// invalid port numbers or an empty label.
func (t *Tagger) Set(port int, tag Tag) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("tagger: invalid port %d", port)
	}
	if tag.Label == "" {
		return fmt.Errorf("tagger: empty label for port %d", port)
	}
	t.mu.Lock()
	t.mapping[port] = tag
	t.mu.Unlock()
	return nil
}
