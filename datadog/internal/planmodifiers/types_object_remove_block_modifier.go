package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

type removeBlockModifier struct {
}

func RemoveBlockModifier() planmodifier.Object {
	return removeBlockModifier{}
}

func (m removeBlockModifier) Description(context.Context) string {
	return "Set removed block to null."
}

func (m removeBlockModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m removeBlockModifier) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	// Older versions of terraform have a bug where removing a block results in 'planned for existence but config wants absence'.
	// To work around this we can set the block to null.
	// Reference: https://github.com/hashicorp/terraform/issues/32460
	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/603#issuecomment-1371358108
	if req.ConfigValue.IsNull() {
		resp.PlanValue = req.ConfigValue
	}
}
