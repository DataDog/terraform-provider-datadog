package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSource = &apiKeyDataSource{}
)

func NewAPIKeyDataSource() datasource.DataSource {
	return &apiKeyDataSource{}
}

type apiKeyDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	ExactMatch types.Bool   `tfsdk:"exact_match"`
	Key        types.String `tfsdk:"key"`
}

type apiKeyDataSource struct {
	Api  *datadogV2.KeyManagementApi
	Auth context.Context
}

func (r *apiKeyDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetKeyManagementApiV2()
	r.Auth = providerData.Auth
}

func (d *apiKeyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "api_key"
}

func (d *apiKeyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about an existing api key. Deprecated. This will be removed in a future release with prior notice. Securely store your API keys using a secret management system or use the datadog_api_key resource to manage API keys in your Datadog account.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name for API Key.",
				Optional:    true,
			},
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Optional:    true,
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
		},
		DeprecationMessage: "Deprecated. This will be removed in a future release with prior notice. Securely store your API keys using a secret management system or use the datadog_api_key resource to manage API keys in your Datadog account.",
	}
}

func (d *apiKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state apiKeyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !state.ID.IsNull() {
		ddResp, _, err := d.Api.GetAPIKey(d.Auth, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting api key"))
			return
		}
		apiKeyData := ddResp.GetData()
		if !d.checkAPIDeprecated(&apiKeyData, resp) {
			d.updateState(&state, &apiKeyData)
		}
	} else if !state.Name.IsNull() {
		optionalParams := datadogV2.NewListAPIKeysOptionalParameters()
		optionalParams.WithFilter(state.Name.ValueString())

		apiKeysResponse, _, err := d.Api.ListAPIKeys(d.Auth, *optionalParams)
		if err != nil {
			resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting api keys"))
			return
		}

		apiKeysData := apiKeysResponse.GetData()

		if len(apiKeysData) > 1 && !state.ExactMatch.ValueBool() {
			resp.Diagnostics.AddError("your query returned more than one result, please try a more specific search criteria", "")
			return
		}
		if len(apiKeysData) == 0 {
			resp.Diagnostics.AddError("your query returned no result, please try a less specific search criteria", "")
			return
		}
		if state.ExactMatch.ValueBool() {
			exact_matches := 0
			var apiKeyData datadogV2.FullAPIKey
			for _, apiKeyPartialData := range apiKeysData {
				apiKeyAttributes := apiKeyPartialData.GetAttributes()
				if state.Name.ValueString() == apiKeyAttributes.GetName() {
					exact_matches++
					id := apiKeyPartialData.GetId()
					ddResp, _, err := d.Api.GetAPIKey(d.Auth, id)
					if err != nil {
						resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting api key"))
						return
					}
					apiKeyData = ddResp.GetData()
				}
			}
			if exact_matches > 1 {
				resp.Diagnostics.AddError("your query returned more than one exact match, please try a more specific search criteria", "")
				return
			}
			if exact_matches == 0 {
				resp.Diagnostics.AddError("your query returned no exact matches, please try a less specific search criteria", "")
				return
			}
			if !d.checkAPIDeprecated(&apiKeyData, resp) {
				d.updateState(&state, &apiKeyData)
			}
		} else {
			id := apiKeysData[0].GetId()
			ddResp, _, err := d.Api.GetAPIKey(d.Auth, id)
			if err != nil {
				resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting api key"))
				return
			}
			apiKeyData := ddResp.GetData()
			if !d.checkAPIDeprecated(&apiKeyData, resp) {
				d.updateState(&state, &apiKeyData)
			}
		}
	} else {
		resp.Diagnostics.AddError("missing id or name parameter", "")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *apiKeyDataSource) updateState(state *apiKeyDataSourceModel, apiKeyData *datadogV2.FullAPIKey) {
	apiKeyAttributes := apiKeyData.GetAttributes()

	state.ID = types.StringValue(apiKeyData.GetId())
	state.Name = types.StringValue(apiKeyAttributes.GetName())
	state.Key = types.StringValue(apiKeyAttributes.GetKey())
}

func (r *apiKeyDataSource) checkAPIDeprecated(apiKeyData *datadogV2.FullAPIKey, resp *datasource.ReadResponse) bool {
	apiKeyAttributes := apiKeyData.GetAttributes()
	if !apiKeyAttributes.HasKey() {
		resp.Diagnostics.AddError("Deprecated", "The datadog_api_key data source is deprecated and will be removed in a future release. Securely store your API key using a secret management system or use the datadog_api_key resource to manage API keys in your Datadog account.")
		return true
	}
	return false
}
