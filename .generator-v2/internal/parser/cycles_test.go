package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

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
