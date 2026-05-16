// Package enricher attaches contextual metadata to alerts before they
// are dispatched through the pipeline. It resolves service names,
// hostnames, and environment labels from a static mapping so that
// downstream consumers receive self-describing payloads.
package enricher

import (
	"fmt"
	"os"
	"sync"

	"github.com/user/portwatch/internal/alert"
)

// Meta holds the extra fields that the Enricher can attach to an alert's
// Tags map.
type Meta struct {
	// ServiceName is a human-readable label for the process expected on the
	// port, e.g. "nginx" or "postgres".
	ServiceName string

	// Environment is an optional deployment environment label such as
	// "production", "staging", or "dev".
	Environment string
}

// Enricher resolves metadata for ports and merges it into alert tags.
type Enricher struct {
	mu       sync.RWMutex
	services map[int]Meta
	hostname string
}

// New creates an Enricher pre-loaded with the provided port-to-Meta mapping.
// A nil map is treated as an empty mapping. The system hostname is resolved
// once at construction time; if resolution fails the hostname field is left
// as an empty string.
func New(services map[int]Meta) *Enricher {
	if services == nil {
		services = make(map[int]Meta)
	}
	host, _ := os.Hostname()
	return &Enricher{
		services: services,
		hostname: host,
	}
}

// Register adds or replaces the Meta entry for the given port.
// It is safe to call concurrently with Enrich.
func (e *Enricher) Register(port int, m Meta) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("enricher: port %d out of range [1, 65535]", port)
	}
	e.mu.Lock()
	e.services[port] = m
	e.mu.Unlock()
	return nil
}

// Enrich copies a into a new alert.Alert and merges any known metadata into
// its Tags field. The original alert is never modified. Tags that already
// exist on the incoming alert take precedence over enriched values so that
// callers can override defaults.
func (e *Enricher) Enrich(a alert.Alert) alert.Alert {
	out := a

	// Shallow-copy the existing tags so we do not mutate the original map.
	merged := make(map[string]string, len(a.Tags)+3)
	for k, v := range a.Tags {
		merged[k] = v
	}

	e.mu.RLock()
	meta, ok := e.services[a.Port]
	e.mu.RUnlock()

	if ok {
		if _, exists := merged["service"]; !exists && meta.ServiceName != "" {
			merged["service"] = meta.ServiceName
		}
		if _, exists := merged["environment"]; !exists && meta.Environment != "" {
			merged["environment"] = meta.Environment
		}
	}

	if _, exists := merged["hostname"]; !exists && e.hostname != "" {
		merged["hostname"] = e.hostname
	}

	out.Tags = merged
	return out
}

// Lookup returns the Meta registered for port and whether it was found.
func (e *Enricher) Lookup(port int) (Meta, bool) {
	e.mu.RLock()
	m, ok := e.services[port]
	e.mu.RUnlock()
	return m, ok
}
