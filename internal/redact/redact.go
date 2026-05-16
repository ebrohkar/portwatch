// Package redact provides utilities for scrubbing sensitive field values
// from alert metadata before they are written to logs or forwarded to
// external notifiers.
package redact

import (
	"strings"
)

const placeholder = "[REDACTED]"

// Redactor scrubs values whose keys match a configured set of sensitive
// field names. Matching is case-insensitive.
type Redactor struct {
	keys map[string]struct{}
}

// New returns a Redactor that will replace the value of any metadata key
// found in sensitiveKeys with the placeholder string "[REDACTED]".
// Duplicate or empty keys are silently ignored.
func New(sensitiveKeys []string) *Redactor {
	keys := make(map[string]struct{}, len(sensitiveKeys))
	for _, k := range sensitiveKeys {
		trimmed := strings.TrimSpace(k)
		if trimmed == "" {
			continue
		}
		keys[strings.ToLower(trimmed)] = struct{}{}
	}
	return &Redactor{keys: keys}
}

// Apply returns a copy of the provided metadata map with sensitive values
// replaced by the placeholder. The original map is never modified.
func (r *Redactor) Apply(meta map[string]string) map[string]string {
	if len(meta) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(meta))
	for k, v := range meta {
		if _, sensitive := r.keys[strings.ToLower(k)]; sensitive {
			out[k] = placeholder
		} else {
			out[k] = v
		}
	}
	return out
}

// IsSensitive reports whether the given key is registered as sensitive.
func (r *Redactor) IsSensitive(key string) bool {
	_, ok := r.keys[strings.ToLower(strings.TrimSpace(key))]
	return ok
}

// Len returns the number of registered sensitive keys.
func (r *Redactor) Len() int {
	return len(r.keys)
}
