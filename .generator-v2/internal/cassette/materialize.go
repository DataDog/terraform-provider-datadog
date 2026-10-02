package cassette

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// ----------------------------------------------------------------------------
// Materialization
//
// Turns a selected example set into the typed values a cassette interaction and
// its generated configuration need. Selection answered "which example"; this
// answers "is it complete", and refuses to make up the difference when it is
// not.
//
// Nothing here invents a value. An optional field with no declared example is
// omitted, not defaulted; an array or map with no example stays absent rather
// than becoming empty; a union with more than one possible branch is reported
// rather than guessed. An invented value would produce a fixture that replays
// perfectly and asserts something the API never said.
// ----------------------------------------------------------------------------

// MaterializedSet is one example set turned into typed values.
type MaterializedSet struct {
	// Body is the API-shaped payload: what the request sends or the response
	// returns, after read-only filtering and sensitive replacement.
	Body any
	// Values are the leaf paths and typed values, carried separately so the
	// HCL renderer can map them through the artifact's own schema rather than
	// re-deriving the envelope flattening here.
	Values []model.MaterializedValue
	// SensitiveReplacements maps a leaf path to the safe value substituted
	// there, so a writer can prove no declared secret survived.
	SensitiveReplacements map[string]string
	// WriteOnlyPaths lists the request leaves the description marks writeOnly.
	// Recorded here, where the path strings are built, so it cannot disagree
	// with Values about what a leaf is called. Consumed by overlaidWith.
	WriteOnlyPaths []string
}

// IncompleteError reports that a set's examples cannot form a complete value.
// It lists the paths rather than just failing, because the remedy is always a
// change to the description and the author needs to know which leaves to fill.
type IncompleteError struct {
	Key SetKey
	// Missing are required leaf paths for which no example or default exists.
	Missing []string
	// Ambiguous are union paths offering more than one branch with nothing to
	// choose between them.
	Ambiguous []string
}

func (e *IncompleteError) Error() string {
	var parts []string
	if len(e.Missing) > 0 {
		parts = append(parts, fmt.Sprintf("no example or default for required %s", joinAnd(e.Missing)))
	}
	if len(e.Ambiguous) > 0 {
		parts = append(parts, fmt.Sprintf("more than one branch possible at %s", joinAnd(e.Ambiguous)))
	}
	return fmt.Sprintf("%s cannot be materialized: %s", e.Key, strings.Join(parts, "; "))
}

// MaterializeSet turns one selected set into typed values against its
// normalized schema.
//
// A whole example takes precedence over assembling property pieces: it is a
// value a human wrote for this operation, so it is used as declared and only
// checked for completeness. Property pieces are assembled recursively and
// succeed only when every required leaf has an example or a usable default.
//
// Read-only values are excluded from a request — sending a server-assigned
// field back is not what the provider does — but kept in a response, where the
// server legitimately returns them.
func MaterializeSet(set SelectedSet, schema *model.Schema) (MaterializedSet, error) {
	out := MaterializedSet{SensitiveReplacements: map[string]string{}}
	if !set.Resolved() {
		return out, &IncompleteError{Key: set.Key, Missing: []string{"the whole value"}}
	}

	walker := &materializer{
		isRequest: set.Key.Role == SetRoleRequest,
		byPath:    fallbacksByPath(set.Fallbacks),
		out:       &out,
	}

	var body any
	if set.Candidate != nil {
		body = walker.fromDeclared(set.Candidate.Value, schema, "")
	} else {
		body = walker.assemble(schema, "")
	}

	if len(walker.missing) > 0 || len(walker.ambiguous) > 0 {
		return MaterializedSet{}, &IncompleteError{
			Key:       set.Key,
			Missing:   walker.missing,
			Ambiguous: walker.ambiguous,
		}
	}
	out.Body = body
	slices.SortStableFunc(out.Values, func(a, b model.MaterializedValue) int {
		return strings.Compare(a.Path, b.Path)
	})
	if len(out.SensitiveReplacements) == 0 {
		out.SensitiveReplacements = nil
	}
	return out, nil
}

// fallbacksByPath indexes property candidates by their dotted path.
func fallbacksByPath(fallbacks []*model.ExampleCandidate) map[string]*model.ExampleCandidate {
	if len(fallbacks) == 0 {
		return nil
	}
	out := make(map[string]*model.ExampleCandidate, len(fallbacks))
	for _, candidate := range fallbacks {
		if candidate == nil || candidate.Location.PropertyPath == "" {
			continue
		}
		out[candidate.Location.PropertyPath] = candidate
	}
	return out
}

// materializer carries the walk's accumulating state.
type materializer struct {
	// isRequest switches on request-side filtering: read-only values are
	// dropped and write-only values are replaced.
	isRequest bool
	byPath    map[string]*model.ExampleCandidate
	out       *MaterializedSet
	missing   []string
	ambiguous []string
}

// fromDeclared walks a value a human declared against its schema, filtering
// and replacing but never adding. A field the example omits stays omitted:
// the author's example is the scenario, and completing it would change what
// the fixture asserts.
func (m *materializer) fromDeclared(value any, schema *model.Schema, path string) any {
	if schema == nil {
		return value
	}
	switch {
	case schema.Kind == model.SchemaKindOneOf:
		return m.declaredOneOf(value, schema, path)
	case schema.Kind == model.SchemaKindObject && schema.Properties != nil:
		return m.declaredObject(value, schema, path)
	case schema.Kind == model.SchemaKindArray && schema.Items != nil:
		return m.declaredArray(value, schema, path)
	default:
		return m.leaf(value, schema, path)
	}
}

func (m *materializer) declaredObject(value any, schema *model.Schema, path string) any {
	declared, ok := value.(map[string]any)
	if !ok {
		// The example disagrees with the schema's shape. Reported as missing
		// rather than coerced: a coerced body would serialize into something
		// the API never described.
		m.missing = append(m.missing, labelPath(path))
		return nil
	}
	out := map[string]any{}
	for _, name := range slices.Sorted(maps.Keys(schema.Properties)) {
		property := schema.Properties[name]
		childPath := joinDotted(path, name)
		present, declaredValue := lookup(declared, name)

		if m.isRequest && property.ReadOnly {
			// Server-assigned; the provider never sends it.
			continue
		}
		if !present {
			// A required field the example omits is incomplete either way: in
			// a request the API would reject it, in a response the generated
			// updateState would leave provider state under-populated.
			if slices.Contains(schema.Required, name) {
				m.missing = append(m.missing, labelPath(childPath))
			}
			continue
		}
		out[name] = m.fromDeclared(declaredValue, property, childPath)
	}
	// Keys the schema does not describe are dropped rather than passed
	// through: the request must validate against the normalized schema.
	return out
}

func (m *materializer) declaredArray(value any, schema *model.Schema, path string) any {
	items, ok := value.([]any)
	if !ok {
		m.missing = append(m.missing, labelPath(path))
		return nil
	}
	out := make([]any, 0, len(items))
	for i, item := range items {
		out = append(out, m.fromDeclared(item, schema.Items, fmt.Sprintf("%s[%d]", path, i)))
	}
	return out
}

// declaredOneOf keeps a declared union value as-is. The example already chose a
// branch, which is precisely the choice materialization must not make itself.
func (m *materializer) declaredOneOf(value any, schema *model.Schema, path string) any {
	if variants := oneOfVariants(schema); len(variants) == 1 {
		return m.fromDeclared(value, variants[0], path)
	}
	// With a declared value the branch is settled even when several exist, so
	// the value passes through unfiltered rather than being reported ambiguous.
	m.leafRecord(value, path, false)
	return value
}

// assemble builds a value from property pieces, which is only sound when every
// required leaf is covered.
func (m *materializer) assemble(schema *model.Schema, path string) any {
	if schema == nil {
		return nil
	}
	switch {
	case schema.Kind == model.SchemaKindOneOf:
		variants := oneOfVariants(schema)
		if len(variants) != 1 {
			// Picking a branch would invent a shape the description never
			// committed to.
			m.ambiguous = append(m.ambiguous, labelPath(path))
			return nil
		}
		return m.assemble(variants[0], path)

	case schema.Kind == model.SchemaKindObject && schema.Properties != nil:
		out := map[string]any{}
		for _, name := range slices.Sorted(maps.Keys(schema.Properties)) {
			property := schema.Properties[name]
			childPath := joinDotted(path, name)
			if m.isRequest && property.ReadOnly {
				continue
			}
			required := slices.Contains(schema.Required, name)
			assembled, ok := m.assembleChild(property, childPath, required)
			if ok {
				out[name] = assembled
			}
		}
		return out

	case schema.Kind == model.SchemaKindArray || schema.Kind == model.SchemaKindMap:
		// An empty array or map is a claim about the API's behavior, not an
		// absence of information. Never invented.
		return nil

	default:
		// A top-level scalar is one leaf, and a required one: the same probe
		// order assembleChild's default arm already applies.
		value, _ := m.assembleChild(schema, path, true)
		return value
	}
}

// assembleChild materializes one property, reporting a required one it cannot
// cover and silently omitting an optional one.
func (m *materializer) assembleChild(schema *model.Schema, path string, required bool) (any, bool) {
	switch schema.Kind {
	case model.SchemaKindObject, model.SchemaKindOneOf:
		before := len(m.missing)
		assembled := m.assemble(schema, path)
		if len(m.missing) > before {
			return nil, false
		}
		if nested, ok := assembled.(map[string]any); ok && len(nested) == 0 && !required {
			// An optional object nothing populated is omitted, not sent empty.
			return nil, false
		}
		return assembled, true

	case model.SchemaKindArray, model.SchemaKindMap:
		if candidate, ok := m.byPath[path]; ok {
			return m.leaf(candidate.Value, schema, path), true
		}
		if required {
			m.missing = append(m.missing, labelPath(path))
		}
		return nil, false

	default:
		if candidate, ok := m.byPath[path]; ok {
			return m.leaf(candidate.Value, schema, path), true
		}
		if value, ok := schemaDefaultValue(schema); ok {
			return m.leaf(value, schema, path), true
		}
		if required {
			m.missing = append(m.missing, labelPath(path))
		}
		return nil, false
	}
}

// leaf records a scalar value, replacing it when the schema marks it
// credential material.
func (m *materializer) leaf(value any, schema *model.Schema, path string) any {
	sensitive := m.isRequest && schema != nil && (schema.WriteOnlySecret || schema.Sensitive)
	if m.isRequest && schema != nil && schema.WriteOnlySecret {
		m.out.WriteOnlyPaths = append(m.out.WriteOnlyPaths, labelPath(path))
	}
	if sensitive {
		if _, isString := value.(string); isString {
			m.out.SensitiveReplacements[labelPath(path)] = model.RedactedPlaceholder
			m.leafRecord(model.RedactedPlaceholder, path, true)
			return model.RedactedPlaceholder
		}
		// A non-string secret keeps its declared value here and is recorded as
		// sensitive; producing a schema-valid replacement for every type is
		// sanitization's job, not materialization's.
		m.leafRecord(value, path, true)
		return value
	}
	m.leafRecord(value, path, false)
	return value
}

func (m *materializer) leafRecord(value any, path string, sensitive bool) {
	m.out.Values = append(m.out.Values, model.MaterializedValue{
		Path:      labelPath(path),
		Value:     value,
		Sensitive: sensitive,
	})
}

// ----------------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------------

// oneOfVariants returns a union's non-null alternatives.
func oneOfVariants(schema *model.Schema) []*model.Schema {
	if schema == nil {
		return nil
	}
	if schema.OneOf != nil && len(schema.OneOf.Variants) > 0 {
		out := make([]*model.Schema, 0, len(schema.OneOf.Variants))
		for _, variant := range schema.OneOf.Variants {
			if variant.Schema != nil {
				out = append(out, variant.Schema)
			}
		}
		return out
	}
	return schema.Variants
}

// schemaDefaultValue returns a schema's usable default, or a single-member
// enum's only possible value. A one-member enum is not a guess: the
// description permits nothing else.
func schemaDefaultValue(schema *model.Schema) (any, bool) {
	if schema == nil {
		return nil, false
	}
	if schema.HasDefault && schema.Default.Value != nil {
		switch schema.Default.Value.Kind {
		case model.ScalarDefaultString:
			return schema.Default.Value.StringValue, true
		case model.ScalarDefaultBool:
			return schema.Default.Value.BoolValue, true
		case model.ScalarDefaultInt64:
			return int(schema.Default.Value.Int64Value), true
		case model.ScalarDefaultFloat64:
			return schema.Default.Value.Float64Value, true
		}
	}
	if len(schema.Enum) == 1 {
		return schema.Enum[0], true
	}
	return nil, false
}

// lookup reads a key, reporting whether it was present so a declared null is
// distinguishable from an omitted field.
func lookup(in map[string]any, name string) (bool, any) {
	value, ok := in[name]
	return ok, value
}

func joinDotted(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}

// labelPath renders the root as something a diagnostic can name.
func labelPath(path string) string {
	if path == "" {
		return "(root)"
	}
	return path
}

// IdentityFrom extracts the canonical identifier a response declares, under
// the artifact's id strategy. One identity is minted once and reused by every
// later interaction: a cassette whose create response and subsequent request
// URLs disagree cannot replay at all.
func IdentityFrom(body any, strategy model.IdStrategy) (string, bool) {
	var path []string
	switch strategy {
	case model.IdStrategyDataID, "":
		path = []string{"data", "id"}
	case model.IdStrategyDataAttributesID:
		path = []string{"data", "attributes", "id"}
	case model.IdStrategyDataAttributesUID:
		path = []string{"data", "attributes", "uuid"}
	default:
		// header.location is carried by a response header, not its body.
		return "", false
	}
	current := body
	for _, segment := range path {
		object, ok := current.(map[string]any)
		if !ok {
			return "", false
		}
		current, ok = object[segment]
		if !ok {
			return "", false
		}
	}
	identity, ok := current.(string)
	if !ok || identity == "" {
		return "", false
	}
	return identity, true
}

// ----------------------------------------------------------------------------
// Overlay
// ----------------------------------------------------------------------------

// overlaidWith returns this set with delta's values layered on top.
//
// An update example describes the API's sparse PATCH delta: only the members
// the caller wants to change. Terraform does not work that way. Its state is
// the complete desired configuration, so the provider serializes every planned
// attribute into the update request — not just the changed ones. A cassette
// recorded from the delta alone therefore never matches the request the
// provider actually sends, and replay fails to find the interaction.
//
// Layering the delta over the create request reconstructs what the provider
// will send: the resource as created, with the update's changes applied. The
// same overlaid set then feeds both the recorded body and the step's
// configuration, so the two cannot drift apart.
//
// Deep merge cannot express removal, so an update that drops a member is not
// representable. That is a limit of describing updates as examples rather than
// as patches, and it fails loudly at replay rather than silently.
func (m MaterializedSet) overlaidWith(delta MaterializedSet) MaterializedSet {
	values := map[string]model.MaterializedValue{}
	for _, value := range m.Values {
		values[value.Path] = value
	}
	for _, value := range delta.Values {
		values[value.Path] = value
	}

	// Body and Values diverge here, and deliberately. A write-only secret
	// stays in Values because the step's configuration must keep declaring it
	// — the attribute is required, and dropping it would fail validation. It
	// is removed from the body because the provider sends a write-only secret
	// only when its _wo_version trigger changes, and a generated update step
	// leaves that version alone. Recording the inherited secret would describe
	// a request the provider never makes.
	body := overlayJSON(m.Body, delta.Body)
	for _, path := range append(slices.Clone(m.WriteOnlyPaths), delta.WriteOnlyPaths...) {
		body = withoutJSONPath(body, strings.Split(path, "."))
	}

	out := MaterializedSet{
		Key:                   delta.Key,
		Body:                  body,
		Values:                make([]model.MaterializedValue, 0, len(values)),
		SensitiveReplacements: map[string]string{},
		WriteOnlyPaths:        delta.WriteOnlyPaths,
	}
	// Sorted so regeneration stays byte-identical: map iteration is not.
	for _, path := range slices.Sorted(maps.Keys(values)) {
		out.Values = append(out.Values, values[path])
	}
	maps.Copy(out.SensitiveReplacements, m.SensitiveReplacements)
	maps.Copy(out.SensitiveReplacements, delta.SensitiveReplacements)
	if len(out.SensitiveReplacements) == 0 {
		out.SensitiveReplacements = nil
	}
	return out
}

// overlayJSON deep-merges delta over base. Objects merge member by member;
// anything else is replaced wholesale. Arrays in particular are replaced
// rather than merged element-wise, because an array is one value to Terraform:
// a config that sets two tags means exactly those two, not an append.
func overlayJSON(base, delta any) any {
	if delta == nil {
		return base
	}
	baseObject, baseOK := base.(map[string]any)
	deltaObject, deltaOK := delta.(map[string]any)
	if !baseOK || !deltaOK {
		return delta
	}
	merged := maps.Clone(baseObject)
	for key, value := range deltaObject {
		merged[key] = overlayJSON(merged[key], value)
	}
	return merged
}

// withoutJSONPath returns value with the member at the given path segments
// removed. A path that does not resolve leaves the value untouched, so a
// write-only leaf the update example never declared is not an error.
func withoutJSONPath(value any, path []string) any {
	object, ok := value.(map[string]any)
	if !ok || len(path) == 0 {
		return value
	}
	head := path[0]
	if _, present := object[head]; !present {
		return value
	}
	out := maps.Clone(object)
	if len(path) == 1 {
		delete(out, head)
		return out
	}
	out[head] = withoutJSONPath(out[head], path[1:])
	return out
}
