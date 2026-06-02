package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// -------------------------------------------------------------------
//  Unit tests
// -------------------------------------------------------------------
// These will exercise the cycleWalker state machine directly, aka
// operating enter/leave/recordCycle functions without a real OpenAPI Schema

func TestCycleWalkerEnterLeaveMaintainsStackAndDepth(t *testing.T) {
	w := newCycleWalker(0) // depth bound disabled

	entered, err := w.enter("a", true)
	if err != nil || !entered {
		t.Fatalf("enter a: entered=%v err=%v", entered, err)
	}
	if !w.onStack["a"] || len(w.stack) != 1 || w.depth != 1 {
		t.Fatalf("after enter a: onStack=%v stack=%v depth=%d", w.onStack["a"], w.stack, w.depth)
	}

	w.leave("a", true, true)
	if w.onStack["a"] || len(w.stack) != 0 || w.depth != 0 {
		t.Fatalf("after leave a: onStack=%v stack=%v depth=%d", w.onStack["a"], w.stack, w.depth)
	}
	if !w.done["a"] {
		t.Error("leave(completed=true) should mark a done")
	}
}

func TestCycleWalkerLeaveIncompleteDoesNotMarkDone(t *testing.T) {
	w := newCycleWalker(0)
	w.enter("a", true)
	w.leave("a", false, true)

	if w.done["a"] {
		t.Error("leave(completed=false) must not mark done")
	}
	if w.depth != 0 || len(w.stack) != 0 {
		t.Errorf("leave must pop and decrement: depth=%d stack=%v", w.depth, w.stack)
	}
}

func TestCycleWalkerSeedDoesNotConsumeDepth(t *testing.T) {
	w := newCycleWalker(2)

	if _, err := w.enter("seed", false); err != nil { // seed is not a $ref edge
		t.Fatalf("enter seed: %v", err)
	}
	if w.depth != 0 {
		t.Fatalf("seed must not consume depth, got depth=%d", w.depth)
	}
	// Two real edges fit within maxDepth=2 precisely because the seed is free.
	if _, err := w.enter("a", true); err != nil {
		t.Fatalf("enter a: %v", err)
	}
	if _, err := w.enter("b", true); err != nil {
		t.Fatalf("enter b: %v", err)
	}
	if _, err := w.enter("c", true); err == nil {
		t.Fatal("enter c should exceed --max-depth 2")
	}
}

func TestCycleWalkerDepthBoundError(t *testing.T) {
	w := newCycleWalker(1)
	if _, err := w.enter("a", true); err != nil {
		t.Fatalf("enter a: %v", err)
	}

	_, err := w.enter("b", true)
	if err == nil {
		t.Fatal("enter b should exceed --max-depth 1")
	}
	if !strings.Contains(err.Error(), "max-depth") {
		t.Errorf("error should mention max-depth: %v", err)
	}
	if w.onStack["b"] || len(w.stack) != 1 {
		t.Errorf("a failed enter must not push: stack=%v", w.stack)
	}
}

func TestCycleWalkerDepthBoundDisabled(t *testing.T) {
	w := newCycleWalker(0)
	for _, ref := range []string{"a", "b", "c", "d", "e"} {
		if _, err := w.enter(ref, true); err != nil {
			t.Fatalf("enter %s with bound disabled: %v", ref, err)
		}
	}
}

func TestCycleWalkerReentryRecordsCycle(t *testing.T) {
	w := newCycleWalker(0)
	w.enter("a", true)
	w.enter("b", true)

	entered, err := w.enter("a", true)
	if entered || err != nil {
		t.Fatalf("re-entering on-stack a: entered=%v err=%v", entered, err)
	}
	if len(w.cycles) != 1 {
		t.Fatalf("expected 1 cycle, got %d", len(w.cycles))
	}
	if got := w.cycles[0]; got.Ref != "a" || !equal(got.Path, []string{"a", "b", "a"}) {
		t.Errorf("cycle = %+v, want Ref=a Path=[a b a]", got)
	}
}

func TestCycleWalkerCycleClosesMidStack(t *testing.T) {
	w := newCycleWalker(0)
	w.enter("a", true)
	w.enter("b", true)
	w.enter("c", true)

	// c -> b closes a cycle on b, not a; the path must start at b.
	w.enter("b", true)
	if len(w.cycles) != 1 {
		t.Fatalf("expected 1 cycle, got %d", len(w.cycles))
	}
	if got := w.cycles[0]; got.Ref != "b" || !equal(got.Path, []string{"b", "c", "b"}) {
		t.Errorf("cycle = %+v, want Ref=b Path=[b c b]", got)
	}
}

func TestCycleWalkerCycleTakesPrecedenceOverDepth(t *testing.T) {
	w := newCycleWalker(1) // at the bound after a single edge
	w.enter("a", true)     // depth now 1 == maxDepth

	// Re-entering a must be reported as a cycle, not rejected as a depth error.
	entered, err := w.enter("a", true)
	if entered {
		t.Fatal("re-entry must not proceed")
	}
	if err != nil {
		t.Fatalf("re-entry at the depth bound must be a cycle, not an error: %v", err)
	}
	if len(w.cycles) != 1 {
		t.Errorf("re-entry should record a cycle, got %d", len(w.cycles))
	}
}

func TestCycleWalkerDoneRefIsPruned(t *testing.T) {
	w := newCycleWalker(0)
	w.done["x"] = true

	entered, err := w.enter("x", true)
	if entered || err != nil {
		t.Fatalf("entering a done ref: entered=%v err=%v", entered, err)
	}
	if len(w.stack) != 0 || len(w.cycles) != 0 {
		t.Errorf("done ref must be skipped without push or cycle: stack=%v cycles=%v", w.stack, w.cycles)
	}
}

func TestCycleWalkerRecordCycleDedupes(t *testing.T) {
	w := newCycleWalker(0)
	w.enter("a", true)

	w.recordCycle("a")
	w.recordCycle("a")
	if len(w.cycles) != 1 {
		t.Errorf("recordCycle should dedupe by ref, got %d cycles", len(w.cycles))
	}
}

// -------------------------------------------------------------------
//  Fixture tests
// -------------------------------------------------------------------

func loadComponents(t *testing.T, fixture string) *v3.Components {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("../testdata/parser", fixture))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	doc, err := libopenapi.NewDocument(data)
	if err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	model, err := doc.BuildV3Model()
	if err != nil {
		t.Fatalf("build v3 model: %v", err)
	}
	return model.Model.Components
}

func componentProxy(t *testing.T, comps *v3.Components, name string) *base.SchemaProxy {
	t.Helper()
	for k, v := range comps.Schemas.FromOldest() {
		if k == name {
			return v
		}
	}
	t.Fatalf("component %q not found", name)
	return nil
}

func TestDetectComponentRefCyclesSelfReference(t *testing.T) {
	cycles, err := DetectComponentRefCycles(loadComponents(t, "cycle_self.yaml"), 16)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 1 {
		t.Fatalf("got %d cycles, want 1: %+v", len(cycles), cycles)
	}
	ref := componentSchemaPrefix + "Node"
	if cycles[0].Ref != ref {
		t.Errorf("cycle ref = %q, want %q", cycles[0].Ref, ref)
	}
	if want := []string{ref, ref}; !equal(cycles[0].Path, want) {
		t.Errorf("cycle path = %v, want %v", cycles[0].Path, want)
	}
}

func TestDetectComponentRefCyclesIndirect(t *testing.T) {
	cycles, err := DetectComponentRefCycles(loadComponents(t, "cycle_indirect.yaml"), 16)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 1 {
		t.Fatalf("got %d cycles, want 1: %+v", len(cycles), cycles)
	}
	// Components iterate in document order (A before B), so the cycle is found
	// from A: A -> B -> A.
	want := []string{componentSchemaPrefix + "A", componentSchemaPrefix + "B", componentSchemaPrefix + "A"}
	if !equal(cycles[0].Path, want) {
		t.Errorf("cycle path = %v, want %v", cycles[0].Path, want)
	}
}

func TestDetectComponentRefCyclesArrayItems(t *testing.T) {
	cycles, err := DetectComponentRefCycles(loadComponents(t, "cycle_array_items.yaml"), 16)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 1 || cycles[0].Ref != componentSchemaPrefix+"Tree" {
		t.Fatalf("got %+v, want a single Tree cycle", cycles)
	}
}

func TestDetectComponentRefCyclesAcyclic(t *testing.T) {
	cycles, err := DetectComponentRefCycles(loadComponents(t, "acyclic_diamond.yaml"), 16)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("got %d cycles, want 0: %+v", len(cycles), cycles)
	}
}

func TestDetectComponentRefCyclesMaxDepthDisabledStillFindsCycle(t *testing.T) {
	cycles, err := DetectComponentRefCycles(loadComponents(t, "cycle_self.yaml"), 0)
	if err != nil {
		t.Fatalf("unexpected error with depth bound disabled: %v", err)
	}
	if len(cycles) != 1 {
		t.Errorf("got %d cycles, want 1", len(cycles))
	}
}

func TestDetectRefCyclesMaxDepth(t *testing.T) {
	comps := loadComponents(t, "deep_chain.yaml")
	root := componentProxy(t, comps, "S0") // S0 -> S1 -> ... -> S8 (8 refs deep)

	if _, err := DetectRefCycles(root, 4); err == nil {
		t.Error("expected an error when the chain exceeds --max-depth 4, got nil")
	}

	cycles, err := DetectRefCycles(root, 8)
	if err != nil {
		t.Errorf("depth 8 should accommodate the chain, got error: %v", err)
	}
	if len(cycles) != 0 {
		t.Errorf("deep chain is acyclic, got %d cycles", len(cycles))
	}

	if _, err := DetectRefCycles(root, 0); err != nil {
		t.Errorf("depth bound disabled should never error on an acyclic chain, got: %v", err)
	}
}

func TestDetectComponentRefCyclesViaAllOf(t *testing.T) {
	cycles, err := DetectComponentRefCycles(loadComponents(t, "cycle_allof.yaml"), 16)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 1 || cycles[0].Ref != componentSchemaPrefix+"Loop" {
		t.Fatalf("got %+v, want a single Loop cycle (allOf edge)", cycles)
	}
}

func TestDetectComponentRefCyclesViaAdditionalProperties(t *testing.T) {
	cycles, err := DetectComponentRefCycles(loadComponents(t, "cycle_additional_properties.yaml"), 16)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cycles) != 1 || cycles[0].Ref != componentSchemaPrefix+"Dict" {
		t.Fatalf("got %+v, want a single Dict cycle (additionalProperties edge)", cycles)
	}
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
