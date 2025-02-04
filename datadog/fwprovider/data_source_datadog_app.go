package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ datasource.DataSource = &appDataSource{}

type appDataSource struct {
	Api  *datadogV2.AppBuilderApi
	Auth context.Context
}

func NewDatadogAppDataSource() datasource.DataSource {
	return &appDataSource{}
}

func (d *appDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetAppBuilderApiV2()
	d.Auth = providerData.Auth
}

func (d *appDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "app"
}

func (d *appDataSource) Schema(_ context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about an existing Datadog App from the App Builder product, for use in other resources.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID for the App.",
				Required:    true,
			},
			"app_json": schema.StringAttribute{
				Computed:    true,
				Description: "The JSON representation of the App.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		},
	}
}

func (d *appDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state appResourceModel
	diags := request.Config.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Error parsing id as uuid", err.Error())
		return
	}

	appModel, err := readApp(d.Auth, d.Api, id)
	if err != nil {
		response.Diagnostics.AddError("Error reading app", err.Error())
		return
	}

	diags = response.State.Set(ctx, appModel)
	response.Diagnostics.Append(diags...)
}
