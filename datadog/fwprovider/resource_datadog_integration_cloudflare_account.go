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
	_ resource.ResourceWithConfigure   = &integrationCloudflareAccountResource{}
	_ resource.ResourceWithImportState = &integrationCloudflareAccountResource{}
)

type integrationCloudflareAccountResource struct {
	Api  *datadogV2.CloudflareIntegrationApi
	Auth context.Context
}

type integrationCloudflareAccountModel struct {
	ID     types.String `tfsdk:"id"`
	ApiKey types.String `tfsdk:"api_key"`
	Email  types.String `tfsdk:"email"`
	Name   types.String `tfsdk:"name"`
}

func NewIntegrationCloudflareAccountResource() resource.Resource {
	return &integrationCloudflareAccountResource{}
}

func (r *integrationCloudflareAccountResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetCloudflareIntegrationApiV2()
	r.Auth = providerData.Auth
}

func (r *integrationCloudflareAccountResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "integration_cloudflare_account"
}

func (r *integrationCloudflareAccountResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog IntegrationCloudflareAccount resource. This can be used to create and manage Datadog integration_cloudflare_account.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Required:    true,
				Description: "The API key (or token) for the Cloudflare account.",
				Sensitive:   true,
			},
			"email": schema.StringAttribute{
				Optional:    true,
				Description: "The email associated with the Cloudflare account. If an API key is provided (and not a token), this field is also required.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the Cloudflare account.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *integrationCloudflareAccountResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *integrationCloudflareAccountResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state integrationCloudflareAccountModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	resp, httpResp, err := r.Api.GetCloudflareAccount(r.Auth, id)
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

func (r *integrationCloudflareAccountResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state integrationCloudflareAccountModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildIntegrationCloudflareAccountRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateCloudflareAccount(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving IntegrationCloudflareAccount"))
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

func (r *integrationCloudflareAccountResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state integrationCloudflareAccountModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	body, diags := r.buildIntegrationCloudflareAccountUpdateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateCloudflareAccount(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving IntegrationCloudflareAccount"))
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

func (r *integrationCloudflareAccountResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state integrationCloudflareAccountModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	httpResp, err := r.Api.DeleteCloudflareAccount(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting integration_cloudflare_account"))
		return
	}
}

func (r *integrationCloudflareAccountResource) updateState(ctx context.Context, state *integrationCloudflareAccountModel, resp *datadogV2.CloudflareAccountResponse) {
	state.ID = types.StringValue(resp.Data.GetId())

	data := resp.GetData()
	attributes := data.GetAttributes()

	if email, ok := attributes.GetEmailOk(); ok {
		state.Email = types.StringValue(*email)
	}

	if name, ok := attributes.GetNameOk(); ok {
		state.Name = types.StringValue(*name)
	}
}

func (r *integrationCloudflareAccountResource) buildIntegrationCloudflareAccountRequestBody(ctx context.Context, state *integrationCloudflareAccountModel) (*datadogV2.CloudflareAccountCreateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewCloudflareAccountCreateRequestAttributesWithDefaults()

	attributes.SetApiKey(state.ApiKey.ValueString())
	if !state.Email.IsNull() {
		attributes.SetEmail(state.Email.ValueString())
	}
	attributes.SetName(state.Name.ValueString())

	req := datadogV2.NewCloudflareAccountCreateRequestWithDefaults()
	req.Data = *datadogV2.NewCloudflareAccountCreateRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}

func (r *integrationCloudflareAccountResource) buildIntegrationCloudflareAccountUpdateRequestBody(ctx context.Context, state *integrationCloudflareAccountModel) (*datadogV2.CloudflareAccountUpdateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewCloudflareAccountUpdateRequestAttributesWithDefaults()

	attributes.SetApiKey(state.ApiKey.ValueString())
	if !state.Email.IsNull() {
		attributes.SetEmail(state.Email.ValueString())
	}

	req := datadogV2.NewCloudflareAccountUpdateRequestWithDefaults()
	req.Data = *datadogV2.NewCloudflareAccountUpdateRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}
