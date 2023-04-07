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
	_ datasource.DataSource = &APIKeyDataSource{}
)

func NewAPIKeyDataSource() datasource.DataSource {
	return &APIKeyDataSource{}
}

type APIKeyDataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Key  types.String `tfsdk:"key"`
}

type APIKeyDataSource struct {
	Api  *datadogV2.KeyManagementApi
	Auth context.Context
}

func (r *APIKeyDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func (d *APIKeyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "api_key"
}

func (d *APIKeyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about an existing api key.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Name for API Key.",
				Optional:    true,
			},
			"id": schema.StringAttribute{
				Description: "The ID of this resource.",
				Optional:    true,
				Computed:    true,
			},
			"key": schema.StringAttribute{
				Description: "The value of the API Key.",
				Computed:    true,
				Sensitive:   true,
			},
		},
	}

}

func (d *APIKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state APIKeyDataSourceModel
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
		d.updateState(&state, &apiKeyData)
	} else {
		optionalParams := datadogV2.NewListAPIKeysOptionalParameters()
		optionalParams.WithFilter(state.Name.ValueString())

		apiKeysResponse, _, err := d.Api.ListAPIKeys(d.Auth, *optionalParams)
		if err != nil {
			resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting api keys"))
			return
		}

		apiKeysData := apiKeysResponse.GetData()

		if len(apiKeysData) > 1 {
			resp.Diagnostics.AddError("your query returned more than one result, please try a more specific search criteria", "")
			return
		}
		if len(apiKeysData) == 0 {
			resp.Diagnostics.AddError("your query returned no result, please try a less specific search criteria", "")
			return
		}

		apiKeyPartialData := apiKeysData[0]

		id := apiKeyPartialData.GetId()
		ddResp, _, err := d.Api.GetAPIKey(d.Auth, id)
		if err != nil {
			resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting api key"))
			return
		}
		apiKeyData := ddResp.GetData()
		d.updateState(&state, &apiKeyData)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *APIKeyDataSource) updateState(state *APIKeyDataSourceModel, apiKeyData *datadogV2.FullAPIKey) {
	apiKeyAttributes := apiKeyData.GetAttributes()
	state.Name = types.StringValue(apiKeyAttributes.GetName())
	state.Key = types.StringValue(apiKeyAttributes.GetKey())
}
