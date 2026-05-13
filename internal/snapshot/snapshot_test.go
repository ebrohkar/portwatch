package snapshot

import (
	"testing"
)

func TestNew_SortsAndDeduplicates(t *testing.T) {
	s, err := New([]int{443, 80, 80, 22})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []int{22, 80, 443}
	if len(s.Ports) != len(want) {
		t.Fatalf("got %d ports, want %d", len(s.Ports), len(want))
	}
	for i, p := range want {
		if s.Ports[i] != p {
			t.Errorf("ports[%d] = %d, want %d", i, s.Ports[i], p)
		}
	}
}

func TestNew_InvalidPort(t *testing.T) {
	cases := [][]int{
		{0},
		{65536},
		{-1},
	}
	for _, ports := range cases {
		_, err := New(ports)
		if err == nil {
			t.Errorf("expected error for ports %v", ports)
		}
	}
}

func TestNew_EmptyPorts(t *testing.T) {
	s, err := New(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Len() != 0 {
		t.Errorf("expected 0 ports, got %d", s.Len())
	}
}

func TestContains(t *testing.T) {
	s, _ := New([]int{22, 80, 443})
	if !s.Contains(80) {
		t.Error("expected Contains(80) = true")
	}
	if s.Contains(8080) {
		t.Error("expected Contains(8080) = false")
	}
}

func TestAdded_NilPrev(t *testing.T) {
	s, _ := New([]int{22, 80})
	added := s.Added(nil)
	if len(added) != 2 {
		t.Errorf("expected 2 added ports, got %d", len(added))
	}
}

func TestAdded_WithPrev(t *testing.T) {
	prev, _ := New([]int{22, 80})
	curr, _ := New([]int{22, 80, 443})
	added := curr.Added(prev)
	if len(added) != 1 || added[0] != 443 {
		t.Errorf("expected [443], got %v", added)
	}
}

func TestRemoved_NilPrev(t *testing.T) {
	s, _ := New([]int{22})
	removed := s.Removed(nil)
	if len(removed) != 0 {
		t.Errorf("expected 0 removed ports with nil prev, got %d", len(removed))
	}
}

func TestRemoved_WithPrev(t *testing.T) {
	prev, _ := New([]int{22, 80, 443})
	curr, _ := New([]int{22, 443})
	removed := curr.Removed(prev)
	if len(removed) != 1 || removed[0] != 80 {
		t.Errorf("expected [80], got %v", removed)
	}
}

func TestAdded_And_Removed_NoChange(t *testing.T) {
	prev, _ := New([]int{22, 80})
	curr, _ := New([]int{22, 80})
	if len(curr.Added(prev)) != 0 {
		t.Error("expected no added ports")
	}
	if len(curr.Removed(prev)) != 0 {
		t.Error("expected no removed ports")
	}
}
