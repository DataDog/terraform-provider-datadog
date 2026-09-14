package model

import (
	"fmt"
	"maps"
	"slices"
	"sort"
)

// ----------------------------------------------------------------------------
// Schema combination helpers
// ----------------------------------------------------------------------------

// MergeNormalizedSchemas intersects the constraints of a oneOf alternative with
// those declared adjacent to the oneOf keyword. Both must hold at once, so
// enums intersect and a kind/type/format conflict yields an Unsupported schema
// carrying the reason rather than an error, confining the failure to the one
// affected alternative.
func MergeNormalizedSchemas(variant, common *Schema) *Schema {
	if variant == nil {
		return common
	}
	if common == nil {
		return variant
	}
	if common.Kind == SchemaKindUnsupported {
		out := CloneSchema(common)
		if out.Description == "" {
			out.Description = variant.Description
		}
		return out
	}
	if variant.Kind == SchemaKindOneOf && variant.OneOf != nil {
		for i := range variant.OneOf.Variants {
			variant.OneOf.Variants[i].Schema = MergeNormalizedSchemas(variant.OneOf.Variants[i].Schema, common)
			variant.OneOf.Variants[i].ValueWrapped = OneOfValueWrapped(variant.OneOf.Variants[i].Schema)
		}
		return variant
	}
	if variant.Kind == SchemaKindUnsupported {
		if variant.UnsupportedReason != "" {
			return variant
		}
		if common.Description == "" {
			common.Description = variant.Description
		}
		return common
	}
	if variant.Kind != common.Kind {
		return &Schema{
			Kind:              SchemaKindUnsupported,
			Description:       variant.Description,
			UnsupportedReason: fmt.Sprintf("oneOf alternative kind %q conflicts with adjacent schema kind %q", variant.Kind, common.Kind),
		}
	}

	switch variant.Kind {
	case SchemaKindObject:
		if variant.Properties == nil {
			variant.Properties = make(map[string]*Schema)
		}
		for key, commonProperty := range common.Properties {
			if property, exists := variant.Properties[key]; exists {
				variant.Properties[key] = MergeNormalizedSchemas(property, commonProperty)
			} else {
				variant.Properties[key] = commonProperty
			}
		}
		variant.Required = sortedUniqueStrings(append(variant.Required, common.Required...))
	case SchemaKindArray, SchemaKindMap:
		variant.Items = MergeNormalizedSchemas(variant.Items, common.Items)
	case SchemaKindPrimitive:
		if variant.Type != "" && common.Type != "" && variant.Type != common.Type {
			return &Schema{
				Kind:              SchemaKindUnsupported,
				Description:       variant.Description,
				UnsupportedReason: fmt.Sprintf("oneOf alternative type %q conflicts with adjacent type %q", variant.Type, common.Type),
			}
		}
		if variant.Type == "" {
			variant.Type = common.Type
		}
		if variant.Format != "" && common.Format != "" && variant.Format != common.Format {
			return &Schema{
				Kind:              SchemaKindUnsupported,
				Description:       variant.Description,
				UnsupportedReason: fmt.Sprintf("oneOf alternative format %q conflicts with adjacent format %q", variant.Format, common.Format),
			}
		}
		if variant.Format == "" {
			variant.Format = common.Format
		}
		switch {
		case len(variant.Enum) == 0:
			variant.Enum = append([]string(nil), common.Enum...)
		case len(common.Enum) > 0:
			intersection := intersectStrings(variant.Enum, common.Enum)
			if len(intersection) == 0 {
				return &Schema{
					Kind:              SchemaKindUnsupported,
					Description:       variant.Description,
					UnsupportedReason: "oneOf alternative enum has no values in common with adjacent enum",
				}
			}
			variant.Enum = intersection
		}
	}
	variant.Sensitive = variant.Sensitive || common.Sensitive
	return variant
}

// OneOfValueWrapped reports whether a oneOf alternative's Terraform variant
// model wraps its value in a single "value" field (primitive, array and map
// alternatives) rather than exposing its own fields directly (objects).
func OneOfValueWrapped(schema *Schema) bool {
	if schema == nil {
		return false
	}
	switch schema.Kind {
	case SchemaKindPrimitive, SchemaKindArray, SchemaKindMap:
		return true
	default:
		return false
	}
}

// CloneSchema returns a deep copy of s, including Items, Properties, Variants
// and OneOf.
func CloneSchema(s *Schema) *Schema {
	if s == nil {
		return nil
	}
	out := *s
	out.Enum = append([]string(nil), s.Enum...)
	out.Required = append([]string(nil), s.Required...)
	out.Items = CloneSchema(s.Items)
	if s.Properties != nil {
		out.Properties = make(map[string]*Schema, len(s.Properties))
		for name, child := range s.Properties {
			out.Properties[name] = CloneSchema(child)
		}
	}
	if s.Variants != nil {
		out.Variants = make([]*Schema, len(s.Variants))
		for i, variant := range s.Variants {
			out.Variants[i] = CloneSchema(variant)
		}
	}
	if s.OneOf != nil {
		oneOf := *s.OneOf
		oneOf.Variants = make([]OneOfVariant, len(s.OneOf.Variants))
		for i, variant := range s.OneOf.Variants {
			oneOf.Variants[i] = variant
			oneOf.Variants[i].Schema = CloneSchema(variant.Schema)
		}
		oneOf.Discriminator = cloneDiscriminator(s.OneOf.Discriminator)
		out.OneOf = &oneOf
	}
	return &out
}

func sortedUniqueStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	sort.Strings(values)
	return slices.Compact(values)
}

func intersectStrings(left, right []string) []string {
	allowed := make(map[string]struct{}, len(right))
	for _, value := range right {
		allowed[value] = struct{}{}
	}
	var intersection []string
	for _, value := range left {
		if _, ok := allowed[value]; ok {
			intersection = append(intersection, value)
		}
	}
	return sortedUniqueStrings(intersection)
}

// ----------------------------------------------------------------------------
// Resource schema merge
// ----------------------------------------------------------------------------

// SchemaMergeError reports a structural disagreement among the Create request,
// Update request and Read response bodies at one schema path: a differing
// Kind, primitive Type, or Format. An array/map element conflict surfaces as
// this same error one recursion deeper, at the "[]"/"{}" path.
type SchemaMergeError struct {
	// Path is the schema path where the bodies disagree, dot-delimited from
	// the merged tree's root, with "[]"/"{}" for an array/map element.
	Path string
	// Aspect names what disagreed: "kind", "type", or "format".
	Aspect string
	// Left and Right are the two conflicting spellings, e.g. "object" and
	// "primitive", or "string" and "integer".
	Left, Right string
}

func (e *SchemaMergeError) Error() string {
	return fmt.Sprintf("model: resource schema merge conflict at %q: %s %q vs %q", e.Path, e.Aspect, e.Left, e.Right)
}

// MergeResourceSchema unions the Create request, Update request and Read
// response bodies of group into one Schema tree, stamping Provenance at every
// correlated position. Nodes correlate by property name at equal depth from
// each body's root; a OneOf/Unsupported/RefCycle/DepthExceeded node instead has
// its subtree cloned verbatim from the preferred side (see mergeVerbatim).
// group.Search and the Create/Update responses are never read: a field only
// they carry would become Computed state refresh can never repopulate. Assumes
// group.Create and group.Read resolve.
func MergeResourceSchema(group *ResolvedGroup) (*Schema, []Diagnostic, error) {
	var updateRequest *Schema
	if group.Update != nil {
		updateRequest = group.Update.RequestSchema
	}

	m := &resourceMerger{}
	merged, err := m.mergeNode(group.Create.RequestSchema, updateRequest, group.Read.ResponseSchema, false, "")
	if err != nil {
		return nil, nil, err
	}
	return merged, m.diagnostics, nil
}

// resourceMerger accumulates the info diagnostics raised while walking the
// three bodies.
type resourceMerger struct {
	diagnostics []Diagnostic
}

// mergeNode combines the three bodies' schemas at one correlated tree
// position. create/update/read are nil when that body does not reach this
// position; createRequired is fixed by the caller from the enclosing object's
// Create-body Required list (a node cannot answer this about itself).
func (m *resourceMerger) mergeNode(create, update, read *Schema, createRequired bool, path string) (*Schema, error) {
	kind, err := kindConflict(create, update, read, path)
	if err != nil {
		return nil, err
	}
	switch kind {
	case SchemaKindObject:
		return m.mergeObject(create, update, read, createRequired, path)
	case SchemaKindArray, SchemaKindMap:
		return m.mergeCollection(kind, create, update, read, createRequired, path)
	case SchemaKindPrimitive:
		return m.mergePrimitive(create, update, read, createRequired, path)
	case SchemaKindOneOf:
		return m.mergeOneOf(create, update, read, createRequired, path)
	default:
		// Unsupported, RefCycle, DepthExceeded: nothing under such a node is
		// representable, so there is no subtree worth correlating — the merged node
		// is the preferred side's clone.
		return m.mergeVerbatim(create, update, read, createRequired, path)
	}
}

func (m *resourceMerger) mergeObject(create, update, read *Schema, createRequired bool, path string) (*Schema, error) {
	// The create body's required list is consulted once per property, so index
	// it up front rather than rescanning the slice for every key.
	createRequiredKeys := map[string]struct{}{}
	if create != nil {
		for _, key := range create.Required {
			createRequiredKeys[key] = struct{}{}
		}
	}

	properties := make(map[string]*Schema)
	for _, key := range unionObjectKeys(create, update, read) {
		var childCreate, childUpdate, childRead *Schema
		childRequired := false
		if create != nil {
			childCreate = create.Properties[key]
			_, childRequired = createRequiredKeys[key]
		}
		if update != nil {
			childUpdate = update.Properties[key]
		}
		if read != nil {
			childRead = read.Properties[key]
		}
		child, err := m.mergeNode(childCreate, childUpdate, childRead, childRequired, ChildPath(path, key))
		if err != nil {
			return nil, err
		}
		properties[key] = child
	}
	return m.stampCommon(&Schema{
		Kind:       SchemaKindObject,
		Properties: properties,
		Required:   requiredFromCreate(create),
	}, create, update, read, createRequired, path), nil
}

// stampCommon fills the fields every merged node carries whatever its kind: the
// cosmetic trio (RefName, Description, Enum, Sensitive), the request-side
// component name, and the provenance stamp. It returns out, so a merge can
// construct its kind-specific fields and stamp the rest in one expression.
func (m *resourceMerger) stampCommon(out, create, update, read *Schema, createRequired bool, path string) *Schema {
	out.RefName, out.Description, out.Enum, out.Sensitive = m.cosmeticFields(create, update, read, path)
	out.RequestRefName = pickRequestRefName(create, update)
	out.Provenance = stampProvenance(create, update, read, createRequired)
	return out
}

func (m *resourceMerger) mergeCollection(kind SchemaKind, create, update, read *Schema, createRequired bool, path string) (*Schema, error) {
	var childCreate, childUpdate, childRead *Schema
	if create != nil {
		childCreate = create.Items
	}
	if update != nil {
		childUpdate = update.Items
	}
	if read != nil {
		childRead = read.Items
	}
	suffix := "[]"
	if kind == SchemaKindMap {
		suffix = "{}"
	}
	// An array/map element has no name of its own to appear in a parent's
	// Required list, so it is never itself request-required.
	items, err := m.mergeNode(childCreate, childUpdate, childRead, false, ChildPath(path, suffix))
	if err != nil {
		return nil, err
	}
	return m.stampCommon(&Schema{
		Kind:  kind,
		Items: items,
	}, create, update, read, createRequired, path), nil
}

func (m *resourceMerger) mergePrimitive(create, update, read *Schema, createRequired bool, path string) (*Schema, error) {
	present := presentSchemas(create, update, read)
	typ, err := reconcileField(path, "type", present, func(s *Schema) string { return s.Type })
	if err != nil {
		return nil, err
	}
	format, err := reconcileField(path, "format", present, func(s *Schema) string { return s.Format })
	if err != nil {
		return nil, err
	}
	return m.stampCommon(&Schema{
		Kind:   SchemaKindPrimitive,
		Type:   typ,
		Format: format,
	}, create, update, read, createRequired, path), nil
}

func (m *resourceMerger) mergeVerbatim(create, update, read *Schema, createRequired bool, path string) (*Schema, error) {
	out := CloneSchema(preferredSchema(create, update, read))
	return m.stampCommon(out, create, update, read, createRequired, path), nil
}

// pickRequestRefName returns the Create body's component name at this node,
// falling back to Update's. Unlike RefName it never prefers Read: a response
// component is routinely named differently from the request one for the same
// field ("...Request" vs "...Response"), so Read's name cannot substitute.
func pickRequestRefName(create, update *Schema) string {
	if create != nil && create.RefName != "" {
		return create.RefName
	}
	if update != nil && update.RefName != "" {
		return update.RefName
	}
	return ""
}

// cosmeticFields reconciles the four cosmetic fields — RefName and Description
// favoring the Read response, Enum members unioned (never intersected: a
// validator must accept everything the response can return), Sensitive the
// disjunction — and records one info Diagnostic when the present bodies
// actually disagreed on any of them.
func (m *resourceMerger) cosmeticFields(create, update, read *Schema, path string) (refName, description string, enum []string, sensitive bool) {
	var refDisagree, descDisagree, enumDisagree, sensDisagree bool
	refName, refDisagree = pickString(func(s *Schema) string { return s.RefName }, create, update, read)
	description, descDisagree = pickString(func(s *Schema) string { return s.Description }, create, update, read)
	enum, enumDisagree = unionEnum(create, update, read)
	sensitive, sensDisagree = anySensitive(create, update, read)

	if refDisagree || descDisagree || enumDisagree || sensDisagree {
		m.diagnostics = append(m.diagnostics, Diagnostic{
			Severity: SeverityInfo,
			Message: fmt.Sprintf(
				"resource schema merge: reconciled a cosmetic difference between the create/update/read bodies at %q — favoring the read response for name/description, unioning enum members, OR-ing sensitive",
				path,
			),
		})
	}
	return refName, description, enum, sensitive
}

func stampProvenance(create, update, read *Schema, createRequired bool) *SchemaProvenance {
	return &SchemaProvenance{
		InRequest:       create != nil || update != nil,
		RequestRequired: createRequired,
		InResponse:      read != nil,
	}
}

// kindConflict reports the Kind every present (non-nil) body agrees on, or a
// SchemaMergeError naming the two conflicting spellings.
func kindConflict(create, update, read *Schema, path string) (SchemaKind, error) {
	present := presentSchemas(create, update, read)
	kind, err := reconcileField(path, "kind", present, func(s *Schema) string { return string(s.Kind) })
	if err != nil {
		return "", err
	}
	return SchemaKind(kind), nil
}

// reconcileField picks the one value every present body that sets it agrees
// on for a structural aspect ("kind", "type", or "format") — skipping a body
// that doesn't set it at all — and returns a SchemaMergeError naming both
// spellings the moment two present bodies disagree.
func reconcileField(path, aspect string, present []*Schema, get func(*Schema) string) (string, error) {
	var value string
	for _, s := range present {
		switch v := get(s); {
		case v == "":
		case value == "":
			value = v
		case v != value:
			return "", &SchemaMergeError{Path: path, Aspect: aspect, Left: value, Right: v}
		}
	}
	return value, nil
}

func presentSchemas(create, update, read *Schema) []*Schema {
	var out []*Schema
	if create != nil {
		out = append(out, create)
	}
	if update != nil {
		out = append(out, update)
	}
	if read != nil {
		out = append(out, read)
	}
	return out
}

func unionObjectKeys(create, update, read *Schema) []string {
	seen := make(map[string]struct{})
	for _, s := range presentSchemas(create, update, read) {
		for key := range s.Properties {
			seen[key] = struct{}{}
		}
	}
	keys := make([]string, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// ChildPath joins parent and child into a dot-delimited schema path, e.g.
// "data.attributes.name", except when child is the array/map element marker
// ("[]"/"{}"), which appends without a dot.
func ChildPath(parent, child string) string {
	if parent == "" {
		return child
	}
	if child == "[]" || child == "{}" {
		return parent + child
	}
	return parent + "." + child
}

func requiredFromCreate(create *Schema) []string {
	if create == nil {
		return nil
	}
	return sortedUniqueStrings(append([]string(nil), create.Required...))
}

// preferenceOrder returns Read, Create, Update in that order — read wins a
// disagreement, else create, else update. An entry is nil when that side is
// absent.
func preferenceOrder(create, update, read *Schema) [3]*Schema {
	return [3]*Schema{read, create, update}
}

// preferredSchema returns the first present side in preferenceOrder.
func preferredSchema(create, update, read *Schema) *Schema {
	for _, s := range preferenceOrder(create, update, read) {
		if s != nil {
			return s
		}
	}
	return nil
}

// pickString returns get's value, preferring the Read side, then Create, then
// Update, skipping an empty value even from the preferred side. disagreed
// reports whether the present sides actually carried more than one distinct
// non-empty value, so the caller knows whether a reconciliation happened or
// every side already agreed.
func pickString(get func(*Schema) string, create, update, read *Schema) (value string, disagreed bool) {
	seen := map[string]bool{}
	for _, s := range presentSchemas(create, update, read) {
		if v := get(s); v != "" {
			seen[v] = true
		}
	}
	disagreed = len(seen) > 1
	for _, s := range preferenceOrder(create, update, read) {
		if s == nil {
			continue
		}
		if v := get(s); v != "" {
			return v, disagreed
		}
	}
	return "", disagreed
}

// unionEnum unions the Enum members of every present side. disagreed reports
// whether at least two sides carried enum members and at least one of them was
// missing a member another side had.
func unionEnum(create, update, read *Schema) (values []string, disagreed bool) {
	var all []string
	withEnum := 0
	// A side's deduped enum is a subset of the union by construction, so a side
	// smaller than the union is missing a member another side had; tracking the
	// largest here avoids re-deduping every side in a second pass.
	maxDeduped := 0
	for _, s := range presentSchemas(create, update, read) {
		if len(s.Enum) == 0 {
			continue
		}
		withEnum++
		all = append(all, s.Enum...)
		if n := len(sortedUniqueStrings(slices.Clone(s.Enum))); n > maxDeduped {
			maxDeduped = n
		}
	}
	values = sortedUniqueStrings(all)
	if withEnum < 2 {
		return values, false
	}
	return values, maxDeduped != len(values)
}

func anySensitive(create, update, read *Schema) (sensitive, disagreed bool) {
	var sawTrue, sawFalse bool
	for _, s := range presentSchemas(create, update, read) {
		if s.Sensitive {
			sawTrue = true
		} else {
			sawFalse = true
		}
	}
	return sawTrue, sawTrue && sawFalse
}

// ----------------------------------------------------------------------------
// oneOf merge
// ----------------------------------------------------------------------------

// OneOfMergeError reports a union the Create request, Update request and Read
// response bodies describe in ways that cannot be correlated into one
// Terraform envelope: the bodies list alternatives that do not line up even
// after their CRUD-role suffixes are removed, or one body names two
// alternatives that collapse onto the same stripped name.
type OneOfMergeError struct {
	// Path is the union's schema path in the merged tree.
	Path string
	// Reason states what could not be correlated.
	Reason string
	// Create, Update and Read are each body's alternatives as that body spells
	// them, sorted; nil when the body does not reach this node. Deliberately the
	// pre-strip names, so the message shows what the specification says.
	Create, Update, Read []string
}

func (e *OneOfMergeError) Error() string {
	return fmt.Sprintf(
		"model: resource schema merge cannot correlate the oneOf at %q: %s (create: %v, update: %v, read: %v)",
		e.Path, e.Reason, e.Create, e.Update, e.Read)
}

// mergeOneOf unions the three bodies' spellings of one union. Unlike the other
// non-object kinds it is deep-merged, because its alternatives are objects
// whose properties differ by role exactly as a nested object's do (a password
// required on Create, optional on Update, absent on the response), so each
// correlated alternative goes back through mergeNode and comes out with
// Provenance at every property. The name alternatives correlate under is also
// the name the variant block is published under, so the two cannot drift. The
// SDK binding stays the preferred (Read) body's, as RefName does.
func (m *resourceMerger) mergeOneOf(create, update, read *Schema, createRequired bool, path string) (*Schema, error) {
	// A node classified oneOf but carrying no normalized union has no
	// alternatives to correlate. That defect is reported elsewhere with its own
	// actionable message, so hand it on untouched.
	for _, s := range presentSchemas(create, update, read) {
		if s.OneOf == nil {
			return m.mergeVerbatim(create, update, read, createRequired, path)
		}
	}

	sides, names, err := correlateOneOf(create, update, read, path)
	if err != nil {
		return nil, err
	}
	// The preferred body does not change per alternative — only the lookup key
	// does — so resolve it once.
	preferred := sides.preferred()

	spec := &OneOfSpec{Variants: make([]OneOfVariant, 0, len(names))}
	for _, name := range names {
		altCreate, altUpdate, altRead := sides.create[name].Schema, sides.update[name].Schema, sides.read[name].Schema
		if altCreate == nil && altUpdate == nil && altRead == nil {
			// mergeNode has no side to prefer and would clone nil.
			return nil, &OneOfMergeError{
				Path:   path,
				Reason: fmt.Sprintf("alternative %q has no normalized schema in any body", name),
			}
		}
		// An alternative is a choice, never an entry in an enclosing object's
		// required list, so it is never itself request-required.
		merged, err := m.mergeNode(altCreate, altUpdate, altRead, false, ChildPath(path, name))
		if err != nil {
			return nil, err
		}
		source := preferred[name]
		spec.Variants = append(spec.Variants, OneOfVariant{
			TFName:         name,
			GoName:         SdkName(name),
			Schema:         merged,
			RefName:        source.RefName,
			SDKField:       source.SDKField,
			SDKConstructor: source.SDKConstructor,
			SDKPointer:     source.SDKPointer,
			ValueWrapped:   OneOfValueWrapped(merged),
		})
	}

	preferredSpec := preferredSchema(create, update, read).OneOf
	// Name is the envelope's generated-model identity, stripped for the same
	// reason a variant's block name is: a resource that later gains an Update
	// endpoint must not rename a struct it already emitted.
	spec.Name = StripOneOfRoleSuffix(preferredSpec.Name)
	spec.Path = preferredSpec.Path
	spec.RefName = preferredSpec.RefName
	spec.SDKType = preferredSpec.SDKType
	spec.Discriminator = cloneDiscriminator(preferredSpec.Discriminator)
	// Absence is permitted wherever any body permits it: a union the Read
	// response may omit must not make refresh fail. The Create body's own
	// requirement travels separately, in the enclosing object's required list.
	for _, s := range presentSchemas(create, update, read) {
		spec.Optional = spec.Optional || s.OneOf.Optional
		spec.Nullable = spec.Nullable || s.OneOf.Nullable
	}

	return m.stampCommon(&Schema{
		Kind:  SchemaKindOneOf,
		OneOf: spec,
	}, create, update, read, createRequired, path), nil
}

// correlateOneOf lines the three bodies' alternatives up under one name each,
// returning a create/update/read triple of name-keyed alternatives plus the
// sorted names they agreed on. A nil map is a body that does not reach the
// union. The bodies' own names are tried first, since a spelling every body
// shares is already role-independent; only if they disagree is
// StripOneOfRoleSuffix applied, to every side at once so the comparison stays
// symmetric.
func correlateOneOf(create, update, read *Schema, path string) (sides oneOfSides, names []string, err error) {
	for _, strip := range []bool{false, true} {
		if sides, err = indexOneOfSides(create, update, read, path, strip); err != nil {
			return sides, nil, err
		}
		if names = sides.names(); sides.agree(len(names)) {
			return sides, names, nil
		}
	}
	return sides, nil, &OneOfMergeError{
		Path: path,
		Reason: "the bodies that reach this union do not list the same alternatives, " +
			"even after their CRUD-role suffixes are removed",
		Create: alternativeNames(create),
		Update: alternativeNames(update),
		Read:   alternativeNames(read),
	}
}

// oneOfSides holds the three bodies' alternatives keyed by the name they are
// being correlated under. A nil map is a body that does not reach the union —
// "no opinion", not "no alternatives".
type oneOfSides struct{ create, update, read map[string]OneOfVariant }

func indexOneOfSides(create, update, read *Schema, path string, strip bool) (oneOfSides, error) {
	var sides oneOfSides
	var err error
	if sides.create, err = oneOfAlternativesByName(create, path, strip); err != nil {
		return sides, err
	}
	if sides.update, err = oneOfAlternativesByName(update, path, strip); err != nil {
		return sides, err
	}
	sides.read, err = oneOfAlternativesByName(read, path, strip)
	return sides, err
}

// preferred returns the body whose alternatives supply the merged node's
// cosmetic and SDK-facing fields: Read, then Create, then Update — the same
// rule preferenceOrder states for schemas.
func (s oneOfSides) preferred() map[string]OneOfVariant {
	switch {
	case s.read != nil:
		return s.read
	case s.create != nil:
		return s.create
	default:
		return s.update
	}
}

// names is every alternative any body carries, sorted so the merged variant
// order cannot depend on which body was walked first.
func (s oneOfSides) names() []string {
	all := map[string]struct{}{}
	for _, side := range []map[string]OneOfVariant{s.create, s.update, s.read} {
		for name := range side {
			all[name] = struct{}{}
		}
	}
	return slices.Sorted(maps.Keys(all))
}

// agree reports whether every body that reaches the union carries all of the
// alternatives. Each side's names are unique by construction, so having the
// full count is the same as having the full set.
func (s oneOfSides) agree(total int) bool {
	for _, side := range []map[string]OneOfVariant{s.create, s.update, s.read} {
		if side != nil && len(side) != total {
			return false
		}
	}
	return true
}

// oneOfAlternativesByName indexes one body's alternatives by name, optionally
// role-stripped, rejecting a body whose own alternatives collapse onto one
// name — two spellings of one variant block cannot both be published.
func oneOfAlternativesByName(s *Schema, path string, strip bool) (map[string]OneOfVariant, error) {
	if s == nil {
		return nil, nil
	}
	out := make(map[string]OneOfVariant, len(s.OneOf.Variants))
	for _, variant := range s.OneOf.Variants {
		name := variant.TFName
		if strip {
			name = StripOneOfRoleSuffix(name)
		}
		if _, duplicate := out[name]; duplicate {
			return nil, &OneOfMergeError{
				Path:   path,
				Reason: fmt.Sprintf("one body has two alternatives whose role-independent name is %q", name),
			}
		}
		out[name] = variant
	}
	return out, nil
}

// alternativeNames lists one body's alternatives as that body spells them,
// for a diagnostic. Nil when the body does not reach the union.
func alternativeNames(s *Schema) []string {
	if s == nil || s.OneOf == nil {
		return nil
	}
	names := make([]string, 0, len(s.OneOf.Variants))
	for _, variant := range s.OneOf.Variants {
		names = append(names, variant.TFName)
	}
	return sortedUniqueStrings(names)
}

// cloneDiscriminator deep-copies a union's discriminator so the merged spec
// does not share a Mapping with the body it was preferred from.
func cloneDiscriminator(d *OneOfDiscriminator) *OneOfDiscriminator {
	if d == nil {
		return nil
	}
	out := *d
	if d.Mapping != nil {
		out.Mapping = make(map[string]string, len(d.Mapping))
		for key, value := range d.Mapping {
			out.Mapping[key] = value
		}
	}
	return &out
}
