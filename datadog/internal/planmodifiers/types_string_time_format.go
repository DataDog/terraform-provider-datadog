package planmodifiers

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type TimeFormatModifier struct {
	format string
}

func (m TimeFormatModifier) Description(context.Context) string {
	return fmt.Sprintf("field is standardized to %v format", m.format)
}

func (m TimeFormatModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m TimeFormatModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// Validates string datetime format and converts it to a standardized output format
	if resp.PlanValue.IsNull() || resp.PlanValue.IsUnknown() {
		return
	}
	planTime, err := time.Parse(m.format, resp.PlanValue.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("property \"%s\" must be of the format %v", req.Path.String(), m.format), fmt.Sprintf("was %v", resp.PlanValue.ValueString()))
		return
	}
	resp.PlanValue = types.StringValue(planTime.Format(m.format))
}

func TimeFormat(format string) planmodifier.String {
	return TimeFormatModifier{format}
}
