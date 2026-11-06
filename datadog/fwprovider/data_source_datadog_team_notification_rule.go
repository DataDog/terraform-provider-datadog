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

	// Computed values - direct attributes (not a list)
	Email      *emailModel      `tfsdk:"email"`
	MsTeams    *msTeamsModel    `tfsdk:"ms_teams"`
	Pagerduty  *pagerdutyModel  `tfsdk:"pagerduty"`
	Servicenow *servicenowModel `tfsdk:"servicenow"`
	Slack      *slackModel      `tfsdk:"slack"`
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
	response.TypeName = "team_notification_rule"
}

func (d *datadogTeamNotificationRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about a specific Datadog team notification rule.",
		Attributes: map[string]schema.Attribute{
			// Datasource ID
			"id": utils.ResourceIDAttribute(),
			// Query Parameters
			"team_id": schema.StringAttribute{
				Required:    true,
				Description: "The team ID to fetch the notification rule for.",
			},
			"rule_id": schema.StringAttribute{
				Required:    true,
				Description: "The notification rule ID to fetch.",
			},
		},
		Blocks: map[string]schema.Block{
			"email": schema.SingleNestedBlock{
				Description: "The email notification settings.",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Computed:    true,
						Description: "Flag indicating whether email notifications should be sent",
					},
					"recipient_email": schema.StringAttribute{
						Computed:    true,
						Description: "Email address notifications are sent to instead of all team members",
					},
				},
			},
			"ms_teams": schema.SingleNestedBlock{
				Description: "The MS Teams notification settings.",
				Attributes: map[string]schema.Attribute{
					"connector_name": schema.StringAttribute{
						Computed:    true,
						Description: "MS Teams connector name",
					},
				},
			},
			"pagerduty": schema.SingleNestedBlock{
				Description: "The PagerDuty notification settings.",
				Attributes: map[string]schema.Attribute{
					"service_name": schema.StringAttribute{
						Computed:    true,
						Description: "PagerDuty service name",
					},
				},
			},
			"servicenow": schema.SingleNestedBlock{
				Description: "The ServiceNow notification settings.",
				Attributes: map[string]schema.Attribute{
					"templates": schema.ListAttribute{
						Computed:    true,
						ElementType: types.StringType,
						Description: "ServiceNow template handle names",
					},
				},
			},
			"slack": schema.SingleNestedBlock{
				Description: "The Slack notification settings.",
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
	}
}

func (d *datadogTeamNotificationRuleDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state datadogTeamNotificationRuleDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	teamId := state.TeamId.ValueString()
	ruleId := state.RuleId.ValueString()

	ddResp, _, err := d.Api.GetTeamNotificationRule(d.Auth, teamId, ruleId)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting TeamNotificationRule"))
		return
	}

	if ddResp.Data == nil {
		response.Diagnostics.AddError("TeamNotificationRule not found", fmt.Sprintf("No notification rule found with ID %s for team %s", ruleId, teamId))
		return
	}

	d.updateState(ctx, &state, ddResp.Data)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (d *datadogTeamNotificationRuleDataSource) updateState(ctx context.Context, state *datadogTeamNotificationRuleDataSourceModel, teamNotificationRuleData *datadogV2.TeamNotificationRule) {
	state.ID = types.StringValue(fmt.Sprintf("%s:%s", state.TeamId.ValueString(), teamNotificationRuleData.GetId()))

	attributes := teamNotificationRuleData.GetAttributes()

	// Always populate email block with default false if not present
	state.Email = &emailModel{Enabled: types.BoolValue(false), RecipientEmail: types.StringNull()}
	if email, ok := attributes.GetEmailOk(); ok {
		state.Email.Enabled = types.BoolValue(email.GetEnabled())
		if recipient := email.GetRecipientEmail(); recipient != "" {
			state.Email.RecipientEmail = types.StringValue(recipient)
		}
	}

	// Only populate other blocks if they exist
	if msTeams, ok := attributes.GetMsTeamsOk(); ok {
		state.MsTeams = &msTeamsModel{
			ConnectorName: types.StringValue(msTeams.GetConnectorName()),
		}
	}

	if pagerduty, ok := attributes.GetPagerdutyOk(); ok {
		state.Pagerduty = &pagerdutyModel{
			ServiceName: types.StringValue(pagerduty.GetServiceName()),
		}
	}

	if servicenow, ok := attributes.GetServicenowOk(); ok {
		templates, _ := types.ListValueFrom(ctx, types.StringType, servicenow.GetTemplates())
		state.Servicenow = &servicenowModel{Templates: templates}
	}

	if slack, ok := attributes.GetSlackOk(); ok {
		slackTf := &slackModel{}
		if slack.Channel != nil {
			slackTf.Channel = types.StringValue(slack.GetChannel())
		}
		if slack.Workspace != nil {
			slackTf.Workspace = types.StringValue(slack.GetWorkspace())
		}
		state.Slack = slackTf
	}
}
