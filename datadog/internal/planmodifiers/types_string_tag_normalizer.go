package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func NormalizeTag() planmodifier.String {
	return normalizeTagModifier{}
}

type normalizeTagModifier struct{}

func (m normalizeTagModifier) Description(_ context.Context) string {
	return "Normalize tag value."
}

func (m normalizeTagModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m normalizeTagModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	val := req.ConfigValue.ValueString()
	resp.PlanValue = types.StringValue(utils.NormalizeTag(val))
}
