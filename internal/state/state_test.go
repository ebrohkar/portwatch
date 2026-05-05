package state

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func tempFile(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "state.json")
}

func TestLoad_NoFile(t *testing.T) {
	store := NewStore(tempFile(t))
	snap, err := store.Load()
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	if snap != nil {
		t.Fatal("expected nil snapshot when file does not exist")
	}
}

func TestSave_And_Load_RoundTrip(t *testing.T) {
	store := NewStore(tempFile(t))
	ports := []int{22, 80, 443}

	if err := store.Save(ports); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	snap, err := store.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if snap == nil {
		t.Fatal("expected non-nil snapshot after save")
	}
	if len(snap.OpenPorts) != len(ports) {
		t.Fatalf("expected %d ports, got %d", len(ports), len(snap.OpenPorts))
	}
}

func TestLoad_MalformedJSON(t *testing.T) {
	path := tempFile(t)
	os.WriteFile(path, []byte("not-json"), 0o644)

	store := NewStore(path)
	_, err := store.Load()
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestCurrent_BeforeLoad(t *testing.T) {
	store := NewStore(tempFile(t))
	if store.Current() != nil {
		t.Fatal("expected nil current before any load or save")
	}
}

func TestDiff_Appeared(t *testing.T) {
	prev := []int{22, 80}
	next := []int{22, 80, 443}

	appeared, disappeared := Diff(prev, next)
	if len(disappeared) != 0 {
		t.Fatalf("expected no disappeared ports, got %v", disappeared)
	}
	if len(appeared) != 1 || appeared[0] != 443 {
		t.Fatalf("expected [443] appeared, got %v", appeared)
	}
}

func TestDiff_Disappeared(t *testing.T) {
	prev := []int{22, 80, 8080}
	next := []int{22, 80}

	appeared, disappeared := Diff(prev, next)
	if len(appeared) != 0 {
		t.Fatalf("expected no appeared ports, got %v", appeared)
	}
	if len(disappeared) != 1 || disappeared[0] != 8080 {
		t.Fatalf("expected [8080] disappeared, got %v", disappeared)
	}
}

func TestDiff_NoChange(t *testing.T) {
	ports := []int{22, 80, 443}
	appeared, disappeared := Diff(ports, ports)
	if len(appeared) != 0 || len(disappeared) != 0 {
		t.Fatalf("expected no diff, got appeared=%v disappeared=%v", appeared, disappeared)
	}
}

func TestDiff_BothChanges(t *testing.T) {
	prev := []int{22, 80}
	next := []int{80, 443}

	appeared, disappeared := Diff(prev, next)
	sort.Ints(appeared)
	sort.Ints(disappeared)

	if len(appeared) != 1 || appeared[0] != 443 {
		t.Fatalf("expected appeared=[443], got %v", appeared)
	}
	if len(disappeared) != 1 || disappeared[0] != 22 {
		t.Fatalf("expected disappeared=[22], got %v", disappeared)
	}
}
