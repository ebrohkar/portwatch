package baseline_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/yourorg/portwatch/internal/baseline"
)

func TestNew_Empty(t *testing.T) {
	b := baseline.New()
	if b.Len() != 0 {
		t.Fatalf("expected 0 ports, got %d", b.Len())
	}
}

func TestSet_And_Contains(t *testing.T) {
	b := baseline.New()
	b.Set([]int{80, 443, 8080})

	for _, p := range []int{80, 443, 8080} {
		if !b.Contains(p) {
			t.Errorf("expected port %d to be in baseline", p)
		}
	}
	if b.Contains(22) {
		t.Error("port 22 should not be in baseline")
	}
}

func TestSet_ReplacesExisting(t *testing.T) {
	b := baseline.New()
	b.Set([]int{80, 443})
	b.Set([]int{22})

	if b.Len() != 1 {
		t.Fatalf("expected 1 port after replacement, got %d", b.Len())
	}
	if !b.Contains(22) {
		t.Error("expected port 22 after replacement")
	}
	if b.Contains(80) {
		t.Error("port 80 should have been replaced")
	}
}

func TestLen_MatchesInput(t *testing.T) {
	b := baseline.New()
	b.Set([]int{1, 2, 3, 4, 5})
	if b.Len() != 5 {
		t.Fatalf("expected 5, got %d", b.Len())
	}
}

func TestSaveToFile_And_LoadFromFile_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "baseline.json")

	original := baseline.New()
	original.Set([]int{22, 80, 443})

	if err := original.SaveToFile(path); err != nil {
		t.Fatalf("SaveToFile: %v", err)
	}

	loaded, err := baseline.LoadFromFile(path)
	if err != nil {
		t.Fatalf("LoadFromFile: %v", err)
	}

	for _, p := range []int{22, 80, 443} {
		if !loaded.Contains(p) {
			t.Errorf("loaded baseline missing port %d", p)
		}
	}
	if loaded.Len() != 3 {
		t.Fatalf("expected 3 ports, got %d", loaded.Len())
	}
}

func TestLoadFromFile_NotFound(t *testing.T) {
	_, err := baseline.LoadFromFile("/nonexistent/baseline.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestSaveToFile_WritesValidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bl.json")

	b := baseline.New()
	b.Set([]int{8080})
	if err := b.SaveToFile(path); err != nil {
		t.Fatalf("SaveToFile: %v", err)
	}

	data, _ := os.ReadFile(path)
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("file is not valid JSON: %v", err)
	}
	if _, ok := raw["captured_at"]; !ok {
		t.Error("missing captured_at field in JSON")
	}
}
