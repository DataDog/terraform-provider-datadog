package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type duplicateRequirementControlValidator struct{}

func (v duplicateRequirementControlValidator) Description(context.Context) string {
	return "checks for duplicate requirement and control names"
}

func (v duplicateRequirementControlValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v duplicateRequirementControlValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	seen := make(map[string]bool)
	for _, requirement := range req.ConfigValue.Elements() {
		reqObj := requirement.(types.Object)
		name := reqObj.Attributes()["name"].(types.String).ValueString()
		if seen[name] {
			resp.Diagnostics.AddError(
				"Each Requirement must have a unique name",
				fmt.Sprintf("Requirement name '%s' is used more than once.", name),
			)
			return
		}
		seen[name] = true
		controls := reqObj.Attributes()["controls"].(types.List)
		controlNames := make(map[string]bool)
		for _, control := range controls.Elements() {
			ctrlObj := control.(types.Object)
			ctrlName := ctrlObj.Attributes()["name"].(types.String).ValueString()
			if controlNames[ctrlName] {
				resp.Diagnostics.AddError(
					"Each Control must have a unique name under the same requirement",
					fmt.Sprintf("Control name '%s' is used more than once under requirement '%s'", ctrlName, name),
				)
				return
			}
			controlNames[ctrlName] = true
		}
	}
}

func DuplicateRequirementControlValidator() validator.List {
	return &duplicateRequirementControlValidator{}
}
