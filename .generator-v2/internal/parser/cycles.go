package parser

import (
	"fmt"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

const componentSchemaPrefix = "#/components/schemas/"

// RefCycle describes a $ref that re-enters a schema already being expanded.
// Path is the chain of $ref targets that closes the loop; its first and last
// elements are the repeated ref.
type RefCycle struct {
	Ref  string
	Path []string
}

// DetectRefCycles walks the schema graph rooted at root. It follows $ref
// references and every structural child (properties, items, prefixItems,
// allOf/oneOf/anyOf, not, additionalProperties) and reports each distinct
// $ref cycle. Callers mark cyclic schemas as terminal (model.SchemaKindRefCycle)
// rather than expanding them forever.
//
// maxDepth bounds how deeply $refs may nest: an acyclic $ref chain longer than
// maxDepth returns an error, guarding against pathological or unbounded specs
// (the --max-depth flag). maxDepth <= 0 disables that bound. Cycles are still
// found, since a re-entered $ref terminates the walk regardless of depth.
func DetectRefCycles(root *base.SchemaProxy, maxDepth int) ([]RefCycle, error) {
	w := newCycleWalker(maxDepth)
	if err := w.walkProxy(root); err != nil {
		return nil, err
	}
	return w.cycles, nil
}

// DetectComponentRefCycles runs cycle detection across every component schema,
// seeding each with its own "#/components/schemas/<name>" ref so a component
// that references itself (directly or transitively) is detected even though its
// top-level node is a definition rather than a $ref. Detection state is shared
// across components, so each schema's subtree is walked at most once.
func DetectComponentRefCycles(components *v3.Components, maxDepth int) ([]RefCycle, error) {
	if components == nil || components.Schemas == nil {
		return nil, nil
	}
	w := newCycleWalker(maxDepth)
	for name, proxy := range components.Schemas.FromOldest() {
		if err := w.walkNamed(componentSchemaPrefix+name, proxy); err != nil {
			return nil, err
		}
	}
	return w.cycles, nil
}

// cycleWalker is a three-color DFS over the schema graph: refs on stack are
// "gray" (re-entry is a cycle), refs in done are "black" (subtree fully
// explored, safe to prune).
type cycleWalker struct {
	maxDepth int
	stack    []string        // $refs currently being expanded (the DFS path)
	onStack  map[string]bool // membership test for stack
	done     map[string]bool // $refs whose subtree is fully explored
	reported map[string]bool // closing refs already recorded, to dedupe
	cycles   []RefCycle
}

func newCycleWalker(maxDepth int) *cycleWalker {
	return &cycleWalker{
		maxDepth: maxDepth,
		onStack:  map[string]bool{},
		done:     map[string]bool{},
		reported: map[string]bool{},
	}
}

// walkNamed expands a named component's schema content as if it were reached
// through its canonical ref.
func (w *cycleWalker) walkNamed(ref string, proxy *base.SchemaProxy) error {
	if proxy == nil {
		return nil
	}
	entered, err := w.enter(ref)
	if err != nil || !entered {
		return err
	}
	walkErr := w.walkSchema(proxy.Schema())
	w.leave(ref, walkErr == nil)
	return walkErr
}

func (w *cycleWalker) walkProxy(p *base.SchemaProxy) error {
	if p == nil {
		return nil
	}
	if !p.IsReference() {
		return w.walkSchema(p.Schema())
	}
	ref := p.GetReference()
	entered, err := w.enter(ref)
	if err != nil || !entered {
		return err
	}
	walkErr := w.walkSchema(p.Schema())
	w.leave(ref, walkErr == nil)
	return walkErr
}

// enter pushes ref onto the path unless it closes a cycle (recorded, then
// skipped), has already been fully explored (skipped), or would exceed maxDepth
// (error). It reports whether the caller should walk ref's children.
func (w *cycleWalker) enter(ref string) (bool, error) {
	switch {
	case w.onStack[ref]:
		w.recordCycle(ref)
		return false, nil
	case w.done[ref]:
		return false, nil
	case w.maxDepth > 0 && len(w.stack) >= w.maxDepth:
		chain := append(append([]string{}, w.stack...), ref)
		return false, fmt.Errorf("parser: $ref expansion exceeded --max-depth %d at %q (chain: %s)",
			w.maxDepth, ref, strings.Join(chain, " -> "))
	}
	w.stack = append(w.stack, ref)
	w.onStack[ref] = true
	return true, nil
}

func (w *cycleWalker) leave(ref string, completed bool) {
	w.onStack[ref] = false
	w.stack = w.stack[:len(w.stack)-1]
	if completed {
		w.done[ref] = true
	}
}

func (w *cycleWalker) walkSchema(s *base.Schema) error {
	if s == nil {
		return nil
	}
	for _, group := range [][]*base.SchemaProxy{s.AllOf, s.OneOf, s.AnyOf, s.PrefixItems} {
		for _, p := range group {
			if err := w.walkProxy(p); err != nil {
				return err
			}
		}
	}
	if err := w.walkProxy(s.Not); err != nil {
		return err
	}
	if s.Properties != nil {
		for _, p := range s.Properties.FromOldest() {
			if err := w.walkProxy(p); err != nil {
				return err
			}
		}
	}
	if s.Items != nil && s.Items.IsA() {
		if err := w.walkProxy(s.Items.A); err != nil {
			return err
		}
	}
	if s.AdditionalProperties != nil && s.AdditionalProperties.IsA() {
		if err := w.walkProxy(s.AdditionalProperties.A); err != nil {
			return err
		}
	}
	return nil
}

func (w *cycleWalker) recordCycle(ref string) {
	if w.reported[ref] {
		return
	}
	w.reported[ref] = true

	start := 0
	for i, r := range w.stack {
		if r == ref {
			start = i
			break
		}
	}
	path := append([]string{}, w.stack[start:]...)
	path = append(path, ref)
	w.cycles = append(w.cycles, RefCycle{Ref: ref, Path: path})
}
