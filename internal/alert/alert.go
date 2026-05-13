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

// String returns a human-readable summary of the alert.
func (a Alert) String() string {
	return a.Timestamp.Format(time.RFC3339) + " [" + string(a.Severity) + "] " +
		a.Protocol + " port " + itoa(a.Port) + " " + a.Event + ": " + a.Message
}

// itoa converts an integer to its decimal string representation without
// importing strconv at the package level.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := [20]byte{}
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[pos:])
}
