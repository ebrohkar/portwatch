package rotation_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/example/portwatch/internal/rotation"
)

func TestNew_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")

	w, err := rotation.New(path, 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer w.Close()

	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected file to exist: %v", err)
	}
}

func TestNew_InvalidPath(t *testing.T) {
	_, err := rotation.New("/no/such/dir/test.log", 0, 0)
	if err == nil {
		t.Fatal("expected error for invalid path")
	}
}

func TestWrite_AppendsData(t *testing.T) {
	dir := t.TempDir()
	w, _ := rotation.New(filepath.Join(dir, "out.log"), 1024, 2)
	defer w.Close()

	msg := "hello portwatch\n"
	n, err := w.Write([]byte(msg))
	if err != nil {
		t.Fatalf("write error: %v", err)
	}
	if n != len(msg) {
		t.Errorf("wrote %d bytes, want %d", n, len(msg))
	}

	data, _ := os.ReadFile(filepath.Join(dir, "out.log"))
	if !strings.Contains(string(data), "hello portwatch") {
		t.Errorf("file content missing expected string")
	}
}

func TestWrite_RotatesWhenFull(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.log")
	w, _ := rotation.New(path, 20, 3)
	defer w.Close()

	// Each write is 10 bytes; second write triggers rotation.
	for i := 0; i < 4; i++ {
		_, err := w.Write([]byte("0123456789"))
		if err != nil {
			t.Fatalf("write %d error: %v", i, err)
		}
	}

	backups := w.BackupPaths()
	if len(backups) == 0 {
		t.Error("expected at least one backup file after rotation")
	}
	for _, bp := range backups {
		if _, err := os.Stat(bp); err != nil {
			t.Errorf("backup file %s not found: %v", bp, err)
		}
	}
}

func TestBackupPaths_EmptyWhenNoRotation(t *testing.T) {
	dir := t.TempDir()
	w, _ := rotation.New(filepath.Join(dir, "out.log"), 1024, 3)
	defer w.Close()

	if paths := w.BackupPaths(); len(paths) != 0 {
		t.Errorf("expected no backups, got %v", paths)
	}
}

func TestClose_Idempotent(t *testing.T) {
	dir := t.TempDir()
	w, _ := rotation.New(filepath.Join(dir, "out.log"), 1024, 2)
	if err := w.Close(); err != nil {
		t.Fatalf("first close: %v", err)
	}
	// Second close should not panic; error is acceptable.
	_ = w.Close()
}
