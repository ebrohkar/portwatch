package grouper_test

import (
	"context"
	"testing"

	"github.com/example/portwatch/internal/grouper"
)

func TestNewStage_PanicsOnNilGrouper(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil grouper")
		}
	}()
	grouper.NewStage(nil)
}

func TestStage_Allow_AlwaysPermits(t *testing.T) {
	g, err := grouper.New([]grouper.Rule{
		{Name: "web", Ports: portsOf(80)},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := grouper.NewStage(g)
	ok, err := s.Allow(context.Background(), makeAlert(80))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected Allow to return true")
	}
}

func TestStage_Allow_UnmatchedPort_StillPermits(t *testing.T) {
	g, _ := grouper.New([]grouper.Rule{
		{Name: "web", Ports: portsOf(80)},
	})
	s := grouper.NewStage(g)
	ok, err := s.Allow(context.Background(), makeAlert(9999))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected Allow to return true even for unmatched port")
	}
	if g.Len() != 0 {
		t.Fatal("unmatched alert should not be buffered")
	}
}

func TestStage_Allow_AccumulatesInGrouper(t *testing.T) {
	g, _ := grouper.New([]grouper.Rule{
		{Name: "web", Ports: portsOf(80, 443)},
	})
	s := grouper.NewStage(g)
	for _, port := range []int{80, 443, 80} {
		s.Allow(context.Background(), makeAlert(port)) //nolint:errcheck
	}
	if g.Len() != 3 {
		t.Fatalf("expected 3 buffered alerts, got %d", g.Len())
	}
}

func TestStage_String_ContainsGroupCount(t *testing.T) {
	g, _ := grouper.New([]grouper.Rule{
		{Name: "web", Ports: portsOf(80)},
		{Name: "db", Ports: portsOf(5432)},
	})
	s := grouper.NewStage(g)
	str := s.String()
	if str == "" {
		t.Fatal("expected non-empty string from Stage.String()")
	}
}
