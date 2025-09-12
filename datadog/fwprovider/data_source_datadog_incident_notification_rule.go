package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSourceWithConfigure = &incidentNotificationRuleDataSource{}

type incidentNotificationRuleDataSource struct {
	Api  *datadogV2.IncidentsApi
	Auth context.Context
}

type incidentNotificationRuleDataSourceModel struct {
	ID                   types.String                             `tfsdk:"id"`
	Conditions           []incidentNotificationRuleConditionModel `tfsdk:"conditions"`
	Enabled              types.Bool                               `tfsdk:"enabled"`
	Handles              []types.String                           `tfsdk:"handles"`
	RenotifyOn           []types.String                           `tfsdk:"renotify_on"`
	Trigger              types.String                             `tfsdk:"trigger"`
	Visibility           types.String                             `tfsdk:"visibility"`
	IncidentType         types.String                             `tfsdk:"incident_type"`
	NotificationTemplate types.String                             `tfsdk:"notification_template"`
	Created              types.String                             `tfsdk:"created"`
	Modified             types.String                             `tfsdk:"modified"`
}

func NewIncidentNotificationRuleDataSource() datasource.DataSource {
	return &incidentNotificationRuleDataSource{}
}

func (d *incidentNotificationRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "incident_notification_rule"
}

func (d *incidentNotificationRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about an existing incident notification rule.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the incident notification rule.",
				Required:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the notification rule is enabled.",
				Computed:    true,
			},
			"handles": schema.ListAttribute{
				Description: "The notification handles (targets) for this rule.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"renotify_on": schema.ListAttribute{
				Description: "List of incident fields that trigger re-notification when changed.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"trigger": schema.StringAttribute{
				Description: "The trigger event for this notification rule.",
				Computed:    true,
			},
			"visibility": schema.StringAttribute{
				Description: "The visibility of the notification rule. Valid values are: all, organization, private.",
				Computed:    true,
			},
			"incident_type": schema.StringAttribute{
				Description: "The ID of the incident type this notification rule is associated with.",
				Computed:    true,
			},
			"notification_template": schema.StringAttribute{
				Description: "The ID of the notification template used by this rule.",
				Computed:    true,
			},
			"created": schema.StringAttribute{
				Description: "Timestamp when the notification rule was created.",
				Computed:    true,
			},
			"modified": schema.StringAttribute{
				Description: "Timestamp when the notification rule was last modified.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"conditions": schema.ListNestedBlock{
				Description: "The conditions that trigger this notification rule.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"field": schema.StringAttribute{
							Description: "The incident field to evaluate. Common values include: state, severity, services, teams. Custom fields are also supported.",
							Computed:    true,
						},
						"values": schema.ListAttribute{
							Description: "The value(s) to compare against.",
							Computed:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

func (d *incidentNotificationRuleDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func (d *incidentNotificationRuleDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var config incidentNotificationRuleDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := uuid.Parse(config.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Error parsing ID",
			"Could not parse incident notification rule ID: "+err.Error(),
		)
		return
	}

	resp, httpResp, err := d.Api.GetIncidentNotificationRule(d.Auth, id)
	if err != nil {
		response.Diagnostics.AddError(
			"Error reading incident notification rule",
			"Could not read incident notification rule ID "+config.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	if httpResp.StatusCode != 200 {
		response.Diagnostics.AddError(
			"Error reading incident notification rule",
			fmt.Sprintf("Received HTTP status %d", httpResp.StatusCode),
		)
		return
	}

	d.updateStateFromResponse(&config, &resp)

	response.Diagnostics.Append(response.State.Set(ctx, &config)...)
}

func (d *incidentNotificationRuleDataSource) updateStateFromResponse(state *incidentNotificationRuleDataSourceModel, resp *datadogV2.IncidentNotificationRule) {
	data := resp.GetData()

	state.ID = types.StringValue(data.GetId().String())

	if attributes, ok := data.GetAttributesOk(); ok && attributes != nil {
		// Convert conditions
		if conditions, conditionsOk := attributes.GetConditionsOk(); conditionsOk && conditions != nil {
			state.Conditions = make([]incidentNotificationRuleConditionModel, len(*conditions))
			for i, condition := range *conditions {
				values := make([]types.String, len(condition.Values))
				for j, value := range condition.Values {
					values[j] = types.StringValue(value)
				}
				state.Conditions[i] = incidentNotificationRuleConditionModel{
					Field:  types.StringValue(condition.Field),
					Values: values,
				}
			}
		}

		if enabled, enabledOk := attributes.GetEnabledOk(); enabledOk && enabled != nil {
			state.Enabled = types.BoolValue(*enabled)
		}

		if handles, handlesOk := attributes.GetHandlesOk(); handlesOk && handles != nil {
			state.Handles = make([]types.String, len(*handles))
			for i, handle := range *handles {
				state.Handles[i] = types.StringValue(handle)
			}
		}

		if renotifyOn, renotifyOnOk := attributes.GetRenotifyOnOk(); renotifyOnOk && renotifyOn != nil {
			state.RenotifyOn = make([]types.String, len(*renotifyOn))
			for i, item := range *renotifyOn {
				state.RenotifyOn[i] = types.StringValue(item)
			}
		}

		state.Trigger = types.StringValue(attributes.GetTrigger())

		if visibility, visibilityOk := attributes.GetVisibilityOk(); visibilityOk && visibility != nil {
			state.Visibility = types.StringValue(string(*visibility))
		}

		if created, createdOk := attributes.GetCreatedOk(); createdOk && created != nil {
			state.Created = types.StringValue(created.Format("2006-01-02T15:04:05Z"))
		}

		if modified, modifiedOk := attributes.GetModifiedOk(); modifiedOk && modified != nil {
			state.Modified = types.StringValue(modified.Format("2006-01-02T15:04:05Z"))
		}
	}

	if relationships, ok := data.GetRelationshipsOk(); ok && relationships != nil {
		if incidentType, ok := relationships.GetIncidentTypeOk(); ok && incidentType != nil {
			if incidentTypeData, ok := incidentType.GetDataOk(); ok && incidentTypeData != nil {
				state.IncidentType = types.StringValue(incidentTypeData.GetId())
			}
		}

		if notificationTemplate, ok := relationships.GetNotificationTemplateOk(); ok && notificationTemplate != nil {
			if templateData, ok := notificationTemplate.GetDataOk(); ok && templateData != nil {
				state.NotificationTemplate = types.StringValue(templateData.GetId().String())
			}
		}
	}
}
