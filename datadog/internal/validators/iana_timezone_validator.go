package validators

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type ianaTimezoneValidator struct{}

func (ianaTimezoneValidator) Description(context.Context) string {
	return "value must be a valid IANA time zone name"
}

func (v ianaTimezoneValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (ianaTimezoneValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if value == "" || value == "Local" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid IANA time zone",
			fmt.Sprintf("%q is not a valid IANA time zone name.", value),
		)
		return
	}
	if _, err := time.LoadLocation(value); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid IANA time zone",
			fmt.Sprintf("%q is not a valid IANA time zone name: %s", value, err),
		)
	}
}

// IANATimezoneValidator validates names from the IANA Time Zone Database,
// including UTC and regional names such as America/New_York.
func IANATimezoneValidator() validator.String {
	return ianaTimezoneValidator{}
}
