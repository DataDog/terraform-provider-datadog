package parser

import (
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
