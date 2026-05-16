package redact

import (
	"testing"
)

func TestNew_EmptySensitiveKeys(t *testing.T) {
	r := New(nil)
	if r.Len() != 0 {
		t.Fatalf("expected 0 keys, got %d", r.Len())
	}
}

func TestNew_DeduplicatesKeys(t *testing.T) {
	r := New([]string{"token", "Token", "TOKEN"})
	if r.Len() != 1 {
		t.Fatalf("expected 1 key after dedup, got %d", r.Len())
	}
}

func TestNew_IgnoresBlankKeys(t *testing.T) {
	r := New([]string{"", "  ", "secret"})
	if r.Len() != 1 {
		t.Fatalf("expected 1 key, got %d", r.Len())
	}
}

func TestIsSensitive_CaseInsensitive(t *testing.T) {
	r := New([]string{"Authorization"})
	for _, k := range []string{"authorization", "AUTHORIZATION", "Authorization"} {
		if !r.IsSensitive(k) {
			t.Errorf("expected %q to be sensitive", k)
		}
	}
}

func TestIsSensitive_UnknownKey(t *testing.T) {
	r := New([]string{"token"})
	if r.IsSensitive("host") {
		t.Fatal("expected 'host' to not be sensitive")
	}
}

func TestApply_RedactsSensitiveValues(t *testing.T) {
	r := New([]string{"password", "token"})
	input := map[string]string{
		"host":     "localhost",
		"password": "s3cr3t",
		"token":    "abc123",
	}
	out := r.Apply(input)
	if out["host"] != "localhost" {
		t.Errorf("expected host to be unchanged, got %q", out["host"])
	}
	if out["password"] != placeholder {
		t.Errorf("expected password to be redacted, got %q", out["password"])
	}
	if out["token"] != placeholder {
		t.Errorf("expected token to be redacted, got %q", out["token"])
	}
}

func TestApply_DoesNotMutateOriginal(t *testing.T) {
	r := New([]string{"secret"})
	orig := map[string]string{"secret": "mysecret", "info": "public"}
	_ = r.Apply(orig)
	if orig["secret"] != "mysecret" {
		t.Fatal("Apply must not modify the original map")
	}
}

func TestApply_EmptyMeta(t *testing.T) {
	r := New([]string{"token"})
	out := r.Apply(map[string]string{})
	if len(out) != 0 {
		t.Fatalf("expected empty output, got %v", out)
	}
}

func TestApply_NilMeta(t *testing.T) {
	r := New([]string{"token"})
	out := r.Apply(nil)
	if out == nil || len(out) != 0 {
		t.Fatal("expected non-nil empty map for nil input")
	}
}
