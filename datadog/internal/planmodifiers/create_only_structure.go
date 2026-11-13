package planmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// RequiresReplaceObjectStructure checks an immutable object's presence without
// comparing its computed descendants. Writable descendants must have their own
// replacement modifiers. Framework marks computed children unknown on updates,
// before running the parent's modifiers, even when its inputs are unchanged.
func RequiresReplaceObjectStructure() planmodifier.Object {
	return objectplanmodifier.RequiresReplaceIf(
		func(_ context.Context, req planmodifier.ObjectRequest, resp *objectplanmodifier.RequiresReplaceIfFuncResponse) {
			resp.RequiresReplace = req.PlanValue.IsNull() != req.StateValue.IsNull() ||
				req.PlanValue.IsUnknown() || req.StateValue.IsUnknown()
		},
		"Changing the presence of this object requires replacing the resource.",
		"Changing the presence of this object requires replacing the resource.",
	)
}

// RequiresReplaceListStructure checks an immutable object list's presence and
// length. Its writable element fields must have their own replacement modifiers;
// comparing complete elements here would also compare computed descendants.
func RequiresReplaceListStructure() planmodifier.List {
	return listplanmodifier.RequiresReplaceIf(
		func(_ context.Context, req planmodifier.ListRequest, resp *listplanmodifier.RequiresReplaceIfFuncResponse) {
			resp.RequiresReplace = req.PlanValue.IsNull() != req.StateValue.IsNull() ||
				req.PlanValue.IsUnknown() || req.StateValue.IsUnknown() ||
				len(req.PlanValue.Elements()) != len(req.StateValue.Elements())
			if resp.RequiresReplace {
				return
			}
			stateElements := req.StateValue.Elements()
			for i, planned := range req.PlanValue.Elements() {
				prior := stateElements[i]
				if planned.IsNull() != prior.IsNull() || planned.IsUnknown() || prior.IsUnknown() {
					resp.RequiresReplace = true
					return
				}
			}
		},
		"Changing the presence or length of this list requires replacing the resource.",
		"Changing the presence or length of this list requires replacing the resource.",
	)
}
