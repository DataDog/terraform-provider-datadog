package model

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// nestingMode tracks whether the current subtree nests as blocks or as attributes.
// Inside a map<object> value blocks are forbidden, so the builder switches to
// nestAttribute and rewrites block forms to their attribute counterparts.
type nestingMode int

const (
	nestBlock nestingMode = iota
	nestAttribute
)

// treeKind distinguishes the three entry points, which differ only in
// presence flags: a response tree is state the provider reads back, so every
// node is Computed; a request tree is practitioner input, so every node is
// Required or Optional instead; a resource tree derives Required, Optional,
// or Computed (or both) per node from its Provenance.
type treeKind int

const (
	responseTree treeKind = iota
	requestTree
	resourceTree
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

// MissingProvenanceError reports a resource-tree node with no Provenance to
// derive presence flags from. Every node a resource schema merge itself
// produces carries one; the one shape that does not is the content nested
// inside a oneOf alternative, which the merge clones verbatim rather than
// walking.
type MissingProvenanceError struct {
	Path string
}

func (e *MissingProvenanceError) Error() string {
	return fmt.Sprintf("model: cannot derive resource presence at %q: node carries no merge provenance", e.Path)
}

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

// BuildResourceTree converts a schema produced by a resource schema merge
// into an AttributeTree rooted at "resource.". Each node's Required, Optional
// and Computed flags come from its Provenance rather than from a single
// request or response direction, and each gets the plan modifiers that
// follow from those flags plus updateUnsupported — true when the resource's
// group resolves no Update role, so every request-settable attribute forces
// replacement instead.
func BuildResourceTree(s *Schema, updateUnsupported bool) (*AttributeTree, []Diagnostic, error) {
	return (&treeBuilder{kind: resourceTree, updateUnsupported: updateUnsupported}).build(s, "resource")
}

// treeBuilder carries the state of one AttributeTree conversion. Only the entry
// point's kind varies across a run; the recursion is otherwise a pure function of
// the schema node, its path, and its nesting context.
type treeBuilder struct {
	kind treeKind
	// oneOfProvenance is the enclosing union's Provenance while walking one of
	// its own alternatives, resourceTree only. A resource schema merge clones
	// a oneOf's content verbatim rather than walking it, so no node inside an
	// alternative carries its own Provenance; applyPresence falls back to
	// this one instead.
	oneOfProvenance *SchemaProvenance
	// updateUnsupported is resourceTree's own input, fixed for the whole walk:
	// true when the resource's group resolves no Update role, so every
	// request-settable attribute gets RequiresReplace(). A path parameter also
	// gets it, unconditionally and from model/artifact.go, since no endpoint
	// re-parents a child — so RequiresReplace() is not by itself evidence that the
	// group resolved no Update role.
	updateUnsupported bool
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
	case SchemaKindPrimitive, SchemaKindObject, SchemaKindArray, SchemaKindMap:
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
	if err := b.applyPresence(attr, s, required); err != nil {
		return nil, err
	}

	// A string enum becomes a OneOf validator; non-string enums produce none for now.
	if s.Kind == SchemaKindPrimitive && isStringEnum(s) {
		attr.IsEnum = true
		attr.RequestModelRefName = s.RequestRefName
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
		attr.Children, attr.ModelRefName, attr.RequestModelRefName = children, s.RefName, s.RequestRefName

	case SchemaKindArray:
		switch s.Items.Kind {
		case SchemaKindObject:
			children, err := b.children(s.Items, path+"[].", mode)
			if err != nil {
				return nil, err
			}
			// The element supplies the struct, so the element's component names it.
			attr.Children, attr.ModelRefName, attr.RequestModelRefName = children, s.Items.RefName, s.Items.RequestRefName
		case SchemaKindOneOf:
			// The list itself carries the envelope: its elements are variant
			// blocks, so no attribute stands at the element path.
			variants, envelope, err := b.oneOfVariants(s.Items, path+"[]", mode)
			if err != nil {
				return nil, err
			}
			attr.Children, attr.OneOf = variants, envelope
		default:
			var err error
			attr.ElementType, attr.ElementFormat, attr.ElementIsEnum, err = elementInfo(s.Items)
			if err != nil {
				return nil, err
			}
		}

	case SchemaKindMap:
		switch s.Items.Kind {
		case SchemaKindObject:
			// A map<object> is a NestedAttributeObject; force everything beneath it
			// into attribute form regardless of the incoming mode.
			children, err := b.children(s.Items, path+"{}.", nestAttribute)
			if err != nil {
				return nil, err
			}
			attr.Children, attr.ModelRefName = children, s.Items.RefName
		case SchemaKindOneOf:
			variants, envelope, err := b.oneOfVariants(s.Items, path+"{}", nestAttribute)
			if err != nil {
				return nil, err
			}
			attr.Children, attr.OneOf = variants, envelope
		default:
			var err error
			attr.ElementType, attr.ElementFormat, attr.ElementIsEnum, err = elementInfo(s.Items)
			if err != nil {
				return nil, err
			}
		}
	}

	return attr, nil
}

// isStringEnum reports whether s is a string constrained to a fixed set of
// values — the same predicate that promotes a primitive attribute itself to
// IsEnum, reused here for a collection element (see elementInfo).
func isStringEnum(s *Schema) bool {
	return s.Type == "string" && len(s.Enum) > 0
}

// elementInfo derives a ListAttribute/MapAttribute's element triple: the
// framework attr.Type expression ElementType renders as, and, since that
// generic mapping collapses a date-time or enum string element to the same
// "types.StringType" as a plain one, the element's own Format and enum-ness
// alongside it — see Attribute.ElementFormat/ElementIsEnum.
func elementInfo(items *Schema) (elementType, format string, isEnum bool, err error) {
	elementType, err = ElementType(items)
	if err != nil {
		return "", "", false, err
	}
	return elementType, items.Format, isStringEnum(items), nil
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
		// Keep the property's own OpenAPI name. SnakeCase is not injective —
		// hostTags, host_tags and host-tags all collapse to host_tags — so a
		// consumer that needs to find this node again in an OpenAPI schema
		// cannot invert the Path, and guessing would silently pick a sibling.
		child.OpenAPIName = key
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
	if err := b.applyPresence(attr, s, required && !envelope.Optional); err != nil {
		return nil, err
	}
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
		block, projected, err := b.oneOfVariant(s, envelope, variant, basePath, mode)
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
// "value" holding the alternative itself. union is the schema of the oneOf node
// itself, carrying the Provenance the block's own presence is derived from —
// no alternative carries its own, since a resource schema merge clones a
// oneOf's content verbatim rather than walking it.
func (b *treeBuilder) oneOfVariant(
	union *Schema,
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
	if err := b.applyPresence(block, union, false); err != nil {
		return fail("", err)
	}

	// The alternative's own content carries no Provenance (see oneOfProvenance).
	// A directly nested union (variant.Schema.Kind == SchemaKindOneOf) has none
	// of its own either, so leave the fallback as whatever enclosing union was
	// already active rather than clobbering it to nil; restore it once this
	// alternative is fully walked either way.
	outerProvenance := b.oneOfProvenance
	if union.Provenance != nil {
		b.oneOfProvenance = union.Provenance
	}
	defer func() { b.oneOfProvenance = outerProvenance }()

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

// applyPresence sets the framework presence flags, and for resourceTree the
// plan modifiers that follow from them. Response state is entirely Computed;
// request input is Required when the schema says so and Optional otherwise;
// resource input reads s.Provenance instead (falling back to oneOfProvenance
// when s carries none), deriving Required, or Optional, or Optional and
// Computed together, or Computed alone — the only one of the three kinds
// that can set two flags at once. required is "is this required by the
// Create body" for a node reached through an object's Required list, but the
// oneOf wrapped-value call site instead passes it unconditionally ("the
// value must be present when its variant is selected"), so resourceTree only
// honors it alongside InRequest — a purely response-only union's wrapped
// value must not come out Required just because it was hardcoded true.
// Returns MissingProvenanceError if resourceTree finds neither source.
func (b *treeBuilder) applyPresence(a *Attribute, s *Schema, required bool) error {
	switch b.kind {
	case responseTree:
		a.Computed = true
	case requestTree:
		a.Required = required
		a.Optional = !required
	case resourceTree:
		p := s.Provenance
		if p == nil {
			p = b.oneOfProvenance
		}
		if p == nil {
			return &MissingProvenanceError{Path: a.Path}
		}
		switch {
		case required && p.InRequest:
			a.Required = true
		case p.InRequest && p.InResponse:
			a.Optional, a.Computed = true, true
		case p.InRequest:
			a.Optional = true
		default:
			a.Computed = true
		}
		a.InResponse = p.InResponse
		a.PlanModifiers = b.resourcePlanModifiers(a, p.InRequest)
	}
	return nil
}

// resourcePlanModifiers derives a's plan modifiers: UseStateForUnknown() when
// applyPresence just set both Optional and Computed on it, RequiresReplace()
// when requestSettable and updateUnsupported — never either one on a
// Computed-only attribute, since the server may change such a value during
// apply and there is no update endpoint to reconcile it through anyway. Both
// are typed from a.GoType, so a Computed-only attribute never even looks one
// up. requestSettable is the caller's Provenance.InRequest, not re-derived
// from a.Required/a.Optional: those are true under the same condition today,
// but only because applyPresence's presence switch happens to be exhaustive
// and mutually exclusive — InRequest is the fact itself.
func (b *treeBuilder) resourcePlanModifiers(a *Attribute, requestSettable bool) []PlanModifierSpec {
	useStateForUnknown := a.Optional && a.Computed
	requiresReplace := b.updateUnsupported && requestSettable
	if !useStateForUnknown && !requiresReplace {
		return nil
	}
	var modifiers []PlanModifierSpec
	if useStateForUnknown {
		modifiers = append(modifiers, PlanModifierSpec{Name: PlanModifierPackage(a.GoType) + ".UseStateForUnknown"})
	}
	if requiresReplace {
		modifiers = append(modifiers, RequiresReplaceSpec(a.GoType))
	}
	return modifiers
}

// RequiresReplaceSpec is the RequiresReplace() plan modifier for an attribute
// of goType. It is the single spelling of that modifier, shared by the two
// places that decide an attribute forces replacement: a request-settable body
// field when the group resolves no Update role (above), and a path parameter,
// which forces replacement whatever the Update role is because no endpoint
// re-parents a child (model/artifact.go).
func RequiresReplaceSpec(goType string) PlanModifierSpec {
	return PlanModifierSpec{Name: PlanModifierPackage(goType) + ".RequiresReplace"}
}

// PlanModifierPackage returns the terraform-plugin-framework planmodifier
// subpackage for goType, e.g. "types.String" -> "stringplanmodifier". goType
// is always one of FrameworkType's seven possible outputs by the time it
// reaches here, all following this same naming convention. Exported so the
// emitter's import block derives the subpackage by the same rule that spelled
// it into PlanModifierSpec.Name, rather than parsing it back out of that name.
func PlanModifierPackage(goType string) string {
	return strings.ToLower(strings.TrimPrefix(goType, "types.")) + "planmodifier"
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
