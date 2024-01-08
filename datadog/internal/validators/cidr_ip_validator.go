package validators

import (
	"context"
	"fmt"
	"net"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type cidrIpValidator struct {
}

func (v cidrIpValidator) Description(ctx context.Context) string {
	return "String must be a valid CIDR block or IP address"
}

func (v cidrIpValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v cidrIpValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	// If the value is unknown or null, there is nothing to validate.
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}
	ip := net.ParseIP(req.ConfigValue.ValueString())
	if _, _, err := net.ParseCIDR(req.ConfigValue.ValueString()); err != nil && ip == nil {
		resp.Diagnostics.AddAttributeError(req.Path, "", fmt.Sprintf("String %s must be a valid CIDR block or IP address", req.ConfigValue.ValueString()))
	}
}

func CidrIpValidator() validator.String {
	return cidrIpValidator{}
}
