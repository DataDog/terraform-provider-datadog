package parser

import (
	"fmt"
	"slices"
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

// MaxDepthError reports that $ref expansion hit the maxDepth bound before a
// path terminated: Ref would have pushed past the limit, Chain is the $ref path
// leading to it. A cycle never produces this error — it terminates the walk
// instead of failing it — so this always means "too deep", never "circular".
type MaxDepthError struct {
	Ref      string
	Chain    []string
	MaxDepth int
}

func (e *MaxDepthError) Error() string {
	return fmt.Sprintf("parser: $ref expansion exceeded --max-depth %d at %q (chain: %s)",
		e.MaxDepth, e.Ref, strings.Join(e.Chain, " -> "))
}

// DetectRefCycles walks the schema graph from root, following $refs and every
// structural child (properties, items, prefixItems, allOf/oneOf/anyOf, not,
// additionalProperties), and returns each distinct $ref cycle. maxDepth bounds
// the $ref edges on one path — a longer acyclic chain is a *MaxDepthError, and
// maxDepth <= 0 disables the bound — but cycles are found at any depth.
func DetectRefCycles(root *base.SchemaProxy, maxDepth int) ([]RefCycle, error) {
	w := newCycleWalker(maxDepth)
	if err := w.walkProxy("", root); err != nil {
		return nil, err
	}
	return w.cycles, nil
}

// DetectComponentRefCycles runs cycle detection over every component schema,
// seeding each with its own "#/components/schemas/<name>" so a component that
// references itself is found even though its top node is a definition, not a
// $ref. The seed costs no depth. Walker state is shared across components, so
// each subtree is walked at most once.
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

// cycleWalker is a three-color DFS: refs on stack are gray (re-entry closes a
// cycle), refs in done are black (subtree fully explored, safe to prune). depth
// counts only the $ref edges on the current path, so the bound is unaffected by
// a seed ref.
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

// walkProxy descends into schemaProxy. A non-empty ref means schemaProxy is the
// named component definition for that ref: it seeds the cycle stack but is not
// a $ref edge and costs no depth. An empty ref means a child node — followed as
// a depth-counted edge when it is a $ref, otherwise descended into inline,
// since an inline schema has no identity and cannot start a cycle.
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

// enter pushes ref onto the path and reports whether to walk its children. It
// records a cycle and skips when ref is already on the stack, skips a ref whose
// subtree is explored, and errors on the depth bound. Cycle and done checks
// come first, so a cyclic spec is reported as a cycle, never as a depth error.
func (w *cycleWalker) enter(ref string, edge bool) (bool, error) {
	if w.onStack[ref] {
		w.recordCycle(ref)
		return false, nil
	}
	if w.done[ref] {
		return false, nil
	}
	if edge && w.maxDepth > 0 && w.depth >= w.maxDepth {
		return false, &MaxDepthError{
			Ref:      ref,
			Chain:    append(append([]string{}, w.stack...), ref),
			MaxDepth: w.maxDepth,
		}
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

	start := slices.Index(w.stack, ref)
	if start < 0 {
		start = 0
	}
	path := append([]string{}, w.stack[start:]...)
	path = append(path, ref)
	w.cycles = append(w.cycles, RefCycle{Ref: ref, Path: path})
}
