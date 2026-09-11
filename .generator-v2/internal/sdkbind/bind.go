// Package sdkbind resolves the go-sdk bindings for every OpenAPI oneOf
// reachable from a normalized operation, writing OneOfSpec.SDKType and each
// variant's SDKField/SDKConstructor/SDKPointer. Names are re-derived by
// reimplementing the go-sdk generator's logic: gotype.go holds the naming
// rules, this file the walk deciding which generated model a union lives in.
package sdkbind

import (
	"fmt"
	"sort"
	"strings"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// BindOperation resolves the SDK oneOf bindings for every union reachable from
// op's request and response schemas, mutating them in place; idempotent.
// Failures are collected rather than returned on first sight: the returned
// *UnresolvedBindingError names every union path and alternative it could not
// bind. A position with no derivable model name is walked past, not abandoned.
func BindOperation(op *model.Operation) error {
	if op == nil {
		return nil
	}
	b := &binder{visited: make(map[visitKey]bool)}
	// Each body root is the SDK type named in the operation's method signature,
	// where the SDK generator's own naming starts too.
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

// BindSpec binds every operation in spec, returning a map from operation to its
// unresolvable-union error, or nil when all of them bound cleanly.
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
	// visited guards termination: the normalized graph is a DAG, not a tree,
	// since one sibling-constraint node can be shared across the alternatives it
	// was merged into. Keying on (node, SDK name) rather than the node alone
	// lets a shared node bind correctly when two positions name it differently,
	// instead of letting whichever was walked first decide for both.
	visited map[visitKey]bool
}

type visitKey struct {
	schema  *model.Schema
	sdkName string
}

// walk descends s carrying sdkName, the generated SDK model this node lives in,
// mirroring openapi.child_models: a $ref restarts the name, an object property
// appends camel_case(key), an array element appends "Item". sdkName is empty at
// a position the rules cannot name, and the walk continues anyway so that only
// a union actually standing there fails.
func (b *binder) walk(s *model.Schema, sdkName, path string) {
	if s == nil {
		return
	}
	// get_name(schema): a component name always wins over the name accumulated
	// from the parent.
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
		// child_models only recurses into additionalProperties for a $ref, so a
		// map value is nameable through its own component name or not at all:
		// pass no accumulated name and let the $ref rule above supply one.
		b.walk(s.Items, "", model.ChildPath(path, "{}"))
	}
}

// bindUnion resolves one union's wrapper and members. The wrapper is the model
// the union node itself became: its component name when it has one, else the
// name the walk accumulated (model_oneof.j2's `name`).
func (b *binder) bindUnion(spec *model.OneOfSpec, sdkName, path string) {
	if spec == nil {
		return
	}
	// Prefer the union's own component name. It agrees with the accumulated one
	// wherever both exist, but is the identity the parser recorded.
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
		// Members are still bound below: their names do not depend on the
		// wrapper, so one wrapper failure beats also reporting every alternative
		// as unbound.
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
		// model_oneof.j2 emits `<Member>As<Union>` for every alternative; with no
		// wrapper name, leave it empty rather than half-formed.
		if wrapper != "" {
			v.SDKConstructor = member + "As" + wrapper
		} else {
			v.SDKConstructor = ""
		}

		// An alternative may itself be a union, or contain one.
		b.walk(v.Schema, v.RefName, model.ChildPath(unionPath, v.TFName))
	}
}

// childModelName appends one path step to an accumulated SDK model name. An
// empty parent stays empty: the child is unnameable rather than named after the
// step alone.
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
// resolve, naming the artifact, operation, union path and alternative.
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
