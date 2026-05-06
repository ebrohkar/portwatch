package history_test

import (
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/history"
)

func makeAlert(port int) alert.Alert {
	return alert.Alert{
		Port:     port,
		Event:    "opened",
		Severity: "warn",
	}
}

func TestNew_DefaultMaxSize(t *testing.T) {
	l := history.New(0)
	if l == nil {
		t.Fatal("expected non-nil Log")
	}
}

func TestAdd_And_Len(t *testing.T) {
	l := history.New(10)
	l.Add(makeAlert(8080))
	l.Add(makeAlert(9090))

	if got := l.Len(); got != 2 {
		t.Fatalf("expected 2 entries, got %d", got)
	}
}

func TestAdd_EvictsOldestWhenFull(t *testing.T) {
	l := history.New(3)
	for port := 1; port <= 5; port++ {
		l.Add(makeAlert(port))
	}

	if got := l.Len(); got != 3 {
		t.Fatalf("expected 3 entries after eviction, got %d", got)
	}

	entries := l.All()
	// oldest three should be ports 3, 4, 5
	if entries[0].Alert.Port != 3 {
		t.Errorf("expected oldest retained port=3, got %d", entries[0].Alert.Port)
	}
	if entries[2].Alert.Port != 5 {
		t.Errorf("expected newest port=5, got %d", entries[2].Alert.Port)
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	l := history.New(10)
	l.Add(makeAlert(8080))

	a := l.All()
	a[0].Alert.Port = 9999 // mutate the copy

	b := l.All()
	if b[0].Alert.Port == 9999 {
		t.Error("All() returned a reference to internal slice, expected a copy")
	}
}

func TestSince_FiltersCorrectly(t *testing.T) {
	l := history.New(10)

	l.Add(makeAlert(1111))
	cutoff := time.Now()
	time.Sleep(2 * time.Millisecond)
	l.Add(makeAlert(2222))

	result := l.Since(cutoff)
	if len(result) != 1 {
		t.Fatalf("expected 1 entry since cutoff, got %d", len(result))
	}
	if result[0].Alert.Port != 2222 {
		t.Errorf("expected port 2222, got %d", result[0].Alert.Port)
	}
}

func TestSince_EmptyWhenAllOld(t *testing.T) {
	l := history.New(10)
	l.Add(makeAlert(3333))

	future := time.Now().Add(time.Hour)
	result := l.Since(future)
	if len(result) != 0 {
		t.Errorf("expected 0 entries, got %d", len(result))
	}
}
