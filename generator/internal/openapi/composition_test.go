package openapi

import (
	"testing"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/orderedmap"
)

func makeSchemaProxy(props map[string]string, required []string) *base.SchemaProxy {
	om := orderedmap.New[string, *base.SchemaProxy]()
	for name, typ := range props {
		s := &base.Schema{
			Type:        []string{typ},
			Description: name + " description",
		}
		om.Set(name, base.CreateSchemaProxy(s))
	}
	return base.CreateSchemaProxy(&base.Schema{
		Type:       []string{"object"},
		Properties: om,
		Required:   required,
	})
}

func TestResolveAllOf_MergesProperties(t *testing.T) {
	s1 := makeSchemaProxy(map[string]string{"name": "string", "version": "string"}, []string{"name"})
	s2 := makeSchemaProxy(map[string]string{"priority": "integer", "config": "object"}, nil)

	merged, err := ResolveAllOf([]*base.SchemaProxy{s1, s2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if merged.Properties == nil {
		t.Fatal("merged properties should not be nil")
	}
	if merged.Properties.Len() != 4 {
		t.Errorf("expected 4 properties, got %d", merged.Properties.Len())
	}

	// Check required was unioned
	reqSet := make(map[string]bool)
	for _, r := range merged.Required {
		reqSet[r] = true
	}
	if !reqSet["name"] {
		t.Error("required should contain 'name'")
	}
}

func TestResolveAllOf_LastWins(t *testing.T) {
	s1 := makeSchemaProxy(map[string]string{"name": "string"}, nil)
	s2 := makeSchemaProxy(map[string]string{"name": "integer"}, nil)

	merged, err := ResolveAllOf([]*base.SchemaProxy{s1, s2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prop, ok := merged.Properties.Get("name")
	if !ok {
		t.Fatal("name property should exist")
	}
	schema, _ := prop.BuildSchema()
	if len(schema.Type) == 0 || schema.Type[0] != "integer" {
		t.Errorf("last definition should win, got type %v", schema.Type)
	}
}

func TestResolveAllOf_Empty(t *testing.T) {
	_, err := ResolveAllOf(nil)
	if err == nil {
		t.Fatal("expected error for empty allOf")
	}
}

func TestResolveOneOf_CollectsAllProperties(t *testing.T) {
	s1 := makeSchemaProxy(map[string]string{"url": "string", "timeout": "integer"}, nil)
	s2 := makeSchemaProxy(map[string]string{"port": "integer", "timeout": "integer"}, nil)

	merged, err := ResolveOneOf([]*base.SchemaProxy{s1, s2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if merged.Properties.Len() != 3 {
		t.Errorf("expected 3 properties (url, port, timeout), got %d", merged.Properties.Len())
	}

	// All fields from oneOf are Optional
	if len(merged.Required) != 0 {
		t.Errorf("expected 0 required, got %d", len(merged.Required))
	}
}

func TestResolveOneOf_CompatibleTypes(t *testing.T) {
	s1 := makeSchemaProxy(map[string]string{"timeout": "integer"}, nil)
	s2 := makeSchemaProxy(map[string]string{"timeout": "integer"}, nil)

	merged, err := ResolveOneOf([]*base.SchemaProxy{s1, s2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prop, _ := merged.Properties.Get("timeout")
	schema, _ := prop.BuildSchema()
	if schema.Type[0] != "integer" {
		t.Errorf("compatible types should keep original, got %v", schema.Type)
	}
}

func TestResolveOneOf_ConflictingTypes(t *testing.T) {
	s1 := makeSchemaProxy(map[string]string{"value": "string"}, nil)
	s2 := makeSchemaProxy(map[string]string{"value": "integer"}, nil)

	merged, err := ResolveOneOf([]*base.SchemaProxy{s1, s2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prop, _ := merged.Properties.Get("value")
	schema, _ := prop.BuildSchema()
	if schema.Type[0] != "string" {
		t.Errorf("conflicting types should fall back to string, got %v", schema.Type)
	}
}

func TestResolveOneOf_EmptyVariant(t *testing.T) {
	s1 := makeSchemaProxy(map[string]string{"name": "string"}, nil)
	empty := base.CreateSchemaProxy(&base.Schema{Type: []string{"object"}})

	merged, err := ResolveOneOf([]*base.SchemaProxy{s1, empty})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if merged.Properties.Len() != 1 {
		t.Errorf("expected 1 property, got %d", merged.Properties.Len())
	}
}

func TestResolveOneOf_Empty(t *testing.T) {
	_, err := ResolveOneOf(nil)
	if err == nil {
		t.Fatal("expected error for empty oneOf")
	}
}

func TestResolveAdditionalProperties_BoolTrue(t *testing.T) {
	boolTrue := true
	schema := &base.Schema{
		Description: "metadata",
		AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{
			N: 1,
			B: boolTrue,
		},
	}

	field, err := ResolveAdditionalProperties(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if field.Type != FieldTypeMapOfStrings {
		t.Errorf("expected FieldTypeMapOfStrings, got %d", field.Type)
	}
}

func TestResolveAdditionalProperties_StringSchema(t *testing.T) {
	elemSchema := base.CreateSchemaProxy(&base.Schema{Type: []string{"string"}})
	schema := &base.Schema{
		Description: "labels",
		AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{
			N: 0,
			A: elemSchema,
		},
	}

	field, err := ResolveAdditionalProperties(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if field.Type != FieldTypeMapOfStrings {
		t.Errorf("expected FieldTypeMapOfStrings, got %d", field.Type)
	}
}

func TestResolveAdditionalProperties_IntegerSchema(t *testing.T) {
	elemSchema := base.CreateSchemaProxy(&base.Schema{Type: []string{"integer"}})
	schema := &base.Schema{
		Description: "counts",
		AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{
			N: 0,
			A: elemSchema,
		},
	}

	field, err := ResolveAdditionalProperties(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if field.Type != FieldTypeMapOfInts {
		t.Errorf("expected FieldTypeMapOfInts, got %d", field.Type)
	}
}

func TestResolveAdditionalProperties_ObjectSchema(t *testing.T) {
	childProps := orderedmap.New[string, *base.SchemaProxy]()
	childProps.Set("key", base.CreateSchemaProxy(&base.Schema{Type: []string{"string"}}))
	elemSchema := base.CreateSchemaProxy(&base.Schema{
		Type:       []string{"object"},
		Properties: childProps,
	})
	schema := &base.Schema{
		Description: "nested map",
		AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{
			N: 0,
			A: elemSchema,
		},
	}

	field, err := ResolveAdditionalProperties(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if field.Type != FieldTypeMapOfObjects {
		t.Errorf("expected FieldTypeMapOfObjects, got %d", field.Type)
	}
	if field.Children == nil {
		t.Error("map of objects should have Children")
	}
}

func TestResolveAdditionalProperties_Nil(t *testing.T) {
	schema := &base.Schema{}
	_, err := ResolveAdditionalProperties(schema)
	if err == nil {
		t.Fatal("expected error for nil additionalProperties")
	}
}

func TestResolveAdditionalProperties_BoolFalse(t *testing.T) {
	boolFalse := false
	schema := &base.Schema{
		AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{
			N: 1,
			B: boolFalse,
		},
	}

	_, err := ResolveAdditionalProperties(schema)
	if err == nil {
		t.Fatal("expected error for additionalProperties: false")
	}
}
