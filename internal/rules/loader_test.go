package rules

import (
	"os"
	"testing"
)

const validYAML = `
rules:
  - port: 80
    description: "HTTP"
    action: alert
    expected: true
  - port: 443
    description: "HTTPS"
    action: alert
    expected: true
  - port: 8080
    description: "Dev server"
    action: ignore
    expected: false
`

const emptyRulesYAML = `
rules: []
`

const invalidActionYAML = `
rules:
  - port: 22
    action: block
    expected: false
`

func TestLoadFromBytes_Valid(t *testing.T) {
	rs, err := LoadFromBytes([]byte(validYAML))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rs == nil {
		t.Fatal("expected non-nil RuleSet")
	}
}

func TestLoadFromBytes_EmptyRules(t *testing.T) {
	_, err := LoadFromBytes([]byte(emptyRulesYAML))
	if err == nil {
		t.Fatal("expected error for empty rules")
	}
}

func TestLoadFromBytes_InvalidAction(t *testing.T) {
	_, err := LoadFromBytes([]byte(invalidActionYAML))
	if err == nil {
		t.Fatal("expected error for invalid action")
	}
}

func TestLoadFromBytes_MalformedYAML(t *testing.T) {
	_, err := LoadFromBytes([]byte(":::not valid yaml:::"))
	if err == nil {
		t.Fatal("expected error for malformed YAML")
	}
}

func TestLoadFromFile_NotFound(t *testing.T) {
	_, err := LoadFromFile("/nonexistent/path/rules.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadFromFile_Valid(t *testing.T) {
	f, err := os.CreateTemp("", "rules-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString(validYAML)
	f.Close()

	rs, err := LoadFromFile(f.Name())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rs == nil {
		t.Fatal("expected non-nil RuleSet")
	}
}
