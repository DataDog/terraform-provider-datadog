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
var _ basetypes.StringTypable = MonitorTypeType{}

type MonitorTypeType struct {
	basetypes.StringType
}

func (t MonitorTypeType) Equal(o attr.Type) bool {
	other, ok := o.(MonitorTypeType)

	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (t MonitorTypeType) String() string {
	return "MonitorTypeType"
}

func (t MonitorTypeType) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	value := MonitorTypeValue{
		StringValue: in,
	}

	return value, nil
}

func (t MonitorTypeType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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

func (t MonitorTypeType) ValueType(ctx context.Context) attr.Value {

	return MonitorTypeValue{}
}
