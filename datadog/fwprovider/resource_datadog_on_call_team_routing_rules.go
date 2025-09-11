package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var (
	_ resource.ResourceWithConfigure   = &onCallTeamRoutingRulesResource{}
	_ resource.ResourceWithImportState = &onCallTeamRoutingRulesResource{}
)

type onCallTeamRoutingRulesResource struct {
	Api  *datadogV2.OnCallApi
	Auth context.Context
}

type onCallTeamRoutingRulesModel struct {
	ID    types.String     `tfsdk:"id"`
	Rules []*teamRuleModel `tfsdk:"rule"`
}

type teamRuleModel struct {
	Id               types.String               `tfsdk:"id"`
	Query            types.String               `tfsdk:"query"`
	Urgency          types.String               `tfsdk:"urgency"`
	EscalationPolicy types.String               `tfsdk:"escalation_policy"`
	TimeRestrictions *teamTimeRestrictionsModel `tfsdk:"time_restrictions"`
	Actions          []*teamRuleActionModel     `tfsdk:"action"`
}

type teamTimeRestrictionsModel struct {
	TimeZone     types.String         `tfsdk:"time_zone"`
	Restrictions []*restrictionsModel `tfsdk:"restriction"`
}
type teamRuleActionModel struct {
	Slack *slackMessageModel `tfsdk:"send_slack_message"`
	Teams *teamsMessageModel `tfsdk:"send_teams_message"`
}

type slackMessageModel struct {
	Workspace types.String `tfsdk:"workspace"`
	Channel   types.String `tfsdk:"channel"`
}

type teamsMessageModel struct {
	Tenant  types.String `tfsdk:"tenant"`
	Team    types.String `tfsdk:"team"`
	Channel types.String `tfsdk:"channel"`
}

func (m *onCallTeamRoutingRulesModel) Validate() diag.Diagnostics {
	diags := diag.Diagnostics{}

	for i, rule := range m.Rules {
		root := path.Root("rule").AtListIndex(i)

		if rule.TimeRestrictions != nil {
			if rule.TimeRestrictions.TimeZone.IsNull() {
				diags.AddAttributeError(root.AtName("time_restrictions"), "missing time_zone", "time_restrictions must specify time_zone")
			}
			if len(rule.TimeRestrictions.Restrictions) == 0 {
				diags.AddAttributeError(root.AtName("time_restrictions"), "missing restrictions", "time_restrictions must specify at least one restriction")
			}
		}

		for actionIdx, action := range rule.Actions {
			actionPath := root.AtName("action").AtListIndex(actionIdx)
			if action.Teams == nil && action.Slack == nil {
				diags.AddAttributeError(actionPath, "missing actions", "action must specify one of send_slack_message or send_teams_message")
			}
			if action.Teams != nil {
				teamsPath := actionPath.AtName("send_teams_message")
				if action.Teams.Team.IsNull() {
					diags.AddAttributeError(teamsPath, "missing team", "team is required")
				}
				if action.Teams.Channel.IsNull() {
					diags.AddAttributeError(teamsPath, "missing channel", "channel is required")
				}
				if action.Teams.Tenant.IsNull() {
					diags.AddAttributeError(teamsPath, "missing tenant", "tenant is required")
				}
			}
			if action.Slack != nil {
				teamsPath := actionPath.AtName("send_slack_message")
				if action.Slack.Workspace.IsNull() {
					diags.AddAttributeError(teamsPath, "missing workspace", "workspace is required")
				}
				if action.Slack.Channel.IsNull() {
					diags.AddAttributeError(teamsPath, "missing channel", "channel is required")
				}
			}
		}
	}

	return diags
}

func NewOnCallTeamRoutingRulesResource() resource.Resource {
	return &onCallTeamRoutingRulesResource{}
}

func (r *onCallTeamRoutingRulesResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetOnCallApiV2()
	r.Auth = providerData.Auth
}

func (r *onCallTeamRoutingRulesResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "on_call_team_routing_rules"
}

func (r *onCallTeamRoutingRulesResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog On-Call team routing rules resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    false,
				Required:    true,
				Description: "ID of the team to associate the routing rules with.",
			},
		},
		Blocks: map[string]schema.Block{
			"rule": schema.ListNestedBlock{
				Description: "List of team routing rules.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of this rule.",
						},
						"urgency": schema.StringAttribute{
							Optional:    true,
							Description: "Defines the urgency for pages created via this rule. Only valid if `escalation_policy` is set.",
							Validators:  []validator.String{stringvalidator.OneOf("high", "low", "dynamic")},
						},
						"query": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString(""),
							Description: "Defines the query or condition that triggers this routing rule.",
						},
						"escalation_policy": schema.StringAttribute{
							Optional:    true,
							Description: "ID of the policy to be applied when this routing rule matches.",
						},
					},
					Blocks: map[string]schema.Block{
						"time_restrictions": schema.SingleNestedBlock{
							Description: "Holds time zone information and a list of time restrictions for a routing rule.",
							Attributes: map[string]schema.Attribute{
								"time_zone": schema.StringAttribute{
									Required:    false,
									Optional:    true,
									Description: "Specifies the time zone applicable to the restrictions, e.g. `America/New_York`.",
								},
							},
							Blocks: map[string]schema.Block{
								"restriction": schema.ListNestedBlock{
									Description: "List of restrictions for the rule.",
									NestedObject: schema.NestedBlockObject{
										Attributes: map[string]schema.Attribute{
											"end_day": schema.StringAttribute{
												Optional:    true,
												Validators:  []validator.String{validators.NewEnumValidator[validator.String](datadogV2.NewWeekdayFromValue)},
												Description: "The weekday when the restriction period ends.",
											},
											"end_time": schema.StringAttribute{
												Optional:    true,
												Description: "The time of day when the restriction ends (hh:mm:ss).",
											},
											"start_day": schema.StringAttribute{
												Optional:    true,
												Validators:  []validator.String{validators.NewEnumValidator[validator.String](datadogV2.NewWeekdayFromValue)},
												Description: "The weekday when the restriction period starts.",
											},
											"start_time": schema.StringAttribute{
												Optional:    true,
												Description: "The time of day when the restriction begins (hh:mm:ss).",
											},
										},
									},
								},
							},
						},
						"action": schema.ListNestedBlock{
							Description: "Specifies the list of actions to perform when the routing rule is matched.",
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"send_slack_message": schema.SingleNestedBlock{
										Attributes: map[string]schema.Attribute{
											"channel": schema.StringAttribute{
												Optional:    true,
												Description: "Slack channel ID.",
											},
											"workspace": schema.StringAttribute{
												Optional:    true,
												Description: "Slack workspace ID.",
											},
										},
									},
									"send_teams_message": schema.SingleNestedBlock{
										Attributes: map[string]schema.Attribute{
											"channel": schema.StringAttribute{
												Optional:    true,
												Description: "Teams channel ID.",
											},
											"tenant": schema.StringAttribute{
												Optional:    true,
												Description: "Teams tenant ID.",
											},
											"team": schema.StringAttribute{
												Optional:    true,
												Description: "Teams team ID.",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *onCallTeamRoutingRulesResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *onCallTeamRoutingRulesResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state onCallTeamRoutingRulesModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	id := state.ID.ValueString()

	include := "rules"
	resp, httpResp, err := r.Api.GetOnCallTeamRoutingRules(r.Auth, id, datadogV2.GetOnCallTeamRoutingRulesOptionalParameters{
		Include: &include,
	})
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving OnCallTeamRoutingRules"))
		return
	}

	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	state = *r.stateFromResponse(&resp)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *onCallTeamRoutingRulesResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan onCallTeamRoutingRulesModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(plan.Validate()...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.teamRoutingRulesRequestFromModel(&plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	include := "rules"
	resp, _, err := r.Api.SetOnCallTeamRoutingRules(r.Auth, plan.ID.ValueString(), *body, datadogV2.SetOnCallTeamRoutingRulesOptionalParameters{
		Include: &include,
	})
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating OnCallTeamRoutingRules"))
		return
	}

	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	state := r.stateFromResponse(&resp)
	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *onCallTeamRoutingRulesResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan onCallTeamRoutingRulesModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(plan.Validate()...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.teamRoutingRulesRequestFromModel(&plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	include := "rules"
	resp, _, err := r.Api.SetOnCallTeamRoutingRules(r.Auth, plan.ID.ValueString(), *body, datadogV2.SetOnCallTeamRoutingRulesOptionalParameters{
		Include: &include,
	})
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating OnCallTeamRoutingRules"))
		return
	}

	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	state := r.stateFromResponse(&resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *onCallTeamRoutingRulesResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state onCallTeamRoutingRulesModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	if id == "" {
		response.Diagnostics.AddError("id is required", "id is required")
		return
	}

	_, httpResp, err := r.Api.SetOnCallTeamRoutingRules(r.Auth, id, r.emptyTeamRoutingRules(id), datadogV2.SetOnCallTeamRoutingRulesOptionalParameters{})
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting on_call_team_routing_rules"))
		return
	}
}

func (r *onCallTeamRoutingRulesResource) stateFromResponse(resp *datadogV2.TeamRoutingRules) *onCallTeamRoutingRulesModel {
	state := &onCallTeamRoutingRulesModel{}
	state.ID = types.StringValue(resp.Data.GetId())

	rulesById := map[string]*datadogV2.RoutingRule{}

	for _, item := range resp.GetIncluded() {
		if item.RoutingRule != nil && item.RoutingRule.Id != nil {
			rulesById[*item.RoutingRule.Id] = item.RoutingRule
		}
	}

	state.Rules = make([]*teamRuleModel, len(resp.Data.Relationships.Rules.Data))
	for i, rule := range resp.Data.Relationships.Rules.Data {
		fullRule := rulesById[rule.Id]
		policyId := types.StringNull()
		relationships := fullRule.GetRelationships()
		attributes := fullRule.GetAttributes()
		if relationships.Policy != nil && relationships.Policy.Data != nil {
			policyId = types.StringValue(fullRule.Relationships.Policy.Data.Id)
		}
		var stateRestrictions *teamTimeRestrictionsModel
		if attributes.TimeRestriction != nil {
			stateRestrictions = &teamTimeRestrictionsModel{
				TimeZone: types.StringValue(attributes.TimeRestriction.TimeZone),
			}
			for _, restriction := range attributes.TimeRestriction.Restrictions {
				stateRestrictions.Restrictions = append(stateRestrictions.Restrictions, &restrictionsModel{
					EndDay:    types.StringValue(string(restriction.GetEndDay())),
					EndTime:   types.StringValue(restriction.GetEndTime()),
					StartDay:  types.StringValue(string(restriction.GetStartDay())),
					StartTime: types.StringValue(restriction.GetStartTime()),
				})
			}
		}
		stateActions := []*teamRuleActionModel{}
		for _, action := range attributes.Actions {
			if action.SendSlackMessageAction != nil {
				stateActions = append(stateActions, &teamRuleActionModel{
					Slack: &slackMessageModel{
						Workspace: types.StringValue(action.SendSlackMessageAction.Workspace),
						Channel:   types.StringValue(action.SendSlackMessageAction.Channel),
					},
				})
			} else if action.SendTeamsMessageAction != nil {
				stateActions = append(stateActions, &teamRuleActionModel{
					Teams: &teamsMessageModel{
						Tenant:  types.StringValue(action.SendTeamsMessageAction.Tenant),
						Team:    types.StringValue(action.SendTeamsMessageAction.Team),
						Channel: types.StringValue(action.SendTeamsMessageAction.Channel),
					},
				})
			}
		}

		state.Rules[i] = &teamRuleModel{
			Id:               types.StringValue(rule.Id),
			Query:            types.StringPointerValue(fullRule.GetAttributes().Query),
			Urgency:          types.StringPointerValue((*string)(attributes.Urgency)),
			EscalationPolicy: policyId,
			TimeRestrictions: stateRestrictions,
			Actions:          stateActions,
		}
	}
	return state
}

func (r *onCallTeamRoutingRulesResource) emptyTeamRoutingRules(id string) datadogV2.TeamRoutingRulesRequest {
	req := datadogV2.NewTeamRoutingRulesRequestWithDefaults()

	data := datadogV2.NewTeamRoutingRulesRequestDataWithDefaults()
	data.Id = &id

	req.Data = data

	return *req
}

func (r *onCallTeamRoutingRulesResource) teamRoutingRulesRequestFromModel(state *onCallTeamRoutingRulesModel) (*datadogV2.TeamRoutingRulesRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	req := datadogV2.NewTeamRoutingRulesRequestWithDefaults()

	data := datadogV2.NewTeamRoutingRulesRequestDataWithDefaults()
	data.Id = state.ID.ValueStringPointer()

	attributes := datadogV2.NewTeamRoutingRulesRequestDataAttributesWithDefaults()

	for ruleIndex, plannedRule := range state.Rules {
		rulePath := path.Root("rule").AtListIndex(ruleIndex)
		actions := []datadogV2.RoutingRuleAction{}
		for actionIndex, plannedAction := range plannedRule.Actions {
			action := datadogV2.RoutingRuleAction{}
			if plannedAction.Teams != nil && plannedAction.Slack != nil {
				diags.AddAttributeError(
					rulePath.AtName("action").AtListIndex(actionIndex),
					"action can only have one configuration",
					"only one of `send_slack_message`, `send_teams_message` is allowed per action. Consider adding a separate `action` block.")
				return nil, diags
			}
			if plannedAction.Teams != nil {
				action.SendTeamsMessageAction = datadogV2.NewSendTeamsMessageActionWithDefaults()
				action.SendTeamsMessageAction.Type = datadogV2.SENDTEAMSMESSAGEACTIONTYPE_SEND_TEAMS_MESSAGE
				action.SendTeamsMessageAction.Team = plannedAction.Teams.Team.ValueString()
				action.SendTeamsMessageAction.Tenant = plannedAction.Teams.Tenant.ValueString()
				action.SendTeamsMessageAction.Channel = plannedAction.Teams.Channel.ValueString()
			}
			if plannedAction.Slack != nil {
				action.SendSlackMessageAction = datadogV2.NewSendSlackMessageActionWithDefaults()
				action.SendSlackMessageAction.Type = datadogV2.SENDSLACKMESSAGEACTIONTYPE_SEND_SLACK_MESSAGE
				action.SendSlackMessageAction.Channel = plannedAction.Slack.Channel.ValueString()
				action.SendSlackMessageAction.Workspace = plannedAction.Slack.Workspace.ValueString()
			}
			actions = append(actions, action)
		}

		var timeRestriction *datadogV2.TimeRestrictions
		if plannedRule.TimeRestrictions != nil {
			timeRestriction = datadogV2.NewTimeRestrictionsWithDefaults()
			timeRestriction.TimeZone = plannedRule.TimeRestrictions.TimeZone.ValueString()
			for _, plannedRestriction := range plannedRule.TimeRestrictions.Restrictions {
				restriction := datadogV2.TimeRestriction{}
				if !plannedRestriction.EndDay.IsNull() {
					restriction.SetEndDay(datadogV2.Weekday(plannedRestriction.EndDay.ValueString()))
				}
				if !plannedRestriction.EndTime.IsNull() {
					restriction.SetEndTime(plannedRestriction.EndTime.ValueString())
				}
				if !plannedRestriction.StartDay.IsNull() {
					restriction.SetStartDay(datadogV2.Weekday(plannedRestriction.StartDay.ValueString()))
				}
				if !plannedRestriction.StartTime.IsNull() {
					restriction.SetStartTime(plannedRestriction.StartTime.ValueString())
				}
				timeRestriction.Restrictions = append(timeRestriction.Restrictions, restriction)
			}
		}

		attributes.Rules = append(attributes.Rules, datadogV2.TeamRoutingRulesRequestRule{
			Actions:         actions,
			PolicyId:        plannedRule.EscalationPolicy.ValueStringPointer(),
			Query:           plannedRule.Query.ValueStringPointer(),
			TimeRestriction: timeRestriction,
			Urgency:         (*datadogV2.Urgency)(plannedRule.Urgency.ValueStringPointer()),
		})
	}

	data.Attributes = attributes
	req.Data = data

	return req, diags
}
