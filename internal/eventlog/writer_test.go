package eventlog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenFile_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "events.jsonl")

	fw, err := OpenFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer fw.Close()

	if fw.Name() != path {
		t.Errorf("name: got %q, want %q", fw.Name(), path)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("file not created: %v", err)
	}
}

func TestOpenFile_InvalidPath(t *testing.T) {
	// Root-owned directory that we cannot write to.
	_, err := OpenFile("/proc/portwatch_test_no_write/events.jsonl")
	if err == nil {
		t.Fatal("expected error for unwritable path")
	}
}

func TestFileWriter_Write_And_Read(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "events.jsonl")

	fw, err := OpenFile(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	el := New(fw)
	_ = el.Record(Entry{Port: 9090, Event: "opened", Severity: "warn", Message: "test"})
	fw.Close()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(string(data), "9090") {
		t.Errorf("expected port 9090 in file, got: %s", data)
	}
}

func TestFileWriter_Close_Idempotent(t *testing.T) {
	dir := t.TempDir()
	fw, err := OpenFile(filepath.Join(dir, "events.jsonl"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := fw.Close(); err != nil {
		t.Fatalf("first close: %v", err)
	}
	// Second close on an already-closed file should surface an error but not panic.
	_ = fw.Close()
}
