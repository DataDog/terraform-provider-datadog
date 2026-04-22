package planmodifiers

import (
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func SortNormalized() planmodifier.List {
	return sortNormalizedModifier{}
}

type sortNormalizedModifier struct{}

func (m sortNormalizedModifier) Description(_ context.Context) string {
	return "Suppress diffs when list elements are the same but in a different order."
}

func (m sortNormalizedModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m sortNormalizedModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() || req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}

	var configVals, stateVals []string
	diags := req.ConfigValue.ElementsAs(ctx, &configVals, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = req.StateValue.ElementsAs(ctx, &stateVals, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(configVals) != len(stateVals) {
		return
	}

	sortedConfig := make([]string, len(configVals))
	copy(sortedConfig, configVals)
	sort.Strings(sortedConfig)

	sortedState := make([]string, len(stateVals))
	copy(sortedState, stateVals)
	sort.Strings(sortedState)

	for i := range sortedConfig {
		if sortedConfig[i] != sortedState[i] {
			return
		}
	}

	// Same elements, different order — suppress the diff by keeping state value.
	elems := make([]types.String, len(stateVals))
	for i, v := range stateVals {
		elems[i] = types.StringValue(v)
	}
	listVal, diags := types.ListValueFrom(ctx, types.StringType, elems)
	resp.Diagnostics.Append(diags...)
	if !resp.Diagnostics.HasError() {
		resp.PlanValue = listVal
	}
}
