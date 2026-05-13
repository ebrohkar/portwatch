// Package rotation implements size-based log file rotation for portwatch
// output streams such as the audit log and notifier sinks.
//
// A rotation.Writer wraps a file path and transparently rotates the underlying
// file once it reaches a configurable byte limit, keeping up to MaxBackups
// numbered copies (e.g. portwatch.log.1, portwatch.log.2, …).
package rotation
