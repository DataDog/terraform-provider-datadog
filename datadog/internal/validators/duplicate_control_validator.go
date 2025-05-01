package validators

import (
	"context"
	"fmt"
	"log"

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

func (v controlNameValidator) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	// Get all control names from the configuration
	var controlNames []string
	for _, control := range req.ConfigValue.Elements() {
		controlObj := control.(types.Object)
		name := controlObj.Attributes()["name"].(types.String).ValueString()
		log.Printf("Found control name in config: %s", name)
		controlNames = append(controlNames, name)
	}

	log.Printf("Found %d control names in config", len(controlNames))

	// Check for duplicates in the list
	seen := make(map[string]bool)
	for _, name := range controlNames {
		log.Printf("Checking control name: %s", name)
		if seen[name] {
			log.Printf("Found duplicate control name: %s", name)
			resp.Diagnostics.AddError(
				"400 Bad Request",
				fmt.Sprintf("Control name '%s' is used more than once. Each control must have a unique name.", name),
			)
			return
		}
		seen[name] = true
	}
}

func ControlNameValidator() validator.Set {
	return &controlNameValidator{}
}
