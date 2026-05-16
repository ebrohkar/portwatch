package eventlog

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// FileWriter wraps an *os.File and implements io.Writer with safe concurrent
// access. It is intended to be passed to New() when persisting events to disk.
type FileWriter struct {
	mu   sync.Mutex
	file *os.File
}

// OpenFile opens or creates the file at path for appending.
func OpenFile(path string) (*FileWriter, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("eventlog: mkdir: %w", err)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("eventlog: open: %w", err)
	}
	return &FileWriter{file: f}, nil
}

// Write implements io.Writer.
func (fw *FileWriter) Write(p []byte) (int, error) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	return fw.file.Write(p)
}

// Close closes the underlying file.
func (fw *FileWriter) Close() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	return fw.file.Close()
}

// Name returns the path of the underlying file.
func (fw *FileWriter) Name() string {
	return fw.file.Name()
}

// Discard is an io.Writer that silently drops all data. Useful in tests.
var Discard io.Writer = io.Discard
