package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &connectionDatasource{}

type connectionDatasource struct {
	Api  *datadogV2.TeamsApi
	Auth context.Context
}

func NewDatadogConnectionDataSource() datasource.DataSource {
	return &connectionDatasource{}
}

type datadogConnectionDatasourceModel struct {
	ID types.String `tfsdk:"id"`
}

func (d *connectionDatasource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetTeamsApiV2()
	d.Auth = providerData.Auth
}

func (d *connectionDatasource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "connection"
}

func (d *connectionDatasource) Schema(_ context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about an existing connection.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID for Connection.",
				Required:    true,
			},
		},
	}
}

func (d *connectionDatasource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state datadogConnectionDatasourceModel
	diags := request.Config.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	diags = response.State.Set(ctx, &state)
	response.Diagnostics.Append(diags...)
}
