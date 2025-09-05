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
var _ basetypes.StringTypable = FloatStringType{}

type FloatStringType struct {
	basetypes.StringType
}

func (t FloatStringType) Equal(o attr.Type) bool {
	other, ok := o.(FloatStringType)

	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (t FloatStringType) String() string {
	return "FloatStringType"
}

func (t FloatStringType) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	value := FloatStringValue{
		StringValue: in,
	}

	return value, nil
}

func (t FloatStringType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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

func (t FloatStringType) ValueType(ctx context.Context) attr.Value {

	return FloatStringValue{}
}
