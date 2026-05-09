package audit_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/user/portwatch/internal/audit"
)

func TestNew_DefaultsToStdout(t *testing.T) {
	// Just ensure New(nil) does not panic.
	l := audit.New(nil)
	if l == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestNew_CustomWriter(t *testing.T) {
	var buf bytes.Buffer
	l := audit.New(&buf)
	if l == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestLog_WritesValidJSON(t *testing.T) {
	var buf bytes.Buffer
	l := audit.New(&buf)

	err := l.Log("port_opened", "unexpected port detected", 8080, "high")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var entry audit.Entry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
	}

	if entry.Event != "port_opened" {
		t.Errorf("event = %q, want %q", entry.Event, "port_opened")
	}
	if entry.Port != 8080 {
		t.Errorf("port = %d, want 8080", entry.Port)
	}
	if entry.Severity != "high" {
		t.Errorf("severity = %q, want %q", entry.Severity, "high")
	}
	if entry.Message != "unexpected port detected" {
		t.Errorf("message = %q, want %q", entry.Message, "unexpected port detected")
	}
}

func TestLog_TimestampIsUTC(t *testing.T) {
	var buf bytes.Buffer
	l := audit.New(&buf)

	_ = l.Log("scan", "scan complete", 0, "")

	var entry audit.Entry
	_ = json.Unmarshal(buf.Bytes(), &entry)

	if entry.Timestamp.Location() != time.UTC {
		t.Errorf("expected UTC timestamp, got %v", entry.Timestamp.Location())
	}
}

func TestLog_EntriesAreNewlineDelimited(t *testing.T) {
	var buf bytes.Buffer
	l := audit.New(&buf)

	_ = l.Log("event_a", "first", 22, "low")
	_ = l.Log("event_b", "second", 443, "medium")

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	for i, line := range lines {
		var e audit.Entry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			t.Errorf("line %d is not valid JSON: %v", i, err)
		}
	}
}

func TestLog_WriterError_ReturnsError(t *testing.T) {
	l := audit.New(&failWriter{})
	err := l.Log("scan", "msg", 0, "")
	if err == nil {
		t.Fatal("expected error from failing writer, got nil")
	}
}

type failWriter struct{}

func (f *failWriter) Write(_ []byte) (int, error) {
	return 0, bytes.ErrTooLarge
}
