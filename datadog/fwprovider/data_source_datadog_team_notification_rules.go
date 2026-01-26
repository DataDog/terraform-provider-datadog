package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSource = &datadogTeamNotificationRulesDataSource{}
)

type datadogTeamNotificationRulesDataSource struct {
	Api  *datadogV2.TeamsApi
	Auth context.Context
}

type datadogTeamNotificationRulesDataSourceModel struct {
	// Datasource ID
	ID types.String `tfsdk:"id"`

	// Query Parameters
	TeamId types.String `tfsdk:"team_id"`

	// Computed values
	NotificationRules []notificationRuleModel `tfsdk:"notification_rules"`
}

// notificationRuleModel represents a notification rule in the data source.
// Uses model types (emailModel, msTeamsModel, pagerdutyModel, slackModel) shared with resource_datadog_team_notification_rule.go
type notificationRuleModel struct {
	ID        types.String    `tfsdk:"id"`
	Email     *emailModel     `tfsdk:"email"`
	MsTeams   *msTeamsModel   `tfsdk:"ms_teams"`
	Pagerduty *pagerdutyModel `tfsdk:"pagerduty"`
	Slack     *slackModel     `tfsdk:"slack"`
}

func NewDatadogTeamNotificationRulesDataSource() datasource.DataSource {
	return &datadogTeamNotificationRulesDataSource{}
}

func (d *datadogTeamNotificationRulesDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetTeamsApiV2()
	d.Auth = providerData.Auth
}

func (d *datadogTeamNotificationRulesDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "team_notification_rules"
}

func (d *datadogTeamNotificationRulesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about all Datadog team notification rules for a specific team.",
		Attributes: map[string]schema.Attribute{
			// Datasource ID
			"id": utils.ResourceIDAttribute(),
			// Query Parameters
			"team_id": schema.StringAttribute{
				Required:    true,
				Description: "The team ID to fetch notification rules for.",
			},
		},
		Blocks: map[string]schema.Block{
			"notification_rules": schema.ListNestedBlock{
				Description: "List of notification rules for the team.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the notification rule.",
						},
					},
					Blocks: map[string]schema.Block{
						"email": schema.SingleNestedBlock{
							Description: "The email notification settings.",
							Attributes: map[string]schema.Attribute{
								"enabled": schema.BoolAttribute{
									Computed:    true,
									Description: "Flag indicating whether email notifications should be sent.",
								},
							},
						},
						"ms_teams": schema.SingleNestedBlock{
							Description: "The MS Teams notification settings.",
							Attributes: map[string]schema.Attribute{
								"connector_name": schema.StringAttribute{
									Computed:    true,
									Description: "MS Teams connector name.",
								},
							},
						},
						"pagerduty": schema.SingleNestedBlock{
							Description: "The PagerDuty notification settings.",
							Attributes: map[string]schema.Attribute{
								"service_name": schema.StringAttribute{
									Computed:    true,
									Description: "PagerDuty service name.",
								},
							},
						},
						"slack": schema.SingleNestedBlock{
							Description: "The Slack notification settings.",
							Attributes: map[string]schema.Attribute{
								"channel": schema.StringAttribute{
									Computed:    true,
									Description: "Slack channel for notifications.",
								},
								"workspace": schema.StringAttribute{
									Computed:    true,
									Description: "Slack workspace for notifications.",
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *datadogTeamNotificationRulesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state datadogTeamNotificationRulesDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	teamId := state.TeamId.ValueString()

	// Fetch all rules for the team
	ddResp, _, err := d.Api.GetTeamNotificationRules(d.Auth, teamId)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error listing TeamNotificationRules"))
		return
	}

	state.NotificationRules = make([]notificationRuleModel, 0, len(ddResp.Data))
	for _, rule := range ddResp.Data {
		state.NotificationRules = append(state.NotificationRules, d.buildNotificationRuleModel(&rule))
	}
	state.ID = types.StringValue(teamId)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (d *datadogTeamNotificationRulesDataSource) buildNotificationRuleModel(teamNotificationRuleData *datadogV2.TeamNotificationRule) notificationRuleModel {
	rule := notificationRuleModel{
		ID: types.StringValue(teamNotificationRuleData.GetId()),
	}

	attributes := teamNotificationRuleData.GetAttributes()

	// Always populate email block with default false if not present
	rule.Email = &emailModel{Enabled: types.BoolValue(false)}
	if email, ok := attributes.GetEmailOk(); ok {
		rule.Email.Enabled = types.BoolValue(email.GetEnabled())
	}

	// Only populate other blocks if they exist
	if msTeams, ok := attributes.GetMsTeamsOk(); ok {
		rule.MsTeams = &msTeamsModel{
			ConnectorName: types.StringValue(msTeams.GetConnectorName()),
		}
	}

	if pagerduty, ok := attributes.GetPagerdutyOk(); ok {
		rule.Pagerduty = &pagerdutyModel{
			ServiceName: types.StringValue(pagerduty.GetServiceName()),
		}
	}

	if slack, ok := attributes.GetSlackOk(); ok {
		slackTf := &slackModel{}
		if slack.Channel != nil {
			slackTf.Channel = types.StringValue(slack.GetChannel())
		}
		if slack.Workspace != nil {
			slackTf.Workspace = types.StringValue(slack.GetWorkspace())
		}
		rule.Slack = slackTf
	}

	return rule
}
