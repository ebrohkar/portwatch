package metrics_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/user/portwatch/internal/metrics"
)

func TestNew_SetsStartTime(t *testing.T) {
	before := time.Now()
	c := metrics.New()
	after := time.Now()

	up := c.Uptime()
	if up < 0 {
		t.Fatalf("expected non-negative uptime, got %s", up)
	}
	_ = before
	_ = after
}

func TestIncScans_Increments(t *testing.T) {
	c := metrics.New()
	c.IncScans()
	c.IncScans()
	if got := c.ScansTotal.Load(); got != 2 {
		t.Fatalf("expected 2 scans, got %d", got)
	}
}

func TestIncAlerts_Increments(t *testing.T) {
	c := metrics.New()
	c.IncAlerts()
	if got := c.AlertsTotal.Load(); got != 1 {
		t.Fatalf("expected 1 alert, got %d", got)
	}
}

func TestIncSuppressed_Increments(t *testing.T) {
	c := metrics.New()
	c.IncSuppressed()
	c.IncSuppressed()
	c.IncSuppressed()
	if got := c.SuppressTotal.Load(); got != 3 {
		t.Fatalf("expected 3 suppressed, got %d", got)
	}
}

func TestIncErrors_Increments(t *testing.T) {
	c := metrics.New()
	c.IncErrors()
	if got := c.ErrorsTotal.Load(); got != 1 {
		t.Fatalf("expected 1 error, got %d", got)
	}
}

func TestWriteTo_ContainsAllFields(t *testing.T) {
	c := metrics.New()
	c.IncScans()
	c.IncAlerts()
	c.IncSuppressed()
	c.IncErrors()

	var buf bytes.Buffer
	_, err := c.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo returned error: %v", err)
	}

	out := buf.String()
	for _, want := range []string{"uptime", "scans_total", "alerts_total", "suppressed_total", "errors_total"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing field %q\ngot:\n%s", want, out)
		}
	}
}

func TestWriteTo_NilWriterDefaultsToStdout(t *testing.T) {
	c := metrics.New()
	// Should not panic when w is nil; stdout is used as fallback.
	_, err := c.WriteTo(nil)
	if err != nil {
		t.Fatalf("unexpected error with nil writer: %v", err)
	}
}
