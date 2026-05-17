package labeler_test

import (
	"testing"
	"time"

	"github.com/example/portwatch/internal/alert"
	"github.com/example/portwatch/internal/labeler"
)

func makeAlert(port int) alert.Alert {
	return alert.Alert{
		Port:      port,
		Event:     "open",
		Severity:  "info",
		Timestamp: time.Now().UTC(),
	}
}

func TestNew_EmptyMappings(t *testing.T) {
	l, err := labeler.New(map[int]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l.Len() != 0 {
		t.Fatalf("expected 0 mappings, got %d", l.Len())
	}
}

func TestNew_ValidMappings(t *testing.T) {
	l, err := labeler.New(map[int]string{80: "http", 443: "https"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l.Len() != 2 {
		t.Fatalf("expected 2 mappings, got %d", l.Len())
	}
}

func TestNew_InvalidPort_Zero(t *testing.T) {
	_, err := labeler.New(map[int]string{0: "invalid"})
	if err == nil {
		t.Fatal("expected error for port 0, got nil")
	}
}

func TestNew_InvalidPort_TooHigh(t *testing.T) {
	_, err := labeler.New(map[int]string{99999: "oob"})
	if err == nil {
		t.Fatal("expected error for port 99999, got nil")
	}
}

func TestNew_EmptyLabel(t *testing.T) {
	_, err := labeler.New(map[int]string{8080: ""})
	if err == nil {
		t.Fatal("expected error for empty label, got nil")
	}
}

func TestLabel_Found(t *testing.T) {
	l, _ := labeler.New(map[int]string{22: "ssh"})
	label, ok := l.Label(22)
	if !ok {
		t.Fatal("expected label to be found")
	}
	if label != "ssh" {
		t.Fatalf("expected \"ssh\", got %q", label)
	}
}

func TestLabel_NotFound(t *testing.T) {
	l, _ := labeler.New(map[int]string{22: "ssh"})
	_, ok := l.Label(9999)
	if ok {
		t.Fatal("expected no label for unmapped port")
	}
}

func TestAnnotate_SetsServiceLabel(t *testing.T) {
	l, _ := labeler.New(map[int]string{80: "http"})
	a := makeAlert(80)
	annotated := l.Annotate(a)
	if annotated.Service != "http" {
		t.Fatalf("expected service \"http\", got %q", annotated.Service)
	}
}

func TestAnnotate_NoMapping_Unchanged(t *testing.T) {
	l, _ := labeler.New(map[int]string{80: "http"})
	a := makeAlert(9999)
	a.Service = "original"
	annotated := l.Annotate(a)
	if annotated.Service != "original" {
		t.Fatalf("expected service to remain \"original\", got %q", annotated.Service)
	}
}

func TestAnnotate_DoesNotMutateOriginal(t *testing.T) {
	l, _ := labeler.New(map[int]string{443: "https"})
	a := makeAlert(443)
	a.Service = ""
	_ = l.Annotate(a)
	if a.Service != "" {
		t.Fatal("original alert should not be mutated")
	}
}
