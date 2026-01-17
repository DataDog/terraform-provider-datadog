package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ ephemeral.EphemeralResource              = &apiKeyEphemeralResource{}
	_ ephemeral.EphemeralResourceWithConfigure = &apiKeyEphemeralResource{}
)

func NewAPIKeyEphemeralResource() ephemeral.EphemeralResource {
	return &apiKeyEphemeralResource{}
}

type apiKeyEphemeralResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	ExactMatch   types.Bool   `tfsdk:"exact_match"`
	Key          types.String `tfsdk:"key"`
	RemoteConfig types.Bool   `tfsdk:"remote_config_read_enabled"`
}

type apiKeyEphemeralResource struct {
	Api  *datadogV2.KeyManagementApi
	Auth context.Context
}

func (r *apiKeyEphemeralResource) Configure(_ context.Context, request ephemeral.ConfigureRequest, response *ephemeral.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}
	providerData, ok := request.ProviderData.(*FrameworkProvider)
	if !ok {
		response.Diagnostics.AddError("Unexpected Configure Type", "Expected *FrameworkProvider")
		return
	}
	r.Api = providerData.DatadogApiInstances.GetKeyManagementApiV2()
	r.Auth = providerData.Auth
}

func (r *apiKeyEphemeralResource) Metadata(_ context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = "_api_key"
}

func (r *apiKeyEphemeralResource) Schema(_ context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this ephemeral resource to retrieve a Datadog API key without storing it in Terraform state. This is the recommended approach for securely accessing API keys, for example to pass to a secrets manager.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name for API Key.",
				Optional:    true,
			},
			"id": schema.StringAttribute{
				Description: "The ID of the API key.",
				Optional:    true,
				Computed:    true,
			},
			"exact_match": schema.BoolAttribute{
				Description: "Whether to use exact match when searching by name.",
				Optional:    true,
			},
			"key": schema.StringAttribute{
				Description: "The value of the API Key.",
				Computed:    true,
				Sensitive:   true,
			},
			"remote_config_read_enabled": schema.BoolAttribute{
				Description: "Whether the API key is used for remote config.",
				Computed:    true,
			},
		},
	}
}

func (r *apiKeyEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var config apiKeyEphemeralResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result apiKeyEphemeralResourceModel

	if !config.ID.IsNull() {
		// Lookup by ID
		ddResp, _, err := r.Api.GetAPIKey(r.Auth, config.ID.ValueString())
		if err != nil {
			resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting api key"))
			return
		}
		apiKeyData := ddResp.GetData()
		if r.updateResult(&result, &apiKeyData, resp) {
			return
		}
	} else if !config.Name.IsNull() {
		// Lookup by name
		optionalParams := datadogV2.NewListAPIKeysOptionalParameters()
		optionalParams.WithFilter(config.Name.ValueString())

		apiKeysResponse, _, err := r.Api.ListAPIKeys(r.Auth, *optionalParams)
		if err != nil {
			resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting api keys"))
			return
		}

		apiKeysData := apiKeysResponse.GetData()

		if len(apiKeysData) > 1 && !config.ExactMatch.ValueBool() {
			resp.Diagnostics.AddError("your query returned more than one result, please try a more specific search criteria", "")
			return
		}
		if len(apiKeysData) == 0 {
			resp.Diagnostics.AddError("your query returned no result, please try a less specific search criteria", "")
			return
		}

		if config.ExactMatch.ValueBool() {
			exactMatches := 0
			var apiKeyData datadogV2.FullAPIKey
			for _, apiKeyPartialData := range apiKeysData {
				apiKeyAttributes := apiKeyPartialData.GetAttributes()
				if config.Name.ValueString() == apiKeyAttributes.GetName() {
					exactMatches++
					id := apiKeyPartialData.GetId()
					ddResp, _, err := r.Api.GetAPIKey(r.Auth, id)
					if err != nil {
						resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting api key"))
						return
					}
					apiKeyData = ddResp.GetData()
				}
			}
			if exactMatches > 1 {
				resp.Diagnostics.AddError("your query returned more than one exact match, please try a more specific search criteria", "")
				return
			}
			if exactMatches == 0 {
				resp.Diagnostics.AddError("your query returned no exact matches, please try a less specific search criteria", "")
				return
			}
			if r.updateResult(&result, &apiKeyData, resp) {
				return
			}
		} else {
			id := apiKeysData[0].GetId()
			ddResp, _, err := r.Api.GetAPIKey(r.Auth, id)
			if err != nil {
				resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting api key"))
				return
			}
			apiKeyData := ddResp.GetData()
			if r.updateResult(&result, &apiKeyData, resp) {
				return
			}
		}
	} else {
		resp.Diagnostics.AddError("missing id or name parameter", "")
		return
	}

	// Preserve input values
	result.ExactMatch = config.ExactMatch

	resp.Diagnostics.Append(resp.Result.Set(ctx, &result)...)
}

// updateResult populates the result model from the API response.
// Returns true if there was an error and the caller should return early.
func (r *apiKeyEphemeralResource) updateResult(result *apiKeyEphemeralResourceModel, apiKeyData *datadogV2.FullAPIKey, resp *ephemeral.OpenResponse) bool {
	apiKeyAttributes := apiKeyData.GetAttributes()

	if !apiKeyAttributes.HasKey() {
		resp.Diagnostics.AddError("API key value not available", "The API key value is not available. This may be due to API restrictions on older keys.")
		return true
	}

	result.ID = types.StringValue(apiKeyData.GetId())
	result.Name = types.StringValue(apiKeyAttributes.GetName())
	result.Key = types.StringValue(apiKeyAttributes.GetKey())
	result.RemoteConfig = types.BoolValue(apiKeyAttributes.GetRemoteConfigReadEnabled())
	return false
}
