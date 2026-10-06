package cassette

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// ----------------------------------------------------------------------------
// Synthesized values
//
// A description that declares no example for a required field used to make its
// whole artifact ineligible. Measured against the upstream spec that is most
// of them: 2xx responses carry examples almost everywhere, request bodies at
// roughly a third. Refusing on that basis means the feature only ever applies
// to the specs that least need help.
//
// So a leaf with no declared example and no schema default gets a synthesized
// one. The value is deterministic, derived from the schema rather than the
// path, and schema-valid as far as the normalized model describes validity —
// an enum member, a format-appropriate string, a number inside its bounds.
//
// What it is not is true. The API never said it. A fixture standing on
// synthesized values can be recorded over, and until it is, every synthesized
// path is reported so nobody mistakes it for evidence.
// ----------------------------------------------------------------------------

// synthesizedUUID is a fixed v4-shaped identifier. Fixed so regeneration stays
// byte-identical, and visibly not a real id.
const synthesizedUUID = "00000000-0000-4000-8000-000000000000"

// synthesizeLeaf invents a schema-valid value for a leaf the description left
// undescribed, reporting false when the schema names no representable type.
func synthesizeLeaf(schema *model.Schema, path string) (any, bool) {
	if schema == nil {
		return nil, false
	}
	// An enum's first member is the only choice that is certainly valid.
	if len(schema.Enum) > 0 {
		return schema.Enum[0], true
	}

	switch schema.Type {
	case "string":
		return synthesizeString(schema.Format, path), true
	case "integer":
		return 0, true
	case "number":
		return float64(0), true
	case "boolean":
		return false, true
	default:
		return nil, false
	}
}

// synthesizeString honors the OpenAPI format, because a field declaring one
// usually has a server-side parser that rejects anything else. Without a
// format the leaf's own name is used, which keeps a generated fixture readable
// and keeps two different fields from colliding on one value.
func synthesizeString(format, path string) string {
	switch format {
	case "date-time":
		// The scenario's freeze instant, so a synthesized timestamp agrees
		// with the clock the harness restores.
		return "2026-01-01T00:00:00Z"
	case "date":
		return "2026-01-01"
	case "uuid":
		return synthesizedUUID
	case "email":
		return "dummy@example.com"
	case "uri", "url":
		return "https://example.com"
	case "ipv4":
		return "192.0.2.1"
	case "hostname":
		return "example.com"
	default:
		return "dummy-" + leafName(path)
	}
}

// leafName is the last segment of a dotted path, with any array index dropped,
// so data.attributes.tags[0] reads as tags.
func leafName(path string) string {
	name := path
	if i := strings.LastIndex(name, "."); i >= 0 {
		name = name[i+1:]
	}
	if i := strings.Index(name, "["); i >= 0 {
		name = name[:i]
	}
	if name == "" || name == "(root)" {
		return "value"
	}
	return name
}

// synthesizeCollection invents the smallest valid collection: one element when
// the schema requires a non-empty one, and an empty collection otherwise. A
// single element is enough to exercise serialization without claiming the API
// returns more than one.
func (m *materializer) synthesizeCollection(schema *model.Schema, path string) (any, bool) {
	switch schema.Kind {
	case model.SchemaKindArray:
		if schema.Items == nil {
			return []any{}, true
		}
		item, ok := m.synthesize(schema.Items, fmt.Sprintf("%s[0]", path))
		if !ok {
			return []any{}, true
		}
		return []any{item}, true
	case model.SchemaKindMap:
		// A map's keys are not described, so an empty one is the only honest
		// synthesis; a required map is satisfied by being present.
		return map[string]any{}, true
	default:
		return nil, false
	}
}

// synthesize invents a value of any kind, recursing into an object's required
// properties. Every leaf it invents is recorded on the set.
func (m *materializer) synthesize(schema *model.Schema, path string) (any, bool) {
	if schema == nil {
		return nil, false
	}
	switch schema.Kind {
	case model.SchemaKindObject:
		out := map[string]any{}
		for _, name := range slices.Sorted(maps.Keys(schema.Properties)) {
			property := schema.Properties[name]
			if m.isRequest && property.ReadOnly {
				continue
			}
			if !slices.Contains(schema.Required, name) {
				continue
			}
			child := model.ChildPath(path, name)
			value, ok := m.synthesize(property, child)
			if !ok {
				continue
			}
			out[name] = value
		}
		return out, true

	case model.SchemaKindArray, model.SchemaKindMap:
		return m.synthesizeCollection(schema, path)

	case model.SchemaKindOneOf:
		// The first variant, so the choice is deterministic rather than a
		// guess at which one the API prefers.
		for _, variant := range oneOfVariants(schema) {
			if value, ok := m.synthesize(variant, path); ok {
				return value, true
			}
		}
		return nil, false

	default:
		value, ok := synthesizeLeaf(schema, path)
		if !ok {
			return nil, false
		}
		m.recordSynthesized(path)
		// Through leaf, so the value reaches Values and the configuration
		// renderer sees it. Recording the path alone would put it in the
		// request body and leave it out of the HCL.
		return m.leaf(value, schema, path), true
	}
}
