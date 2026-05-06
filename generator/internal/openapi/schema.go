package openapi

import (
	"fmt"
	"log"

	"github.com/pb33f/libopenapi/datamodel/high/base"
)

// FieldType represents the resolved type of a schema field.
type FieldType int

const (
	FieldTypeString FieldType = iota
	FieldTypeInt64
	FieldTypeInt32
	FieldTypeFloat64
	FieldTypeFloat32
	FieldTypeBool
	FieldTypeObject
	FieldTypeArrayOfObjects
	FieldTypeArrayOfStrings
	FieldTypeArrayOfInts
	FieldTypeArrayOfFloats
	FieldTypeArrayOfBools
	FieldTypeMapOfStrings
	FieldTypeMapOfInts
	FieldTypeMapOfFloats
	FieldTypeMapOfBools
	FieldTypeMapOfObjects
)

// SchemaField represents a single property in a schema.
type SchemaField struct {
	Name        string
	Description string
	Type        FieldType
	Required    bool
	ReadOnly    bool
	Nullable    bool
	Format      string
	EnumValues  []string
	Children    *SchemaObject
}

// SchemaObject represents a parsed schema with its fields.
type SchemaObject struct {
	Name        string
	Description string
	Fields      []SchemaField
}

// ParseSchema parses a SchemaProxy into a SchemaObject, recursively resolving
// nested objects and arrays. Returns an error on unsupported constructs
// (oneOf, anyOf, allOf).
func ParseSchema(proxy *base.SchemaProxy) (*SchemaObject, error) {
	if proxy == nil {
		return nil, fmt.Errorf("nil schema proxy")
	}

	schema, err := proxy.BuildSchema()
	if err != nil {
		return nil, fmt.Errorf("building schema: %w", err)
	}

	return parseSchemaFromResolved(schema, "")
}

// parseSchemaFromResolved parses a resolved Schema into a SchemaObject.
// Handles composition constructs (allOf, oneOf, anyOf) by resolving them
// before processing properties. Handles additionalProperties as map fields.
func parseSchemaFromResolved(schema *base.Schema, name string) (*SchemaObject, error) {
	// Handle composition constructs in priority order
	if len(schema.AllOf) > 0 {
		log.Printf("Resolving allOf for schema %q with %d sub-schemas", name, len(schema.AllOf))
		merged, err := ResolveAllOf(schema.AllOf)
		if err != nil {
			return nil, fmt.Errorf("resolving allOf for %q: %w", name, err)
		}
		return parseSchemaFromResolved(merged, name)
	}

	if len(schema.OneOf) > 0 {
		log.Printf("Resolving oneOf for schema %q with %d variants", name, len(schema.OneOf))
		merged, err := ResolveOneOf(schema.OneOf)
		if err != nil {
			return nil, fmt.Errorf("resolving oneOf for %q: %w", name, err)
		}
		return parseSchemaFromResolved(merged, name)
	}

	if len(schema.AnyOf) > 0 {
		log.Printf("Resolving anyOf for schema %q with %d variants", name, len(schema.AnyOf))
		merged, err := ResolveOneOf(schema.AnyOf)
		if err != nil {
			return nil, fmt.Errorf("resolving anyOf for %q: %w", name, err)
		}
		return parseSchemaFromResolved(merged, name)
	}

	obj := &SchemaObject{
		Name:        name,
		Description: schema.Description,
	}

	// Handle additionalProperties as map fields
	if schema.AdditionalProperties != nil {
		mapField, err := ResolveAdditionalProperties(schema)
		if err == nil {
			mapField.Name = name
			if mapField.Description == "" {
				mapField.Description = schema.Description
			}
			obj.Fields = append(obj.Fields, *mapField)
		}
		// If additionalProperties is the only thing (no properties), return early
		if schema.Properties == nil || schema.Properties.Len() == 0 {
			return obj, nil
		}
	}

	if schema.Properties == nil {
		return obj, nil
	}

	requiredSet := make(map[string]bool, len(schema.Required))
	for _, r := range schema.Required {
		requiredSet[r] = true
	}

	for fieldName, fieldProxy := range schema.Properties.FromOldest() {
		field, err := parseField(fieldName, fieldProxy, requiredSet[fieldName])
		if err != nil {
			return nil, fmt.Errorf("parsing field %q: %w", fieldName, err)
		}
		obj.Fields = append(obj.Fields, *field)
	}

	return obj, nil
}

// parseField parses a single field from a property SchemaProxy.
func parseField(name string, proxy *base.SchemaProxy, required bool) (*SchemaField, error) {
	schema, err := proxy.BuildSchema()
	if err != nil {
		return nil, fmt.Errorf("building schema: %w", err)
	}

	// Handle composition constructs at the field level
	if len(schema.AllOf) > 0 {
		log.Printf("Resolving allOf for field %q", name)
		merged, err := ResolveAllOf(schema.AllOf)
		if err != nil {
			return nil, fmt.Errorf("resolving allOf for field %q: %w", name, err)
		}
		// Parse the merged schema as an object field
		children, err := parseSchemaFromResolved(merged, name)
		if err != nil {
			return nil, fmt.Errorf("parsing merged allOf for field %q: %w", name, err)
		}
		return &SchemaField{
			Name:        name,
			Description: schema.Description,
			Type:        FieldTypeObject,
			Required:    required,
			Children:    children,
		}, nil
	}

	if len(schema.OneOf) > 0 {
		log.Printf("Resolving oneOf for field %q", name)
		merged, err := ResolveOneOf(schema.OneOf)
		if err != nil {
			return nil, fmt.Errorf("resolving oneOf for field %q: %w", name, err)
		}
		children, err := parseSchemaFromResolved(merged, name)
		if err != nil {
			return nil, fmt.Errorf("parsing merged oneOf for field %q: %w", name, err)
		}
		return &SchemaField{
			Name:        name,
			Description: schema.Description,
			Type:        FieldTypeObject,
			Required:    required,
			Children:    children,
		}, nil
	}

	if len(schema.AnyOf) > 0 {
		log.Printf("Resolving anyOf for field %q", name)
		merged, err := ResolveOneOf(schema.AnyOf)
		if err != nil {
			return nil, fmt.Errorf("resolving anyOf for field %q: %w", name, err)
		}
		children, err := parseSchemaFromResolved(merged, name)
		if err != nil {
			return nil, fmt.Errorf("parsing merged anyOf for field %q: %w", name, err)
		}
		return &SchemaField{
			Name:        name,
			Description: schema.Description,
			Type:        FieldTypeObject,
			Required:    required,
			Children:    children,
		}, nil
	}

	// Handle additionalProperties at the field level
	if schema.AdditionalProperties != nil {
		mapField, err := ResolveAdditionalProperties(schema)
		if err == nil {
			mapField.Name = name
			mapField.Required = required
			if mapField.Description == "" {
				mapField.Description = schema.Description
			}
			return mapField, nil
		}
	}

	field := &SchemaField{
		Name:        name,
		Description: schema.Description,
		Format:      schema.Format,
		Required:    required,
		ReadOnly:    schema.ReadOnly != nil && *schema.ReadOnly,
		Nullable:    schema.Nullable != nil && *schema.Nullable,
	}

	// Extract enum values
	for _, e := range schema.Enum {
		field.EnumValues = append(field.EnumValues, e.Value)
	}

	// Determine type
	schemaType := ""
	if len(schema.Type) > 0 {
		schemaType = schema.Type[0]
	}

	switch schemaType {
	case "string":
		field.Type = FieldTypeString
	case "integer":
		switch schema.Format {
		case "int32":
			field.Type = FieldTypeInt32
		default:
			field.Type = FieldTypeInt64
		}
	case "number":
		switch schema.Format {
		case "float":
			field.Type = FieldTypeFloat32
		default:
			field.Type = FieldTypeFloat64
		}
	case "boolean":
		field.Type = FieldTypeBool
	case "object":
		field.Type = FieldTypeObject
		children, err := parseSchemaFromResolved(schema, name)
		if err != nil {
			return nil, fmt.Errorf("parsing nested object %q: %w", name, err)
		}
		field.Children = children
	case "array":
		if err := resolveArrayType(field, schema); err != nil {
			return nil, err
		}
	default:
		field.Type = FieldTypeString // fallback
	}

	return field, nil
}

// resolveArrayType determines the array element type and sets the field type.
func resolveArrayType(field *SchemaField, schema *base.Schema) error {
	if schema.Items == nil || schema.Items.A == nil {
		return fmt.Errorf("array field %q has no items schema", field.Name)
	}

	itemProxy := schema.Items.A
	itemSchema, err := itemProxy.BuildSchema()
	if err != nil {
		return fmt.Errorf("building array items schema for %q: %w", field.Name, err)
	}

	itemType := ""
	if len(itemSchema.Type) > 0 {
		itemType = itemSchema.Type[0]
	}

	switch itemType {
	case "object":
		field.Type = FieldTypeArrayOfObjects
		children, err := parseSchemaFromResolved(itemSchema, field.Name)
		if err != nil {
			return fmt.Errorf("parsing array items for %q: %w", field.Name, err)
		}
		field.Children = children
	case "string":
		field.Type = FieldTypeArrayOfStrings
	case "integer":
		field.Type = FieldTypeArrayOfInts
	case "number":
		field.Type = FieldTypeArrayOfFloats
	case "boolean":
		field.Type = FieldTypeArrayOfBools
	default:
		field.Type = FieldTypeArrayOfStrings // fallback
	}

	return nil
}
