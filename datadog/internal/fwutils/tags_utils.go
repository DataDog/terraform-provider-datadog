package fwutils

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

// ApplyIgnoreTagKeys re-injects the values of ignored tag keys from state into planTags, so Terraform neither reports drift on those keys nor strips them on apply.
func ApplyIgnoreTagKeys(ctx context.Context, planTags, stateTags, ignoreKeys types.Set) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	if ignoreKeys.IsNull() || ignoreKeys.IsUnknown() {
		return planTags, diags
	}
	// On create there is no prior state, so state.EffectiveTags is a zero-value Set with no
	// element type. Calling ElementsAs on it panics in ToTerraformValue. There is nothing to
	// pin against anyway, so return the plan tags unchanged (matches StripIgnoredTags' empty-state path).
	if stateTags.IsNull() || stateTags.IsUnknown() {
		return planTags, diags
	}

	var plan, state, keys []string
	diags.Append(planTags.ElementsAs(ctx, &plan, false)...)
	diags.Append(stateTags.ElementsAs(ctx, &state, false)...)
	diags.Append(ignoreKeys.ElementsAs(ctx, &keys, false)...)
	if diags.HasError() {
		return planTags, diags
	}

	result, d := types.SetValueFrom(ctx, types.StringType, utils.StripIgnoredTags(plan, state, keys))
	diags.Append(d...)
	return result, diags
}

func CombineTags(ctx context.Context, rawInputTags types.Set, defaultTags map[string]string) (types.Set, diag.Diagnostics) {
	if len(defaultTags) == 0 && rawInputTags.IsNull() {
		return types.SetValueMust(types.StringType, []attr.Value{}), nil
	} else if len(defaultTags) == 0 {
		return rawInputTags, nil
	}

	var inputTags []string
	rawInputTags.ElementsAs(ctx, &inputTags, false)

	combinedTagMap := make(map[string][]string)
	for _, tag := range inputTags {
		key, value, _ := strings.Cut(tag, ":")
		oldVals, ok := combinedTagMap[key]
		if !ok {
			oldVals = []string{}
		}
		combinedTagMap[key] = append(oldVals, value)
	}
	for k, v := range defaultTags {
		if _, alreadyDefined := combinedTagMap[k]; !alreadyDefined {
			combinedTagMap[k] = []string{v}
		}
	}

	var resultTags []string
	for k, vals := range combinedTagMap {
		for _, v := range vals {
			tag := fmt.Sprintf("%s:%v", k, v)
			if v == "" {
				tag = k
			}
			resultTags = append(resultTags, tag)
		}
	}
	return types.SetValueFrom(ctx, types.StringType, resultTags)
}
