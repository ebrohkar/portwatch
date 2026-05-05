package notifier

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestNew_DefaultsToStdout(t *testing.T) {
	n := New()
	if len(n.writers) != 1 {
		t.Fatalf("expected 1 writer, got %d", len(n.writers))
	}
}

func TestNew_CustomWriter(t *testing.T) {
	var buf bytes.Buffer
	n := New(&buf)
	if len(n.writers) != 1 {
		t.Fatalf("expected 1 writer, got %d", len(n.writers))
	}
}

func TestSend_FormatsOutput(t *testing.T) {
	var buf bytes.Buffer
	n := New(&buf)

	ts := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	err := n.Send(Event{
		Timestamp: ts,
		Level:     LevelAlert,
		Port:      8080,
		Action:    "alert",
		Message:   "unexpected open port",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	for _, want := range []string{"2024-01-15T10:00:00Z", "[ALERT]", "port=8080", "action=alert", "unexpected open port"} {
		if !strings.Contains(out, want) {
			t.Errorf("output %q missing %q", out, want)
		}
	}
}

func TestSend_SetsTimestampWhenZero(t *testing.T) {
	var buf bytes.Buffer
	n := New(&buf)

	err := n.Send(Event{
		Level:   LevelWarn,
		Port:    443,
		Action:  "log",
		Message: "port closed unexpectedly",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected output, got empty buffer")
	}
}

func TestSend_MultipleWriters(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	n := New(&buf1, &buf2)

	err := n.Send(Event{
		Level:   LevelInfo,
		Port:    22,
		Action:  "log",
		Message: "port state unchanged",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf1.String() != buf2.String() {
		t.Errorf("writers received different output:\nbuf1=%q\nbuf2=%q", buf1.String(), buf2.String())
	}
}
