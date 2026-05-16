package sampler

import (
	"testing"
)

func alwaysPass() float64  { return 0.0 }
func alwaysBlock() float64 { return 1.0 }

func TestNew_ValidRate(t *testing.T) {
	s, err := New(0.5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil sampler")
	}
}

func TestNew_InvalidRate_Zero(t *testing.T) {
	_, err := New(0.0)
	if err == nil {
		t.Fatal("expected error for rate=0")
	}
}

func TestNew_InvalidRate_AboveOne(t *testing.T) {
	_, err := New(1.1)
	if err == nil {
		t.Fatal("expected error for rate>1")
	}
}

func TestNew_RateOne_AlwaysPasses(t *testing.T) {
	s, err := withRand(1.0, alwaysPass)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !s.Allow(80) {
		t.Error("rate=1.0 should always allow")
	}
}

func TestAllow_UsesDefaultRate(t *testing.T) {
	s, _ := withRand(0.5, alwaysPass) // randf returns 0.0 < 0.5
	if !s.Allow(443) {
		t.Error("expected Allow=true when randf < rate")
	}

	s2, _ := withRand(0.5, alwaysBlock) // randf returns 1.0 >= 0.5
	if s2.Allow(443) {
		t.Error("expected Allow=false when randf >= rate")
	}
}

func TestAllow_UsesPortSpecificRate(t *testing.T) {
	s, _ := withRand(0.1, alwaysPass) // default 10%, randf=0.0 always passes
	if err := s.SetPortRate(8080, 0.9); err != nil {
		t.Fatalf("SetPortRate: %v", err)
	}
	// port 8080 has rate 0.9, randf=0.0 → 0.0 < 0.9 → allow
	if !s.Allow(8080) {
		t.Error("expected Allow=true for port-specific rate")
	}
}

func TestSetPortRate_InvalidPort(t *testing.T) {
	s, _ := New(1.0)
	if err := s.SetPortRate(0, 0.5); err == nil {
		t.Error("expected error for port=0")
	}
	if err := s.SetPortRate(65536, 0.5); err == nil {
		t.Error("expected error for port=65536")
	}
}

func TestSetPortRate_InvalidRate(t *testing.T) {
	s, _ := New(1.0)
	if err := s.SetPortRate(80, 0.0); err == nil {
		t.Error("expected error for rate=0")
	}
	if err := s.SetPortRate(80, 1.5); err == nil {
		t.Error("expected error for rate=1.5")
	}
}

func TestRate_ReturnsPortSpecific(t *testing.T) {
	s, _ := New(0.3)
	_ = s.SetPortRate(22, 0.8)
	if got := s.Rate(22); got != 0.8 {
		t.Errorf("expected 0.8, got %v", got)
	}
	if got := s.Rate(9999); got != 0.3 {
		t.Errorf("expected default 0.3, got %v", got)
	}
}
