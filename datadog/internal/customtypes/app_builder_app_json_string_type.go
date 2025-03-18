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
	_ basetypes.StringTypable = (*AppBuilderAppStringType)(nil)
)

// AppBuilderAppStringType is an attribute type that represents a JSON string (RFC 7159). Semantic equality logic is defined for AppBuilderAppStringType
// such that inconsequential differences between JSON strings are ignored (whitespace, property order, etc), similar to jsontypes.NormalizedType,
// but also ignores other differences such as the App's ID, which is ignored in the App Builder API.
type AppBuilderAppStringType struct {
	basetypes.StringType
}

// String returns a human readable string of the type name.
func (t AppBuilderAppStringType) String() string {
	return "AppBuilderAppStringType"
}

// ValueType returns the Value type.
func (t AppBuilderAppStringType) ValueType(ctx context.Context) attr.Value {
	// AppBuilderAppStringValue defined in the value type section
	return AppBuilderAppStringValue{}
}

// Equal returns true if the given type is equivalent.
func (t AppBuilderAppStringType) Equal(o attr.Type) bool {
	other, ok := o.(AppBuilderAppStringType)

	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

// ValueFromString returns a StringValuable type given a StringValue.
func (t AppBuilderAppStringType) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	// AppBuilderAppStringValue defined in the value type section
	return AppBuilderAppStringValue{
		StringValue: in,
	}, nil
}

// ValueFromTerraform returns a Value given a tftypes.Value.  This is meant to convert the tftypes.Value into a more convenient Go type
// for the provider to consume the data with.
func (t AppBuilderAppStringType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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
