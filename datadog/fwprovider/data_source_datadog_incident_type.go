package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSourceWithConfigure = &incidentTypeDataSource{}
)

type incidentTypeDataSource struct {
	Api  *datadogV2.IncidentsApi
	Auth context.Context
}

type incidentTypeDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	IsDefault   types.Bool   `tfsdk:"is_default"`
}

func NewIncidentTypeDataSource() datasource.DataSource {
	return &incidentTypeDataSource{}
}

func (d *incidentTypeDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetIncidentsApiV2()
	d.Auth = providerData.Auth
}

func (d *incidentTypeDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "incident_type"
}

func (d *incidentTypeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about an existing incident type.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the incident type.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the incident type.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the incident type.",
				Computed:    true,
			},
			"is_default": schema.BoolAttribute{
				Description: "Whether this incident type is the default type.",
				Computed:    true,
			},
		},
	}
}

func (d *incidentTypeDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state incidentTypeDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, httpResp, err := d.Api.GetIncidentType(d.Auth, state.ID.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.Diagnostics.AddError(
				"Incident type not found",
				fmt.Sprintf("Incident type with ID %s not found", state.ID.ValueString()),
			)
			return
		}
		response.Diagnostics.AddError(
			"Error reading incident type",
			"Could not read incident type, unexpected error: "+err.Error(),
		)
		return
	}

	d.updateStateFromResponse(&state, &resp)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (d *incidentTypeDataSource) updateStateFromResponse(state *incidentTypeDataSourceModel, resp *datadogV2.IncidentTypeResponse) {
	state.ID = types.StringValue(resp.Data.GetId())

	if attributes, ok := resp.Data.GetAttributesOk(); ok {
		state.Name = types.StringValue(attributes.GetName())
		state.Description = types.StringValue(attributes.GetDescription())
		state.IsDefault = types.BoolValue(attributes.GetIsDefault())
	}
}
