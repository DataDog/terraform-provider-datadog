package fwprovider

import (
	"context"

	frameworkDiag "github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/secretbridge"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &apiKeyResource{}
	_ resource.ResourceWithImportState = &apiKeyResource{}
	_ resource.ResourceWithModifyPlan  = &apiKeyResource{}
)

type encryptionTransition int

const (
	noEncryptionTransition encryptionTransition = iota
	addingEncryption
	removingEncryption
)

func detectEncryptionTransition(encryptionKeyWO, stateKey, stateEncryptedKey types.String) encryptionTransition {
	encryptionRequested := !encryptionKeyWO.IsNull() && encryptionKeyWO.ValueString() != ""
	hadPlaintextKey := !stateKey.IsNull() && stateKey.ValueString() != ""
	hadEncryptedKey := !stateEncryptedKey.IsNull() && stateEncryptedKey.ValueString() != ""

	if encryptionRequested && hadPlaintextKey && !hadEncryptedKey {
		return addingEncryption
	}
	if !encryptionRequested && hadEncryptedKey {
		return removingEncryption
	}
	return noEncryptionTransition
}

func NewAPIKeyResource() resource.Resource {
	return &apiKeyResource{}
}

type apiKeyResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Key             types.String `tfsdk:"key"`
	EncryptedKey    types.String `tfsdk:"encrypted_key"`
	EncryptionKeyWO types.String `tfsdk:"encryption_key_wo"`
	RemoteConfig    types.Bool   `tfsdk:"remote_config_read_enabled"`
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
				Description:   "The value of the API Key. Mutually exclusive with `encrypted_key` when `encryption_key_wo` is set.",
				Computed:      true,
				Sensitive:     true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"encrypted_key": schema.StringAttribute{
				Description:   "The encrypted value of the API Key. Only populated when `encryption_key_wo` is provided. Use the `datadog_secret_decrypt` ephemeral resource to decrypt this value. **Warning:** This attribute is intended only for a transfer to a secret manager. The encryption format may change in future provider versions without notice resulting in decryption failure if attempted.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"encryption_key_wo": secretbridge.EncryptionKeyAttribute(),
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

	// Write-only attributes must be read from Config, not Plan (always null in plan)
	response.Diagnostics.Append(request.Config.GetAttribute(ctx, frameworkPath.Root("encryption_key_wo"), &state.EncryptionKeyWO)...)
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

	// Handle encryption if encryption_key_wo is provided
	apiKeyAttrs := apiKeyData.GetAttributes()
	if apiKeyAttrs.HasKey() {
		plaintextKey := apiKeyAttrs.GetKey()
		if !state.EncryptionKeyWO.IsNull() {
			// Encrypt the key and store in encrypted_key, set key to null
			encrypted, diags := secretbridge.Encrypt(ctx, plaintextKey, []byte(state.EncryptionKeyWO.ValueString()))
			response.Diagnostics.Append(diags...)
			if response.Diagnostics.HasError() {
				return
			}
			state.EncryptedKey = types.StringValue(encrypted)
			state.Key = types.StringNull()

			response.Diagnostics.AddWarning(
				"Encrypted Key for Transit Only",
				"The encrypted_key attribute is intended only for a transfer to a secret manager. "+
					"Do not rely on decrypting it across different provider versions. The encryption format may "+
					"change in future releases without notice. After transferring the secret, "+
					"consider removing the encryption_key_wo to clear the encrypted value from state.",
			)
		} else {
			// No encryption - store key in plaintext, set encrypted_key to null
			state.Key = types.StringValue(plaintextKey)
			state.EncryptedKey = types.StringNull()
		}
	}

	updateStateDiag := r.updateStateMetadata(&state, &apiKeyData)
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
	// Only update metadata - key/encrypted_key are preserved from state
	updateStateDiag := r.updateStateMetadata(&state, &apiKeyData)
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

	var encryptionKeyWO types.String
	response.Diagnostics.Append(request.Config.GetAttribute(ctx, frameworkPath.Root("encryption_key_wo"), &encryptionKeyWO)...)
	if response.Diagnostics.HasError() {
		return
	}

	var priorState apiKeyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &priorState)...)
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

	switch detectEncryptionTransition(encryptionKeyWO, priorState.Key, priorState.EncryptedKey) {
	case addingEncryption:
		encrypted, diags := secretbridge.Encrypt(ctx, priorState.Key.ValueString(), []byte(encryptionKeyWO.ValueString()))
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
		state.EncryptedKey = types.StringValue(encrypted)
		state.Key = types.StringNull()
	case removingEncryption:
		state.Key = types.StringNull()
		state.EncryptedKey = types.StringNull()
		response.Diagnostics.AddWarning(
			"Encryption Removed",
			"The encryption key has been removed. The API key value is no longer accessible from Terraform state. "+
				"Retrieve the key from Datadog console or recreate the resource if needed.",
		)
	case noEncryptionTransition:
		// No transition: preserve key fields from prior state
		state.Key = priorState.Key
		state.EncryptedKey = priorState.EncryptedKey
	}

	updateStateDiag := r.updateStateMetadata(&state, &apiKeyData)
	if updateStateDiag != nil {
		response.Diagnostics.Append(updateStateDiag)
		return
	}
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

func (r *apiKeyResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var encryptionKeyWO, stateKey, stateEncryptedKey types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, frameworkPath.Root("encryption_key_wo"), &encryptionKeyWO)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, frameworkPath.Root("key"), &stateKey)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, frameworkPath.Root("encrypted_key"), &stateEncryptedKey)...)
	if resp.Diagnostics.HasError() {
		return
	}

	switch detectEncryptionTransition(encryptionKeyWO, stateKey, stateEncryptedKey) {
	case addingEncryption:
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, frameworkPath.Root("key"), types.StringNull())...)
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, frameworkPath.Root("encrypted_key"), types.StringUnknown())...)
	case removingEncryption:
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, frameworkPath.Root("key"), types.StringNull())...)
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, frameworkPath.Root("encrypted_key"), types.StringNull())...)
	case noEncryptionTransition:
		// No change needed
	}
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

// updateStateMetadata updates only non-key fields from API response.
// Key fields (key, encrypted_key) are handled separately in Create and preserved in Read/Update.
func (r *apiKeyResource) updateStateMetadata(state *apiKeyResourceModel, apiKeyData *datadogV2.FullAPIKey) frameworkDiag.Diagnostic {
	var d frameworkDiag.Diagnostic
	apiKeyAttributes := apiKeyData.GetAttributes()
	state.Name = types.StringValue(apiKeyAttributes.GetName())
	if state.RemoteConfig.ValueBool() && !apiKeyAttributes.GetRemoteConfigReadEnabled() {
		d = frameworkDiag.NewErrorDiagnostic("remote_config_read_enabled is true but Remote config is not enabled at org level", "Please either remove remote_config_read_enabled from the resource configuration or enable Remote config at org level")
	}
	state.RemoteConfig = types.BoolValue(apiKeyAttributes.GetRemoteConfigReadEnabled())
	return d
}
