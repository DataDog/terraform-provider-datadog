package openapi

import (
	"fmt"
	"log"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/orderedmap"
)

// ResolveAllOf merges properties from all allOf sub-schemas into a single Schema.
// Resolves $ref targets before merging. Last definition wins on property name collisions.
// Required arrays are unioned.
func ResolveAllOf(allOfSchemas []*base.SchemaProxy) (*base.Schema, error) {
	if len(allOfSchemas) == 0 {
		return nil, fmt.Errorf("allOf has no sub-schemas")
	}

	log.Printf("Flattening allOf with %d sub-schemas", len(allOfSchemas))

	merged := &base.Schema{
		Properties: orderedmap.New[string, *base.SchemaProxy](),
	}

	requiredSet := make(map[string]bool)

	for i, proxy := range allOfSchemas {
		resolved, err := proxy.BuildSchema()
		if err != nil {
			return nil, fmt.Errorf("resolving allOf sub-schema %d: %w", i, err)
		}

		// If the resolved schema itself has allOf, recursively resolve it
		if len(resolved.AllOf) > 0 {
			resolved, err = ResolveAllOf(resolved.AllOf)
			if err != nil {
				return nil, fmt.Errorf("resolving nested allOf in sub-schema %d: %w", i, err)
			}
		}

		// Merge properties (last definition wins)
		if resolved.Properties != nil {
			for name, propProxy := range resolved.Properties.FromOldest() {
				merged.Properties.Set(name, propProxy)
			}
		}

		// Union required arrays
		for _, r := range resolved.Required {
			requiredSet[r] = true
		}
	}

	for r := range requiredSet {
		merged.Required = append(merged.Required, r)
	}

	// Copy type from first sub-schema if set
	if len(merged.Type) == 0 {
		merged.Type = []string{"object"}
	}

	return merged, nil
}

// ResolveOneOf collects properties from all variant schemas as optional attributes.
// Merges compatible-type properties. Uses string-typed property as fallback when
// variants have conflicting types for the same property name.
// anyOf uses the same logic.
func ResolveOneOf(oneOfSchemas []*base.SchemaProxy) (*base.Schema, error) {
	if len(oneOfSchemas) == 0 {
		return nil, fmt.Errorf("oneOf has no variant schemas")
	}

	log.Printf("Resolving oneOf/anyOf with %d variants", len(oneOfSchemas))

	merged := &base.Schema{
		Properties: orderedmap.New[string, *base.SchemaProxy](),
	}
	// Track types seen for each property name to detect conflicts
	typeMap := make(map[string]string) // property name -> first seen type

	for i, proxy := range oneOfSchemas {
		resolved, err := proxy.BuildSchema()
		if err != nil {
			return nil, fmt.Errorf("resolving oneOf variant %d: %w", i, err)
		}

		if resolved.Properties == nil {
			continue
		}

		for name, propProxy := range resolved.Properties.FromOldest() {
			propSchema, err := propProxy.BuildSchema()
			if err != nil {
				continue
			}

			propType := ""
			if len(propSchema.Type) > 0 {
				propType = propSchema.Type[0]
			}

			existingType, exists := typeMap[name]
			if exists {
				if existingType != propType {
					// Type conflict — use string fallback
					log.Printf("Warning: type conflict for property %q in oneOf (%s vs %s), falling back to string", name, existingType, propType)
					// Create a string-typed schema proxy as fallback
					merged.Properties.Set(name, createStringSchemaProxy(propSchema.Description))
					typeMap[name] = "string"
				}
				// Compatible types: keep existing (already set)
			} else {
				typeMap[name] = propType
				merged.Properties.Set(name, propProxy)
			}
		}
	}

	// All fields from oneOf/anyOf variants are Optional (never Required)
	merged.Required = nil

	if len(merged.Type) == 0 {
		merged.Type = []string{"object"}
	}

	return merged, nil
}

// ResolveAdditionalProperties converts additionalProperties to a map-typed SchemaField.
// When additionalProperties is boolean true, returns FieldTypeMapOfStrings.
// When it is a schema, parses the element type and returns the appropriate map FieldType.
func ResolveAdditionalProperties(schema *base.Schema) (*SchemaField, error) {
	if schema.AdditionalProperties == nil {
		return nil, fmt.Errorf("schema has no additionalProperties")
	}

	// Check if additionalProperties is a boolean (true)
	if schema.AdditionalProperties.IsB() {
		if schema.AdditionalProperties.B {
			log.Printf("Resolving additionalProperties: boolean true -> MapOfStrings")
			return &SchemaField{
				Type:        FieldTypeMapOfStrings,
				Description: schema.Description,
			}, nil
		}
		// additionalProperties: false — no map field
		return nil, fmt.Errorf("additionalProperties is false")
	}

	// additionalProperties is a schema
	if schema.AdditionalProperties.IsA() && schema.AdditionalProperties.A != nil {
		elemSchema, err := schema.AdditionalProperties.A.BuildSchema()
		if err != nil {
			return nil, fmt.Errorf("building additionalProperties schema: %w", err)
		}

		elemType := ""
		if len(elemSchema.Type) > 0 {
			elemType = elemSchema.Type[0]
		}

		var fieldType FieldType
		switch elemType {
		case "string":
			fieldType = FieldTypeMapOfStrings
		case "integer":
			fieldType = FieldTypeMapOfInts
		case "number":
			fieldType = FieldTypeMapOfFloats
		case "boolean":
			fieldType = FieldTypeMapOfBools
		case "object":
			fieldType = FieldTypeMapOfObjects
		default:
			fieldType = FieldTypeMapOfStrings // fallback
		}

		log.Printf("Resolving additionalProperties: %s schema -> %v", elemType, fieldType)

		field := &SchemaField{
			Type:        fieldType,
			Description: schema.Description,
		}

		// For object-type additionalProperties, parse the children
		if elemType == "object" {
			children, err := parseSchemaFromResolved(elemSchema, "")
			if err != nil {
				return nil, fmt.Errorf("parsing additionalProperties object schema: %w", err)
			}
			field.Children = children
		}

		return field, nil
	}

	// Fallback for unexpected additionalProperties shape
	return &SchemaField{
		Type:        FieldTypeMapOfStrings,
		Description: schema.Description,
	}, nil
}

// createStringSchemaProxy creates a base.SchemaProxy wrapping a simple string-typed schema.
// Used as a fallback when oneOf variants have conflicting types.
func createStringSchemaProxy(description string) *base.SchemaProxy {
	// Build a minimal schema proxy for a string type
	// We use the low-level approach to create a schema proxy that resolves to a string schema
	s := &base.Schema{
		Type:        []string{"string"},
		Description: description,
	}
	return base.CreateSchemaProxy(s)
}
