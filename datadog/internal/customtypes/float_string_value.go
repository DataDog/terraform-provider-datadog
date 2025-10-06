package customtypes

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisfies the expected interfaces
var _ basetypes.StringValuable = FloatStringValue{}
var _ basetypes.StringValuableWithSemanticEquals = FloatStringValue{}

type FloatStringValue struct {
	basetypes.StringValue
}

func (v FloatStringValue) Equal(o attr.Value) bool {
	other, ok := o.(FloatStringValue)

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (v FloatStringValue) Type(ctx context.Context) attr.Type {
	return FloatStringType{}
}

func (v FloatStringValue) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(FloatStringValue)
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
	// Exit early if string matches
	if newValue.Equal(v) {
		return true, diags
	}

	newVal := newValue.ValueString()
	oldVal := v.ValueString()
	oldFloat, err := strconv.ParseFloat(oldVal, 64)
	if err != nil {
		diags.AddError("Error parsing float of value: %s", oldVal)
		return false, diags
	}
	newFloat, err := strconv.ParseFloat(newVal, 64)
	if err != nil {
		diags.AddError("Error parsing float of value: %s", newVal)
		return false, diags
	}
	return newFloat == oldFloat, diags
}
