package modifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func NormalizeTagPlanModifierSet() *tagTypeSetAttributePlanModifier {
	return &tagTypeSetAttributePlanModifier{}
}

type tagTypeSetAttributePlanModifier struct {
}

func (m *tagTypeSetAttributePlanModifier) Description(_ context.Context) string {
	return "Normalizes tags."
}

func (m *tagTypeSetAttributePlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m *tagTypeSetAttributePlanModifier) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	if req.StateValue.IsNull() {
		return
	}

	if req.PlanValue.ElementType(ctx) != types.StringType {
		return
	}

	var vals []string
	req.PlanValue.ElementsAs(ctx, &vals, false)

	for i, _ := range vals {
		vals[i] = utils.NormalizeTag(vals[i])
	}

	resp.PlanValue, resp.Diagnostics = types.SetValueFrom(ctx, types.StringType, vals)
}
