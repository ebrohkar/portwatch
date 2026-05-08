// Package healthcheck provides a simple HTTP health endpoint for portwatch.
// It exposes a /healthz route that returns daemon status and basic metrics.
package healthcheck

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/user/portwatch/internal/metrics"
)

// Response is the JSON body returned by the health endpoint.
type Response struct {
	Status    string    `json:"status"`
	Uptime    string    `json:"uptime"`
	Scans     int64     `json:"scans"`
	Alerts    int64     `json:"alerts"`
	Errors    int64     `json:"errors"`
	CheckedAt time.Time `json:"checked_at"`
}

// Handler holds dependencies for the health check HTTP handler.
type Handler struct {
	metrics *metrics.Metrics
}

// New returns a new Handler backed by the provided Metrics instance.
func New(m *metrics.Metrics) *Handler {
	if m == nil {
		panic("healthcheck: metrics must not be nil")
	}
	return &Handler{metrics: m}
}

// ServeHTTP handles GET /healthz requests.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	snap := h.metrics.Snapshot()
	resp := Response{
		Status:    "ok",
		Uptime:    snap.Uptime.Round(time.Second).String(),
		Scans:     snap.Scans,
		Alerts:    snap.Alerts,
		Errors:    snap.Errors,
		CheckedAt: time.Now().UTC(),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// Register mounts the health handler on the given mux at /healthz.
func Register(mux *http.ServeMux, m *metrics.Metrics) {
	mux.Handle("/healthz", New(m))
}
