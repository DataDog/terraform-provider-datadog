package fwprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/fwutils"
)

// Interface assertions for FrameworkEphemeralResourceWrapper
var (
	_ ephemeral.EphemeralResource                     = &FrameworkEphemeralResourceWrapper{}
	_ ephemeral.EphemeralResourceWithConfigure        = &FrameworkEphemeralResourceWrapper{}
	_ ephemeral.EphemeralResourceWithValidateConfig   = &FrameworkEphemeralResourceWrapper{}
	_ ephemeral.EphemeralResourceWithConfigValidators = &FrameworkEphemeralResourceWrapper{}
	_ ephemeral.EphemeralResourceWithRenew            = &FrameworkEphemeralResourceWrapper{}
	_ ephemeral.EphemeralResourceWithClose            = &FrameworkEphemeralResourceWrapper{}
)

// NewFrameworkEphemeralResourceWrapper creates a new ephemeral resource wrapper following
// the same pattern as the existing FrameworkResourceWrapper
func NewFrameworkEphemeralResourceWrapper(i *ephemeral.EphemeralResource) ephemeral.EphemeralResource {
	return &FrameworkEphemeralResourceWrapper{
		innerResource: i,
	}
}

// FrameworkEphemeralResourceWrapper wraps ephemeral resources to provide consistent behavior
// across all ephemeral resources, following the existing FrameworkResourceWrapper pattern
type FrameworkEphemeralResourceWrapper struct {
	innerResource *ephemeral.EphemeralResource
}

// Metadata implements the core ephemeral.EphemeralResource interface
// Adds provider type name prefix to the resource type name, following existing pattern
func (r *FrameworkEphemeralResourceWrapper) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	(*r.innerResource).Metadata(ctx, req, resp)
	resp.TypeName = req.ProviderTypeName + resp.TypeName
}

// Schema implements the core ephemeral.EphemeralResource interface
// Enriches schema with common framework patterns
func (r *FrameworkEphemeralResourceWrapper) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	(*r.innerResource).Schema(ctx, req, resp)
	fwutils.EnrichFrameworkEphemeralResourceSchema(&resp.Schema)
}

// Open implements the core ephemeral.EphemeralResource interface
// This is where ephemeral resources create/acquire their temporary resources
func (r *FrameworkEphemeralResourceWrapper) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	(*r.innerResource).Open(ctx, req, resp)
}

// Configure implements the optional ephemeral.EphemeralResourceWithConfigure interface
// Uses interface detection to only call if the inner resource supports configuration
func (r *FrameworkEphemeralResourceWrapper) Configure(ctx context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	rCasted, ok := (*r.innerResource).(ephemeral.EphemeralResourceWithConfigure)
	if ok {
		if req.ProviderData == nil {
			return
		}
		_, ok := req.ProviderData.(*FrameworkProvider)
		if !ok {
			resp.Diagnostics.AddError("Unexpected Ephemeral Resource Configure Type", "")
			return
		}

		rCasted.Configure(ctx, req, resp)
	}
}

// ValidateConfig implements the optional ephemeral.EphemeralResourceWithValidateConfig interface
// Uses interface detection to only call if the inner resource supports validation
func (r *FrameworkEphemeralResourceWrapper) ValidateConfig(ctx context.Context, req ephemeral.ValidateConfigRequest, resp *ephemeral.ValidateConfigResponse) {
	if rCasted, ok := (*r.innerResource).(ephemeral.EphemeralResourceWithValidateConfig); ok {
		rCasted.ValidateConfig(ctx, req, resp)
	}
}

// ConfigValidators implements the optional ephemeral.EphemeralResourceWithConfigValidators interface
// Uses interface detection to only call if the inner resource supports declarative validators
func (r *FrameworkEphemeralResourceWrapper) ConfigValidators(ctx context.Context) []ephemeral.ConfigValidator {
	if rCasted, ok := (*r.innerResource).(ephemeral.EphemeralResourceWithConfigValidators); ok {
		return rCasted.ConfigValidators(ctx)
	}
	return nil
}

// Renew implements the optional ephemeral.EphemeralResourceWithRenew interface
// Uses interface detection to only call if the inner resource supports renewal
func (r *FrameworkEphemeralResourceWrapper) Renew(ctx context.Context, req ephemeral.RenewRequest, resp *ephemeral.RenewResponse) {
	if rCasted, ok := (*r.innerResource).(ephemeral.EphemeralResourceWithRenew); ok {
		rCasted.Renew(ctx, req, resp)
	}
}

// Close implements the optional ephemeral.EphemeralResourceWithClose interface
// Uses interface detection to only call if the inner resource supports cleanup
func (r *FrameworkEphemeralResourceWrapper) Close(ctx context.Context, req ephemeral.CloseRequest, resp *ephemeral.CloseResponse) {
	if rCasted, ok := (*r.innerResource).(ephemeral.EphemeralResourceWithClose); ok {
		rCasted.Close(ctx, req, resp)
	}
}
