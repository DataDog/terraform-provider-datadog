package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TagsSetIsNormalized() tagsSetIsNormalized {
	return tagsSetIsNormalized{}
}

type tagsSetIsNormalized struct{}

func (v tagsSetIsNormalized) Description(_ context.Context) string {
	return "Tags must be normalized. See docs https://docs.datadoghq.com/getting_started/tagging/#define-tags"
}
func (v tagsSetIsNormalized) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v tagsSetIsNormalized) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	for _, i := range req.ConfigValue.Elements() {
		var val types.String
		diags := tfsdk.ValueAs(ctx, i, &val)
		resp.Diagnostics.Append(diags...)
		if diags.HasError() {
			return
		}

		normalizedVal := utils.NormalizeTag(val.ValueString())
		if val.ValueString() != normalizedVal {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				v.Description(ctx),
				fmt.Sprintf("'%s' should be '%s'", val.ValueString(), normalizedVal),
			)
		}
	}
}
