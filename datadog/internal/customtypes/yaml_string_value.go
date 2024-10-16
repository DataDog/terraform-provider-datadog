package customtypes

import (
	"context"
	"fmt"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"gopkg.in/yaml.v3"
)

// Ensure the implementation satisfies the expected interfaces
var _ basetypes.StringValuable = YAMLStringValue{}
var _ basetypes.StringValuableWithSemanticEquals = YAMLStringValue{}

type YAMLStringValue struct {
	basetypes.StringValue
}

func (v YAMLStringValue) Equal(o attr.Value) bool {
	other, ok := o.(YAMLStringValue)

	if !ok {
		return false
	}
	return v.StringValue.Equal(other.StringValue)
}

func (v YAMLStringValue) Type(ctx context.Context) attr.Type {
	// YAMLStringType defined in the schema type section
	return YAMLStringType{}
}

func (v YAMLStringValue) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	// The framework should always pass the correct value type, but always check
	other, ok := newValuable.(YAMLStringValue)

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

	var prev interface{}
	var next interface{}
	yaml.Unmarshal([]byte(v.StringValue.ValueString()), &prev)
	yaml.Unmarshal([]byte(other.StringValue.ValueString()), &next)
	return cmp.Equal(prev, next), diags
}
