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

// RefCycleError reports one or more $ref cycles found while loading a spec. It
// implements error so LoadSpec can fail fast, and exposes the offending refs so
// callers can inspect them. The message names the OpenAPI path of the offending
// $ref (e.g. "#/components/schemas/Node").
type RefCycleError struct {
	Cycles []RefCycle
}

func (e *RefCycleError) Error() string {
	first := e.Cycles[0]
	msg := fmt.Sprintf("parser: circular $ref at %s (cycle: %s)", first.Ref, strings.Join(first.Path, " -> "))
	if len(e.Cycles) > 1 {
		msg += fmt.Sprintf(" (and %d more)", len(e.Cycles)-1)
	}
	return msg
}

// DetectRefCycles walks the schema graph rooted at root. It follows $ref
// references and every structural child (properties, items, prefixItems,
// allOf/oneOf/anyOf, not, additionalProperties) and reports each distinct $ref
// cycle. Callers mark cyclic schemas as terminal (model.SchemaKindRefCycle)
// rather than expanding them forever.
//
// maxDepth bounds how many $ref edges may be followed on a single path: an
// acyclic chain longer than maxDepth returns an error, guarding against
// pathological or unbounded specs (the --max-depth flag). maxDepth <= 0 disables
// that bound. Cycles are still found, since a re-entered $ref terminates the
// walk regardless of depth.
func DetectRefCycles(root *base.SchemaProxy, maxDepth int) ([]RefCycle, error) {
	w := newCycleWalker(maxDepth)
	if err := w.walkProxy("", root); err != nil {
		return nil, err
	}
	return w.cycles, nil
}

// DetectComponentRefCycles runs cycle detection across every component schema,
// seeding each with its own "#/components/schemas/<name>" ref so a component
// that references itself (directly or transitively) is detected even though its
// top-level node is a definition rather than a $ref. The seed does not count
// toward maxDepth, only the $ref edges followed from it do. Detection state is
// shared across components, so each schema's subtree is walked at most once.
func DetectComponentRefCycles(components *v3.Components, maxDepth int) ([]RefCycle, error) {
	if components == nil || components.Schemas == nil {
		return nil, nil
	}
	w := newCycleWalker(maxDepth)
	for name, proxy := range components.Schemas.FromOldest() {
		if err := w.walkProxy(componentSchemaPrefix+name, proxy); err != nil {
			return nil, err
		}
	}
	return w.cycles, nil
}

// cycleWalker is a three-color DFS over the schema graph: refs on stack are
// "gray" (re-entry is a cycle), refs in done are "black" (subtree fully
// explored, safe to prune). depth counts only the $ref edges currently on the
// path, so the maxDepth bound is independent of whether the walk was seeded
// with a component's own ref.
type cycleWalker struct {
	maxDepth int
	depth    int             // count of $ref edges on the current path
	stack    []string        // refs on the current path (for cycle detection)
	onStack  map[string]bool // membership test for stack
	done     map[string]bool // refs whose subtree is fully explored
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

// walkProxy descends into the schemaProxy for cycle detection. The ref argument distinguishes the
// two ways a node is reached:
//
//   - ref != "": schemaProxy is a named component definition and ref is its canonical
//     "#/components/schemas/<name>". It seeds the cycle stack (so the component
//     can be detected referencing itself) but is not a $ref edge, so it does not
//     consume depth budget.
//   - ref == "": schemaProxy is a child node. If it is a $ref, it is followed as a
//     depth-counted edge; otherwise it is an inline schema, which has no
//     identity and cannot start a cycle, so the walkProxy simply descends into it.
func (w *cycleWalker) walkProxy(ref string, schemaProxy *base.SchemaProxy) error {
	if schemaProxy == nil {
		return nil
	}

	edge := false
	if ref == "" {
		if !schemaProxy.IsReference() {
			return w.walkSchema(schemaProxy.Schema())
		}
		ref, edge = schemaProxy.GetReference(), true
	}

	entered, err := w.enter(ref, edge)
	if err != nil || !entered {
		return err
	}
	walkErr := w.walkSchema(schemaProxy.Schema())
	w.leave(ref, walkErr == nil, edge)
	return walkErr
}

// enter handles a node on the path. It reports a cycle (and skips) when ref is
// already on the stack, skips refs whose subtree is already explored, and
// errors if the depth bound would be exceeded. Cycle and done checks take
// precedence over the depth bound so a cyclic spec is reported as a cycle,
// never as a depth error. It returns whether the caller should walk ref's children.
func (w *cycleWalker) enter(ref string, edge bool) (bool, error) {
	if w.onStack[ref] {
		w.recordCycle(ref)
		return false, nil
	}
	if w.done[ref] {
		return false, nil
	}
	if edge && w.maxDepth > 0 && w.depth >= w.maxDepth {
		chain := append(append([]string{}, w.stack...), ref)
		return false, fmt.Errorf("parser: $ref expansion exceeded --max-depth %d at %q (chain: %s)",
			w.maxDepth, ref, strings.Join(chain, " -> "))
	}
	w.stack = append(w.stack, ref)
	w.onStack[ref] = true
	if edge {
		w.depth++
	}
	return true, nil
}

func (w *cycleWalker) leave(ref string, completed, edge bool) {
	if edge {
		w.depth--
	}
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
			if err := w.walkProxy("", p); err != nil {
				return err
			}
		}
	}
	if err := w.walkProxy("", s.Not); err != nil {
		return err
	}
	if s.Properties != nil {
		for _, p := range s.Properties.FromOldest() {
			if err := w.walkProxy("", p); err != nil {
				return err
			}
		}
	}
	if s.Items != nil && s.Items.IsA() {
		if err := w.walkProxy("", s.Items.A); err != nil {
			return err
		}
	}
	if s.AdditionalProperties != nil && s.AdditionalProperties.IsA() {
		if err := w.walkProxy("", s.AdditionalProperties.A); err != nil {
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
