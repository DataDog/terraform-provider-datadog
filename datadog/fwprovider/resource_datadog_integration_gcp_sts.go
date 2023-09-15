package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &integrationGcpStsResource{}
	_ resource.ResourceWithImportState = &integrationGcpStsResource{}
)

type integrationGcpStsResource struct {
	Api  *datadogV2.GCPIntegrationApi
	Auth context.Context
}

type integrationGcpStsModel struct {
	ID                   types.String `tfsdk:"id"`
	Automute             types.Bool   `tfsdk:"automute"`
	ClientEmail          types.String `tfsdk:"client_email"`
	DelegateAccountEmail types.String `tfsdk:"delegate_account_email"`
	IsCspmEnabled        types.Bool   `tfsdk:"is_cspm_enabled"`
	HostFilters          types.Set    `tfsdk:"host_filters"`
}

func NewIntegrationGcpStsResource() resource.Resource {
	return &integrationGcpStsResource{}
}

func (r *integrationGcpStsResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetGCPIntegrationApiV2()
	r.Auth = providerData.Auth
}

func (r *integrationGcpStsResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "integration_gcp_sts"
}

func (r *integrationGcpStsResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Integration GCP Sts resource. This can be used to create and manage Datadog - Google Cloud Platform integration.",
		Attributes: map[string]schema.Attribute{
			"automute": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Silence monitors for expected GCE instance shutdowns.",
			},
			"client_email": schema.StringAttribute{
				Required:    true,
				Description: "Your service account email address.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"delegate_account_email": schema.StringAttribute{
				Computed:    true,
				Description: "Datadog's STS Delegate Email.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_cspm_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "When enabled, Datadog performs configuration checks across your Google Cloud environment by continuously scanning every resource, which may incur additional charges.",
			},
			"host_filters": schema.SetAttribute{
				Optional:    true,
				Description: "Your Host Filters.",
				ElementType: types.StringType,
			},
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *integrationGcpStsResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *integrationGcpStsResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state integrationGcpStsModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	resp, httpResp, err := r.Api.ListGCPSTSAccounts(r.Auth)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving Integration Gcp Sts"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	found := false
	for _, account := range resp.GetData() {
		if account.GetId() == state.ID.ValueString() {
			found = true
			r.updateState(ctx, &state, &account)
			break
		}
	}

	if !found {
		response.State.RemoveResource(ctx)
		return
	}

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationGcpStsResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state integrationGcpStsModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// This resource is special and uses datadog delagate account.
	// The datadog delegate account cannot mutated after creation hence it is safe
	// to call MakeGCPSTSDelegate multiple times. And to ensure it is created, we call it once before creating
	// gcp sts resource.
	delegateResponse, _, err := r.Api.MakeGCPSTSDelegate(r.Auth, *datadogV2.NewMakeGCPSTSDelegateOptionalParameters())
	if err != nil {
		response.Diagnostics.AddError("Error creating GCP Delegate within Datadog",
			"Could not create Delegate Service Account, unexpected error: "+err.Error())
		return
	}
	delegateEmail := delegateResponse.Data.Attributes.GetDelegateAccountEmail()
	state.DelegateAccountEmail = types.StringValue(delegateEmail)

	body, diags := r.buildIntegrationGcpStsRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateGCPSTSAccount(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving Integration Gcp Sts"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateState(ctx, &state, resp.Data)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationGcpStsResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state integrationGcpStsModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	body, diags := r.buildIntegrationGcpStsUpdateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateGCPSTSAccount(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving Integration Gcp Sts"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateState(ctx, &state, resp.Data)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationGcpStsResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state integrationGcpStsModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	httpResp, err := r.Api.DeleteGCPSTSAccount(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting integration_gcp_sts"))
		return
	}
}

func (r *integrationGcpStsResource) updateState(ctx context.Context, state *integrationGcpStsModel, resp *datadogV2.GCPSTSServiceAccount) {
	state.ID = types.StringValue(resp.GetId())

	attributes := resp.GetAttributes()
	if automute, ok := attributes.GetAutomuteOk(); ok {
		state.Automute = types.BoolValue(*automute)
	}
	if clientEmail, ok := attributes.GetClientEmailOk(); ok {
		state.ClientEmail = types.StringValue(*clientEmail)
	}
	if hostFilters, ok := attributes.GetHostFiltersOk(); ok && len(*hostFilters) > 0 {
		state.HostFilters, _ = types.SetValueFrom(ctx, types.StringType, *hostFilters)
	}
	if isCspmEnabled, ok := attributes.GetIsCspmEnabledOk(); ok {
		state.IsCspmEnabled = types.BoolValue(*isCspmEnabled)
	}
}

func (r *integrationGcpStsResource) buildIntegrationGcpStsRequestBody(ctx context.Context, state *integrationGcpStsModel) (*datadogV2.GCPSTSServiceAccountCreateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewGCPSTSServiceAccountAttributesWithDefaults()

	if !state.Automute.IsNull() {
		attributes.SetAutomute(state.Automute.ValueBool())
	}
	if !state.ClientEmail.IsNull() {
		attributes.SetClientEmail(state.ClientEmail.ValueString())
	}
	if !state.IsCspmEnabled.IsNull() {
		attributes.SetIsCspmEnabled(state.IsCspmEnabled.ValueBool())
	}

	hostFilters := make([]string, 0)
	if !state.HostFilters.IsNull() {
		diags.Append(state.HostFilters.ElementsAs(ctx, &hostFilters, false)...)
	}
	attributes.SetHostFilters(hostFilters)

	req := datadogV2.NewGCPSTSServiceAccountCreateRequestWithDefaults()
	req.Data = datadogV2.NewGCPSTSServiceAccountDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}

func (r *integrationGcpStsResource) buildIntegrationGcpStsUpdateRequestBody(ctx context.Context, state *integrationGcpStsModel) (*datadogV2.GCPSTSServiceAccountUpdateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewGCPSTSServiceAccountAttributesWithDefaults()

	if !state.Automute.IsNull() {
		attributes.SetAutomute(state.Automute.ValueBool())
	}
	if !state.ClientEmail.IsNull() {
		attributes.SetClientEmail(state.ClientEmail.ValueString())
	}
	if !state.IsCspmEnabled.IsNull() {
		attributes.SetIsCspmEnabled(state.IsCspmEnabled.ValueBool())
	}

	hostFilters := make([]string, 0)
	if !state.HostFilters.IsNull() {
		diags.Append(state.HostFilters.ElementsAs(ctx, &hostFilters, false)...)
	}
	attributes.SetHostFilters(hostFilters)

	req := datadogV2.NewGCPSTSServiceAccountUpdateRequestWithDefaults()
	req.Data = datadogV2.NewGCPSTSServiceAccountUpdateRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}
