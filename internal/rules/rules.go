package rules

import (
	"fmt"
	"slices"
)

// Action defines what to do when a rule matches.
type Action string

const (
	ActionAlert  Action = "alert"
	ActionIgnore Action = "ignore"
)

// Rule defines a condition for a specific port or port range.
type Rule struct {
	Port        int    `yaml:"port"`
	Description string `yaml:"description"`
	Action      Action `yaml:"action"`
	Expected    bool   `yaml:"expected"`
}

// RuleSet holds a collection of rules and evaluates port states.
type RuleSet struct {
	rules []Rule
}

// NewRuleSet creates a RuleSet from a slice of rules.
func NewRuleSet(rules []Rule) (*RuleSet, error) {
	for _, r := range rules {
		if r.Port < 1 || r.Port > 65535 {
			return nil, fmt.Errorf("invalid port %d in rule: must be 1-65535", r.Port)
		}
		if r.Action != ActionAlert && r.Action != ActionIgnore {
			return nil, fmt.Errorf("invalid action %q for port %d", r.Action, r.Port)
		}
	}
	return &RuleSet{rules: rules}, nil
}

// Evaluate checks a list of open ports against the rule set and returns violations.
func (rs *RuleSet) Evaluate(openPorts []int) []Violation {
	var violations []Violation

	for _, rule := range rs.rules {
		if rule.Action == ActionIgnore {
			continue
		}
		isOpen := slices.Contains(openPorts, rule.Port)
		if rule.Expected && !isOpen {
			violations = append(violations, Violation{
				Port:    rule.Port,
				Message: fmt.Sprintf("expected port %d to be open, but it is closed", rule.Port),
				Rule:    rule,
			})
		} else if !rule.Expected && isOpen {
			violations = append(violations, Violation{
				Port:    rule.Port,
				Message: fmt.Sprintf("unexpected port %d is open", rule.Port),
				Rule:    rule,
			})
		}
	}
	return violations
}

// Violation represents a rule breach.
type Violation struct {
	Port    int
	Message string
	Rule    Rule
}
