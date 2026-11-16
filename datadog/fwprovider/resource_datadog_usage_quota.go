package fwprovider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var _ resource.ResourceWithConfigure = &usageQuotaResource{}
var _ resource.ResourceWithImportState = &usageQuotaResource{}

type usageQuotaResource struct {
	api       *datadogV2.UsageMeteringApi
	createAPI *datadogV2.UsageMeteringApi
	auth      context.Context
}

func NewUsageQuotaResource() resource.Resource { return &usageQuotaResource{} }
func (r *usageQuotaResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "usage_quota"
}
func (r *usageQuotaResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = usageQuotaSchema()
}

func (r *usageQuotaResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	p, ok := req.ProviderData.(*FrameworkProvider)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider configuration", "Expected FrameworkProvider.")
		return
	}
	r.api = p.DatadogApiInstances.GetUsageMeteringApiV2()
	// Copy the configuration by value, preserving provider settings and future SDK
	// fields. Shared maps remain read-only; preview operations are enabled centrally.
	cfg := *p.DatadogApiInstances.HttpClient.GetConfig()
	// A create can commit before returning a 5xx. Retrying would report a conflict
	// for the quota just created. Require explicit import after ambiguous failures.
	cfg.RetryConfiguration.EnableRetry = false
	r.createAPI = datadogV2.NewUsageMeteringApi(datadog.NewAPIClient(&cfg))
	r.auth = p.Auth
}

func (r *usageQuotaResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var state usageQuotaModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body, diags := buildUsageQuotaCreate(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, httpResp, err := r.createAPI.CreateQuotas(r.auth, state.QuotaNamespace.ValueString(), body, *datadogV2.NewCreateQuotasOptionalParameters().WithCreateOnly(true).WithIncludeDescendants(false))
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusConflict {
			resp.Diagnostics.AddError("Usage quota already exists", "Import the existing quota using <quota_namespace>/<quota_id> before managing it. Terraform has not adopted or changed the existing quota.")
		} else {
			resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "Error creating usage quota. If the request may have succeeded, inspect the quotas and import the created quota before retrying."))
		}
		return
	}
	if result.UnparsedObject != nil || len(result.Data) != 1 {
		resp.Diagnostics.AddError("Invalid quota create response", "Expected exactly one quota. Inspect the quotas and import any created quota before retrying.")
		return
	}
	item := result.Data[0]
	if item.Attributes.Error != nil {
		resp.Diagnostics.AddError("Error creating usage quota", item.Attributes.GetError())
		return
	}
	if item.Id == "" {
		resp.Diagnostics.AddError("Invalid quota create response", "The response did not include a quota ID. Inspect quotas before retrying.")
		return
	}
	// Save the identity even if a subsequent response validation fails.
	state.QuotaID = types.StringValue(item.Id)
	state.ID = types.StringValue(state.QuotaNamespace.ValueString() + "/" + item.Id)
	state.OrgPublicID = types.StringValue(item.Attributes.GetOrgPublicId())
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	if item.UnparsedObject != nil || item.Attributes.UnparsedObject != nil || item.Attributes.UsageLimit == nil || item.Attributes.Enforced == nil || item.Attributes.OrgPublicId == nil {
		resp.Diagnostics.AddError("Incomplete quota create response", "The quota ID was saved. Refresh state before making further changes.")
		return
	}
	data := datadogV2.UsageQuotaResponseData{Id: item.Id, Type: item.Type, Attributes: datadogV2.UsageQuotaResponseAttributes{UsageLimit: item.Attributes.GetUsageLimit(), Enforced: item.Attributes.GetEnforced(), OrgPublicId: item.Attributes.GetOrgPublicId(), Scope: item.Attributes.Scope, PendingUsageLimit: item.Attributes.PendingUsageLimit}}
	resp.Diagnostics.Append(flattenUsageQuota(ctx, &state, data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *usageQuotaResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state usageQuotaModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	result, h, err := r.api.GetQuota(r.auth, state.QuotaNamespace.ValueString(), state.QuotaID.ValueString(), *datadogV2.NewGetQuotaOptionalParameters().WithIncludeDescendants(false))
	if err != nil {
		if h != nil && h.StatusCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "Error reading usage quota"))
		return
	}
	if result.UnparsedObject != nil {
		resp.Diagnostics.AddError("Invalid quota response", "Unable to decode quota response.")
		return
	}
	resp.Diagnostics.Append(flattenUsageQuota(ctx, &state, result.Data)...)
	if !resp.Diagnostics.HasError() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	}
}

func (r *usageQuotaResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state usageQuotaModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body, changed, diags := buildUsageQuotaUpdate(plan, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if changed {
		result, _, err := r.api.UpdateQuota(r.auth, state.QuotaNamespace.ValueString(), state.QuotaID.ValueString(), body, *datadogV2.NewUpdateQuotaOptionalParameters().WithIncludeDescendants(false))
		if err != nil {
			resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "Error updating usage quota"))
			return
		}
		if result.UnparsedObject != nil {
			resp.Diagnostics.AddError("Invalid quota response", "Unable to decode quota response.")
			return
		}
		resp.Diagnostics.Append(flattenUsageQuota(ctx, &plan, result.Data)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *usageQuotaResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state usageQuotaModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	h, err := r.api.DeleteQuota(r.auth, state.QuotaNamespace.ValueString(), state.QuotaID.ValueString(), *datadogV2.NewDeleteQuotaOptionalParameters().WithIncludeDescendants(false))
	if err != nil && (h == nil || h.StatusCode != http.StatusNotFound) {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "Error deleting usage quota"))
	}
}

func (r *usageQuotaResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	namespace, id, err := parseUsageQuotaImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid usage quota import ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("quota_namespace"), namespace)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("quota_id"), id)...)
}

func flattenUsageQuota(ctx context.Context, state *usageQuotaModel, data datadogV2.UsageQuotaResponseData) diag.Diagnostics {
	var diags diag.Diagnostics
	if data.UnparsedObject != nil || data.Attributes.UnparsedObject != nil || data.Id == "" || data.Attributes.OrgPublicId == "" {
		diags.AddError("Invalid quota response", "The response did not contain a valid quota and owning organization.")
		return diags
	}
	if !state.QuotaID.IsNull() && !state.QuotaID.IsUnknown() && state.QuotaID.ValueString() != data.Id {
		diags.AddError("Unexpected quota ID", "The API returned a different quota ID.")
		return diags
	}
	state.QuotaID = types.StringValue(data.Id)
	state.ID = types.StringValue(state.QuotaNamespace.ValueString() + "/" + data.Id)
	state.OrgPublicID = types.StringValue(data.Attributes.OrgPublicId)
	state.Enforced = types.BoolValue(data.Attributes.Enforced)
	state.UsageLimit = types.Float64Value(data.Attributes.UsageLimit)
	state.Scope = types.MapNull(types.StringType)
	if len(data.Attributes.Scope) > 0 {
		var d diag.Diagnostics
		state.Scope, d = types.MapValueFrom(ctx, types.StringType, data.Attributes.Scope)
		diags.Append(d...)
	}
	if data.Attributes.PendingUsageLimit != nil {
		diags.AddWarning("Scheduled quota change exists", fmt.Sprintf("A pending limit of %g is configured outside Terraform. It will change the effective limit when it takes effect. Disabling enforcement clears it; other updates preserve it.", *data.Attributes.PendingUsageLimit))
	}
	return diags
}
