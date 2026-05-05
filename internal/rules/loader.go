package rules

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the top-level YAML configuration structure.
type Config struct {
	Rules []Rule `yaml:"rules"`
}

// LoadFromFile reads and parses a YAML rules file, returning a validated RuleSet.
func LoadFromFile(path string) (*RuleSet, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading rules file %q: %w", path, err)
	}
	return LoadFromBytes(data)
}

// LoadFromBytes parses YAML rule definitions from a byte slice.
func LoadFromBytes(data []byte) (*RuleSet, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing rules YAML: %w", err)
	}
	if len(cfg.Rules) == 0 {
		return nil, fmt.Errorf("no rules defined in configuration")
	}
	return NewRuleSet(cfg.Rules)
}
