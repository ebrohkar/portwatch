// Package rotation provides log rotation for audit and notifier outputs,
// capping file size and keeping a configurable number of backups.
package rotation

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const (
	DefaultMaxBytes   = 10 * 1024 * 1024 // 10 MiB
	DefaultMaxBackups = 3
)

// Writer is an io.Writer that rotates the underlying file when it exceeds
// MaxBytes. Old files are renamed with a numeric suffix (.1, .2, …).
type Writer struct {
	mu         sync.Mutex
	path       string
	maxBytes   int64
	maxBackups int
	file       *os.File
	size       int64
}

// New creates a new rotation Writer. The file at path is opened (or created)
// in append mode. MaxBytes and MaxBackups default to package-level constants
// when zero.
func New(path string, maxBytes int64, maxBackups int) (*Writer, error) {
	if maxBytes <= 0 {
		maxBytes = DefaultMaxBytes
	}
	if maxBackups <= 0 {
		maxBackups = DefaultMaxBackups
	}
	w := &Writer{path: path, maxBytes: maxBytes, maxBackups: maxBackups}
	if err := w.openOrCreate(); err != nil {
		return nil, err
	}
	return w, nil
}

// Write implements io.Writer. It rotates the file before writing when the
// current size would exceed MaxBytes.
func (w *Writer) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.size+int64(len(p)) > w.maxBytes {
		if err := w.rotate(); err != nil {
			return 0, fmt.Errorf("rotation: rotate: %w", err)
		}
	}
	n, err := w.file.Write(p)
	w.size += int64(n)
	return n, err
}

// Close closes the underlying file.
func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

func (w *Writer) openOrCreate() error {
	f, err := os.OpenFile(w.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("rotation: open %s: %w", w.path, err)
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return fmt.Errorf("rotation: stat %s: %w", w.path, err)
	}
	w.file = f
	w.size = info.Size()
	return nil
}

func (w *Writer) rotate() error {
	if err := w.file.Close(); err != nil {
		return err
	}
	// Shift existing backups: .3 removed, .2->.3, .1->.2, base->.1
	for i := w.maxBackups - 1; i >= 1; i-- {
		old := fmt.Sprintf("%s.%d", w.path, i)
		new := fmt.Sprintf("%s.%d", w.path, i+1)
		_ = os.Rename(old, new)
	}
	if err := os.Rename(w.path, w.path+".1"); err != nil && !os.IsNotExist(err) {
		return err
	}
	return w.openOrCreate()
}

// BackupPaths returns the paths of existing backup files in rotation order.
func (w *Writer) BackupPaths() []string {
	w.mu.Lock()
	defer w.mu.Unlock()
	var paths []string
	for i := 1; i <= w.maxBackups; i++ {
		p := fmt.Sprintf("%s.%d", w.path, i)
		if _, err := os.Stat(p); err == nil {
			paths = append(paths, filepath.Clean(p))
		}
	}
	return paths
}
