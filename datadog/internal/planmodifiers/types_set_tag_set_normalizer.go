package planmodifiers

import (
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func NormalizeTagSet() planmodifier.Set {
	return normalizeTagSetModifier{}
}

type normalizeTagSetModifier struct{}

func (m normalizeTagSetModifier) Description(_ context.Context) string {
	return "Normalize tag set value."
}

func (m normalizeTagSetModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m normalizeTagSetModifier) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	elementsList := req.ConfigValue.Elements()
	for i, v := range elementsList {
		elementsList[i] = types.StringValue(utils.NormalizeTag(v.String()))
	}

	sortedList := sortAttrValueString(elementsList)
	resp.PlanValue = types.SetValueMust(types.StringType, sortedList)
}

// sort []attr.Value by converting each value to string, sorting the strings, and converting back to []attr.Value
func sortAttrValueString(v []attr.Value) []attr.Value {
	s := make([]string, len(v))
	for i, val := range v {
		s[i] = val.String()
	}
	sort.Strings(s)

	result := make([]attr.Value, len(s))
	for i, val := range s {
		v[i] = types.StringValue(val)
	}
	return result
}
