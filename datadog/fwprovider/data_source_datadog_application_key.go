package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var _ datasource.DataSource = &applicationKeyDataSource{}

type applicationKeyDataSourceModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Key  types.String `tfsdk:"key"`
}

type applicationKeyDataSource struct {
	Api  *datadogV2.KeyManagementApi
	Auth context.Context
}

func NewApplicationKeyDataSource() datasource.DataSource {
	return &applicationKeyDataSource{}
}

// Metadata implements datasource.DataSource.
func (d *applicationKeyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "application_key"
}

// Schema implements datasource.DataSource.
func (d *applicationKeyDataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about an existing application key.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Id for Application Key.",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name for Application Key.",
				Optional:    true,
			},
			"key": schema.StringAttribute{
				Description: "The value of the Application Key.",
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

func (r *applicationKeyDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetKeyManagementApiV2()
	r.Auth = providerData.Auth
}

// Read implements datasource.DataSource.
func (d *applicationKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state applicationKeyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !state.Id.IsNull() {
		ddResp, _, err := d.Api.GetCurrentUserApplicationKey(d.Auth, state.Id.ValueString())
		if err != nil {
			resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting application key"))
			return
		}
		apiKeyData := ddResp.GetData()
		d.updateState(&state, &apiKeyData)
	} else if !state.Name.IsNull() {
		optionalParams := datadogV2.NewListCurrentUserApplicationKeysOptionalParameters()
		optionalParams.WithFilter(state.Name.ValueString())
		applicationKeysResponse, _, err := d.Api.ListCurrentUserApplicationKeys(d.Auth, *optionalParams)
		if err != nil {
			resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting application keys"))
			return
		}
		applicationKeysData := applicationKeysResponse.GetData()
		if len(applicationKeysData) > 1 {
			resp.Diagnostics.AddError("your query returned more than one result, please try a more specific search criteria", "")
			return
		}
		if len(applicationKeysData) == 0 {
			resp.Diagnostics.AddError("your query returned no result, please try a less specific search criteria", "")
			return
		}
		applicationKeyPartialData := applicationKeysData[0]
		id := applicationKeyPartialData.GetId()
		applicationKeyResponse, _, err := d.Api.GetCurrentUserApplicationKey(d.Auth, id)
		if err != nil {
			resp.Diagnostics.AddError("error getting application key", "")
			return
		}
		applicationKeyFullData := applicationKeyResponse.GetData()
		d.updateState(&state, &applicationKeyFullData)
	} else {
		resp.Diagnostics.AddError("missing id or name parameter", "")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *applicationKeyDataSource) updateState(state *applicationKeyDataSourceModel, applicationKeyData *datadogV2.FullApplicationKey) {
	applicationKeyAttributes := applicationKeyData.GetAttributes()

	state.Id = types.StringValue(applicationKeyData.GetId())
	state.Name = types.StringValue(applicationKeyAttributes.GetName())
	state.Key = types.StringValue(applicationKeyAttributes.GetKey())
}
