package validators

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type timeFormatValidator struct {
	expectedFormat string
}

func (m timeFormatValidator) Description(context.Context) string {
	return fmt.Sprintf("field is standardized to %v format", m.expectedFormat)
}

func (m timeFormatValidator) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m timeFormatValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if _, err := time.Parse(m.expectedFormat, req.ConfigValue.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("property \"%s\" must be of the format %v", req.Path.String(), m.expectedFormat),
			fmt.Sprintf("was %v", req.ConfigValue.ValueString()),
		)
		return
	}
}

func TimeFormatValidator(expectedFormat string) validator.String {
	return timeFormatValidator{expectedFormat}
}
