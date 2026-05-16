package pipeline_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/user/portwatch/internal/debounce"
	"github.com/user/portwatch/internal/filter"
	"github.com/user/portwatch/internal/pipeline"
	"github.com/user/portwatch/internal/ratelimit"
	"github.com/user/portwatch/internal/suppress"
)

func buildParts(t *testing.T) (
	*filter.Filter,
	*debounce.Debouncer,
	*ratelimit.RateLimiter,
	*suppress.Suppressor,
) {
	t.Helper()
	f, err := filter.New(nil, nil)
	if err != nil {
		t.Fatalf("filter.New: %v", err)
	}
	d := debounce.New(debounce.WithCooldown(time.Millisecond))
	rl := ratelimit.New(ratelimit.WithMaxHits(10), ratelimit.WithWindow(time.Second))
	su := suppress.New()
	return f, d, rl, su
}

func TestFromParts_PanicsOnNilNotifier(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil notifier")
		}
	}()
	f, d, rl, su := buildParts(t)
	pipeline.FromParts(nil, f, d, rl, su)
}

func TestFromParts_AlertPassesAllStages(t *testing.T) {
	var buf bytes.Buffer
	n := buildNotifier(t, &buf)
	f, d, rl, su := buildParts(t)
	p := pipeline.FromParts(n, f, d, rl, su)

	a := makeAlert(9090, "open")
	if err := p.Run(context.Background(), a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected alert output")
	}
}

func TestFromParts_FilterExcludesPort(t *testing.T) {
	var buf bytes.Buffer
	n := buildNotifier(t, &buf)

	f, err := filter.New(nil, []int{9090})
	if err != nil {
		t.Fatalf("filter.New: %v", err)
	}
	d := debounce.New(debounce.WithCooldown(time.Millisecond))
	rl := ratelimit.New(ratelimit.WithMaxHits(10), ratelimit.WithWindow(time.Second))
	su := suppress.New()

	p := pipeline.FromParts(n, f, d, rl, su)
	a := makeAlert(9090, "open")
	if err := p.Run(context.Background(), a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Fatalf("excluded port should not produce output, got %d bytes", buf.Len())
	}
}

func TestFromParts_SuppressedAlertDropped(t *testing.T) {
	var buf bytes.Buffer
	n := buildNotifier(t, &buf)
	f, d, rl, su := buildParts(t)

	su.Add(8080, "open", time.Hour)

	p := pipeline.FromParts(n, f, d, rl, su)
	a := makeAlert(8080, "open")
	if err := p.Run(context.Background(), a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Fatalf("suppressed alert should not produce output, got %d bytes", buf.Len())
	}
}
