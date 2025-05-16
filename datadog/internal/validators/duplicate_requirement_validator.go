package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type requirementNameValidator struct{}

func (v requirementNameValidator) Description(context.Context) string {
	return "checks for duplicate requirement names"
}

func (v requirementNameValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v requirementNameValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var requirementNames []string
	for _, requirement := range req.ConfigValue.Elements() {
		reqObj := requirement.(types.Object)
		name := reqObj.Attributes()["name"].(types.String).ValueString()
		requirementNames = append(requirementNames, name)
	}

	seen := make(map[string]bool)
	for _, name := range requirementNames {
		if seen[name] {
			resp.Diagnostics.AddError(
				"Each Requirement must have a unique name",
				fmt.Sprintf("Requirement name '%s' is used more than once.", name),
			)
			return
		}
		seen[name] = true
	}
}

func RequirementNameValidator() validator.Set {
	return &requirementNameValidator{}
}
