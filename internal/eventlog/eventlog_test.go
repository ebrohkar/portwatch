package eventlog

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestNew_DefaultsToStdout(t *testing.T) {
	el := New(nil)
	if el.out == nil {
		t.Fatal("expected non-nil writer")
	}
}

func TestNew_CustomWriter(t *testing.T) {
	var buf bytes.Buffer
	el := New(&buf)
	if el.out != &buf {
		t.Fatal("expected custom writer to be set")
	}
}

func TestRecord_WritesValidJSON(t *testing.T) {
	var buf bytes.Buffer
	el := New(&buf)

	err := el.Record(Entry{
		Port:     8080,
		Event:    "opened",
		Severity: "warn",
		Message:  "unexpected port",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got Entry
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if got.Port != 8080 {
		t.Errorf("port: got %d, want 8080", got.Port)
	}
	if got.Event != "opened" {
		t.Errorf("event: got %q, want \"opened\"", got.Event)
	}
}

func TestRecord_SetsTimestampWhenZero(t *testing.T) {
	var buf bytes.Buffer
	fixed := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	el := New(&buf)
	el.now = func() time.Time { return fixed }

	_ = el.Record(Entry{Port: 443, Event: "closed", Severity: "info"})

	var got Entry
	_ = json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &got)
	if !got.Timestamp.Equal(fixed) {
		t.Errorf("timestamp: got %v, want %v", got.Timestamp, fixed)
	}
}

func TestRecord_PreservesExplicitTimestamp(t *testing.T) {
	var buf bytes.Buffer
	explicit := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)
	el := New(&buf)
	// now should not override an already-set timestamp
	el.now = func() time.Time { return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC) }

	_ = el.Record(Entry{Port: 80, Event: "opened", Severity: "info", Timestamp: explicit})

	var got Entry
	_ = json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &got)
	if !got.Timestamp.Equal(explicit) {
		t.Errorf("timestamp: got %v, want explicit %v", got.Timestamp, explicit)
	}
}

func TestRecord_InvalidPort(t *testing.T) {
	var buf bytes.Buffer
	el := New(&buf)

	if err := el.Record(Entry{Port: 0, Event: "opened"}); err == nil {
		t.Error("expected error for port 0")
	}
	if err := el.Record(Entry{Port: 99999, Event: "opened"}); err == nil {
		t.Error("expected error for port 99999")
	}
}

func TestRecord_EmptyEvent(t *testing.T) {
	var buf bytes.Buffer
	el := New(&buf)

	if err := el.Record(Entry{Port: 80, Event: ""}); err == nil {
		t.Error("expected error for empty event")
	}
}

func TestRecord_EntriesAreNewlineDelimited(t *testing.T) {
	var buf bytes.Buffer
	el := New(&buf)

	_ = el.Record(Entry{Port: 80, Event: "opened", Severity: "info"})
	_ = el.Record(Entry{Port: 443, Event: "opened", Severity: "info"})

	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	for i, line := range lines {
		var e Entry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			t.Errorf("line %d is not valid JSON: %v", i, err)
		}
	}
}
