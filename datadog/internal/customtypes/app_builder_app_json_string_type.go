// reference: https://github.com/hashicorp/terraform-plugin-framework-jsontypes/blob/v0.2.0/jsontypes/normalized_type.go

package customtypes

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Ensure the implementation satisfies the expected interfaces
// Not adding validation because that will be handled in the App Builder API
var (
	_ basetypes.StringTypable = (*AppBuilderAppJSONStringType)(nil)
)

// AppBuilderAppJSONStringType is an attribute type that represents a JSON string (RFC 7159). Semantic equality logic is defined for AppBuilderAppJSONStringType
// such that inconsequential differences between JSON strings are ignored (whitespace, property order, etc), similar to jsontypes.NormalizedType,
// but also ignores other differences such as the App's ID, which is ignored in the App Builder API.
type AppBuilderAppJSONStringType struct {
	basetypes.StringType
}

// String returns a human readable string of the type name.
func (t AppBuilderAppJSONStringType) String() string {
	return "AppBuilderAppJSONStringType"
}

// ValueType returns the Value type.
func (t AppBuilderAppJSONStringType) ValueType(ctx context.Context) attr.Value {
	// AppBuilderAppJSONStringValue defined in the value type section
	return AppBuilderAppJSONStringValue{}
}

// Equal returns true if the given type is equivalent.
func (t AppBuilderAppJSONStringType) Equal(o attr.Type) bool {
	other, ok := o.(AppBuilderAppJSONStringType)

	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

// ValueFromString returns a StringValuable type given a StringValue.
func (t AppBuilderAppJSONStringType) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	// AppBuilderAppJSONStringValue defined in the value type section
	return AppBuilderAppJSONStringValue{
		StringValue: in,
	}, nil
}

// ValueFromTerraform returns a Value given a tftypes.Value.  This is meant to convert the tftypes.Value into a more convenient Go type
// for the provider to consume the data with.
func (t AppBuilderAppJSONStringType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromString(ctx, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}

	return stringValuable, nil
}
