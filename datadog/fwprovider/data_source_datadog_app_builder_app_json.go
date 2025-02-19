package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &appBuilderAppJSONDataSource{}

type appBuilderAppJSONDataSource struct {
	Api  *datadogV2.AppBuilderApi
	Auth context.Context
}

func NewDatadogAppBuilderAppJSONDataSource() datasource.DataSource {
	return &appBuilderAppJSONDataSource{}
}

func (d *appBuilderAppJSONDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetAppBuilderApiV2()
	d.Auth = providerData.Auth
}

func (d *appBuilderAppJSONDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "app_builder_app_json"
}

func (d *appBuilderAppJSONDataSource) Schema(_ context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "This data source retrieves the JSON definition of an existing Datadog App from App Builder for use in other resources.",
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
				CustomType: jsontypes.NormalizedType{},
			},
			"action_query_ids_to_connection_ids": schema.MapAttribute{
				Computed:    true,
				Description: "A map of the App's Action Query IDs to Action Connection IDs.",
				ElementType: types.StringType,
			},
		},
	}
}

func (d *appBuilderAppJSONDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state appBuilderAppJSONResourceModel
	diags := request.Config.Get(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("error parsing id as uuid", err.Error())
		return
	}

	appBuilderAppJSONModel, err := readAppBuilderAppJSON(d.Auth, d.Api, id)
	if err != nil {
		response.Diagnostics.AddError("error reading app", err.Error())
		return
	}

	diags = response.State.Set(ctx, appBuilderAppJSONModel)
	response.Diagnostics.Append(diags...)
}
