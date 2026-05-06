package types

import (
	"testing"

	"github.com/DataDog/terraform-provider-datadog/generator/internal/openapi"
)

func TestToGoType(t *testing.T) {
	tests := []struct {
		name      string
		fieldType openapi.FieldType
		want      string
	}{
		{"string", openapi.FieldTypeString, "string"},
		{"int64", openapi.FieldTypeInt64, "int64"},
		{"int32", openapi.FieldTypeInt32, "int32"},
		{"float64", openapi.FieldTypeFloat64, "float64"},
		{"float32", openapi.FieldTypeFloat32, "float32"},
		{"bool", openapi.FieldTypeBool, "bool"},
		{"object", openapi.FieldTypeObject, "object"},
		{"array of objects", openapi.FieldTypeArrayOfObjects, "[]object"},
		{"array of strings", openapi.FieldTypeArrayOfStrings, "[]string"},
		{"array of ints", openapi.FieldTypeArrayOfInts, "[]int64"},
		{"array of floats", openapi.FieldTypeArrayOfFloats, "[]float64"},
		{"array of bools", openapi.FieldTypeArrayOfBools, "[]bool"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := openapi.SchemaField{Type: tt.fieldType}
			got := ToGoType(field)
			if got != tt.want {
				t.Errorf("ToGoType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestToTerraformSchemaType(t *testing.T) {
	tests := []struct {
		name      string
		fieldType openapi.FieldType
		want      string
	}{
		{"string", openapi.FieldTypeString, "schema.StringAttribute"},
		{"int64", openapi.FieldTypeInt64, "schema.Int64Attribute"},
		{"int32", openapi.FieldTypeInt32, "schema.Int64Attribute"},
		{"float64", openapi.FieldTypeFloat64, "schema.Float64Attribute"},
		{"float32", openapi.FieldTypeFloat32, "schema.Float64Attribute"},
		{"bool", openapi.FieldTypeBool, "schema.BoolAttribute"},
		{"array of strings", openapi.FieldTypeArrayOfStrings, "schema.ListAttribute"},
		{"array of ints", openapi.FieldTypeArrayOfInts, "schema.ListAttribute"},
		{"array of floats", openapi.FieldTypeArrayOfFloats, "schema.ListAttribute"},
		{"array of bools", openapi.FieldTypeArrayOfBools, "schema.ListAttribute"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := openapi.SchemaField{Type: tt.fieldType}
			got := ToTerraformSchemaType(field)
			if got != tt.want {
				t.Errorf("ToTerraformSchemaType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestToTerraformValueType(t *testing.T) {
	tests := []struct {
		name      string
		fieldType openapi.FieldType
		want      string
	}{
		{"string", openapi.FieldTypeString, "types.String"},
		{"int64", openapi.FieldTypeInt64, "types.Int64"},
		{"int32", openapi.FieldTypeInt32, "types.Int64"},
		{"float64", openapi.FieldTypeFloat64, "types.Float64"},
		{"float32", openapi.FieldTypeFloat32, "types.Float64"},
		{"bool", openapi.FieldTypeBool, "types.Bool"},
		{"array of strings", openapi.FieldTypeArrayOfStrings, "types.List"},
		{"array of ints", openapi.FieldTypeArrayOfInts, "types.List"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := openapi.SchemaField{Type: tt.fieldType}
			got := ToTerraformValueType(field)
			if got != tt.want {
				t.Errorf("ToTerraformValueType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestToModelType(t *testing.T) {
	field := openapi.SchemaField{Type: openapi.FieldTypeString}
	got := ToModelType(field)
	if got != "types.String" {
		t.Errorf("ToModelType() = %q, want %q", got, "types.String")
	}
}

func TestTypeValueConstructor(t *testing.T) {
	tests := []struct {
		name      string
		fieldType openapi.FieldType
		want      string
	}{
		{"string", openapi.FieldTypeString, "types.StringValue"},
		{"int64", openapi.FieldTypeInt64, "types.Int64Value"},
		{"int32", openapi.FieldTypeInt32, "types.Int64Value"},
		{"float64", openapi.FieldTypeFloat64, "types.Float64Value"},
		{"bool", openapi.FieldTypeBool, "types.BoolValue"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := openapi.SchemaField{Type: tt.fieldType}
			got := TypeValueConstructor(field)
			if got != tt.want {
				t.Errorf("TypeValueConstructor() = %q, want %q", got, tt.want)
			}
		})
	}
}

// T103: Map type conversion tests
func TestToGoType_Maps(t *testing.T) {
	tests := []struct {
		name      string
		fieldType openapi.FieldType
		want      string
	}{
		{"map of strings", openapi.FieldTypeMapOfStrings, "map[string]string"},
		{"map of ints", openapi.FieldTypeMapOfInts, "map[string]int64"},
		{"map of floats", openapi.FieldTypeMapOfFloats, "map[string]float64"},
		{"map of bools", openapi.FieldTypeMapOfBools, "map[string]bool"},
		{"map of objects", openapi.FieldTypeMapOfObjects, "map[string]object"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := openapi.SchemaField{Type: tt.fieldType}
			got := ToGoType(field)
			if got != tt.want {
				t.Errorf("ToGoType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestToTerraformSchemaType_Maps(t *testing.T) {
	tests := []struct {
		name      string
		fieldType openapi.FieldType
		want      string
	}{
		{"map of strings", openapi.FieldTypeMapOfStrings, "schema.MapAttribute"},
		{"map of ints", openapi.FieldTypeMapOfInts, "schema.MapAttribute"},
		{"map of floats", openapi.FieldTypeMapOfFloats, "schema.MapAttribute"},
		{"map of bools", openapi.FieldTypeMapOfBools, "schema.MapAttribute"},
		{"map of objects", openapi.FieldTypeMapOfObjects, "schema.MapNestedAttribute"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := openapi.SchemaField{Type: tt.fieldType}
			got := ToTerraformSchemaType(field)
			if got != tt.want {
				t.Errorf("ToTerraformSchemaType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestToTerraformValueType_Maps(t *testing.T) {
	tests := []struct {
		name      string
		fieldType openapi.FieldType
		want      string
	}{
		{"map of strings", openapi.FieldTypeMapOfStrings, "types.Map"},
		{"map of ints", openapi.FieldTypeMapOfInts, "types.Map"},
		{"map of floats", openapi.FieldTypeMapOfFloats, "types.Map"},
		{"map of bools", openapi.FieldTypeMapOfBools, "types.Map"},
		{"map of objects", openapi.FieldTypeMapOfObjects, "types.Map"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := openapi.SchemaField{Type: tt.fieldType}
			got := ToTerraformValueType(field)
			if got != tt.want {
				t.Errorf("ToTerraformValueType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMapElementType(t *testing.T) {
	tests := []struct {
		name      string
		fieldType openapi.FieldType
		want      string
	}{
		{"map of strings", openapi.FieldTypeMapOfStrings, "types.StringType"},
		{"map of ints", openapi.FieldTypeMapOfInts, "types.Int64Type"},
		{"map of floats", openapi.FieldTypeMapOfFloats, "types.Float64Type"},
		{"map of bools", openapi.FieldTypeMapOfBools, "types.BoolType"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := openapi.SchemaField{Type: tt.fieldType}
			got := MapElementType(field)
			if got != tt.want {
				t.Errorf("MapElementType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsMapType(t *testing.T) {
	mapTypes := []openapi.FieldType{
		openapi.FieldTypeMapOfStrings, openapi.FieldTypeMapOfInts,
		openapi.FieldTypeMapOfFloats, openapi.FieldTypeMapOfBools,
		openapi.FieldTypeMapOfObjects,
	}
	for _, ft := range mapTypes {
		if !IsMapType(openapi.SchemaField{Type: ft}) {
			t.Errorf("IsMapType should be true for %d", ft)
		}
	}

	nonMapTypes := []openapi.FieldType{
		openapi.FieldTypeString, openapi.FieldTypeInt64, openapi.FieldTypeBool,
		openapi.FieldTypeArrayOfStrings, openapi.FieldTypeObject,
	}
	for _, ft := range nonMapTypes {
		if IsMapType(openapi.SchemaField{Type: ft}) {
			t.Errorf("IsMapType should be false for %d", ft)
		}
	}
}

func TestNeedsCast(t *testing.T) {
	tests := []struct {
		name      string
		fieldType openapi.FieldType
		want      string
	}{
		{"string", openapi.FieldTypeString, ""},
		{"int64", openapi.FieldTypeInt64, ""},
		{"int32", openapi.FieldTypeInt32, "int64"},
		{"float64", openapi.FieldTypeFloat64, ""},
		{"float32", openapi.FieldTypeFloat32, "float64"},
		{"bool", openapi.FieldTypeBool, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := openapi.SchemaField{Type: tt.fieldType}
			got := NeedsCast(field)
			if got != tt.want {
				t.Errorf("NeedsCast() = %q, want %q", got, tt.want)
			}
		})
	}
}
