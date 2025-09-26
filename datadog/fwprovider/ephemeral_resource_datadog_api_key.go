package fwprovider

import (
	"context"
	"log"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Interface assertions for EphemeralAPIKeyResource
var (
	_ ephemeral.EphemeralResource              = &EphemeralAPIKeyResource{}
	_ ephemeral.EphemeralResourceWithConfigure = &EphemeralAPIKeyResource{}
)

// EphemeralAPIKeyResource implements ephemeral API key resource
type EphemeralAPIKeyResource struct {
	Api  *datadogV2.KeyManagementApi
	Auth context.Context
}

// EphemeralAPIKeyModel represents the data model for the ephemeral API key resource
type EphemeralAPIKeyModel struct {
	ID                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	Key                     types.String `tfsdk:"key"`
	RemoteConfigReadEnabled types.Bool   `tfsdk:"remote_config_read_enabled"`
}

// NewEphemeralAPIKeyResource creates a new ephemeral API key resource
func NewEphemeralAPIKeyResource() ephemeral.EphemeralResource {
	return &EphemeralAPIKeyResource{}
}

// Metadata implements the core ephemeral.EphemeralResource interface
func (r *EphemeralAPIKeyResource) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = "api_key" // Will become "datadog_api_key" via wrapper
}

// Schema implements the core ephemeral.EphemeralResource interface
func (r *EphemeralAPIKeyResource) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves an existing Datadog API key as an ephemeral resource. The API key value is retrieved securely and made available for use in other resources without being stored in state.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the API key to retrieve.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the API key.",
			},
			"key": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The actual API key value (sensitive).",
			},
			"remote_config_read_enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether remote configuration reads are enabled for this key.",
			},
		},
	}
}

// Open implements the core ephemeral.EphemeralResource interface
// This is where the ephemeral resource acquires the API key data
func (r *EphemeralAPIKeyResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	// 1. Extract API key ID from config
	var config EphemeralAPIKeyModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 2. Fetch API key from Datadog API
	apiKey, httpResp, err := r.Api.GetAPIKey(r.Auth, config.ID.ValueString())
	if err != nil {
		log.Printf("[ERROR] Ephemeral open operation failed for api_key: %v", err)
		resp.Diagnostics.AddError(
			"API Key Retrieval Failed",
			"Unable to fetch API key data from Datadog API",
		)
		return
	}

	// Check HTTP response status
	if httpResp != nil && httpResp.StatusCode >= 400 {
		log.Printf("[WARN] Ephemeral open operation failed for api_key")
		resp.Diagnostics.AddError(
			"API Key Retrieval Failed",
			"Received error response from Datadog API",
		)
		return
	}

	// 3. Extract API key data from response
	apiKeyData := apiKey.GetData()
	apiKeyAttributes := apiKeyData.GetAttributes()

	// 4. Set result data (including the sensitive key value)
	result := EphemeralAPIKeyModel{
		ID:                      config.ID,
		Name:                    types.StringValue(apiKeyAttributes.GetName()),
		Key:                     types.StringValue(apiKeyAttributes.GetKey()), // SENSITIVE
		RemoteConfigReadEnabled: types.BoolValue(apiKeyAttributes.GetRemoteConfigReadEnabled()),
	}

	resp.Diagnostics.Append(resp.Result.Set(ctx, &result)...)
	if resp.Diagnostics.HasError() {
		return
	}

	log.Printf("[DEBUG] Ephemeral open operation succeeded for api_key")
}

// Configure implements the optional ephemeral.EphemeralResourceWithConfigure interface
func (r *EphemeralAPIKeyResource) Configure(ctx context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*FrameworkProvider)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Configure Type",
			"Expected *FrameworkProvider",
		)
		return
	}

	r.Api = providerData.DatadogApiInstances.GetKeyManagementApiV2()
	r.Auth = providerData.Auth
}
