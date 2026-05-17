package shedder_test

import (
	"sync"
	"testing"

	"github.com/example/portwatch/internal/shedder"
)

func TestNew_PanicsOnZeroMax(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for max=0")
		}
	}()
	shedder.New(0)
}

func TestNew_PanicsOnNegativeMax(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for max=-1")
		}
	}()
	shedder.New(-1)
}

func TestAllow_PermitsUpToMax(t *testing.T) {
	s := shedder.New(3)
	for i := 0; i < 3; i++ {
		if err := s.Allow(); err != nil {
			t.Fatalf("expected nil on call %d, got %v", i+1, err)
		}
	}
	if got := s.Inflight(); got != 3 {
		t.Fatalf("inflight = %d, want 3", got)
	}
}

func TestAllow_RejectsBeyondMax(t *testing.T) {
	s := shedder.New(2)
	_ = s.Allow()
	_ = s.Allow()

	if err := s.Allow(); err != shedder.ErrShedded {
		t.Fatalf("expected ErrShedded, got %v", err)
	}
	if got := s.Dropped(); got != 1 {
		t.Fatalf("dropped = %d, want 1", got)
	}
}

func TestDone_DecrementsInflight(t *testing.T) {
	s := shedder.New(1)
	if err := s.Allow(); err != nil {
		t.Fatal(err)
	}
	s.Done()
	if got := s.Inflight(); got != 0 {
		t.Fatalf("inflight = %d after Done, want 0", got)
	}
	// Should be allowed again after Done.
	if err := s.Allow(); err != nil {
		t.Fatalf("expected nil after Done, got %v", err)
	}
}

func TestAllow_ConcurrentSafety(t *testing.T) {
	const ceiling = 10
	s := shedder.New(ceiling)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := s.Allow(); err == nil {
				defer s.Done()
			}
		}()
	}
	wg.Wait()

	if got := s.Inflight(); got != 0 {
		t.Fatalf("inflight = %d after all goroutines finished, want 0", got)
	}
}

func TestString_ContainsFields(t *testing.T) {
	s := shedder.New(5)
	_ = s.Allow()
	str := s.String()
	for _, want := range []string{"max=5", "inflight=1", "dropped=0"} {
		if !containsSubstr(str, want) {
			t.Errorf("String() = %q, missing %q", str, want)
		}
	}
}

func containsSubstr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub ||
		len(s) > 0 && func() bool {
			for i := 0; i+len(sub) <= len(s); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
