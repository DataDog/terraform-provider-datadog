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
var _ basetypes.StringTypable = YAMLStringType{}

type YAMLStringType struct {
	basetypes.StringType
}

func (t YAMLStringType) Equal(o attr.Type) bool {
	other, ok := o.(YAMLStringType)

	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (t YAMLStringType) String() string {
	return "YAMLStringType"
}

func (t YAMLStringType) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	// YAMLStringValue defined in the value type section
	value := YAMLStringValue{
		StringValue: in,
	}

	return value, nil
}

func (t YAMLStringType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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

func (t YAMLStringType) ValueType(ctx context.Context) attr.Value {
	// YAMLStringValue defined in the value type section
	return YAMLStringValue{}
}
