package watchdog_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/example/portwatch/internal/watchdog"
)

func TestNew_ValidParams(t *testing.T) {
	w, err := watchdog.New(3, func(int, error) {})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w == nil {
		t.Fatal("expected non-nil watchdog")
	}
}

func TestNew_ZeroThreshold(t *testing.T) {
	_, err := watchdog.New(0, func(int, error) {})
	if err == nil {
		t.Fatal("expected error for zero threshold")
	}
}

func TestNew_NilAlertFn(t *testing.T) {
	_, err := watchdog.New(1, nil)
	if err == nil {
		t.Fatal("expected error for nil alertFn")
	}
}

func TestRecordFailure_BelowThreshold_NoAlert(t *testing.T) {
	called := false
	w, _ := watchdog.New(3, func(int, error) { called = true })
	w.RecordFailure(errors.New("oops"))
	w.RecordFailure(errors.New("oops"))
	if called {
		t.Fatal("alert should not fire below threshold")
	}
	if w.Consecutive() != 2 {
		t.Fatalf("expected consecutive=2, got %d", w.Consecutive())
	}
}

func TestRecordFailure_AtThreshold_FiresAlert(t *testing.T) {
	var mu sync.Mutex
	var gotFailures int
	var gotErr error
	w, _ := watchdog.New(2, func(n int, err error) {
		mu.Lock()
		gotFailures = n
		gotErr = err
		mu.Unlock()
	})
	scanErr := errors.New("scan failed")
	w.RecordFailure(scanErr)
	w.RecordFailure(scanErr)
	mu.Lock()
	defer mu.Unlock()
	if gotFailures != 2 {
		t.Fatalf("expected 2 failures reported, got %d", gotFailures)
	}
	if !errors.Is(gotErr, scanErr) {
		t.Fatalf("expected scan error, got %v", gotErr)
	}
}

func TestRecordFailure_AlertFiresOnceUntilReset(t *testing.T) {
	calls := 0
	w, _ := watchdog.New(2, func(int, error) { calls++ })
	err := errors.New("e")
	w.RecordFailure(err)
	w.RecordFailure(err)
	w.RecordFailure(err) // still above threshold but already alerted
	if calls != 1 {
		t.Fatalf("expected alert called once, got %d", calls)
	}
}

func TestRecordSuccess_ResetsState(t *testing.T) {
	calls := 0
	w, _ := watchdog.New(2, func(int, error) { calls++ })
	err := errors.New("e")
	w.RecordFailure(err)
	w.RecordFailure(err)
	w.RecordSuccess()
	if w.Consecutive() != 0 {
		t.Fatalf("expected consecutive=0 after success, got %d", w.Consecutive())
	}
	// After reset a new breach should fire again.
	w.RecordFailure(err)
	w.RecordFailure(err)
	if calls != 2 {
		t.Fatalf("expected alert called twice (once per breach), got %d", calls)
	}
}

func TestLastFailure_ZeroBeforeAnyFailure(t *testing.T) {
	w, _ := watchdog.New(1, func(int, error) {})
	if !w.LastFailure().IsZero() {
		t.Fatal("expected zero time before any failure")
	}
}

func TestLastFailure_SetAfterFailure(t *testing.T) {
	w, _ := watchdog.New(1, func(int, error) {})
	w.RecordFailure(errors.New("boom"))
	if w.LastFailure().IsZero() {
		t.Fatal("expected non-zero last failure time")
	}
}
