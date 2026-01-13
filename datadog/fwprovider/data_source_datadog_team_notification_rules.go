package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSource = &datadogTeamNotificationRuleDataSource{}
)

type datadogTeamNotificationRuleDataSource struct {
	Api  *datadogV2.TeamsApi
	Auth context.Context
}

type datadogTeamNotificationRuleDataSourceModel struct {
	// Datasource ID
	ID types.String `tfsdk:"id"`

	// Query Parameters
	TeamId types.String `tfsdk:"team_id"`
	RuleId types.String `tfsdk:"rule_id"`

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

func NewDatadogTeamNotificationRuleDataSource() datasource.DataSource {
	return &datadogTeamNotificationRuleDataSource{}
}

func (d *datadogTeamNotificationRuleDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetTeamsApiV2()
	d.Auth = providerData.Auth
}

func (d *datadogTeamNotificationRuleDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "team_notification_rules"
}

func (d *datadogTeamNotificationRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about Datadog team notification rules.",
		Attributes: map[string]schema.Attribute{
			// Datasource ID
			"id": utils.ResourceIDAttribute(),
			// Query Parameters
			"team_id": schema.StringAttribute{
				Required:    true,
				Description: "The team ID to fetch notification rules for.",
			},
			"rule_id": schema.StringAttribute{
				Optional:    true,
				Description: "Optional rule ID to filter to a specific notification rule. If not provided, all notification rules for the team will be returned.",
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
							Attributes: map[string]schema.Attribute{
								"enabled": schema.BoolAttribute{
									Computed:    true,
									Description: "Flag indicating whether email notifications should be sent",
								},
							},
						},
						"ms_teams": schema.SingleNestedBlock{
							Attributes: map[string]schema.Attribute{
								"connector_name": schema.StringAttribute{
									Computed:    true,
									Description: "MS Teams connector name",
								},
							},
						},
						"pagerduty": schema.SingleNestedBlock{
							Attributes: map[string]schema.Attribute{
								"service_name": schema.StringAttribute{
									Computed:    true,
									Description: "PagerDuty service name",
								},
							},
						},
						"slack": schema.SingleNestedBlock{
							Attributes: map[string]schema.Attribute{
								"channel": schema.StringAttribute{
									Computed:    true,
									Description: "Slack channel for notifications",
								},
								"workspace": schema.StringAttribute{
									Computed:    true,
									Description: "Slack workspace for notifications",
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *datadogTeamNotificationRuleDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state datadogTeamNotificationRuleDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	teamId := state.TeamId.ValueString()

	if !state.RuleId.IsNull() {
		// Fetch a specific rule
		teamNotificationRuleId := state.RuleId.ValueString()

		ddResp, _, err := d.Api.GetTeamNotificationRule(d.Auth, teamId, teamNotificationRuleId)
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting TeamNotificationRule"))
			return
		}

		if ddResp.Data != nil {
			state.NotificationRules = []notificationRuleModel{d.buildNotificationRuleModel(ddResp.Data)}
		}
		state.ID = types.StringValue(fmt.Sprintf("%s:%s", teamId, teamNotificationRuleId))
	} else {
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
	}

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (d *datadogTeamNotificationRuleDataSource) buildNotificationRuleModel(teamNotificationRuleData *datadogV2.TeamNotificationRule) notificationRuleModel {
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
