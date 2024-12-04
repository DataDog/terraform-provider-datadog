package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var _ datasource.DataSource = &datadogConnectionDatasource{}

type datadogConnectionDatasource struct {
	Api  *datadogV2.TeamsApi
	Auth context.Context
}

func NewDatadogConnectionDataSource() datasource.DataSource {
	return &datadogConnectionDatasource{}
}

type datadogConnectionDatasourceModel struct {
	ID   types.String `tfsdk:"id"`
	name types.String `tfsdk:"name"`
}

func (d *datadogConnectionDatasource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetTeamsApiV2()
	d.Auth = providerData.Auth
}

func (d *datadogConnectionDatasource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "connection"
}

func (d *datadogConnectionDatasource) Schema(_ context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about an existing connection, for use in other resources.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (d *datadogConnectionDatasource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state datadogConnectionDatasourceModel
	diags := request.Config.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	state.ID = types.StringValue("hi")
	diags = response.State.Set(ctx, &state)
	response.Diagnostics.Append(diags...)
}
