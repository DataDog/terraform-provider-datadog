package fwprovider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// aiGatewayPositiveFloat64Validator rejects zero and negative values. The schema enricher appends no text for it,
// so the attribute's own description states the range, which a Between or AtLeast(0) validator would contradict.
type aiGatewayPositiveFloat64Validator struct{}

func (aiGatewayPositiveFloat64Validator) Description(context.Context) string {
	return "value must be greater than 0"
}

func (v aiGatewayPositiveFloat64Validator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (aiGatewayPositiveFloat64Validator) ValidateFloat64(_ context.Context, req validator.Float64Request, resp *validator.Float64Response) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if value := req.ConfigValue.ValueFloat64(); value <= 0 {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid Attribute Value",
			fmt.Sprintf("Attribute %s value must be greater than 0, got: %v", req.Path, value))
	}
}
