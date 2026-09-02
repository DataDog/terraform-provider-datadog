package model

import (
	"fmt"
	"slices"
	"sort"
)

// ----------------------------------------------------------------------------
// Schema combination helpers
// ----------------------------------------------------------------------------

// MergeNormalizedSchemas intersects the subset of OpenAPI constraints retained
// by Schema. It combines a oneOf alternative with the constraints declared
// adjacent to the oneOf keyword (its "sibling" schema) — both must hold
// simultaneously, so enums intersect and a kind/type/format conflict makes the
// alternative Unsupported rather than erroring, letting the affected variant
// alone report that precise failure instead of the whole union being dropped.
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
// model wraps its value in a single "value" field (every primitive, array and
// map alternative) rather than exposing its own fields directly (every object
// alternative).
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
		if s.OneOf.Discriminator != nil {
			discriminator := *s.OneOf.Discriminator
			if s.OneOf.Discriminator.Mapping != nil {
				discriminator.Mapping = make(map[string]string, len(s.OneOf.Discriminator.Mapping))
				for key, value := range s.OneOf.Discriminator.Mapping {
					discriminator.Mapping[key] = value
				}
			}
			oneOf.Discriminator = &discriminator
		}
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
// Kind, primitive Type, or Format. An element/value shape conflict (an array
// or map whose element disagrees) surfaces as this same error one level
// deeper, at the "[]"/"{}" path, since Items disagreement is just a Kind (or
// Type/Format) mismatch one recursion step further down — no separate check
// is needed for it.
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
// correlated position. Nodes are correlated by property name at equal depth
// from each body's root, so a JSON:API data.attributes.<field> lines up
// across all three even though the enclosing request/response components
// differ by name. The one exception is a OneOf/Unsupported/RefCycle/
// DepthExceeded node (see mergeVerbatim): its own position is stamped, but
// its subtree is cloned verbatim from the preferred side rather than walked,
// so nodes inside it carry whatever Provenance (typically none) they already
// had.
//
// group.Search and the Create/Update *response* bodies are never read: a
// field only they carry would become Computed state that refresh can never
// repopulate — the search element can be a narrower shape than the by-id
// record, and a create-response-only field is never seen again.
//
// It returns a plain error when group is nil or is missing a Create or Read
// operation.
func MergeResourceSchema(group *ResolvedGroup) (*Schema, []Diagnostic, error) {
	if group == nil || group.Create == nil || group.Read == nil {
		return nil, nil, fmt.Errorf("model: MergeResourceSchema requires a resolved Create and Read operation")
	}

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
	default:
		// OneOf, Unsupported, RefCycle, DepthExceeded: the cosmetic-reconciliation
		// list names RefName/Description/Enum/Sensitive, not oneOf-specific fields,
		// so these kinds are not deep-merged — the merged node is the preferred
		// side's clone with those four fields reconciled the same way as every
		// other kind.
		return m.mergeVerbatim(create, update, read, createRequired, path)
	}
}

func (m *resourceMerger) mergeObject(create, update, read *Schema, createRequired bool, path string) (*Schema, error) {
	properties := make(map[string]*Schema)
	for _, key := range unionObjectKeys(create, update, read) {
		var childCreate, childUpdate, childRead *Schema
		childRequired := false
		if create != nil {
			childCreate = create.Properties[key]
			childRequired = slices.Contains(create.Required, key)
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
	refName, description, enum, sensitive := m.cosmeticFields(create, update, read, path)
	return &Schema{
		Kind:        SchemaKindObject,
		Properties:  properties,
		Required:    requiredFromCreate(create),
		RefName:     refName,
		Description: description,
		Enum:        enum,
		Sensitive:   sensitive,
		Provenance:  stampProvenance(create, update, read, createRequired),
	}, nil
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
	refName, description, enum, sensitive := m.cosmeticFields(create, update, read, path)
	return &Schema{
		Kind:        kind,
		Items:       items,
		RefName:     refName,
		Description: description,
		Enum:        enum,
		Sensitive:   sensitive,
		Provenance:  stampProvenance(create, update, read, createRequired),
	}, nil
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
	refName, description, enum, sensitive := m.cosmeticFields(create, update, read, path)
	return &Schema{
		Kind:        SchemaKindPrimitive,
		Type:        typ,
		Format:      format,
		Enum:        enum,
		RefName:     refName,
		Description: description,
		Sensitive:   sensitive,
		Provenance:  stampProvenance(create, update, read, createRequired),
	}, nil
}

func (m *resourceMerger) mergeVerbatim(create, update, read *Schema, createRequired bool, path string) (*Schema, error) {
	out := CloneSchema(preferredSchema(create, update, read))
	out.RefName, out.Description, out.Enum, out.Sensitive = m.cosmeticFields(create, update, read, path)
	out.Provenance = stampProvenance(create, update, read, createRequired)
	return out, nil
}

// cosmeticFields reconciles the four fields treated as cosmetic —
// RefName and Description favoring the Read response, Enum members unioned
// (never intersected: a validator must accept everything the response can
// return, or refresh fails on a value the practitioner never chose), Sensitive
// the disjunction — and records one info Diagnostic when the present bodies
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
// whether at least two sides carried enum members and at least one of them
// was missing a member another side had — a side whose own deduped enum is
// already the same size as the union necessarily equals it, since it is a
// subset by construction.
func unionEnum(create, update, read *Schema) (values []string, disagreed bool) {
	present := presentSchemas(create, update, read)
	var all []string
	withEnum := 0
	for _, s := range present {
		if len(s.Enum) > 0 {
			withEnum++
			all = append(all, s.Enum...)
		}
	}
	values = sortedUniqueStrings(all)
	if withEnum < 2 {
		return values, false
	}
	for _, s := range present {
		if len(s.Enum) == 0 {
			continue
		}
		if len(sortedUniqueStrings(append([]string(nil), s.Enum...))) != len(values) {
			return values, true
		}
	}
	return values, false
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
