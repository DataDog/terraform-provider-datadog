package validators

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// securityMonitoringDataSourceWarningValidator emits a deprecation warning when
// the value matches a known deprecated security_monitoring data_source. Pair
// with NewEnumValidator for accept/reject; this validator never errors.
type securityMonitoringDataSourceWarningValidator struct{}

func (securityMonitoringDataSourceWarningValidator) Description(context.Context) string {
	return "warns when data_source is set to a deprecated value"
}

func (v securityMonitoringDataSourceWarningValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (securityMonitoringDataSourceWarningValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if req.ConfigValue.ValueString() == "app_sec_spans" {
		resp.Diagnostics.AddAttributeWarning(
			req.Path,
			"app_sec_spans datasource is deprecated",
			"Use data_source = \"spans\" and add @appsec.security_activity:* to your query to keep the same behavior",
		)
	}
}

// SecurityMonitoringDataSourceWarningValidator returns a String validator that
// emits a deprecation warning for the legacy `app_sec_spans` data_source value.
// It does not enforce membership in the enum — chain it with NewEnumValidator
// to preserve accept/reject behaviour.
func SecurityMonitoringDataSourceWarningValidator() validator.String {
	return securityMonitoringDataSourceWarningValidator{}
}
