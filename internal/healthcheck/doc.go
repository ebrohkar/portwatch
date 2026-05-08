// Package healthcheck exposes a lightweight HTTP endpoint that reports the
// current operational status of the portwatch daemon.
//
// Usage:
//
//	mux := http.NewServeMux()
//	healthcheck.Register(mux, metricsInstance)
//	http.ListenAndServe(":9090", mux)
//
// The /healthz endpoint returns a JSON object with uptime, scan counts,
// alert counts, and error counts sourced from the shared Metrics store.
package healthcheck
