package tagger_test

import (
	"strings"
	"testing"

	"github.com/example/portwatch/internal/alert"
	"github.com/example/portwatch/internal/tagger"
)

func TestNew_ValidMapping(t *testing.T) {
	_, err := tagger.New(map[int]tagger.Tag{
		80:  {Label: "http", Severity: "low"},
		443: {Label: "https"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNew_InvalidPort(t *testing.T) {
	_, err := tagger.New(map[int]tagger.Tag{
		0: {Label: "bad"},
	})
	if err == nil {
		t.Fatal("expected error for port 0")
	}
}

func TestNew_EmptyLabel(t *testing.T) {
	_, err := tagger.New(map[int]tagger.Tag{
		8080: {Label: ""},
	})
	if err == nil {
		t.Fatal("expected error for empty label")
	}
}

func TestLookup_Found(t *testing.T) {
	tgr, _ := tagger.New(map[int]tagger.Tag{
		22: {Label: "ssh", Severity: "high"},
	})
	tag, ok := tgr.Lookup(22)
	if !ok {
		t.Fatal("expected mapping to be found")
	}
	if tag.Label != "ssh" {
		t.Errorf("label: got %q, want %q", tag.Label, "ssh")
	}
}

func TestLookup_NotFound(t *testing.T) {
	tgr, _ := tagger.New(nil)
	_, ok := tgr.Lookup(9999)
	if ok {
		t.Fatal("expected no mapping")
	}
}

func TestAnnotate_AddsPrefixAndSeverity(t *testing.T) {
	tgr, _ := tagger.New(map[int]tagger.Tag{
		3306: {Label: "mysql", Severity: "critical"},
	})
	a := alert.Alert{Port: 3306, Message: "port opened", Severity: "low", Event: "open"}
	got := tgr.Annotate(a)
	if !strings.HasPrefix(got.Message, "[mysql]") {
		t.Errorf("message prefix missing: %q", got.Message)
	}
	if got.Severity != "critical" {
		t.Errorf("severity: got %q, want %q", got.Severity, "critical")
	}
}

func TestAnnotate_NoMapping_ReturnsUnchanged(t *testing.T) {
	tgr, _ := tagger.New(nil)
	a := alert.Alert{Port: 1234, Message: "original", Severity: "info", Event: "open"}
	got := tgr.Annotate(a)
	if got.Message != "original" {
		t.Errorf("message changed unexpectedly: %q", got.Message)
	}
}

func TestSet_AddsNewMapping(t *testing.T) {
	tgr, _ := tagger.New(nil)
	if err := tgr.Set(8443, tagger.Tag{Label: "alt-https"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tag, ok := tgr.Lookup(8443)
	if !ok || tag.Label != "alt-https" {
		t.Errorf("expected alt-https mapping, got %+v ok=%v", tag, ok)
	}
}

func TestSet_InvalidPort_ReturnsError(t *testing.T) {
	tgr, _ := tagger.New(nil)
	if err := tgr.Set(70000, tagger.Tag{Label: "bad"}); err == nil {
		t.Fatal("expected error for port 70000")
	}
}
