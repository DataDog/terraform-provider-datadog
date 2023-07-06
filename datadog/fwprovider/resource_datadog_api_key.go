package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &apiKeyResource{}
	_ resource.ResourceWithImportState = &apiKeyResource{}
)

func NewAPIKeyResource() resource.Resource {
	return &apiKeyResource{}
}

type apiKeyResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Key  types.String `tfsdk:"key"`
}

type apiKeyResource struct {
	Api  *datadogV2.KeyManagementApi
	Auth context.Context
}

func (r *apiKeyResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetKeyManagementApiV2()
	r.Auth = providerData.Auth
}

func (r *apiKeyResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "api_key"
}

func (r *apiKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog API Key resource. This can be used to create and manage Datadog API Keys.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name for API Key.",
				Required:    true,
			},
			"key": schema.StringAttribute{
				Description: "The value of the API Key.",
				Computed:    true,
				Sensitive:   true,
			},
			// Resource ID
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *apiKeyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state apiKeyResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateAPIKey(r.Auth, *r.buildDatadogApiKeyCreateV2Struct(&state))
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating api key"))
		return
	}

	apiKeyData := resp.GetData()
	state.ID = types.StringValue(apiKeyData.GetId())
	r.updateState(&state, &apiKeyData)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *apiKeyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state apiKeyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, httpResp, err := r.Api.GetAPIKey(r.Auth, state.ID.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving API Key"))
		return
	}

	apiKeyData := resp.GetData()
	r.updateState(&state, &apiKeyData)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *apiKeyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state apiKeyResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateAPIKey(r.Auth, state.ID.ValueString(), *r.buildDatadogApiKeyUpdateV2Struct(&state))
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating api key"))
		return
	}

	apiKeyData := resp.GetData()
	state.ID = types.StringValue(apiKeyData.GetId())
	r.updateState(&state, &apiKeyData)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *apiKeyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state apiKeyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if _, err := r.Api.DeleteAPIKey(r.Auth, state.ID.ValueString()); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting api key"))
	}
}

func (r *apiKeyResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *apiKeyResource) buildDatadogApiKeyCreateV2Struct(state *apiKeyResourceModel) *datadogV2.APIKeyCreateRequest {
	apiKeyAttributes := datadogV2.NewAPIKeyCreateAttributes(state.Name.ValueString())
	apiKeyData := datadogV2.NewAPIKeyCreateData(*apiKeyAttributes, datadogV2.APIKEYSTYPE_API_KEYS)
	apiKeyRequest := datadogV2.NewAPIKeyCreateRequest(*apiKeyData)

	return apiKeyRequest
}

func (r *apiKeyResource) buildDatadogApiKeyUpdateV2Struct(state *apiKeyResourceModel) *datadogV2.APIKeyUpdateRequest {
	apiKeyAttributes := datadogV2.NewAPIKeyUpdateAttributes(state.Name.ValueString())
	apiKeyData := datadogV2.NewAPIKeyUpdateData(*apiKeyAttributes, state.ID.ValueString(), datadogV2.APIKEYSTYPE_API_KEYS)
	apiKeyRequest := datadogV2.NewAPIKeyUpdateRequest(*apiKeyData)

	return apiKeyRequest
}

func (r *apiKeyResource) updateState(state *apiKeyResourceModel, apiKeyData *datadogV2.FullAPIKey) {
	apiKeyAttributes := apiKeyData.GetAttributes()
	state.Name = types.StringValue(apiKeyAttributes.GetName())
	state.Key = types.StringValue(apiKeyAttributes.GetKey())
}
