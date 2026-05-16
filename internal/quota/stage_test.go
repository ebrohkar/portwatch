package quota

import (
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
)

func makeAlert(port int) alert.Alert {
	a, _ := alert.New(port, "open", "info")
	return a
}

func TestNewStage_PanicsOnNilQuota(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil quota")
		}
	}()
	NewStage(nil)
}

func TestStage_Allow_WithinQuota_Passes(t *testing.T) {
	q := newWithClock(3, time.Minute, fixed)
	s := NewStage(q)
	a := makeAlert(80)
	if !s.Allow(a) {
		t.Fatal("first alert should pass")
	}
}

func TestStage_Allow_ExceedsQuota_Drops(t *testing.T) {
	q := newWithClock(1, time.Minute, fixed)
	s := NewStage(q)
	a := makeAlert(443)
	s.Allow(a)
	if s.Allow(a) {
		t.Fatal("second alert should be dropped")
	}
}

func TestStage_Allow_DifferentPorts_Independent(t *testing.T) {
	q := newWithClock(1, time.Minute, fixed)
	s := NewStage(q)
	if !s.Allow(makeAlert(22)) {
		t.Fatal("port 22 should pass")
	}
	if !s.Allow(makeAlert(8080)) {
		t.Fatal("port 8080 should pass independently")
	}
}

func TestStage_String_ContainsQuota(t *testing.T) {
	q := New(5, time.Minute)
	s := NewStage(q)
	if s.String() == "" {
		t.Fatal("String should not be empty")
	}
}
