package fwutils

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.ResourceWithConfigure        = &FrameworkResourceWrapper{}
	_ resource.ResourceWithImportState      = &FrameworkResourceWrapper{}
	_ resource.ResourceWithConfigValidators = &FrameworkResourceWrapper{}
	_ resource.ResourceWithModifyPlan       = &FrameworkResourceWrapper{}
	_ resource.ResourceWithUpgradeState     = &FrameworkResourceWrapper{}
	_ resource.ResourceWithValidateConfig   = &FrameworkResourceWrapper{}
)

func NewFrameworkResourceWrapper(i *resource.Resource) resource.Resource {
	return &FrameworkResourceWrapper{
		innerResource: i,
	}
}

type FrameworkResourceWrapper struct {
	innerResource *resource.Resource
}

func (r *FrameworkResourceWrapper) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	rCasted, ok := (*r.innerResource).(resource.ResourceWithConfigure)
	if ok {
		if req.ProviderData == nil {
			return
		}
		rCasted.Configure(ctx, req, resp)
	}
}

func (r *FrameworkResourceWrapper) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	(*r.innerResource).Metadata(ctx, req, resp)
	resp.TypeName = req.ProviderTypeName + resp.TypeName
}

func (r *FrameworkResourceWrapper) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	(*r.innerResource).Schema(ctx, req, resp)
	resp.Schema = enrichFrameworkResourceSchema(resp.Schema)
}

func (r *FrameworkResourceWrapper) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	(*r.innerResource).Create(ctx, req, resp)
}

func (r *FrameworkResourceWrapper) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	(*r.innerResource).Read(ctx, req, resp)
}

func (r *FrameworkResourceWrapper) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	(*r.innerResource).Update(ctx, req, resp)
}

func (r *FrameworkResourceWrapper) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	(*r.innerResource).Delete(ctx, req, resp)
}

func (r *FrameworkResourceWrapper) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if rCasted, ok := (*r.innerResource).(resource.ResourceWithImportState); ok {
		rCasted.ImportState(ctx, req, resp)
		return
	}

	resp.Diagnostics.AddError(
		"Resource Import Not Implemented",
		"This resource does not support import. Please contact the provider developer for additional information.",
	)
}

func (r *FrameworkResourceWrapper) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	if rCasted, ok := (*r.innerResource).(resource.ResourceWithConfigValidators); ok {
		return rCasted.ConfigValidators(ctx)
	}
	return nil
}

func (r *FrameworkResourceWrapper) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if v, ok := (*r.innerResource).(resource.ResourceWithModifyPlan); ok {
		v.ModifyPlan(ctx, req, resp)
	}
}

func (r *FrameworkResourceWrapper) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	if v, ok := (*r.innerResource).(resource.ResourceWithUpgradeState); ok {
		return v.UpgradeState(ctx)
	}
	return nil
}

func (r *FrameworkResourceWrapper) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	if v, ok := (*r.innerResource).(resource.ResourceWithValidateConfig); ok {
		v.ValidateConfig(ctx, req, resp)
	}
}
