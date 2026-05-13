// Package baseline captures and compares a known-good set of open ports
// so that portwatch can distinguish expected from unexpected changes.
package baseline

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Baseline holds a snapshot of ports that are considered "normal".
type Baseline struct {
	mu        sync.RWMutex
	ports     map[int]struct{}
	CapturedAt time.Time `json:"captured_at"`
	Ports      []int     `json:"ports"`
}

// New returns an empty Baseline.
func New() *Baseline {
	return &Baseline{
		ports: make(map[int]struct{}),
	}
}

// Set replaces the current baseline with the given port list.
func (b *Baseline) Set(ports []int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.ports = make(map[int]struct{}, len(ports))
	for _, p := range ports {
		b.ports[p] = struct{}{}
	}
	b.CapturedAt = time.Now().UTC()
	b.Ports = append([]int(nil), ports...)
}

// Contains reports whether port is part of the baseline.
func (b *Baseline) Contains(port int) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	_, ok := b.ports[port]
	return ok
}

// Len returns the number of ports in the baseline.
func (b *Baseline) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.ports)
}

// SaveToFile persists the baseline to a JSON file.
func (b *Baseline) SaveToFile(path string) error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return fmt.Errorf("baseline: marshal: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("baseline: write %s: %w", path, err)
	}
	return nil
}

// LoadFromFile reads a previously saved baseline from disk.
func LoadFromFile(path string) (*Baseline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("baseline: read %s: %w", path, err)
	}
	var b Baseline
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, fmt.Errorf("baseline: unmarshal: %w", err)
	}
	b.ports = make(map[int]struct{}, len(b.Ports))
	for _, p := range b.Ports {
		b.ports[p] = struct{}{}
	}
	return &b, nil
}
