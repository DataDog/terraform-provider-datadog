package fwprovider

import (
	"context"

	frameworkDiag "github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Key          types.String `tfsdk:"key"`
	StoreKey     types.Bool   `tfsdk:"store_key"`
	RemoteConfig types.Bool   `tfsdk:"remote_config_read_enabled"`
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
		Description: "Provides a Datadog API Key resource. This can be used to create and manage Datadog API Keys. Import functionality for this resource is deprecated and will be removed in a future release with prior notice. Securely store your API keys using a secret management system or use this resource to create and manage new API keys.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name for API Key.",
				Required:    true,
			},
			"key": schema.StringAttribute{
				Description:   "The value of the API Key.",
				Computed:      true,
				Sensitive:     true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"store_key": schema.BoolAttribute{
				Description: "Whether to store the API key value in Terraform state. Set to `false` to avoid storing the key in state for security purposes. Defaults to `true` for backwards compatibility. When set to `false`, the `key` attribute will be empty.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"remote_config_read_enabled": schema.BoolAttribute{
				Description: "Whether the API key is used for remote config. Set to true only if remote config is enabled in `/organization-settings/remote-config`.",
				Optional:    true,
				Computed:    true,
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
	updateStateDiag := r.updateState(&state, &apiKeyData)
	if updateStateDiag != nil {
		response.Diagnostics.Append(updateStateDiag)
	}
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
	updateStateDiag := r.updateState(&state, &apiKeyData)
	if updateStateDiag != nil {
		response.Diagnostics.Append(updateStateDiag)
		return
	}

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
	updateStateDiag := r.updateState(&state, &apiKeyData)
	if updateStateDiag != nil {
		response.Diagnostics.Append(updateStateDiag)
		return
	}
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
	response.Diagnostics.AddWarning(
		"Deprecated",
		"The import functionality for datadog_api_key resources is deprecated and will be removed in a future release with prior notice. Securely store your API keys using a secret management system or use the datadog_api_key resource to create and manage new API keys.",
	)
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *apiKeyResource) buildDatadogApiKeyCreateV2Struct(state *apiKeyResourceModel) *datadogV2.APIKeyCreateRequest {
	apiKeyAttributes := datadogV2.NewAPIKeyCreateAttributes(state.Name.ValueString())
	if !(state.RemoteConfig.IsUnknown() || state.RemoteConfig.IsNull()) {
		apiKeyAttributes.SetRemoteConfigReadEnabled(state.RemoteConfig.ValueBool())
	}
	apiKeyData := datadogV2.NewAPIKeyCreateData(*apiKeyAttributes, datadogV2.APIKEYSTYPE_API_KEYS)
	apiKeyRequest := datadogV2.NewAPIKeyCreateRequest(*apiKeyData)
	return apiKeyRequest
}

func (r *apiKeyResource) buildDatadogApiKeyUpdateV2Struct(state *apiKeyResourceModel) *datadogV2.APIKeyUpdateRequest {
	apiKeyAttributes := datadogV2.NewAPIKeyUpdateAttributes(state.Name.ValueString())
	if !(state.RemoteConfig.IsUnknown() || state.RemoteConfig.IsNull()) {
		apiKeyAttributes.SetRemoteConfigReadEnabled(state.RemoteConfig.ValueBool())
	}
	apiKeyData := datadogV2.NewAPIKeyUpdateData(*apiKeyAttributes, state.ID.ValueString(), datadogV2.APIKEYSTYPE_API_KEYS)
	apiKeyRequest := datadogV2.NewAPIKeyUpdateRequest(*apiKeyData)
	return apiKeyRequest
}

func (r *apiKeyResource) updateState(state *apiKeyResourceModel, apiKeyData *datadogV2.FullAPIKey) frameworkDiag.Diagnostic {
	var d frameworkDiag.Diagnostic
	apiKeyAttributes := apiKeyData.GetAttributes()
	state.Name = types.StringValue(apiKeyAttributes.GetName())
	if state.RemoteConfig.ValueBool() && !apiKeyAttributes.GetRemoteConfigReadEnabled() {
		d = frameworkDiag.NewErrorDiagnostic("remote_config_read_enabled is true but Remote config is not enabled at org level", "Please either remove remote_config_read_enabled from the resource configuration or enable Remote config at org level")
	}
	state.RemoteConfig = types.BoolValue(apiKeyAttributes.GetRemoteConfigReadEnabled())

	// Only store the key if store_key is true (default behavior for backwards compatibility)
	if state.StoreKey.ValueBool() && apiKeyAttributes.HasKey() {
		state.Key = types.StringValue(apiKeyAttributes.GetKey())
	} else if !state.StoreKey.ValueBool() {
		// Explicitly set to null when store_key is false
		state.Key = types.StringNull()
	}
	return d
}
