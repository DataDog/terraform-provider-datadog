package planmodifiers

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

type objectObjectRequired struct {
}

func ObjectRequired() planmodifier.Object {
	return objectObjectRequired{}
}

func (m objectObjectRequired) Description(context.Context) string {
	return "Mark object as required."
}

func (m objectObjectRequired) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m objectObjectRequired) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		resp.Diagnostics.AddError("object is required", fmt.Sprintf("property \"%s\" must be defined", req.Path.String()))
	}
}
