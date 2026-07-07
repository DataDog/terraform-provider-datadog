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
	ID            types.String                    `tfsdk:"id"`
	Name          types.String                    `tfsdk:"name"`
	Description   types.String                    `tfsdk:"description"`
	IsDefault     types.Bool                      `tfsdk:"is_default"`
	Configuration *incidentTypeConfigurationModel `tfsdk:"configuration"`
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
			"configuration": schema.SingleNestedAttribute{
				Description: "The incident type's behavior settings.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"private_incidents":                          schema.BoolAttribute{Description: "Whether responders can create private incidents of this type.", Computed: true},
					"private_incidents_by_default":               schema.BoolAttribute{Description: "Whether incidents of this type are created as private by default.", Computed: true},
					"allow_workflows":                            schema.BoolAttribute{Description: "Whether automation workflows can be triggered for incidents of this type.", Computed: true},
					"allow_incident_deletion":                    schema.BoolAttribute{Description: "Whether incidents of this type can be deleted.", Computed: true},
					"editable_timestamps":                        schema.BoolAttribute{Description: "Whether responders can edit incident timestamps for incidents of this type.", Computed: true},
					"test_incidents":                             schema.BoolAttribute{Description: "Whether incidents of this type are treated as test incidents.", Computed: true},
					"create_message":                             schema.StringAttribute{Description: "An optional message shown to users when they declare an incident of this type.", Computed: true},
					"disable_out_of_the_box_postmortem_template": schema.BoolAttribute{Description: "Whether the out-of-the-box postmortem template is disabled for incidents of this type.", Computed: true},
					"slug_source":                                schema.StringAttribute{Description: "The source used to derive the incident slug (`default` or `servicenow`).", Computed: true},
				},
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

		if cfg, ok := attributes.GetConfigurationOk(); ok {
			state.Configuration = &incidentTypeConfigurationModel{
				PrivateIncidents:                     types.BoolValue(cfg.GetPrivateIncidents()),
				PrivateIncidentsByDefault:            types.BoolValue(cfg.GetPrivateIncidentsByDefault()),
				AllowWorkflows:                       types.BoolValue(cfg.GetAllowWorkflows()),
				AllowIncidentDeletion:                types.BoolValue(cfg.GetAllowIncidentDeletion()),
				EditableTimestamps:                   types.BoolValue(cfg.GetEditableTimestamps()),
				TestIncidents:                        types.BoolValue(cfg.GetTestIncidents()),
				CreateMessage:                        types.StringValue(cfg.GetCreateMessage()),
				DisableOutOfTheBoxPostmortemTemplate: types.BoolValue(cfg.GetDisableOutOfTheBoxPostmortemTemplate()),
				SlugSource:                           types.StringValue(string(cfg.GetSlugSource())),
			}
		}
	}
}
