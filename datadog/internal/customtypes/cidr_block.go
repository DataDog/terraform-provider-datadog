package customtypes

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var _ basetypes.StringTypable = CidrBlockType{}
var _ basetypes.StringValuable = CidrBlockValue{}
var _ basetypes.StringValuableWithSemanticEquals = CidrBlockValue{}

type CidrBlockValue struct {
	basetypes.StringValue
}

func (v CidrBlockValue) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	// The framework should always pass the correct value type, but always check
	newValue, ok := newValuable.(CidrBlockValue)

	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected value type was received while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Expected Value Type: "+fmt.Sprintf("%T", v)+"\n"+
				"Got Value Type: "+fmt.Sprintf("%T", newValuable),
		)

		return false, diags
	}

	// Skipping error checking if CustomStringValue already implemented RFC3339 validation
	priorTime, _ := time.Parse(time.RFC3339, v.StringValue.ValueString())

	// Skipping error checking if CustomStringValue already implemented RFC3339 validation
	newTime, _ := time.Parse(time.RFC3339, newValue.ValueString())

	// If the times are equivalent, keep the prior value
	return priorTime.Equal(newTime), diags
}

func (v CidrBlockValue) Equal(o attr.Value) bool {
	other, ok := o.(CidrBlockValue)

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (v CidrBlockValue) Type(ctx context.Context) attr.Type {
	return CidrBlockType{}
}

type CidrBlockType struct {
	basetypes.StringType
}

func (t CidrBlockType) Equal(o attr.Type) bool {
	other, ok := o.(CidrBlockType)

	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (t CidrBlockType) String() string {
	return "CidrBlockType"
}

func (t CidrBlockType) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	value := CidrBlockType{
		StringValue: in,
	}

	return value, nil
}

func (t CidrBlockType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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

func (t CidrBlockType) ValueType(ctx context.Context) attr.Value {
	// CustomStringValue defined in the value type section
	return CidrBlockType{}
}
