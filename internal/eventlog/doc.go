// Package eventlog provides a structured, append-only event log for
// portwatch. Each port-change event is serialised as a newline-delimited
// JSON object and written to a configurable io.Writer (default: stdout).
//
// Typical usage:
//
//	f, _ := os.OpenFile("events.jsonl", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
//	log := eventlog.New(f)
//	log.Record(eventlog.Entry{Port: 8080, Event: "opened", Severity: "warn"})
package eventlog
