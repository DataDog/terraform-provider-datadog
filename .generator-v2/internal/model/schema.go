package model

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// treeKind selects how presence flags are assigned: responseTree marks every
// node Computed, requestTree marks each Required or Optional, resourceTree
// derives the flags per node from its Provenance.
type treeKind int

const (
	responseTree treeKind = iota
	requestTree
	resourceTree
)

// oneOfValueField is the single child name a non-object oneOf alternative
// (scalar, list, map, or nested union) wraps its whole value under.
const oneOfValueField = "value"

// UnsupportedKindError reports a schema kind that cannot become a Terraform
// attribute: a ref_cycle, an anyOf, or any other node classified unsupported.
type UnsupportedKindError struct {
	Path   string
	Kind   SchemaKind
	Reason string
}

func (e *UnsupportedKindError) Error() string {
	return fmt.Sprintf("model: cannot build attribute at %q: schema kind %q is not representable%s",
		e.Path, e.Kind, reasonText(e.Reason))
}

// OneOfProjectionError reports a union that cannot be projected into a Terraform
// envelope, naming the envelope, the offending alternative and the schema path,
// and wrapping the underlying per-alternative failure when there is one.
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
// derive presence flags from. A merge gives every node it walks one; content
// inside a oneOf alternative is cloned verbatim and so has none.
type MissingProvenanceError struct {
	Path string
}

func (e *MissingProvenanceError) Error() string {
	return fmt.Sprintf("model: cannot derive resource presence at %q: node carries no merge provenance", e.Path)
}

// BuildResponseTree converts a response-body schema into an AttributeTree,
// rooting every attribute path at "response." and marking every node Computed.
// The diagnostics are the walk's non-fatal notes; none are produced today,
// since a node either projects or errors.
func BuildResponseTree(s *Schema) (*AttributeTree, []Diagnostic, error) {
	return (&treeBuilder{kind: responseTree}).build(s, "response")
}

// BuildRequestTree converts a request-body schema into an AttributeTree, rooting
// every attribute path at "request." and marking each node Required or Optional
// rather than Computed.
func BuildRequestTree(s *Schema) (*AttributeTree, []Diagnostic, error) {
	return (&treeBuilder{kind: requestTree}).build(s, "request")
}

// BuildResourceTree converts a merged resource schema into an AttributeTree
// rooted at "resource.". Presence flags come from each node's Provenance, and
// plan modifiers follow from those flags plus updateUnsupported — when true,
// every request-settable attribute gets RequiresReplace().
func BuildResourceTree(s *Schema, updateUnsupported bool) (*AttributeTree, []Diagnostic, error) {
	return (&treeBuilder{kind: resourceTree, updateUnsupported: updateUnsupported}).build(s, "resource")
}

// treeBuilder carries the state of one AttributeTree conversion; the recursion
// is otherwise a function of the schema node, its path and its nesting context.
type treeBuilder struct {
	kind treeKind
	// oneOfProvenance is the enclosing union's Provenance while walking one of
	// its alternatives, resourceTree only. A merge clones a oneOf's content
	// verbatim, so nodes inside an alternative carry none of their own and
	// applyPresence falls back to this.
	oneOfProvenance *SchemaProvenance
	// updateUnsupported is fixed for the whole walk: true when no Update role
	// exists, so every request-settable attribute gets RequiresReplace(). Path
	// parameters get it unconditionally elsewhere, so its presence on an
	// attribute does not imply this flag.
	updateUnsupported bool
}

// build converts s into a tree rooted at root. A root object explodes its
// properties into top-level attributes; any other kind — a bare union or a
// collection included — becomes one attribute at Path == root. A nil schema
// yields an empty tree.
func (b *treeBuilder) build(s *Schema, root string) (*AttributeTree, []Diagnostic, error) {
	tree := &AttributeTree{}
	if s == nil {
		return tree, nil, nil
	}
	if s.Kind == SchemaKindObject {
		attrs, err := b.children(s, root+".")
		if err != nil {
			return nil, nil, err
		}
		tree.Attributes = attrs
		return tree, nil, nil
	}
	// A body root is always present, so it is required whenever it is input —
	// except a union, which states its own optionality, there being no
	// enclosing object's required list to consult.
	attr, err := b.attribute(s, root, !rootUnionOptional(s))
	if err != nil {
		return nil, nil, err
	}
	tree.Attributes = []*Attribute{attr}
	return tree, nil, nil
}

// attribute converts one schema node at path into an Attribute, recursing into
// its properties, element or value schema. required says whether the node must
// be configured. A non-representable kind errors rather than being skipped, so
// no field can silently vanish from the schema.
func (b *treeBuilder) attribute(s *Schema, path string, required bool) (*Attribute, error) {
	// A union has no framework type of its own; it projects into a synthetic
	// envelope instead.
	if s.Kind == SchemaKindOneOf {
		return b.envelope(s, path, required)
	}

	// The remaining kinds (unsupported, ref_cycle, depth_exceeded) have no
	// Terraform representation.
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

	// A string enum gets a stringvalidator.OneOf; other enums get no validator.
	if s.Kind == SchemaKindPrimitive && isStringEnum(s) {
		attr.IsEnum = true
		attr.RequestModelRefName = s.RequestRefName
		args := make([]string, len(s.Enum))
		for i, v := range s.Enum {
			args[i] = strconv.Quote(v)
		}
		attr.Validators = []ValidatorSpec{{Name: "stringvalidator.OneOf", Args: args}}
	}

	// Recurse into object shapes, or record the element type for a collection
	// chain terminating in a primitive.
	switch s.Kind {
	case SchemaKindObject:
		children, err := b.children(s, path+".")
		if err != nil {
			return nil, err
		}
		attr.Children, attr.ModelRefName, attr.RequestModelRefName = children, s.RefName, s.RequestRefName

	case SchemaKindArray:
		switch s.Items.Kind {
		case SchemaKindObject:
			children, err := b.children(s.Items, path+"[].")
			if err != nil {
				return nil, err
			}
			// The element supplies the struct, so it names it.
			attr.Children, attr.ModelRefName, attr.RequestModelRefName = children, s.Items.RefName, s.Items.RequestRefName
		case SchemaKindOneOf:
			// The list carries the envelope, so its children are the variant
			// attributes and no attribute stands at the element path.
			variants, envelope, err := b.oneOfVariants(s.Items, path+"[]")
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
			// A map<object> is a NestedAttributeObject.
			children, err := b.children(s.Items, path+"{}.")
			if err != nil {
				return nil, err
			}
			attr.Children, attr.ModelRefName = children, s.Items.RefName
		case SchemaKindOneOf:
			variants, envelope, err := b.oneOfVariants(s.Items, path+"{}")
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
// values.
func isStringEnum(s *Schema) bool {
	return s.Type == "string" && len(s.Enum) > 0
}

// elementInfo derives a collection's element triple: the attr.Type expression,
// plus the element's own Format and enum-ness, which that expression would
// otherwise lose (a date-time or enum string maps to plain types.StringType).
func elementInfo(items *Schema) (elementType, format string, isEnum bool, err error) {
	elementType, err = ElementType(items)
	if err != nil {
		return "", "", false, err
	}
	return elementType, items.Format, isStringEnum(items), nil
}

// children builds one child attribute per property of parent, each pathed
// prefix+SnakeCase(key). Keys are visited sorted, so the result is Path-sorted
// and the recursion deterministic. Required-ness comes from parent.Required.
func (b *treeBuilder) children(parent *Schema, prefix string) ([]*Attribute, error) {
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
		// Terraform attribute names must be snake_case; SnakeCase is idempotent
		// on names already in that form.
		child, err := b.attribute(props[key], prefix+SnakeCase(key), required[key])
		if err != nil {
			return nil, err
		}
		// Keep the raw OpenAPI key: SnakeCase is not injective (hostTags,
		// host_tags and host-tags all give host_tags), so Path cannot be
		// inverted back to it.
		child.OpenAPIName = key
		children = append(children, child)
	}
	return children, nil
}

// envelope projects a union standing at its own position in the tree — a root
// or an object property — into the synthetic block holding its variants.
func (b *treeBuilder) envelope(s *Schema, path string, required bool) (*Attribute, error) {
	variants, envelope, err := b.oneOfVariants(s, path)
	if err != nil {
		return nil, err
	}
	attr := &Attribute{
		Path:        path,
		TfType:      "schema.SingleNestedAttribute",
		GoType:      "types.Object",
		Sensitive:   s.Sensitive,
		Description: s.Description,
		Children:    variants,
		OneOf:       envelope,
	}
	// Required only when the containing field demands a value and the union may
	// not be absent — a nullable union is an absent envelope, not a null
	// variant. Keyed on Nullable alone: OneOf.Optional either restates
	// `required`, or (on a resource tree) has been OR-ed across three bodies
	// into the weaker "some body may omit this". Only a root reads it.
	if err := b.applyPresence(attr, s, required && !s.OneOf.Nullable); err != nil {
		return nil, err
	}
	return attr, nil
}

// oneOfVariants projects a normalized union's alternatives into one nested
// attribute each and returns them with the envelope metadata. basePath is the
// union's own schema path: the envelope attribute's path for a root or property
// union, the element path ("choices[]", "choices{}") for a collection's
// element, where the collection attribute carries the envelope itself.
func (b *treeBuilder) oneOfVariants(s *Schema, basePath string) ([]*Attribute, *OneOfEnvelope, error) {
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
		// A nullable union maps to an absent envelope, so it is as optional
		// as one whose containing field is optional.
		Optional: spec.Optional || spec.Nullable,
		Computed: b.kind == responseTree,
		Variants: make([]OneOfEnvelopeVariant, 0, len(spec.Variants)),
	}

	// Order by Terraform name so no OpenAPI or caller construction order reaches
	// the output. The parser already sorts; a copy is sorted here so any caller
	// is correct without the spec being mutated.
	ordered := make([]OneOfVariant, len(spec.Variants))
	copy(ordered, spec.Variants)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].TFName < ordered[j].TFName })

	blocks := make([]*Attribute, 0, len(ordered))
	for _, variant := range ordered {
		block, projected, err := b.oneOfVariant(s, envelope, variant, basePath)
		if err != nil {
			return nil, nil, err
		}
		blocks = append(blocks, block)
		envelope.Variants = append(envelope.Variants, projected)
	}
	return blocks, envelope, nil
}

// oneOfVariant projects one alternative into its nested attribute. An object
// alternative exposes its own fields; every other shape gets a single child
// named "value" holding the alternative itself. The block's presence derives
// from union's Provenance, not its own, a variant being a choice; its children
// do carry their own.
func (b *treeBuilder) oneOfVariant(
	union *Schema,
	envelope *OneOfEnvelope,
	variant OneOfVariant,
	basePath string,
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
		TfType:      "schema.SingleNestedAttribute",
		GoType:      "types.Object",
		Sensitive:   variant.Schema.Sensitive,
		Description: variant.Schema.Description,
	}
	// A variant is a choice, never mandatory: exactly-one selection is enforced
	// by the envelope's validator, not by marking every branch Required.
	if err := b.applyPresence(block, union, false); err != nil {
		return fail("", err)
	}

	// The alternative's content carries no Provenance, so publish the union's as
	// the fallback. A directly nested union has none either: keep whatever
	// enclosing union was active rather than clobbering it to nil. Restored on
	// the way out either way.
	outerProvenance := b.oneOfProvenance
	if union.Provenance != nil {
		b.oneOfProvenance = union.Provenance
	}
	defer func() { b.oneOfProvenance = outerProvenance }()

	valueWrapped := variant.Schema.Kind != SchemaKindObject
	if valueWrapped {
		// The wrapped value is present whenever its variant is selected.
		value, err := b.attribute(variant.Schema, path+"."+oneOfValueField, true)
		if err != nil {
			return fail("", err)
		}
		block.Children = []*Attribute{value}
	} else {
		children, err := b.children(variant.Schema, path+".")
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

// applyPresence sets a's presence flags, plus resourceTree's plan modifiers.
// responseTree is all Computed; requestTree is Required per `required` and
// Optional otherwise; resourceTree reads s.Provenance (falling back to
// b.oneOfProvenance) and is the only kind that can set two flags at once
// (Optional+Computed). resourceTree honors `required` only alongside InRequest,
// since the oneOf wrapped-value call site passes it unconditionally. Returns
// MissingProvenanceError when resourceTree finds no Provenance at all.
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

// resourcePlanModifiers derives a's plan modifiers, typed from a.GoType:
// UseStateForUnknown() when a is both Optional and Computed, RequiresReplace()
// when requestSettable and updateUnsupported. A Computed-only attribute gets
// neither, since the server may change such a value during apply.
// requestSettable is the caller's Provenance.InRequest — the fact itself rather
// than a re-derivation from a.Required/a.Optional.
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
// of goType — the single spelling of that modifier.
func RequiresReplaceSpec(goType string) PlanModifierSpec {
	return PlanModifierSpec{Name: PlanModifierPackage(goType) + ".RequiresReplace"}
}

// PlanModifierPackage returns the planmodifier subpackage for goType, e.g.
// "types.String" -> "stringplanmodifier". Every GoType FrameworkType produces
// follows this convention.
func PlanModifierPackage(goType string) string {
	return strings.ToLower(strings.TrimPrefix(goType, "types.")) + "planmodifier"
}

// oneOfModelName is the generated Go struct name for an envelope or variant.
// Deriving it from the envelope name rather than the use site lets two uses of
// one reusable union share a single model.
func oneOfModelName(name string) string { return name + "Model" }

// rootUnionOptional reports whether s is a union its own OpenAPI field left
// optional or nullable — the one presence fact a body root cannot read off an
// enclosing object's required list, having no enclosing object.
func rootUnionOptional(s *Schema) bool {
	return s.Kind == SchemaKindOneOf && s.OneOf != nil && (s.OneOf.Optional || s.OneOf.Nullable)
}
