package digest

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestNewStage_PanicsOnNilDigest(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on nil digest")
		}
	}()
	NewStage(nil)
}

func TestStage_Allow_AlwaysReturnsTrue(t *testing.T) {
	var buf bytes.Buffer
	d := New(&buf, time.Minute)
	s := NewStage(d)
	if !s.Allow(makeAlert(8080, "opened")) {
		t.Error("expected Allow to return true")
	}
}

func TestStage_Allow_FeedsDigest(t *testing.T) {
	var buf bytes.Buffer
	d := New(&buf, time.Minute)
	s := NewStage(d)
	s.Allow(makeAlert(9090, "opened"))
	s.Allow(makeAlert(9090, "closed"))
	entries := d.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Total != 2 {
		t.Errorf("expected total=2, got %d", entries[0].Total)
	}
}

func TestStage_Allow_MultiplePortsTracked(t *testing.T) {
	var buf bytes.Buffer
	d := New(&buf, time.Minute)
	s := NewStage(d)
	s.Allow(makeAlert(80, "opened"))
	s.Allow(makeAlert(443, "opened"))
	s.Allow(makeAlert(80, "closed"))
	entries := d.Entries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestStage_String_ContainsEntryCount(t *testing.T) {
	var buf bytes.Buffer
	d := New(&buf, time.Minute)
	s := NewStage(d)
	s.Allow(makeAlert(8080, "opened"))
	str := s.String()
	if !strings.Contains(str, "DigestStage") {
		t.Errorf("unexpected String output: %s", str)
	}
	if !strings.Contains(str, "entries=1") {
		t.Errorf("expected entries=1 in String, got: %s", str)
	}
}
