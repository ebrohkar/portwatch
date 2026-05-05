package rules

import (
	"testing"
)

func TestNewRuleSet_Valid(t *testing.T) {
	rules := []Rule{
		{Port: 80, Action: ActionAlert, Expected: true},
		{Port: 8080, Action: ActionIgnore, Expected: false},
	}
	rs, err := NewRuleSet(rules)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rs == nil {
		t.Fatal("expected non-nil RuleSet")
	}
}

func TestNewRuleSet_InvalidPort(t *testing.T) {
	rules := []Rule{
		{Port: 0, Action: ActionAlert, Expected: true},
	}
	_, err := NewRuleSet(rules)
	if err == nil {
		t.Fatal("expected error for invalid port 0")
	}
}

func TestNewRuleSet_InvalidAction(t *testing.T) {
	rules := []Rule{
		{Port: 443, Action: "block", Expected: true},
	}
	_, err := NewRuleSet(rules)
	if err == nil {
		t.Fatal("expected error for invalid action")
	}
}

func TestEvaluate_UnexpectedOpenPort(t *testing.T) {
	rs, _ := NewRuleSet([]Rule{
		{Port: 9999, Action: ActionAlert, Expected: false},
	})
	violations := rs.Evaluate([]int{9999})
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if violations[0].Port != 9999 {
		t.Errorf("expected violation on port 9999, got %d", violations[0].Port)
	}
}

func TestEvaluate_ExpectedPortClosed(t *testing.T) {
	rs, _ := NewRuleSet([]Rule{
		{Port: 80, Action: ActionAlert, Expected: true},
	})
	violations := rs.Evaluate([]int{})
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
}

func TestEvaluate_IgnoredPort(t *testing.T) {
	rs, _ := NewRuleSet([]Rule{
		{Port: 8080, Action: ActionIgnore, Expected: false},
	})
	violations := rs.Evaluate([]int{8080})
	if len(violations) != 0 {
		t.Fatalf("expected 0 violations for ignored port, got %d", len(violations))
	}
}

func TestEvaluate_NoViolations(t *testing.T) {
	rs, _ := NewRuleSet([]Rule{
		{Port: 443, Action: ActionAlert, Expected: true},
	})
	violations := rs.Evaluate([]int{443})
	if len(violations) != 0 {
		t.Fatalf("expected 0 violations, got %d", len(violations))
	}
}
