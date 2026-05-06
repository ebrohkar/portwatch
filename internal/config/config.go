// Package config handles loading and validation of portwatch configuration.
package config

import (
	"errors"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the top-level portwatch configuration.
type Config struct {
	ScanInterval time.Duration `yaml:"scan_interval"`
	PortRange    PortRange     `yaml:"port_range"`
	StateFile    string        `yaml:"state_file"`
	LogFile      string        `yaml:"log_file"`
	RulesFile    string        `yaml:"rules_file"`
}

// PortRange defines the inclusive start and end ports to scan.
type PortRange struct {
	Start int `yaml:"start"`
	End   int `yaml:"end"`
}

// defaults applies sensible default values to any zero-value fields.
func (c *Config) defaults() {
	if c.ScanInterval == 0 {
		c.ScanInterval = 60 * time.Second
	}
	if c.PortRange.Start == 0 {
		c.PortRange.Start = 1
	}
	if c.PortRange.End == 0 {
		c.PortRange.End = 65535
	}
	if c.StateFile == "" {
		c.StateFile = "/var/lib/portwatch/state.json"
	}
}

// Validate returns an error if the configuration is invalid.
func (c *Config) Validate() error {
	if c.ScanInterval < time.Second {
		return errors.New("scan_interval must be at least 1s")
	}
	if c.PortRange.Start < 1 || c.PortRange.Start > 65535 {
		return errors.New("port_range.start must be between 1 and 65535")
	}
	if c.PortRange.End < 1 || c.PortRange.End > 65535 {
		return errors.New("port_range.end must be between 1 and 65535")
	}
	if c.PortRange.Start > c.PortRange.End {
		return errors.New("port_range.start must be less than or equal to port_range.end")
	}
	return nil
}

// LoadFromFile reads and parses a YAML config file at the given path.
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadFromBytes(data)
}

// LoadFromBytes parses YAML config bytes and returns a validated Config.
func LoadFromBytes(data []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	cfg.defaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}
