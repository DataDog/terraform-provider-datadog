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
var _ basetypes.StringTypable = TagFilterExpressionStringType{}

type TagFilterExpressionStringType struct {
	basetypes.StringType
}

func (t TagFilterExpressionStringType) Equal(o attr.Type) bool {
	other, ok := o.(TagFilterExpressionStringType)

	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (t TagFilterExpressionStringType) String() string {
	return "TagFilterExpressionStringType"
}

func (t TagFilterExpressionStringType) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	// TagFilterExpressionStringValue defined in the value type section
	value := TagFilterExpressionStringValue{
		StringValue: in,
	}

	return value, nil
}

func (t TagFilterExpressionStringType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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

func (t TagFilterExpressionStringType) ValueType(ctx context.Context) attr.Value {

	return TagFilterExpressionStringValue{}
}
