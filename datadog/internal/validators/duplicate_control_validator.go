package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type controlNameValidator struct{}

func (v controlNameValidator) Description(context.Context) string {
	return "checks for duplicate control names"
}

func (v controlNameValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v controlNameValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var controlNames []string
	for _, control := range req.ConfigValue.Elements() {
		controlObj := control.(types.Object)
		name := controlObj.Attributes()["name"].(types.String).ValueString()
		controlNames = append(controlNames, name)
	}

	seen := make(map[string]bool)
	for _, name := range controlNames {
		if seen[name] {
			resp.Diagnostics.AddError(
				"Each Control must have a unique name under the same requirement",
				fmt.Sprintf("Control name '%s' is used more than once under the same requirement", name),
			)
			return
		}
		seen[name] = true
	}
}

func ControlNameValidator() validator.List {
	return &controlNameValidator{}
}
