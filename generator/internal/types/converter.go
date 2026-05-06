package types

import (
	"github.com/DataDog/terraform-provider-datadog/generator/internal/openapi"
)

// ToGoType returns the Go type string for a schema field.
func ToGoType(field openapi.SchemaField) string {
	switch field.Type {
	case openapi.FieldTypeString:
		return "string"
	case openapi.FieldTypeInt64:
		return "int64"
	case openapi.FieldTypeInt32:
		return "int32"
	case openapi.FieldTypeFloat64:
		return "float64"
	case openapi.FieldTypeFloat32:
		return "float32"
	case openapi.FieldTypeBool:
		return "bool"
	case openapi.FieldTypeObject:
		return "object"
	case openapi.FieldTypeArrayOfObjects:
		return "[]object"
	case openapi.FieldTypeArrayOfStrings:
		return "[]string"
	case openapi.FieldTypeArrayOfInts:
		return "[]int64"
	case openapi.FieldTypeArrayOfFloats:
		return "[]float64"
	case openapi.FieldTypeArrayOfBools:
		return "[]bool"
	case openapi.FieldTypeMapOfStrings:
		return "map[string]string"
	case openapi.FieldTypeMapOfInts:
		return "map[string]int64"
	case openapi.FieldTypeMapOfFloats:
		return "map[string]float64"
	case openapi.FieldTypeMapOfBools:
		return "map[string]bool"
	case openapi.FieldTypeMapOfObjects:
		return "map[string]object"
	default:
		return "string"
	}
}

// ToTerraformSchemaType returns the Terraform schema attribute type string.
func ToTerraformSchemaType(field openapi.SchemaField) string {
	switch field.Type {
	case openapi.FieldTypeString:
		return "schema.StringAttribute"
	case openapi.FieldTypeInt64, openapi.FieldTypeInt32:
		return "schema.Int64Attribute"
	case openapi.FieldTypeFloat64, openapi.FieldTypeFloat32:
		return "schema.Float64Attribute"
	case openapi.FieldTypeBool:
		return "schema.BoolAttribute"
	case openapi.FieldTypeArrayOfStrings:
		return "schema.ListAttribute"
	case openapi.FieldTypeArrayOfInts:
		return "schema.ListAttribute"
	case openapi.FieldTypeArrayOfFloats:
		return "schema.ListAttribute"
	case openapi.FieldTypeArrayOfBools:
		return "schema.ListAttribute"
	case openapi.FieldTypeMapOfStrings, openapi.FieldTypeMapOfInts,
		openapi.FieldTypeMapOfFloats, openapi.FieldTypeMapOfBools:
		return "schema.MapAttribute"
	case openapi.FieldTypeMapOfObjects:
		return "schema.MapNestedAttribute"
	default:
		return "schema.StringAttribute"
	}
}

// ToTerraformValueType returns the Terraform types value type string.
func ToTerraformValueType(field openapi.SchemaField) string {
	switch field.Type {
	case openapi.FieldTypeString:
		return "types.String"
	case openapi.FieldTypeInt64, openapi.FieldTypeInt32:
		return "types.Int64"
	case openapi.FieldTypeFloat64, openapi.FieldTypeFloat32:
		return "types.Float64"
	case openapi.FieldTypeBool:
		return "types.Bool"
	case openapi.FieldTypeArrayOfStrings, openapi.FieldTypeArrayOfInts,
		openapi.FieldTypeArrayOfFloats, openapi.FieldTypeArrayOfBools:
		return "types.List"
	case openapi.FieldTypeMapOfStrings, openapi.FieldTypeMapOfInts,
		openapi.FieldTypeMapOfFloats, openapi.FieldTypeMapOfBools,
		openapi.FieldTypeMapOfObjects:
		return "types.Map"
	default:
		return "types.String"
	}
}

// ToModelType returns the Go model struct type for a Terraform attribute.
func ToModelType(field openapi.SchemaField) string {
	return ToTerraformValueType(field)
}

// ListElementType returns the Terraform element type string for list attributes.
func ListElementType(field openapi.SchemaField) string {
	switch field.Type {
	case openapi.FieldTypeArrayOfStrings:
		return "types.StringType"
	case openapi.FieldTypeArrayOfInts:
		return "types.Int64Type"
	case openapi.FieldTypeArrayOfFloats:
		return "types.Float64Type"
	case openapi.FieldTypeArrayOfBools:
		return "types.BoolType"
	default:
		return "types.StringType"
	}
}

// MapElementType returns the Terraform element type string for map attributes.
func MapElementType(field openapi.SchemaField) string {
	switch field.Type {
	case openapi.FieldTypeMapOfStrings:
		return "types.StringType"
	case openapi.FieldTypeMapOfInts:
		return "types.Int64Type"
	case openapi.FieldTypeMapOfFloats:
		return "types.Float64Type"
	case openapi.FieldTypeMapOfBools:
		return "types.BoolType"
	default:
		return "types.StringType"
	}
}

// IsMapType returns true if the field is a map type.
func IsMapType(field openapi.SchemaField) bool {
	switch field.Type {
	case openapi.FieldTypeMapOfStrings, openapi.FieldTypeMapOfInts,
		openapi.FieldTypeMapOfFloats, openapi.FieldTypeMapOfBools,
		openapi.FieldTypeMapOfObjects:
		return true
	}
	return false
}

// TypeValueConstructor returns the types.XxxValue() function name for a field.
func TypeValueConstructor(field openapi.SchemaField) string {
	switch field.Type {
	case openapi.FieldTypeString:
		return "types.StringValue"
	case openapi.FieldTypeInt64, openapi.FieldTypeInt32:
		return "types.Int64Value"
	case openapi.FieldTypeFloat64, openapi.FieldTypeFloat32:
		return "types.Float64Value"
	case openapi.FieldTypeBool:
		return "types.BoolValue"
	default:
		return "types.StringValue"
	}
}

// NeedsCast returns the cast type needed for SDK value assignment, or empty if none.
func NeedsCast(field openapi.SchemaField) string {
	switch field.Type {
	case openapi.FieldTypeInt32:
		return "int64"
	case openapi.FieldTypeFloat32:
		return "float64"
	default:
		return ""
	}
}
