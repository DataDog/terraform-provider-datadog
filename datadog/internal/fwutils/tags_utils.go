package fwutils

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
