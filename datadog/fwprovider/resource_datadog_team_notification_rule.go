package fwprovider

import (
	"context"
	"fmt"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure        = &teamNotificationRuleResource{}
	_ resource.ResourceWithImportState      = &teamNotificationRuleResource{}
	_ resource.ResourceWithConfigValidators = &teamNotificationRuleResource{}
	_ resource.ResourceWithModifyPlan       = &teamNotificationRuleResource{}
)

type teamNotificationRuleResource struct {
	Api  *datadogV2.TeamsApi
	Auth context.Context
}

type teamNotificationRuleModel struct {
	ID     types.String `tfsdk:"id"`
	TeamId types.String `tfsdk:"team_id"`

	Email     *emailModel     `tfsdk:"email"`
	MsTeams   *msTeamsModel   `tfsdk:"ms_teams"`
	Pagerduty *pagerdutyModel `tfsdk:"pagerduty"`
	Slack     *slackModel     `tfsdk:"slack"`
}
type emailModel struct {
	Enabled types.Bool `tfsdk:"enabled"`
}
type msTeamsModel struct {
	ConnectorName types.String `tfsdk:"connector_name"`
}
type pagerdutyModel struct {
	ServiceName types.String `tfsdk:"service_name"`
}
type slackModel struct {
	Channel   types.String `tfsdk:"channel"`
	Workspace types.String `tfsdk:"workspace"`
}

func NewTeamNotificationRuleResource() resource.Resource {
	return &teamNotificationRuleResource{}
}

func (r *teamNotificationRuleResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetTeamsApiV2()
	r.Auth = providerData.Auth
}

func (r *teamNotificationRuleResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "team_notification_rule"
}

func (r *teamNotificationRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog team notification rule resource. This can be used to create and manage notification rules for Datadog teams.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"team_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the team that this notification rule belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"email": schema.SingleNestedBlock{
				Description: "The email notification settings.",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Optional:    true,
						Description: "Whether to send email notifications to team members when alerts are triggered.",
					},
				},
			},
			"ms_teams": schema.SingleNestedBlock{
				Description: "The MS Teams notification settings.",
				Attributes: map[string]schema.Attribute{
					"connector_name": schema.StringAttribute{
						Optional:    true,
						Description: "MS Teams connector name used to route notifications to the appropriate channel.",
					},
				},
			},
			"pagerduty": schema.SingleNestedBlock{
				Description: "The PagerDuty notification settings.",
				Attributes: map[string]schema.Attribute{
					"service_name": schema.StringAttribute{
						Optional:    true,
						Description: "PagerDuty service name to send incident notifications to. The service name can be found in your PagerDuty service settings.",
					},
				},
			},
			"slack": schema.SingleNestedBlock{
				Description: "The Slack notification settings.",
				Attributes: map[string]schema.Attribute{
					"channel": schema.StringAttribute{
						Optional:    true,
						Description: "Slack channel name for notifications (for example, #alerts or #team-notifications).",
					},
					"workspace": schema.StringAttribute{
						Optional:    true,
						Description: "Slack workspace name where the channel is located.",
					},
				},
			},
		},
	}
}

func (r *teamNotificationRuleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {

	result := strings.SplitN(request.ID, ":", 2)
	if len(result) != 2 {
		response.Diagnostics.AddError("error retrieving team_id or rule_id from given ID. Format is team_id:rule_id", "")
		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, frameworkPath.Root("team_id"), result[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, frameworkPath.Root("id"), result[1])...)

}

func (r *teamNotificationRuleResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		&teamNotificationRuleValidator{},
	}
}

func (r *teamNotificationRuleResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	// If the resource is being destroyed, no need to modify the plan
	if request.Plan.Raw.IsNull() {
		return
	}

	var config teamNotificationRuleModel
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}

	var plan teamNotificationRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// For each optional notification block, if it's not in the config but is in the plan,
	// explicitly set it to null in the plan. This ensures proper handling in Terraform 1.1.2
	// where optional nested blocks aren't automatically planned for removal.
	if config.MsTeams == nil && plan.MsTeams != nil {
		plan.MsTeams = nil
	}
	if config.Pagerduty == nil && plan.Pagerduty != nil {
		plan.Pagerduty = nil
	}
	if config.Slack == nil && plan.Slack != nil {
		plan.Slack = nil
	}

	response.Diagnostics.Append(response.Plan.Set(ctx, &plan)...)
}

func (r *teamNotificationRuleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state teamNotificationRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	teamId := state.TeamId.ValueString()
	ruleId := state.ID.ValueString()

	resp, httpResp, err := r.Api.GetTeamNotificationRule(r.Auth, teamId, ruleId)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving TeamNotificationRule"))
		return
	}

	err = r.updateStateFromResponse(ctx, &state, resp)

	if err != nil {
		response.Diagnostics.AddError("error updating state from response", err.Error())
		return
	}

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *teamNotificationRuleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state teamNotificationRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	teamId := state.TeamId.ValueString()

	body, diags := r.buildTeamNotificationRuleRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateTeamNotificationRule(r.Auth, teamId, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating TeamNotificationRule"))
		return
	}

	err = r.updateStateFromResponse(ctx, &state, resp)

	if err != nil {
		response.Diagnostics.AddError("error updating state from response", err.Error())
		return
	}

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *teamNotificationRuleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state teamNotificationRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	teamId := state.TeamId.ValueString()

	id := state.ID.ValueString()

	body, diags := r.buildTeamNotificationRuleRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateTeamNotificationRule(r.Auth, teamId, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating TeamNotificationRule"))
		return
	}

	err = r.updateStateFromResponse(ctx, &state, resp)

	if err != nil {
		response.Diagnostics.AddError("error updating state from response", err.Error())
		return
	}

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *teamNotificationRuleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state teamNotificationRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	teamId := state.TeamId.ValueString()

	id := state.ID.ValueString()

	httpResp, err := r.Api.DeleteTeamNotificationRule(r.Auth, teamId, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting team-notification-rule"))
		return
	}
}

func (r *teamNotificationRuleResource) updateStateFromResponse(ctx context.Context, state *teamNotificationRuleModel, resp datadogV2.TeamNotificationRuleResponse) error {

	// Try to get data normally first
	if resp.Data != nil {
		if id := resp.Data.GetId(); id != "" {
			state.ID = types.StringValue(id)
		}
		data := resp.GetData()
		r.updateState(ctx, state, &data)
		return nil
	}

	// If normal fields failed, try UnparsedObject
	// This happens when the API returns minimal data (just id and type).
	// This happens when a notification rule is created with just the email property disabled.
	// The API will return something like: {"data":{"id":"de603ad9-9955-410e-a213-88612989a848","type":"team_notification_rules"}}
	//  For some reason, the code in the client will have resp.Data = nil and resp.UnparsedObject != nil when no attributes are provided
	if resp.UnparsedObject != nil {
		if dataRaw, ok := resp.UnparsedObject["data"].(map[string]interface{}); ok {
			if idVal, idOk := dataRaw["id"].(string); idOk && idVal != "" {
				state.ID = types.StringValue(idVal)
			}

			// Fail if there are attributes set (as they should have been parsed)
			if attributesRaw, attrsOk := dataRaw["attributes"].(map[string]interface{}); attrsOk {
				return fmt.Errorf("error parsing the response. Attributes should have been parsed but were not: %s", attributesRaw)
			}
		}
	}

	return nil
}

func (r *teamNotificationRuleResource) updateState(ctx context.Context, state *teamNotificationRuleModel, resp *datadogV2.TeamNotificationRule) {

	if resp == nil {
		return
	}

	notificationRule := *resp

	state.ID = types.StringValue(notificationRule.GetId())

	// Default values. If attributes are not present, we will use these
	// The API will NOT return the email.Enabled field if it is false
	state.Email = &emailModel{Enabled: types.BoolValue(false)}
	state.MsTeams = nil
	state.Pagerduty = nil
	state.Slack = nil

	if attributes, ok := notificationRule.GetAttributesOk(); ok {

		if email, ok := attributes.GetEmailOk(); ok {

			emailTf := emailModel{}
			if enabled, ok := email.GetEnabledOk(); ok {
				emailTf.Enabled = types.BoolValue(*enabled)
			}

			state.Email = &emailTf
		}
		if msTeams, ok := attributes.GetMsTeamsOk(); ok {

			msTeamsTf := msTeamsModel{}
			if connectorName, ok := msTeams.GetConnectorNameOk(); ok {
				msTeamsTf.ConnectorName = types.StringValue(*connectorName)
			}
			state.MsTeams = &msTeamsTf
		}
		if pagerduty, ok := attributes.GetPagerdutyOk(); ok {

			pagerdutyTf := pagerdutyModel{}
			if serviceName, ok := pagerduty.GetServiceNameOk(); ok {
				pagerdutyTf.ServiceName = types.StringValue(*serviceName)
			}
			state.Pagerduty = &pagerdutyTf
		}
		if slack, ok := attributes.GetSlackOk(); ok {

			slackTf := slackModel{}
			if channel, ok := slack.GetChannelOk(); ok {
				slackTf.Channel = types.StringValue(*channel)
			}
			if workspace, ok := slack.GetWorkspaceOk(); ok {
				slackTf.Workspace = types.StringValue(*workspace)
			}
			state.Slack = &slackTf
		}
	}
}

func (r *teamNotificationRuleResource) buildTeamNotificationRuleRequestBody(ctx context.Context, state *teamNotificationRuleModel) (*datadogV2.TeamNotificationRuleRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	req := datadogV2.NewTeamNotificationRuleRequestWithDefaults()
	req.Data = *datadogV2.NewTeamNotificationRuleWithDefaults()
	req.Data.Attributes = *datadogV2.NewTeamNotificationRuleAttributesWithDefaults()

	if state.Email != nil && !state.Email.Enabled.IsNull() {
		req.Data.Attributes.Email = datadogV2.NewTeamNotificationRuleAttributesEmailWithDefaults()
		req.Data.Attributes.Email.SetEnabled(state.Email.Enabled.ValueBool())
	}

	if state.MsTeams != nil && !state.MsTeams.ConnectorName.IsNull() {
		req.Data.Attributes.MsTeams = datadogV2.NewTeamNotificationRuleAttributesMsTeamsWithDefaults()
		req.Data.Attributes.MsTeams.SetConnectorName(state.MsTeams.ConnectorName.ValueString())
	}

	if state.Pagerduty != nil && !state.Pagerduty.ServiceName.IsNull() {
		req.Data.Attributes.Pagerduty = datadogV2.NewTeamNotificationRuleAttributesPagerdutyWithDefaults()
		req.Data.Attributes.Pagerduty.SetServiceName(state.Pagerduty.ServiceName.ValueString())
	}

	if state.Slack != nil && (!state.Slack.Channel.IsNull() || !state.Slack.Workspace.IsNull()) {
		req.Data.Attributes.Slack = datadogV2.NewTeamNotificationRuleAttributesSlackWithDefaults()
		if !state.Slack.Channel.IsNull() {
			req.Data.Attributes.Slack.SetChannel(state.Slack.Channel.ValueString())
		}
		if !state.Slack.Workspace.IsNull() {
			req.Data.Attributes.Slack.SetWorkspace(state.Slack.Workspace.ValueString())
		}
	}

	return req, diags
}

// teamNotificationRuleValidator validates that at least one notification type is configured
type teamNotificationRuleValidator struct{}

func (v *teamNotificationRuleValidator) Description(ctx context.Context) string {
	return "Validates that at least one notification type (email, ms_teams, pagerduty, or slack) is configured"
}

func (v *teamNotificationRuleValidator) MarkdownDescription(ctx context.Context) string {
	return "Validates that at least one notification type (`email`, `ms_teams`, `pagerduty`, or `slack`) is configured"
}

func (v *teamNotificationRuleValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config teamNotificationRuleModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if at least one notification type is configured
	hasEmail := config.Email != nil && !config.Email.Enabled.IsNull()
	hasMsTeams := config.MsTeams != nil && !config.MsTeams.ConnectorName.IsNull()
	hasPagerduty := config.Pagerduty != nil && !config.Pagerduty.ServiceName.IsNull()
	hasSlack := config.Slack != nil && (!config.Slack.Channel.IsNull() || !config.Slack.Workspace.IsNull())

	if !hasEmail && !hasMsTeams && !hasPagerduty && !hasSlack {
		resp.Diagnostics.AddError(
			"Missing Notification Configuration",
			"At least one notification type must be configured. Please configure one of: email, ms_teams, pagerduty, or slack.",
		)
	}
}
