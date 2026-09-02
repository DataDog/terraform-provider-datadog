// Package sdkbind resolves the Datadog go-sdk bindings for every OpenAPI oneOf
// reachable from a normalized operation: the generated wrapper struct, and for
// each alternative the wrapper member that selects it plus the convenience
// constructor that builds it.
//
// It runs after parser normalization and before the Terraform projection, and it
// writes only into OneOfSpec.SDKType and OneOfVariant.SDKField/SDKConstructor/
// SDKPointer. The fields model.BuildResponseTree carries through onto each
// OneOfEnvelope.
//
// Method of record : every name here is **re-derived** by reimplementing the go-sdk
// generator's own logic over the same OpenAPI input. gotype.go holds the naming rules;
// this file holds the walk that decides which generated model a union lives in.
//
// The walk exists because a wrapper's identity is positional, not local. The SDK
// generator's child_models() names a model after its $ref component when it has
// one, and otherwise after the accumulated path from the enclosing named model,
// parent name plus camel_case(property), plus "Item" per array step. A Terraform
// envelope name must never stand in for that: an inline union's envelope name is
// path-derived for generated-model stability and names no SDK struct.
package sdkbind

import (
	"fmt"
	"sort"
	"strings"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// BindOperation resolves the SDK oneOf bindings for every union reachable from
// op's request and response schemas, mutating the normalized schemas in place. It
// is idempotent: re-running it over an already-bound operation recomputes the same
// values.
//
// Failures are collected rather than returned on first sight, and the returned
// *UnresolvedBindingError names the artifact, the operation, and every union path
// and alternative it could not bind. Because binding is per operation, one
// artifact's unresolvable union cannot stop another artifact from generating.
//
// A position whose SDK model name cannot be determined is walked *past*, not
// abandoned: only a union actually standing at such a position fails, so an
// unnameable branch holding no union costs nothing.
func BindOperation(op *model.Operation) error {
	if op == nil {
		return nil
	}
	b := &binder{visited: make(map[visitKey]bool)}
	// Each body root is the SDK type the operation's method signature names, which
	// is where the SDK generator's own naming starts too.
	b.walk(op.RequestSchema, op.RequestRefName, "request")
	b.walk(op.ResponseSchema, op.ResponseRefName, "response")

	if len(b.failures) == 0 {
		return nil
	}
	return &UnresolvedBindingError{
		Artifact:  artifactName(op),
		Operation: op.OperationId,
		Failures:  b.failures,
	}
}

// BindSpec binds every operation in spec, returning one error per operation that
// had an unresolvable union, keyed by operation. It exists for callers that want
// to bind the whole spec up front; the generate path binds per artifact so a
// failure lands on that artifact's report entry.
func BindSpec(spec *model.Spec) map[*model.Operation]error {
	if spec == nil {
		return nil
	}
	var failed map[*model.Operation]error
	for _, op := range spec.Operations {
		if err := BindOperation(op); err != nil {
			if failed == nil {
				failed = make(map[*model.Operation]error)
			}
			failed[op] = err
		}
	}
	return failed
}

type binder struct {
	failures []Failure
	// visited guards termination over the normalized graph, which is a DAG rather
	// than a tree: mergeOneOfSiblings shares one sibling-constraint node across
	// every alternative it was merged into. Keying on (node, SDK name) rather than
	// on the node alone keeps a shared node bound correctly when two positions give
	// it two different SDK names, instead of letting whichever position was walked
	// first decide for both.
	visited map[visitKey]bool
}

type visitKey struct {
	schema  *model.Schema
	sdkName string
}

// walk descends s, carrying the name of the generated SDK model this node lives
// in. sdkName is empty at a position whose model the rules cannot name; the walk
// continues so that only a union standing there fails.
//
// This mirrors openapi.child_models: a $ref restarts the name, an object property
// appends camel_case(key), and an array element appends "Item".
func (b *binder) walk(s *model.Schema, sdkName, path string) {
	if s == nil {
		return
	}
	// get_name(schema) or alternative_name: a component name always wins over the
	// name accumulated from the parent.
	if s.RefName != "" {
		sdkName = s.RefName
	}

	key := visitKey{schema: s, sdkName: sdkName}
	if b.visited[key] {
		return
	}
	b.visited[key] = true

	switch s.Kind {
	case model.SchemaKindOneOf:
		b.bindUnion(s.OneOf, sdkName, path)

	case model.SchemaKindObject:
		for _, name := range sortedKeys(s.Properties) {
			b.walk(s.Properties[name], childModelName(sdkName, model.SdkName(name)), model.ChildPath(path, name))
		}

	case model.SchemaKindArray:
		b.walk(s.Items, childModelName(sdkName, "Item"), model.ChildPath(path, "[]"))

	case model.SchemaKindMap:
		// child_models only recurses into additionalProperties when the value is a
		// $ref, so a map value is nameable through its own component name or not at
		// all: pass no accumulated name and let the $ref rule above supply one.
		b.walk(s.Items, "", model.ChildPath(path, "{}"))
	}
}

// bindUnion resolves one union's wrapper and members. The wrapper is the generated
// model the union node itself became: its component name when it has one, else the
// name the walk accumulated (model_oneof.j2's `name`).
func (b *binder) bindUnion(spec *model.OneOfSpec, sdkName, path string) {
	if spec == nil {
		return
	}
	// Prefer the union's own component name over the accumulated one. They agree
	// wherever both exist — the walk sets sdkName from the same RefName — but the
	// spec's copy is the identity the parser recorded, so read it directly.
	wrapper := spec.RefName
	if wrapper == "" {
		wrapper = sdkName
	}
	unionPath := spec.Path
	if unionPath == "" {
		unionPath = path
	}

	if wrapper == "" {
		b.failures = append(b.failures, Failure{
			Path: unionPath,
			Reason: "inline union sits at a position with no generated SDK model to name its " +
				"wrapper (the go-sdk names an inline model after the enclosing model plus the " +
				"property path, which does not resolve here); replace the inline oneOf with a " +
				"$ref to a named schema component",
		})
		// Still bind the members below: their names do not depend on the wrapper, and
		// reporting one wrapper failure is more useful than also reporting every
		// alternative as unbound.
	}
	spec.SDKType = wrapper

	for i := range spec.Variants {
		v := &spec.Variants[i]
		member, pointer, err := memberBinding(*v)
		if err != nil {
			b.failures = append(b.failures, Failure{
				Path:    unionPath,
				Variant: v.TFName,
				Reason:  err.Error(),
			})
			v.SDKField, v.SDKConstructor, v.SDKPointer = "", "", false
			continue
		}
		v.SDKField = member
		v.SDKPointer = pointer
		// model_oneof.j2 emits `<Member>As<Union>` for every alternative,
		// unconditionally — so this is never optional, and an empty wrapper name
		// leaves it empty rather than half-formed.
		if wrapper != "" {
			v.SDKConstructor = member + "As" + wrapper
		} else {
			v.SDKConstructor = ""
		}

		// An alternative may itself be a union, or contain one.
		b.walk(v.Schema, v.RefName, model.ChildPath(unionPath, v.TFName))
	}
}

// childModelName appends one path step to an accumulated SDK model name. An empty
// parent stays empty: child_models has no name to build on either, so the child is
// unnameable rather than named after the step alone.
func childModelName(parent, step string) string {
	if parent == "" {
		return ""
	}
	return parent + step
}

func sortedKeys(m map[string]*model.Schema) []string {
	if len(m) == 0 {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func artifactName(op *model.Operation) string {
	if op.Tracking != nil {
		return op.Tracking.ArtifactName
	}
	return ""
}

// Failure is one union (or one of its alternatives) the derivation could not bind.
type Failure struct {
	// Path is the union's schema path, e.g. "response.data.attributes.integration".
	Path string
	// Variant is the Terraform variant name when the failure is specific to one
	// alternative, empty when it concerns the wrapper itself.
	Variant string
	// Reason states what could not be derived and what the maintainer can do.
	Reason string
}

// UnresolvedBindingError reports every SDK oneOf binding an operation could not
// resolve. It names the artifact, operation, union path and alternative so the
// diagnostic is actionable without reading the specification.
type UnresolvedBindingError struct {
	Artifact  string
	Operation string
	Failures  []Failure
}

func (e *UnresolvedBindingError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "sdkbind: artifact %q operation %q: %d unresolved SDK oneOf binding(s):",
		e.Artifact, e.Operation, len(e.Failures))
	for _, f := range e.Failures {
		b.WriteString("\n  ")
		b.WriteString(f.Path)
		if f.Variant != "" {
			b.WriteString(" variant ")
			b.WriteString(f.Variant)
		}
		b.WriteString(": ")
		b.WriteString(f.Reason)
	}
	return b.String()
}
