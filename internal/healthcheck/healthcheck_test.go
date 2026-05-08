package healthcheck_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/portwatch/internal/healthcheck"
	"github.com/user/portwatch/internal/metrics"
)

func newMetrics() *metrics.Metrics {
	return metrics.New()
}

func TestNew_PanicsOnNilMetrics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil metrics")
		}
	}()
	healthcheck.New(nil)
}

func TestServeHTTP_StatusOK(t *testing.T) {
	m := newMetrics()
	h := healthcheck.New(m)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestServeHTTP_ContentTypeJSON(t *testing.T) {
	h := healthcheck.New(newMetrics())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Fatalf("expected application/json, got %s", ct)
	}
}

func TestServeHTTP_BodyFields(t *testing.T) {
	m := newMetrics()
	m.IncScans()
	m.IncAlerts()
	m.IncErrors()

	h := healthcheck.New(m)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	var resp healthcheck.Response
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Status != "ok" {
		t.Errorf("expected status ok, got %s", resp.Status)
	}
	if resp.Scans != 1 {
		t.Errorf("expected 1 scan, got %d", resp.Scans)
	}
	if resp.Alerts != 1 {
		t.Errorf("expected 1 alert, got %d", resp.Alerts)
	}
	if resp.Errors != 1 {
		t.Errorf("expected 1 error, got %d", resp.Errors)
	}
	if resp.CheckedAt.IsZero() {
		t.Error("expected non-zero checked_at")
	}
}

func TestServeHTTP_UptimeNonEmpty(t *testing.T) {
	m := newMetrics()
	time.Sleep(10 * time.Millisecond)
	h := healthcheck.New(m)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	var resp healthcheck.Response
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Uptime == "" || resp.Uptime == "0s" {
		t.Errorf("expected non-zero uptime, got %q", resp.Uptime)
	}
}

func TestRegister_MountsHandler(t *testing.T) {
	mux := http.NewServeMux()
	healthcheck.Register(mux, newMetrics())

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 from mux, got %d", rec.Code)
	}
}
