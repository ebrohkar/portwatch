package summary_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/history"
	"github.com/user/portwatch/internal/summary"
)

func makeAlert(port int, event, severity string) alert.Alert {
	return alert.Alert{
		Port:      port,
		Event:     event,
		Severity:  severity,
		Timestamp: time.Now().UTC(),
	}
}

func TestNew_DefaultsToStdoutAndInterval(t *testing.T) {
	h := history.New(0)
	r := summary.New(h, nil, 0)
	if r == nil {
		t.Fatal("expected non-nil reporter")
	}
	if r.Interval() != 5*time.Minute {
		t.Errorf("expected default interval 5m, got %v", r.Interval())
	}
}

func TestNew_CustomInterval(t *testing.T) {
	h := history.New(0)
	r := summary.New(h, &bytes.Buffer{}, 10*time.Minute)
	if r.Interval() != 10*time.Minute {
		t.Errorf("expected 10m interval, got %v", r.Interval())
	}
}

func TestWrite_NoAlerts(t *testing.T) {
	h := history.New(0)
	var buf bytes.Buffer
	r := summary.New(h, &buf, time.Minute)
	r.Write()

	out := buf.String()
	if !strings.Contains(out, "Total alerts: 0") {
		t.Errorf("expected zero alert count, got: %s", out)
	}
	if !strings.Contains(out, "No alerts recorded.") {
		t.Errorf("expected no-alerts message, got: %s", out)
	}
}

func TestWrite_WithAlerts_ShowsCounts(t *testing.T) {
	h := history.New(0)
	h.Add(makeAlert(80, "opened", "warning"))
	h.Add(makeAlert(443, "opened", "critical"))
	h.Add(makeAlert(8080, "closed", "info"))

	var buf bytes.Buffer
	r := summary.New(h, &buf, time.Minute)
	r.Write()

	out := buf.String()
	if !strings.Contains(out, "Total alerts: 3") {
		t.Errorf("expected total 3, got: %s", out)
	}
	if !strings.Contains(out, "Critical:") {
		t.Errorf("expected Critical count, got: %s", out)
	}
	if !strings.Contains(out, "Warning:") {
		t.Errorf("expected Warning count, got: %s", out)
	}
}

func TestWrite_ShowsRecentAlerts(t *testing.T) {
	h := history.New(0)
	for i := 0; i < 7; i++ {
		h.Add(makeAlert(8000+i, "opened", "info"))
	}

	var buf bytes.Buffer
	r := summary.New(h, &buf, time.Minute)
	r.Write()

	out := buf.String()
	// Should show last 5; port 8006 is the last
	if !strings.Contains(out, "port=8006") {
		t.Errorf("expected most recent port in output, got: %s", out)
	}
	// port=8000 and port=8001 should not appear (evicted from recent view)
	if strings.Contains(out, "port=8000") {
		t.Errorf("did not expect oldest port in recent output, got: %s", out)
	}
}

// TestWrite_MultipleWrites ensures the summary output is not cumulative across
// successive Write calls — each call should reflect the current history state.
func TestWrite_MultipleWrites(t *testing.T) {
	h := history.New(0)
	h.Add(makeAlert(80, "opened", "info"))

	var buf bytes.Buffer
	r := summary.New(h, &buf, time.Minute)
	r.Write()

	first := buf.String()
	if !strings.Contains(first, "Total alerts: 1") {
		t.Errorf("first write: expected total 1, got: %s", first)
	}

	buf.Reset()
	h.Add(makeAlert(443, "opened", "warning"))
	r.Write()

	second := buf.String()
	if !strings.Contains(second, "Total alerts: 2") {
		t.Errorf("second write: expected total 2, got: %s", second)
	}
}
