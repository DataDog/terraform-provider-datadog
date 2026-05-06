package planmodifiers

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// ImmutableString returns a plan modifier that rejects any change to the
// attribute after the resource has been created. Use it in place of
// `stringplanmodifier.RequiresReplace()` when you want terraform plan to fail
// loudly instead of silently destroying and recreating the resource — a
// common preference for credential-bearing fields where an unexpected
// rotation has wider blast radius than a noisy plan failure.
func ImmutableString(attribute string) planmodifier.String {
	return immutableStringPlanModifier{attribute: attribute}
}

type immutableStringPlanModifier struct {
	attribute string
}

func (m immutableStringPlanModifier) Description(_ context.Context) string {
	return fmt.Sprintf("%s cannot be modified after the resource is created.", m.attribute)
}

func (m immutableStringPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m immutableStringPlanModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Use the whole-resource state/plan flags to decide create vs destroy vs
	// update. Looking at req.StateValue alone is misleading because an
	// optional attribute can be null on an existing resource.
	if req.State.Raw.IsNull() {
		// Resource is being created.
		return
	}
	if req.Plan.Raw.IsNull() {
		// Resource is being destroyed.
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		// No change.
		return
	}
	resp.Diagnostics.AddAttributeError(
		req.Path,
		fmt.Sprintf("%s cannot be modified after creation", m.attribute),
		fmt.Sprintf("The Datadog API does not allow updating %s after the resource is created. To change this value, destroy and re-create the resource.", m.attribute),
	)
}
