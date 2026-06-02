package parser

import (
	"errors"
	"path/filepath"
	"slices"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// scrambledSpecPath is a spec whose paths and methods are declared out of
// (path, method) order on purpose (paths zebra/alpha/mid; /alpha lists
// post/get/delete), so these tests fail if LoadSpec ever stops sorting.
var scrambledSpecPath = filepath.Join("../testdata/parser", "scrambled_ordering.yaml")

// operationOrder renders the operation sequence as "METHOD path" tokens so
// orderings are easy to compare and to read in failure output.
func operationOrder(s *model.Spec) []string {
	out := make([]string, len(s.Operations))
	for i, op := range s.Operations {
		out[i] = op.Method + " " + op.Path
	}
	return out
}

// TestLoadSpecSortsByPathThenMethod pins the documented ordering: operations
// come out sorted by path, then by method, regardless of source order.
func TestLoadSpecSortsByPathThenMethod(t *testing.T) {
	spec, err := LoadSpec(scrambledSpecPath)
	if err != nil {
		t.Fatalf("LoadSpec: %v", err)
	}

	want := []string{
		"DELETE /alpha",
		"GET /alpha",
		"POST /alpha",
		"PUT /mid",
		"GET /zebra",
	}
	if got := operationOrder(spec); !slices.Equal(got, want) {
		t.Errorf("operation order = %v, want %v", got, want)
	}
}

// TestLoadSpecOrderingIsDeterministic loads the same spec many times and
// asserts the ordering never changes. Go randomizes map iteration per range,
// so if a future change iterated a plain map instead of the ordered structures
// (or dropped the explicit sort), at least one of these runs would diverge.
func TestLoadSpecOrderingIsDeterministic(t *testing.T) {
	first, err := LoadSpec(scrambledSpecPath)
	if err != nil {
		t.Fatalf("LoadSpec: %v", err)
	}
	baseline := operationOrder(first)
	if len(baseline) == 0 {
		t.Fatal("expected operations, got none")
	}

	const runs = 5
	for i := range runs {
		got, err := LoadSpec(scrambledSpecPath)
		if err != nil {
			t.Fatalf("LoadSpec run %d: %v", i, err)
		}
		if order := operationOrder(got); !slices.Equal(order, baseline) {
			t.Fatalf("run %d diverged:\n baseline = %v\n got      = %v", i, baseline, order)
		}
	}
}

// TestLoadSpecCircularRefReturnsTypedError covers the ticket AC: a deliberately
// circular $ref makes LoadSpec fail with a typed *RefCycleError that names the
// offending $ref path.
func TestLoadSpecCircularRefReturnsTypedError(t *testing.T) {
	_, err := LoadSpec(filepath.Join("../testdata/parser", "cycle_self.yaml"))
	if err == nil {
		t.Fatal("expected an error for a circular $ref, got nil")
	}
	var cycleErr *RefCycleError
	if !errors.As(err, &cycleErr) {
		t.Fatalf("error %v (%T) is not a *RefCycleError", err, err)
	}
	if len(cycleErr.Cycles) == 0 {
		t.Fatal("RefCycleError carries no cycles")
	}
	if want := "#/components/schemas/Node"; cycleErr.Cycles[0].Ref != want {
		t.Errorf("offending ref = %q, want %q", cycleErr.Cycles[0].Ref, want)
	}
}

// TestLoadSpecIndirectCircularRefReturnsTypedError checks an A -> B -> A cycle
// also surfaces as the typed error.
func TestLoadSpecIndirectCircularRefReturnsTypedError(t *testing.T) {
	_, err := LoadSpec(filepath.Join("../testdata/parser", "cycle_indirect.yaml"))
	var cycleErr *RefCycleError
	if !errors.As(err, &cycleErr) {
		t.Fatalf("error %v (%T) is not a *RefCycleError", err, err)
	}
}

// TestLoadSpecMaxDepthFailsFastNotAsCycle covers the ticket AC: --max-depth caps
// recursion, and a deep-but-acyclic chain is never misclassified as a cycle.
func TestLoadSpecMaxDepthFailsFastNotAsCycle(t *testing.T) {
	deep := filepath.Join("../testdata/parser", "deep_chain.yaml")

	// The chain is 8 $refs deep; the default bound (8) and an explicit 8 both
	// accommodate it without error.
	if _, err := LoadSpec(deep); err != nil {
		t.Errorf("default --max-depth should accommodate the 8-deep chain, got: %v", err)
	}
	if _, err := LoadSpec(deep, WithMaxDepth(8)); err != nil {
		t.Errorf("--max-depth 8 should accommodate the 8-deep chain, got: %v", err)
	}

	// A tighter bound fails fast — but as a depth error, not a cycle.
	_, err := LoadSpec(deep, WithMaxDepth(4))
	if err == nil {
		t.Fatal("expected a depth error at --max-depth 4, got nil")
	}
	var cycleErr *RefCycleError
	if errors.As(err, &cycleErr) {
		t.Errorf("deep-but-acyclic refs must not be reported as a cycle: %v", err)
	}
}
