package cassette

import (
	"fmt"
	"math"
	"slices"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// declaredOneOf resolves the example's branch before walking its children.
// Keeping the branch identity prevents shared properties from configuring
// several Terraform alternatives or being validated against the wrong schema.
func (m *materializer) declaredOneOf(value any, schema *model.Schema, path string) any {
	variants := oneOfVariants(schema)
	selected := -1
	for i, variant := range variants {
		if len(variants) != 1 && !m.matchesDeclared(value, variant) {
			continue
		}
		if selected >= 0 {
			m.ambiguous = append(m.ambiguous, labelPath(path))
			return nil
		}
		selected = i
	}
	if selected < 0 {
		m.missing = append(m.missing, labelPath(path)+" (no matching union branch)")
		return nil
	}
	branch := variants[selected]
	body := m.fromDeclared(value, branch, path)
	if schema.OneOf != nil {
		for _, variant := range schema.OneOf.Variants {
			if variant.Schema == branch {
				m.out.Values = append(m.out.Values, model.MaterializedValue{
					Path: labelPath(path), Value: body, Schema: branch, Variant: variant.TFName,
				})
				break
			}
		}
	}
	return body
}

// matchesDeclared uses shape, required members and enum discriminators. If
// these cannot distinguish the alternatives, declining is safer than choosing
// one and silently dropping the fields another alternative would recognize.
func (m *materializer) matchesDeclared(value any, schema *model.Schema) bool {
	if schema == nil || schema.Kind == model.SchemaKindUnsupported {
		return false
	}
	if len(schema.Enum) > 0 && !slices.Contains(schema.Enum, fmt.Sprint(value)) {
		return false
	}
	switch schema.Kind {
	case model.SchemaKindOneOf:
		matches := 0
		for _, branch := range oneOfVariants(schema) {
			if m.matchesDeclared(value, branch) {
				matches++
			}
		}
		return matches == 1
	case model.SchemaKindObject, model.SchemaKindMap:
		object, ok := value.(map[string]any)
		if !ok {
			return false
		}
		for _, key := range schema.Required {
			if child := schema.Properties[key]; child != nil && m.isRequest && child.ReadOnly {
				continue
			}
			if _, present := object[key]; !present {
				return false
			}
		}
		for key, item := range object {
			child := schema.Properties[key]
			if schema.Kind == model.SchemaKindMap {
				child = schema.Items
			}
			if child != nil && !m.matchesDeclared(item, child) {
				return false
			}
		}
		return true
	case model.SchemaKindArray:
		items, ok := value.([]any)
		if !ok {
			return false
		}
		for _, item := range items {
			if schema.Items != nil && !m.matchesDeclared(item, schema.Items) {
				return false
			}
		}
		return true
	}
	switch schema.Type {
	case "string":
		_, ok := value.(string)
		return ok
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "integer", "number":
		switch number := value.(type) {
		case int, int64:
			return true
		case float64:
			return schema.Type == "number" || number == math.Trunc(number)
		default:
			return false
		}
	default:
		return false
	}
}
