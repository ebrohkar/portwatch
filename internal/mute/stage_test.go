package mute

import (
	"context"
	"testing"
	"time"

	"github.com/example/portwatch/internal/alert"
)

func makeAlert(port int) alert.Alert {
	return alert.Alert{
		Port:     port,
		Event:    "open",
		Severity: "info",
	}
}

func TestNewStage_PanicsOnNilStore(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on nil store")
		}
	}()
	NewStage(nil)
}

func TestStage_Allow_NotMuted(t *testing.T) {
	s := withClock(fixedClock(epoch))
	stage := NewStage(s)
	a := makeAlert(8080)
	out, ok, err := stage.Allow(context.Background(), a)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected alert to pass through")
	}
	if out.Port != 8080 {
		t.Fatalf("expected port 8080, got %d", out.Port)
	}
}

func TestStage_Allow_MutedPort_Drops(t *testing.T) {
	s := withClock(fixedClock(epoch))
	s.Add(8080, time.Hour, "maintenance")
	stage := NewStage(s)
	_, ok, err := stage.Allow(context.Background(), makeAlert(8080))
	if ok {
		t.Fatal("expected muted alert to be dropped")
	}
	if err == nil {
		t.Fatal("expected non-nil error for muted port")
	}
}

func TestStage_Allow_DifferentPort_Passes(t *testing.T) {
	s := withClock(fixedClock(epoch))
	s.Add(8080, time.Hour, "")
	stage := NewStage(s)
	_, ok, err := stage.Allow(context.Background(), makeAlert(9090))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected unmuted port to pass")
	}
}

func TestStage_Allow_ExpiredMute_Passes(t *testing.T) {
	now := epoch
	s := withClock(func() time.Time { return now })
	s.Add(8080, time.Second, "")
	now = epoch.Add(5 * time.Second)
	stage := NewStage(s)
	_, ok, err := stage.Allow(context.Background(), makeAlert(8080))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected expired mute to allow alert")
	}
}
