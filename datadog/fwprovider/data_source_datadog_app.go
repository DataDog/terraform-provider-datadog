package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var _ datasource.DataSource = &appDataSource{}

type appDataSource struct {
	Api  *datadogV2.AppsApi
	Auth context.Context
}

func NewDatadogAppDataSource() datasource.DataSource {
	return &appDataSource{}
}

// TODO: figure out rest of model
type datadogAppDatasourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (d *appDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetAppsApiV2()
	d.Auth = providerData.Auth
}

func (d *appDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "app"
}

// TODO: figure out rest of schema
func (d *appDataSource) Schema(_ context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about an existing App, for use in other resources.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
		},
	}
}

// TODO
func (d *appDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state datadogAppDatasourceModel
	diags := request.Config.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	state.ID = types.StringValue("hi")
	diags = response.State.Set(ctx, &state)
	response.Diagnostics.Append(diags...)
}
