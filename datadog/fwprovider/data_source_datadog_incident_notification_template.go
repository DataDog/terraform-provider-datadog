package fwprovider

import (
	"context"
	"fmt"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSourceWithConfigure = &incidentNotificationTemplateDataSource{}

type incidentNotificationTemplateDataSource struct {
	Api  *datadogV2.IncidentsApi
	Auth context.Context
}

type incidentNotificationTemplateDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Subject      types.String `tfsdk:"subject"`
	Content      types.String `tfsdk:"content"`
	Category     types.String `tfsdk:"category"`
	IncidentType types.String `tfsdk:"incident_type"`
	Created      types.String `tfsdk:"created"`
	Modified     types.String `tfsdk:"modified"`
}

func NewIncidentNotificationTemplateDataSource() datasource.DataSource {
	return &incidentNotificationTemplateDataSource{}
}

func (d *incidentNotificationTemplateDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "incident_notification_template"
}

func (d *incidentNotificationTemplateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about an existing incident notification template.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the incident notification template.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the notification template.",
				Optional:    true,
				Computed:    true,
			},
			"subject": schema.StringAttribute{
				Description: "The subject line of the notification template.",
				Computed:    true,
			},
			"content": schema.StringAttribute{
				Description: "The content body of the notification template.",
				Computed:    true,
			},
			"category": schema.StringAttribute{
				Description: "The category of the notification template.",
				Computed:    true,
			},
			"incident_type": schema.StringAttribute{
				Description: "The ID of the incident type this notification template is associated with.",
				Computed:    true,
			},
			"created": schema.StringAttribute{
				Description: "Timestamp when the notification template was created.",
				Computed:    true,
			},
			"modified": schema.StringAttribute{
				Description: "Timestamp when the notification template was last modified.",
				Computed:    true,
			},
		},
	}
}

func (d *incidentNotificationTemplateDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	providerData, ok := request.ProviderData.(*FrameworkProvider)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *FrameworkProvider, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)
		return
	}

	d.Api = providerData.DatadogApiInstances.GetIncidentsApiV2()
	d.Auth = providerData.Auth
}

func (d *incidentNotificationTemplateDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state incidentNotificationTemplateDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// If ID is provided, fetch by ID
	if !state.ID.IsNull() && !state.ID.IsUnknown() && state.ID.ValueString() != "" {
		id, err := uuid.Parse(state.ID.ValueString())
		if err != nil {
			response.Diagnostics.AddError(
				"Error parsing ID",
				"Could not parse incident notification template ID: "+err.Error(),
			)
			return
		}

		resp, httpResp, err := d.Api.GetIncidentNotificationTemplate(d.Auth, id)
		if err != nil {
			response.Diagnostics.AddError(
				"Error reading incident notification template",
				"Could not read incident notification template ID "+state.ID.ValueString()+": "+err.Error(),
			)
			return
		}
		if httpResp.StatusCode != 200 {
			response.Diagnostics.AddError(
				"Error reading incident notification template",
				fmt.Sprintf("Received HTTP status %d", httpResp.StatusCode),
			)
			return
		}

		d.updateStateFromResponse(&state, &resp)
	} else if !state.Name.IsNull() && !state.Name.IsUnknown() && state.Name.ValueString() != "" {
		// If name is provided, search by name
		resp, httpResp, err := d.Api.ListIncidentNotificationTemplates(d.Auth)
		if err != nil {
			response.Diagnostics.AddError(
				"Error listing incident notification templates",
				"Could not list incident notification templates: "+err.Error(),
			)
			return
		}
		if httpResp.StatusCode != 200 {
			response.Diagnostics.AddError(
				"Error listing incident notification templates",
				fmt.Sprintf("Received HTTP status %d", httpResp.StatusCode),
			)
			return
		}

		// Find template by name
		templates := resp.GetData()
		var foundTemplate *datadogV2.IncidentNotificationTemplateResponseData
		for _, template := range templates {
			if attrs, ok := template.GetAttributesOk(); ok && attrs != nil {
				if attrs.GetName() == state.Name.ValueString() {
					foundTemplate = &template
					break
				}
			}
		}

		if foundTemplate == nil {
			response.Diagnostics.AddError(
				"Incident notification template not found",
				"Could not find incident notification template with name: "+state.Name.ValueString(),
			)
			return
		}

		// Create a single template response for consistency with the updateStateFromResponse function
		singleResp := datadogV2.IncidentNotificationTemplate{
			Data:     *foundTemplate,
			Included: resp.GetIncluded(),
		}
		d.updateStateFromResponse(&state, &singleResp)
	} else {
		response.Diagnostics.AddError(
			"Missing search criteria",
			"Either 'id' or 'name' must be provided to look up the incident notification template.",
		)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (d *incidentNotificationTemplateDataSource) updateStateFromResponse(state *incidentNotificationTemplateDataSourceModel, resp *datadogV2.IncidentNotificationTemplate) {
	data := resp.GetData()

	state.ID = types.StringValue(data.GetId().String())

	if attributes, ok := data.GetAttributesOk(); ok && attributes != nil {
		state.Name = types.StringValue(attributes.GetName())
		state.Subject = types.StringValue(attributes.GetSubject())
		// Normalize content by trimming trailing newlines to match resource behavior
		content := strings.TrimRight(attributes.GetContent(), "\n")
		state.Content = types.StringValue(content)
		state.Category = types.StringValue(attributes.GetCategory())
		state.Created = types.StringValue(attributes.GetCreated().Format("2006-01-02T15:04:05Z"))
		state.Modified = types.StringValue(attributes.GetModified().Format("2006-01-02T15:04:05Z"))
	}

	// Extract incident type ID from relationships
	if relationships, ok := data.GetRelationshipsOk(); ok && relationships != nil {
		if incidentType, ok := relationships.GetIncidentTypeOk(); ok && incidentType != nil {
			if incidentTypeData, ok := incidentType.GetDataOk(); ok && incidentTypeData != nil {
				state.IncidentType = types.StringValue(incidentTypeData.GetId())
			}
		}
	}
}
