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
var _ basetypes.StringTypable = TrimSpaceStringType{}

type TrimSpaceStringType struct {
	basetypes.StringType
}

func (t TrimSpaceStringType) Equal(o attr.Type) bool {
	other, ok := o.(TrimSpaceStringType)

	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (t TrimSpaceStringType) String() string {
	return "TrimSpaceStringType"
}

func (t TrimSpaceStringType) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	// TrimSpaceStringValue defined in the value type section
	value := TrimSpaceStringValue{
		StringValue: in,
	}

	return value, nil
}

func (t TrimSpaceStringType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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

func (t TrimSpaceStringType) ValueType(ctx context.Context) attr.Value {

	return TrimSpaceStringValue{}
}
