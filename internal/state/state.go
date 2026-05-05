package state

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// Snapshot holds the recorded open ports at a point in time.
type Snapshot struct {
	Timestamp time.Time `json:"timestamp"`
	OpenPorts []int     `json:"open_ports"`
}

// Store manages persistence of port scan snapshots.
type Store struct {
	mu       sync.RWMutex
	filePath string
	current  *Snapshot
}

// NewStore creates a Store backed by the given file path.
func NewStore(filePath string) *Store {
	return &Store{filePath: filePath}
}

// Load reads the last snapshot from disk. Returns nil if no snapshot exists yet.
func (s *Store) Load() (*Snapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, err
	}
	s.current = &snap
	return &snap, nil
}

// Save persists a new snapshot to disk and updates the in-memory current.
func (s *Store) Save(ports []int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	snap := &Snapshot{
		Timestamp: time.Now().UTC(),
		OpenPorts: ports,
	}

	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(s.filePath, data, 0o644); err != nil {
		return err
	}
	s.current = snap
	return nil
}

// Current returns the in-memory snapshot without reading from disk.
func (s *Store) Current() *Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.current
}

// Diff computes ports that appeared or disappeared between two snapshots.
func Diff(prev, next []int) (appeared, disappeared []int) {
	prevSet := toSet(prev)
	nextSet := toSet(next)

	for p := range nextSet {
		if !prevSet[p] {
			appeared = append(appeared, p)
		}
	}
	for p := range prevSet {
		if !nextSet[p] {
			disappeared = append(disappeared, p)
		}
	}
	return
}

func toSet(ports []int) map[int]bool {
	s := make(map[int]bool, len(ports))
	for _, p := range ports {
		s[p] = true
	}
	return s
}
