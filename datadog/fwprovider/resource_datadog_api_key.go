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
	_ resource.ResourceWithConfigure   = &APIKeyResource{}
	_ resource.ResourceWithImportState = &APIKeyResource{}
)

func NewAPIKeyResource() resource.Resource {
	return &APIKeyResource{}
}

type APIKeyResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Key  types.String `tfsdk:"key"`
}

type APIKeyResource struct {
	Api  *datadogV2.KeyManagementApi
	Auth context.Context
}

func (r *APIKeyResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	providerData, ok := request.ProviderData.(*FrameworkProvider)
	if !ok {
		response.Diagnostics.AddError("Unexpected Resource Configure Type", "")
		return
	}

	r.Api = providerData.DatadogApiInstances.GetKeyManagementApiV2()
	r.Auth = providerData.Auth
}

func (r *APIKeyResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "api_key"
}

func (r *APIKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
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

func (r *APIKeyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state APIKeyResourceModel
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

func (r *APIKeyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state APIKeyResourceModel
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

func (r *APIKeyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state APIKeyResourceModel
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

func (r *APIKeyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state APIKeyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if _, err := r.Api.DeleteAPIKey(r.Auth, state.ID.ValueString()); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting api key"))
	}
}

func (r *APIKeyResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *APIKeyResource) buildDatadogApiKeyCreateV2Struct(state *APIKeyResourceModel) *datadogV2.APIKeyCreateRequest {
	apiKeyAttributes := datadogV2.NewAPIKeyCreateAttributes(state.Name.ValueString())
	apiKeyData := datadogV2.NewAPIKeyCreateData(*apiKeyAttributes, datadogV2.APIKEYSTYPE_API_KEYS)
	apiKeyRequest := datadogV2.NewAPIKeyCreateRequest(*apiKeyData)

	return apiKeyRequest
}

func (r *APIKeyResource) buildDatadogApiKeyUpdateV2Struct(state *APIKeyResourceModel) *datadogV2.APIKeyUpdateRequest {
	apiKeyAttributes := datadogV2.NewAPIKeyUpdateAttributes(state.Name.ValueString())
	apiKeyData := datadogV2.NewAPIKeyUpdateData(*apiKeyAttributes, state.ID.ValueString(), datadogV2.APIKEYSTYPE_API_KEYS)
	apiKeyRequest := datadogV2.NewAPIKeyUpdateRequest(*apiKeyData)

	return apiKeyRequest
}

func (r *APIKeyResource) updateState(state *APIKeyResourceModel, apiKeyData *datadogV2.FullAPIKey) {
	apiKeyAttributes := apiKeyData.GetAttributes()
	state.Name = types.StringValue(apiKeyAttributes.GetName())
	state.Key = types.StringValue(apiKeyAttributes.GetKey())
}
