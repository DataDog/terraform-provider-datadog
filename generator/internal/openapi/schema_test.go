package openapi

import (
	"testing"
)

func TestParseSchema_FlatPrimitives(t *testing.T) {
	model, err := LoadSpec("../../testdata/minimal.yaml")
	if err != nil {
		t.Fatalf("failed to load test spec: %v", err)
	}

	// Use the simple endpoint (flat JSON, no JSON:API)
	op, err := ExtractOperation(&model.Model, "/api/v2/simple/{id}", "get")
	if err != nil {
		t.Fatalf("extracting operation: %v", err)
	}

	obj, err := ParseSchema(op.ResponseSchemaProxy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedFields := map[string]FieldType{
		"name":    FieldTypeString,
		"value":   FieldTypeFloat64,
		"enabled": FieldTypeBool,
		"tags":    FieldTypeArrayOfStrings,
	}

	if len(obj.Fields) != len(expectedFields) {
		t.Fatalf("expected %d fields, got %d", len(expectedFields), len(obj.Fields))
	}

	for _, f := range obj.Fields {
		want, ok := expectedFields[f.Name]
		if !ok {
			t.Errorf("unexpected field %q", f.Name)
			continue
		}
		if f.Type != want {
			t.Errorf("field %q: type = %d, want %d", f.Name, f.Type, want)
		}
	}
}

func TestParseSchema_JSONAPIAttributes(t *testing.T) {
	model, err := LoadSpec("../../testdata/minimal.yaml")
	if err != nil {
		t.Fatalf("failed to load test spec: %v", err)
	}

	// Extract the team endpoint, unwrap JSON:API
	op, err := ExtractOperation(&model.Model, "/api/v2/team/{team_id}", "get")
	if err != nil {
		t.Fatalf("extracting operation: %v", err)
	}

	schema, err := op.ResponseSchemaProxy.BuildSchema()
	if err != nil {
		t.Fatalf("building schema: %v", err)
	}

	attrsProxy, _, err := UnwrapJSONAPI(schema)
	if err != nil {
		t.Fatalf("unwrapping JSON:API: %v", err)
	}

	obj, err := ParseSchema(attrsProxy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedTypes := map[string]FieldType{
		"name":            FieldTypeString,
		"handle":          FieldTypeString,
		"description":     FieldTypeString,
		"summary":         FieldTypeString,
		"user_count":      FieldTypeInt64,
		"link_count":      FieldTypeInt32,
		"is_active":       FieldTypeBool,
		"status":          FieldTypeString,
		"created_at":      FieldTypeString,
		"modified_at":     FieldTypeString,
		"hidden_modules":  FieldTypeArrayOfStrings,
		"visible_modules": FieldTypeArrayOfStrings,
	}

	if len(obj.Fields) != len(expectedTypes) {
		t.Fatalf("expected %d fields, got %d (fields: %v)", len(expectedTypes), len(obj.Fields), fieldNames(obj.Fields))
	}

	for _, f := range obj.Fields {
		want, ok := expectedTypes[f.Name]
		if !ok {
			t.Errorf("unexpected field %q", f.Name)
			continue
		}
		if f.Type != want {
			t.Errorf("field %q: type = %d, want %d", f.Name, f.Type, want)
		}
	}

	// Verify nullable fields
	for _, f := range obj.Fields {
		switch f.Name {
		case "description", "summary":
			if !f.Nullable {
				t.Errorf("field %q should be nullable", f.Name)
			}
		default:
			if f.Nullable {
				t.Errorf("field %q should not be nullable", f.Name)
			}
		}
	}

	// Verify date-time format
	for _, f := range obj.Fields {
		if f.Name == "created_at" || f.Name == "modified_at" {
			if f.Format != "date-time" {
				t.Errorf("field %q: format = %q, want %q", f.Name, f.Format, "date-time")
			}
		}
	}
}

func TestParseSchema_RequiredFields(t *testing.T) {
	model, err := LoadSpec("../../testdata/minimal.yaml")
	if err != nil {
		t.Fatalf("failed to load test spec: %v", err)
	}

	op, err := ExtractOperation(&model.Model, "/api/v2/team/{team_id}", "get")
	if err != nil {
		t.Fatalf("extracting operation: %v", err)
	}

	schema, err := op.ResponseSchemaProxy.BuildSchema()
	if err != nil {
		t.Fatalf("building schema: %v", err)
	}

	attrsProxy, _, err := UnwrapJSONAPI(schema)
	if err != nil {
		t.Fatalf("unwrapping JSON:API: %v", err)
	}

	obj, err := ParseSchema(attrsProxy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	requiredFields := map[string]bool{
		"name":   true,
		"handle": true,
	}

	for _, f := range obj.Fields {
		if requiredFields[f.Name] && !f.Required {
			t.Errorf("field %q should be required", f.Name)
		}
		if !requiredFields[f.Name] && f.Required {
			t.Errorf("field %q should not be required", f.Name)
		}
	}
}

func TestParseSchema_EnumField(t *testing.T) {
	model, err := LoadSpec("../../testdata/minimal.yaml")
	if err != nil {
		t.Fatalf("failed to load test spec: %v", err)
	}

	op, err := ExtractOperation(&model.Model, "/api/v2/team/{team_id}", "get")
	if err != nil {
		t.Fatalf("extracting operation: %v", err)
	}

	schema, err := op.ResponseSchemaProxy.BuildSchema()
	if err != nil {
		t.Fatalf("building schema: %v", err)
	}

	attrsProxy, _, err := UnwrapJSONAPI(schema)
	if err != nil {
		t.Fatalf("unwrapping JSON:API: %v", err)
	}

	obj, err := ParseSchema(attrsProxy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var statusField *SchemaField
	for i := range obj.Fields {
		if obj.Fields[i].Name == "status" {
			statusField = &obj.Fields[i]
			break
		}
	}

	if statusField == nil {
		t.Fatal("status field not found")
	}

	if statusField.Type != FieldTypeString {
		t.Errorf("status type = %d, want FieldTypeString", statusField.Type)
	}

	expectedEnums := []string{"active", "disabled", "paused"}
	if len(statusField.EnumValues) != len(expectedEnums) {
		t.Fatalf("expected %d enum values, got %d", len(expectedEnums), len(statusField.EnumValues))
	}
	for i, v := range statusField.EnumValues {
		if v != expectedEnums[i] {
			t.Errorf("enum[%d] = %q, want %q", i, v, expectedEnums[i])
		}
	}
}

func TestParseSchema_NestedObject(t *testing.T) {
	model, err := LoadSpec("../../testdata/minimal.yaml")
	if err != nil {
		t.Fatalf("failed to load test spec: %v", err)
	}

	op, err := ExtractOperation(&model.Model, "/api/v2/complex/{id}", "get")
	if err != nil {
		t.Fatalf("extracting operation: %v", err)
	}

	schema, err := op.ResponseSchemaProxy.BuildSchema()
	if err != nil {
		t.Fatalf("building schema: %v", err)
	}

	attrsProxy, _, err := UnwrapJSONAPI(schema)
	if err != nil {
		t.Fatalf("unwrapping JSON:API: %v", err)
	}

	obj, err := ParseSchema(attrsProxy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Find the config field (nested object)
	var configField *SchemaField
	for i := range obj.Fields {
		if obj.Fields[i].Name == "config" {
			configField = &obj.Fields[i]
			break
		}
	}

	if configField == nil {
		t.Fatal("config field not found")
	}

	if configField.Type != FieldTypeObject {
		t.Errorf("config type = %d, want FieldTypeObject", configField.Type)
	}

	if configField.Children == nil {
		t.Fatal("config should have children")
	}

	if len(configField.Children.Fields) != 2 {
		t.Errorf("config children count = %d, want 2", len(configField.Children.Fields))
	}
}

func TestParseSchema_ArrayOfObjects(t *testing.T) {
	model, err := LoadSpec("../../testdata/minimal.yaml")
	if err != nil {
		t.Fatalf("failed to load test spec: %v", err)
	}

	op, err := ExtractOperation(&model.Model, "/api/v2/complex/{id}", "get")
	if err != nil {
		t.Fatalf("extracting operation: %v", err)
	}

	schema, err := op.ResponseSchemaProxy.BuildSchema()
	if err != nil {
		t.Fatalf("building schema: %v", err)
	}

	attrsProxy, _, err := UnwrapJSONAPI(schema)
	if err != nil {
		t.Fatalf("unwrapping JSON:API: %v", err)
	}

	obj, err := ParseSchema(attrsProxy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Find the endpoints field (array of objects)
	var endpointsField *SchemaField
	for i := range obj.Fields {
		if obj.Fields[i].Name == "endpoints" {
			endpointsField = &obj.Fields[i]
			break
		}
	}

	if endpointsField == nil {
		t.Fatal("endpoints field not found")
	}

	if endpointsField.Type != FieldTypeArrayOfObjects {
		t.Errorf("endpoints type = %d, want FieldTypeArrayOfObjects", endpointsField.Type)
	}

	if endpointsField.Children == nil {
		t.Fatal("endpoints should have children")
	}

	if len(endpointsField.Children.Fields) != 2 {
		t.Errorf("endpoints children count = %d, want 2", len(endpointsField.Children.Fields))
	}
}

// T101: Composition constructs now parse successfully
func TestParseSchema_ComposedAllOf(t *testing.T) {
	model, err := LoadSpec("../../testdata/minimal.yaml")
	if err != nil {
		t.Fatalf("failed to load test spec: %v", err)
	}

	op, err := ExtractOperation(&model.Model, "/api/v2/composed/{id}", "get")
	if err != nil {
		t.Fatalf("extracting operation: %v", err)
	}

	schema, err := op.ResponseSchemaProxy.BuildSchema()
	if err != nil {
		t.Fatalf("building schema: %v", err)
	}

	attrsProxy, _, err := UnwrapJSONAPI(schema)
	if err != nil {
		t.Fatalf("unwrapping JSON:API: %v", err)
	}

	obj, err := ParseSchema(attrsProxy)
	if err != nil {
		t.Fatalf("expected allOf schema to parse successfully, got error: %v", err)
	}

	// Should have merged properties from ComposedBase (name, version) +
	// ComposedExtension (priority, config) + inline (metadata, labels)
	fieldMap := make(map[string]FieldType)
	for _, f := range obj.Fields {
		fieldMap[f.Name] = f.Type
	}

	// From ComposedBase
	if _, ok := fieldMap["name"]; !ok {
		t.Error("missing 'name' from ComposedBase")
	}
	if _, ok := fieldMap["version"]; !ok {
		t.Error("missing 'version' from ComposedBase")
	}

	// From ComposedExtension
	if _, ok := fieldMap["priority"]; !ok {
		t.Error("missing 'priority' from ComposedExtension")
	}
	if ft, ok := fieldMap["config"]; !ok {
		t.Error("missing 'config' from ComposedExtension")
	} else if ft != FieldTypeObject {
		t.Errorf("config should be FieldTypeObject, got %d", ft)
	}

	// From inline additionalProperties
	if ft, ok := fieldMap["metadata"]; !ok {
		t.Error("missing 'metadata' (additionalProperties: true)")
	} else if ft != FieldTypeMapOfStrings {
		t.Errorf("metadata should be FieldTypeMapOfStrings, got %d", ft)
	}
	if ft, ok := fieldMap["labels"]; !ok {
		t.Error("missing 'labels' (additionalProperties: string)")
	} else if ft != FieldTypeMapOfStrings {
		t.Errorf("labels should be FieldTypeMapOfStrings, got %d", ft)
	}
}

func TestParseSchema_OneOfResolvesSuccessfully(t *testing.T) {
	model, err := LoadSpec("../../testdata/minimal.yaml")
	if err != nil {
		t.Fatalf("failed to load test spec: %v", err)
	}

	op, err := ExtractOperation(&model.Model, "/api/v2/composed/{id}", "get")
	if err != nil {
		t.Fatalf("extracting operation: %v", err)
	}

	schema, err := op.ResponseSchemaProxy.BuildSchema()
	if err != nil {
		t.Fatalf("building schema: %v", err)
	}

	attrsProxy, _, err := UnwrapJSONAPI(schema)
	if err != nil {
		t.Fatalf("unwrapping JSON:API: %v", err)
	}

	obj, err := ParseSchema(attrsProxy)
	if err != nil {
		t.Fatalf("expected oneOf in schema to parse successfully, got error: %v", err)
	}

	// Find the config field (from oneOf: ConfigTypeA + ConfigTypeB)
	var configField *SchemaField
	for i := range obj.Fields {
		if obj.Fields[i].Name == "config" {
			configField = &obj.Fields[i]
			break
		}
	}
	if configField == nil {
		t.Fatal("config field not found")
	}

	// config should be an object with merged oneOf children
	if configField.Type != FieldTypeObject {
		t.Errorf("config type = %d, want FieldTypeObject", configField.Type)
	}
	if configField.Children == nil {
		t.Fatal("config should have children from oneOf resolution")
	}

	// Should have url (from A), port (from B), timeout (shared)
	childMap := make(map[string]bool)
	for _, c := range configField.Children.Fields {
		childMap[c.Name] = true
	}
	if !childMap["url"] {
		t.Error("config children should include 'url' from ConfigTypeA")
	}
	if !childMap["port"] {
		t.Error("config children should include 'port' from ConfigTypeB")
	}
	if !childMap["timeout"] {
		t.Error("config children should include 'timeout' (shared)")
	}
}

func TestParseSchema_NilProxy(t *testing.T) {
	_, err := ParseSchema(nil)
	if err == nil {
		t.Fatal("expected error for nil proxy")
	}
}

func fieldNames(fields []SchemaField) []string {
	names := make([]string, len(fields))
	for i, f := range fields {
		names[i] = f.Name
	}
	return names
}
