package customtypes

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisfies the expected interfaces
var _ basetypes.StringValuable = MonitorTypeValue{}
var _ basetypes.StringValuableWithSemanticEquals = MonitorTypeValue{}

type MonitorTypeValue struct {
	basetypes.StringValue
}

func (v MonitorTypeValue) Equal(o attr.Value) bool {
	other, ok := o.(MonitorTypeValue)

	if !ok {
		return false
	}
	return v.StringValue.Equal(other.StringValue)
}

func (v MonitorTypeValue) Type(ctx context.Context) attr.Type {
	return MonitorTypeType{}
}

// Datadog API quirk, see https://github.com/hashicorp/terraform/issues/13784
func (v MonitorTypeValue) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(MonitorTypeValue)
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
	newVal := newValue.ValueString()
	oldVal := v.ValueString()
	if (oldVal == "query alert" && newVal == "metric alert") ||
		(oldVal == "metric alert" && newVal == "query alert") {
		return true, diags
	}

	return newVal == oldVal, diags
}
