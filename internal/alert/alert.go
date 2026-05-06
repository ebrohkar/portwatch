// Package alert defines the Alert type and severity levels used across portwatch.
package alert

import "time"

// Severity represents the urgency of an alert.
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

// Alert represents a single detected port change event.
type Alert struct {
	Timestamp time.Time `json:"timestamp"`
	Port      int       `json:"port"`
	Protocol  string    `json:"protocol"`
	Event     string    `json:"event"` // "opened" or "closed"
	Severity  Severity  `json:"severity"`
	Message   string    `json:"message"`
}

// New constructs an Alert with the current timestamp.
func New(port int, protocol, event string, severity Severity, message string) Alert {
	return Alert{
		Timestamp: time.Now().UTC(),
		Port:      port,
		Protocol:  protocol,
		Event:     event,
		Severity:  severity,
		Message:   message,
	}
}

// IsValid returns true when the alert contains the minimum required fields.
func (a Alert) IsValid() bool {
	return a.Port > 0 && a.Port <= 65535 &&
		a.Protocol != "" &&
		(a.Event == "opened" || a.Event == "closed") &&
		(a.Severity == SeverityInfo || a.Severity == SeverityWarning || a.Severity == SeverityCritical)
}
