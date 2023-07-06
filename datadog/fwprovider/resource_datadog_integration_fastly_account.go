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
	_ resource.ResourceWithConfigure   = &integrationFastlyAccountResource{}
	_ resource.ResourceWithImportState = &integrationFastlyAccountResource{}
)

type integrationFastlyAccountResource struct {
	Api  *datadogV2.FastlyIntegrationApi
	Auth context.Context
}

type integrationFastlyAccountModel struct {
	ID     types.String `tfsdk:"id"`
	ApiKey types.String `tfsdk:"api_key"`
	Name   types.String `tfsdk:"name"`
}

func NewIntegrationFastlyAccountResource() resource.Resource {
	return &integrationFastlyAccountResource{}
}

func (r *integrationFastlyAccountResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetFastlyIntegrationApiV2()
	r.Auth = providerData.Auth
}

func (r *integrationFastlyAccountResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "integration_fastly_account"
}

func (r *integrationFastlyAccountResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog IntegrationFastlyAccount resource. This can be used to create and manage Datadog integration_fastly_account.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Required:    true,
				Description: "The API key for the Fastly account.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the Fastly account.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *integrationFastlyAccountResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *integrationFastlyAccountResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state integrationFastlyAccountModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	resp, httpResp, err := r.Api.GetFastlyAccount(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving API Key"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationFastlyAccountResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state integrationFastlyAccountModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildIntegrationFastlyAccountRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateFastlyAccount(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving IntegrationFastlyAccount"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationFastlyAccountResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state integrationFastlyAccountModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	body, diags := r.buildIntegrationFastlyAccountUpdateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateFastlyAccount(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving IntegrationFastlyAccount"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *integrationFastlyAccountResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state integrationFastlyAccountModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	httpResp, err := r.Api.DeleteFastlyAccount(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting integration_fastly_account"))
		return
	}
}

func (r *integrationFastlyAccountResource) updateState(ctx context.Context, state *integrationFastlyAccountModel, resp *datadogV2.FastlyAccountResponse) {
	state.ID = types.StringValue(resp.Data.GetId())

	data := resp.GetData()
	attributes := data.GetAttributes()

	if name, ok := attributes.GetNameOk(); ok {
		state.Name = types.StringValue(*name)
	}
}

func (r *integrationFastlyAccountResource) buildIntegrationFastlyAccountRequestBody(ctx context.Context, state *integrationFastlyAccountModel) (*datadogV2.FastlyAccountCreateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewFastlyAccountCreateRequestAttributesWithDefaults()

	attributes.SetApiKey(state.ApiKey.ValueString())
	attributes.SetName(state.Name.ValueString())
	// TODO: Api marks this as required for now. Remove once fixed.
	attributes.SetServices([]datadogV2.FastlyService{})

	req := datadogV2.NewFastlyAccountCreateRequestWithDefaults()
	req.Data = *datadogV2.NewFastlyAccountCreateRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}

func (r *integrationFastlyAccountResource) buildIntegrationFastlyAccountUpdateRequestBody(ctx context.Context, state *integrationFastlyAccountModel) (*datadogV2.FastlyAccountUpdateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewFastlyAccountUpdateRequestAttributesWithDefaults()

	if !state.ApiKey.IsNull() {
		attributes.SetApiKey(state.ApiKey.ValueString())
	}

	req := datadogV2.NewFastlyAccountUpdateRequestWithDefaults()
	req.Data = *datadogV2.NewFastlyAccountUpdateRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}
