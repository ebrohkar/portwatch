// Package config provides loading and validation of portwatch runtime
// configuration from YAML files or raw bytes.
//
// A Config defines the scan interval, port range, state file path, optional
// log file, and rules file path. Sensible defaults are applied for any fields
// that are omitted from the configuration source.
package config
