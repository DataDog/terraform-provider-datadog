package model

import (
	"fmt"
	"sort"
	"strconv"
)

// nestingMode tracks whether the current subtree nests as blocks or as attributes.
// Inside a map<object> value blocks are forbidden, so the builder switches to
// nestAttribute and rewrites block forms to their attribute counterparts.
type nestingMode int

const (
	nestBlock nestingMode = iota
	nestAttribute
)

// treeKind distinguishes the two entry points, which differ only in presence
// flags: a response tree is state the provider reads back, so every node is
// Computed, whereas a request tree is practitioner input, so every node is
// Required or Optional instead.
type treeKind int

const (
	responseTree treeKind = iota
	requestTree
)

// oneOfValueField is the single child a non-object oneOf alternative exposes.
// A scalar, list, map, or directly nested union has no fields of its own to
// surface, so its variant block wraps the whole alternative under this name.
const oneOfValueField = "value"

// UnsupportedKindError reports a schema kind that cannot become a Terraform
// attribute — anyOf (classified unsupported), a ref_cycle, or any other
// unsupported node. The attribute-tree builder fails the artifact when it reaches
// one rather than emitting a types.Dynamic escape hatch.
type UnsupportedKindError struct {
	Path   string
	Kind   SchemaKind
	Reason string
}

func (e *UnsupportedKindError) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf(
			"model: cannot build attribute at %q: schema kind %q is not representable: %s",
			e.Path,
			e.Kind,
			e.Reason,
		)
	}
	return fmt.Sprintf("model: cannot build attribute at %q: schema kind %q is not representable", e.Path, e.Kind)
}

// OneOfProjectionError reports a union that cannot be projected into a Terraform
// envelope. It names the envelope, the offending alternative and the schema path
// so a maintainer can find the union in the OpenAPI document, and it wraps the
// underlying per-alternative failure when there is one. A oneOf is never dropped
// from the tree: either it projects, or its artifact fails with this error.
type OneOfProjectionError struct {
	// Envelope is the generated envelope name (OneOfSpec.Name).
	Envelope string
	// Variant is the Terraform variant name, empty when the whole envelope failed.
	Variant string
	// Path is the union's schema path.
	Path string
	// Reason states the failure directly; when empty, Err carries it.
	Reason string
	Err    error
}

func (e *OneOfProjectionError) Error() string {
	reason := e.Reason
	if reason == "" && e.Err != nil {
		reason = e.Err.Error()
	}
	if e.Variant != "" {
		return fmt.Sprintf(
			"model: cannot project oneOf envelope %q alternative %q at %q: %s",
			e.Envelope, e.Variant, e.Path, reason,
		)
	}
	return fmt.Sprintf("model: cannot project oneOf envelope %q at %q: %s", e.Envelope, e.Path, reason)
}

func (e *OneOfProjectionError) Unwrap() error { return e.Err }

// BuildResponseTree converts a response-body schema into an AttributeTree,
// rooting every attribute path at "response." and marking every node Computed.
//
// The returned diagnostics are the non-fatal notes raised during the walk; the
// conversion currently produces none, since every node either projects into the
// tree or fails the artifact.
func BuildResponseTree(s *Schema) (*AttributeTree, []Diagnostic, error) {
	return (&treeBuilder{kind: responseTree}).build(s, "response")
}

// BuildRequestTree converts a request-body schema into an AttributeTree, rooting
// every attribute path at "request." and marking each node Required or Optional
// rather than Computed. Like BuildResponseTree it returns the diagnostics raised
// during the walk.
func BuildRequestTree(s *Schema) (*AttributeTree, []Diagnostic, error) {
	return (&treeBuilder{kind: requestTree}).build(s, "request")
}

// treeBuilder carries the state of one AttributeTree conversion. Only the entry
// point's kind varies across a run; the recursion is otherwise a pure function of
// the schema node, its path, and its nesting context.
type treeBuilder struct {
	kind treeKind
}

// build is the shared recursion behind both entry points, differing only in root.
// A root object explodes its properties into top-level attributes; any other kind
// — including a bare union or a collection of unions — becomes one attribute at
// Path == root. A nil schema yields an empty tree.
func (b *treeBuilder) build(s *Schema, root string) (*AttributeTree, []Diagnostic, error) {
	tree := &AttributeTree{}
	if s == nil {
		return tree, nil, nil
	}
	if s.Kind == SchemaKindObject {
		attrs, err := b.children(s, root+".", nestBlock)
		if err != nil {
			return nil, nil, err
		}
		tree.Attributes = attrs
		return tree, nil, nil
	}
	// A body root is always present, so it is required when it is input at all.
	attr, err := b.attribute(s, root, nestBlock, true)
	if err != nil {
		return nil, nil, err
	}
	tree.Attributes = []*Attribute{attr}
	return tree, nil, nil
}

// attribute converts one schema node at path into an Attribute, recursing into its
// properties, element, or value schema. mode threads the nesting world down, and
// required says whether the node must be configured (a request-tree concern only).
// Every non-representable kind fails here rather than being skipped, so no marked
// field can disappear from the generated schema.
func (b *treeBuilder) attribute(s *Schema, path string, mode nestingMode, required bool) (*Attribute, error) {
	// A union has no framework type of its own: it projects into a synthetic
	// envelope whose form depends on where it sits, so it is handled separately.
	if s.Kind == SchemaKindOneOf {
		return b.envelope(s, path, mode, required)
	}

	// The remaining non-representable kinds (anyOf and other unsupported nodes,
	// ref_cycle) have no Terraform representation: fail the artifact here rather
	// than emit garbage.
	switch s.Kind {
	case SchemaKindPrimitive, SchemaKindJSON, SchemaKindObject, SchemaKindArray, SchemaKindMap:
		// representable — continue
	default:
		return nil, &UnsupportedKindError{
			Path:   path,
			Kind:   s.Kind,
			Reason: s.UnsupportedReason,
		}
	}

	tfType, goType, err := FrameworkType(s)
	if err != nil {
		return nil, err
	}
	// Inside a map value the framework forbids blocks, so rewrite block forms to
	// attribute forms (leaf types are unaffected).
	if mode == nestAttribute {
		tfType = attributeForm(tfType)
	}

	attr := &Attribute{
		Path:        path,
		TfType:      tfType,
		GoType:      goType,
		Format:      s.Format,
		Sensitive:   s.Sensitive,
		Description: s.Description,
	}
	if s.Kind == SchemaKindJSON {
		attr.CustomType = "jsontypes.NormalizedType{}"
	}
	b.applyPresence(attr, required)

	// A string enum becomes a OneOf validator; non-string enums produce none for now.
	if s.Kind == SchemaKindPrimitive && s.Type == "string" && len(s.Enum) > 0 {
		attr.IsEnum = true
		args := make([]string, len(s.Enum))
		for i, v := range s.Enum {
			args[i] = strconv.Quote(v)
		}
		attr.Validators = []ValidatorSpec{{Name: "stringvalidator.OneOf", Args: args}}
	}

	// Recurse into object shapes, or record the recursive element type for a
	// collection chain that terminates in a primitive.
	switch s.Kind {
	case SchemaKindObject:
		children, err := b.children(s, path+".", mode)
		if err != nil {
			return nil, err
		}
		attr.Children, attr.ModelRefName = children, s.RefName

	case SchemaKindArray:
		switch s.Items.Kind {
		case SchemaKindObject:
			children, err := b.children(s.Items, path+"[].", mode)
			if err != nil {
				return nil, err
			}
			// The element supplies the struct, so the element's component names it.
			attr.Children, attr.ModelRefName = children, s.Items.RefName
		case SchemaKindOneOf:
			// The list itself carries the envelope: its elements are variant
			// blocks, so no attribute stands at the element path.
			variants, envelope, err := b.oneOfVariants(s.Items, path+"[]", mode)
			if err != nil {
				return nil, err
			}
			attr.Children, attr.OneOf = variants, envelope
		default:
			elem, err := ElementType(s.Items)
			if err != nil {
				return nil, err
			}
			attr.ElementType = elem
		}

	case SchemaKindMap:
		switch s.Items.Kind {
		case SchemaKindObject:
			// A map<object> is a NestedAttributeObject; force everything beneath it
			// into attribute form regardless of the incoming mode.
			children, err := b.children(s.Items, path+"{}.", nestAttribute)
			if err != nil {
				return b.responseMapFallback(attr, err)
			}
			attr.Children, attr.ModelRefName = children, s.Items.RefName
		case SchemaKindOneOf:
			variants, envelope, err := b.oneOfVariants(s.Items, path+"{}", nestAttribute)
			if err != nil {
				return b.responseMapFallback(attr, err)
			}
			attr.Children, attr.OneOf = variants, envelope
		default:
			elem, err := ElementType(s.Items)
			if err != nil {
				return b.responseMapFallback(attr, err)
			}
			attr.ElementType = elem
		}
	}

	return attr, nil
}

// responseMapFallback preserves a computed dynamic map as normalized JSON when
// one of its value descendants cannot be represented as a Terraform nested
// attribute. The map remains lossless in state, while request trees stay strict:
// accepting opaque practitioner input would discard the OpenAPI validation shape.
func (b *treeBuilder) responseMapFallback(attr *Attribute, projectionErr error) (*Attribute, error) {
	if b.kind != responseTree {
		return nil, projectionErr
	}
	attr.TfType = "schema.StringAttribute"
	attr.GoType = "jsontypes.Normalized"
	attr.CustomType = "jsontypes.NormalizedType{}"
	attr.ElementType = ""
	attr.Children = nil
	attr.OneOf = nil
	attr.ModelRefName = ""
	return attr, nil
}

// children builds one child attribute per property of parent, each pathed
// prefix+key. Keys are visited sorted, making recursion deterministic and the
// result Path-sorted. Required-ness comes from the parent's required list, which
// only reaches the output in a request tree.
func (b *treeBuilder) children(parent *Schema, prefix string, mode nestingMode) ([]*Attribute, error) {
	props := parent.Properties
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	required := make(map[string]bool, len(parent.Required))
	for _, name := range parent.Required {
		required[name] = true
	}

	children := make([]*Attribute, 0, len(props))
	for _, key := range keys {
		// Terraform attribute names must be snake_case; SnakeCase normalizes camelCase
		// OAS names and is idempotent on already-snake names (SdkName recovers the getter).
		child, err := b.attribute(props[key], prefix+SnakeCase(key), mode, required[key])
		if err != nil {
			return nil, err
		}
		children = append(children, child)
	}
	return children, nil
}

// envelope projects a union standing at its own position in the tree — the schema
// root, or an object property — into the synthetic block that holds its variants.
func (b *treeBuilder) envelope(s *Schema, path string, mode nestingMode, required bool) (*Attribute, error) {
	variants, envelope, err := b.oneOfVariants(s, path, mode)
	if err != nil {
		return nil, err
	}
	attr := &Attribute{
		Path:        path,
		TfType:      singleNestedForm(mode),
		GoType:      "types.Object",
		Sensitive:   s.Sensitive,
		Description: s.Description,
		Children:    variants,
		OneOf:       envelope,
	}
	// The envelope is required only when its containing field demands a value and
	// the union itself is neither optional nor nullable; a nullable union is
	// represented by an absent envelope rather than a null variant.
	b.applyPresence(attr, required && !envelope.Optional)
	return attr, nil
}

// oneOfVariants projects the alternatives of a normalized union into one nested
// block each, pathed under basePath, and returns them with the envelope metadata
// the emit layer needs. basePath is the union's own schema path: the envelope
// attribute's path for a root or property union, and the element path
// ("choices[]", "choices{}") when the union is a collection's element — in that
// case the collection attribute carries the envelope and these blocks are its
// children directly.
func (b *treeBuilder) oneOfVariants(s *Schema, basePath string, mode nestingMode) ([]*Attribute, *OneOfEnvelope, error) {
	spec := s.OneOf
	if spec == nil {
		return nil, nil, &UnsupportedKindError{
			Path:   basePath,
			Kind:   s.Kind,
			Reason: "oneOf node carries no normalized union",
		}
	}
	if spec.Name == "" {
		return nil, nil, &OneOfProjectionError{
			Path:   basePath,
			Reason: "union has no generated envelope name to build a model from",
		}
	}
	if len(spec.Variants) == 0 {
		return nil, nil, &OneOfProjectionError{
			Envelope: spec.Name,
			Path:     basePath,
			Reason:   "union has no non-null alternative to select",
		}
	}

	envelope := &OneOfEnvelope{
		Name:    spec.Name,
		GoModel: oneOfModelName(spec.Name),
		SDKType: spec.SDKType,
		Path:    basePath,
		// A nullable union maps to an absent envelope, so it is as optional as one
		// whose containing field is optional.
		Optional: spec.Optional || spec.Nullable,
		Computed: b.kind == responseTree,
		Variants: make([]OneOfEnvelopeVariant, 0, len(spec.Variants)),
	}

	// Order by Terraform name so neither OpenAPI alternative order nor a caller's
	// construction order can reach the generated schema. The parser already sorts;
	// sorting a copy here keeps the projection correct for any caller without
	// mutating the spec.
	ordered := make([]OneOfVariant, len(spec.Variants))
	copy(ordered, spec.Variants)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].TFName < ordered[j].TFName })

	blocks := make([]*Attribute, 0, len(ordered))
	for _, variant := range ordered {
		block, projected, err := b.oneOfVariant(envelope, variant, basePath, mode)
		if err != nil {
			return nil, nil, err
		}
		blocks = append(blocks, block)
		envelope.Variants = append(envelope.Variants, projected)
	}
	return blocks, envelope, nil
}

// oneOfVariant projects one alternative into its nested block. An object
// alternative exposes its own fields; every other shape — scalar, list, map, or a
// directly nested union — has no fields to expose, so it gets a single child named
// "value" holding the alternative itself.
func (b *treeBuilder) oneOfVariant(
	envelope *OneOfEnvelope,
	variant OneOfVariant,
	basePath string,
	mode nestingMode,
) (*Attribute, OneOfEnvelopeVariant, error) {
	fail := func(reason string, err error) (*Attribute, OneOfEnvelopeVariant, error) {
		return nil, OneOfEnvelopeVariant{}, &OneOfProjectionError{
			Envelope: envelope.Name,
			Variant:  variant.TFName,
			Path:     basePath,
			Reason:   reason,
			Err:      err,
		}
	}
	if variant.TFName == "" {
		return fail("alternative has no stable Terraform variant name", nil)
	}
	if variant.Schema == nil {
		return fail("alternative has no normalized schema", nil)
	}

	path := basePath + "." + variant.TFName
	block := &Attribute{
		Path:        path,
		TfType:      singleNestedForm(mode),
		GoType:      "types.Object",
		Sensitive:   variant.Schema.Sensitive,
		Description: variant.Schema.Description,
	}
	// A variant is a choice, never a mandatory field: exactly-one selection is
	// enforced by the envelope's validator and by request mapping, not by marking
	// every branch Required.
	b.applyPresence(block, false)

	valueWrapped := variant.Schema.Kind != SchemaKindObject
	if valueWrapped {
		// The wrapped value is present whenever its variant is selected.
		value, err := b.attribute(variant.Schema, path+"."+oneOfValueField, mode, true)
		if err != nil {
			return fail("", err)
		}
		block.Children = []*Attribute{value}
	} else {
		children, err := b.children(variant.Schema, path+".", mode)
		if err != nil {
			return fail("", err)
		}
		block.Children = children
	}

	goField := variant.GoName
	if goField == "" {
		goField = SdkName(variant.TFName)
	}
	return block, OneOfEnvelopeVariant{
		TFName:  variant.TFName,
		GoField: goField,
		GoModel: oneOfModelName(envelope.Name + goField),
		// Carried through, never derived: SDK and Terraform naming diverge.
		SDKField:       variant.SDKField,
		SDKConstructor: variant.SDKConstructor,
		SDKPointer:     variant.SDKPointer,
		ValueWrapped:   valueWrapped,
		Attribute:      block,
	}, nil
}

// applyPresence sets the framework presence flags. Response state is entirely
// Computed; request input is Required when the schema says so and Optional
// otherwise. Exactly one flag ends up set either way, as the framework requires.
func (b *treeBuilder) applyPresence(a *Attribute, required bool) {
	if b.kind == responseTree {
		a.Computed = true
		return
	}
	a.Required = required
	a.Optional = !required
}

// attributeForm rewrites a block framework type into its nested-attribute
// counterpart, leaving leaf and already-attribute forms alone.
func attributeForm(tfType string) string {
	switch tfType {
	case "schema.SingleNestedBlock":
		return "schema.SingleNestedAttribute"
	case "schema.ListNestedBlock":
		return "schema.ListNestedAttribute"
	default:
		return tfType
	}
}

// singleNestedForm is the single-nested framework type valid in mode's nesting
// world. Envelopes and variant blocks are always single-nested: the envelope holds
// one variant, the variant one alternative.
func singleNestedForm(mode nestingMode) string {
	if mode == nestAttribute {
		return "schema.SingleNestedAttribute"
	}
	return "schema.SingleNestedBlock"
}

// oneOfModelName is the generated Go struct name for an envelope or variant.
// Deriving it from the envelope's name (rather than from the use site) is what
// lets two uses of one reusable union share a single generated model.
func oneOfModelName(name string) string { return name + "Model" }
