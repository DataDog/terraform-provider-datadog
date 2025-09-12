package customtypes

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisfies the expected interfaces
var _ basetypes.StringValuable = TrimSpaceStringValue{}
var _ basetypes.StringValuableWithSemanticEquals = TrimSpaceStringValue{}

type TrimSpaceStringValue struct {
	basetypes.StringValue
}

func (v TrimSpaceStringValue) Equal(o attr.Value) bool {
	other, ok := o.(TrimSpaceStringValue)

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (v TrimSpaceStringValue) Type(ctx context.Context) attr.Type {
	return TrimSpaceStringType{}
}

func (v TrimSpaceStringValue) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(TrimSpaceStringValue)
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
	return strings.TrimSpace(newValue.ValueString()) == strings.TrimSpace(v.ValueString()), diags
}
