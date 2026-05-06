package filter_test

import (
	"testing"

	"github.com/yourorg/portwatch/internal/filter"
)

func TestNew_EmptyLists(t *testing.T) {
	f, err := filter.New(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f == nil {
		t.Fatal("expected non-nil filter")
	}
}

func TestNew_InvalidIncludePort(t *testing.T) {
	_, err := filter.New([]int{0}, nil)
	if err == nil {
		t.Fatal("expected error for port 0")
	}
}

func TestNew_InvalidExcludePort(t *testing.T) {
	_, err := filter.New(nil, []int{65536})
	if err == nil {
		t.Fatal("expected error for port 65536")
	}
}

func TestAllow_EmptyInclude_AllowsAll(t *testing.T) {
	f, _ := filter.New(nil, nil)
	for _, p := range []int{1, 80, 443, 65535} {
		if !f.Allow(p) {
			t.Errorf("expected port %d to be allowed", p)
		}
	}
}

func TestAllow_ExcludeOverridesAll(t *testing.T) {
	f, _ := filter.New(nil, []int{80, 443})
	if f.Allow(80) {
		t.Error("expected port 80 to be excluded")
	}
	if f.Allow(443) {
		t.Error("expected port 443 to be excluded")
	}
	if !f.Allow(8080) {
		t.Error("expected port 8080 to be allowed")
	}
}

func TestAllow_IncludeList_RestrictsToSet(t *testing.T) {
	f, _ := filter.New([]int{22, 80}, nil)
	if !f.Allow(22) {
		t.Error("expected port 22 to be allowed")
	}
	if f.Allow(443) {
		t.Error("expected port 443 to be denied (not in include list)")
	}
}

func TestAllow_IncludeAndExclude_ExcludeTakesPrecedence(t *testing.T) {
	f, _ := filter.New([]int{22, 80}, []int{80})
	if !f.Allow(22) {
		t.Error("expected port 22 to be allowed")
	}
	if f.Allow(80) {
		t.Error("expected port 80 to be excluded despite being in include list")
	}
}

func TestApply_FiltersSlice(t *testing.T) {
	f, _ := filter.New([]int{22, 80, 443}, []int{80})
	result := f.Apply([]int{22, 80, 443, 8080})
	if len(result) != 2 {
		t.Fatalf("expected 2 ports, got %d: %v", len(result), result)
	}
	if result[0] != 22 || result[1] != 443 {
		t.Errorf("unexpected result: %v", result)
	}
}

func TestApply_EmptyInput(t *testing.T) {
	f, _ := filter.New(nil, nil)
	result := f.Apply([]int{})
	if len(result) != 0 {
		t.Errorf("expected empty result, got %v", result)
	}
}
