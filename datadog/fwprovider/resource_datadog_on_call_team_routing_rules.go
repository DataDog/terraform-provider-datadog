package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
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
	Slack            *slackMessageModel              `tfsdk:"send_slack_message"`
	Teams            *teamsMessageModel              `tfsdk:"send_teams_message"`
	Workflow         *triggerWorkflowAutomationModel `tfsdk:"trigger_workflow_automation"`
	EscalationPolicy *escalationPolicyActionModel    `tfsdk:"escalation_policy"`
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

type triggerWorkflowAutomationModel struct {
	Handle types.String `tfsdk:"handle"`
}

type escalationPolicyActionModel struct {
	PolicyId          types.String                       `tfsdk:"policy_id"`
	AckTimeoutMinutes types.Int64                        `tfsdk:"ack_timeout_minutes"`
	Urgency           types.String                       `tfsdk:"urgency"`
	SupportHours      *escalationPolicySupportHoursModel `tfsdk:"support_hours"`
}

type escalationPolicySupportHoursModel struct {
	TimeZone     types.String         `tfsdk:"time_zone"`
	Restrictions []*restrictionsModel `tfsdk:"restriction"`
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

		hasEscalationPolicyAction := false

		for actionIdx, action := range rule.Actions {
			actionPath := root.AtName("action").AtListIndex(actionIdx)
			if action.Teams == nil && action.Slack == nil && action.Workflow == nil && action.EscalationPolicy == nil {
				diags.AddAttributeError(actionPath, "missing actions", "action must specify one of send_slack_message, send_teams_message, trigger_workflow_automation, or escalation_policy")
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
			if action.Workflow != nil {
				workflowPath := actionPath.AtName("trigger_workflow_automation")
				if action.Workflow.Handle.IsNull() {
					diags.AddAttributeError(workflowPath, "missing handle", "handle is required")
				}
			}
			if action.EscalationPolicy != nil {
				escalationPolicyPath := actionPath.AtName("escalation_policy")
				if hasEscalationPolicyAction {
					diags.AddAttributeError(escalationPolicyPath, "duplicate escalation_policy action", "at most one escalation_policy action is allowed per rule")
				}
				hasEscalationPolicyAction = true
				if action.EscalationPolicy.PolicyId.IsNull() {
					diags.AddAttributeError(escalationPolicyPath, "missing policy_id", "policy_id is required")
				}
				if action.EscalationPolicy.SupportHours != nil {
					supportHoursPath := escalationPolicyPath.AtName("support_hours")
					if action.EscalationPolicy.SupportHours.TimeZone.IsNull() {
						diags.AddAttributeError(supportHoursPath, "missing time_zone", "support_hours must specify time_zone")
					}
					if len(action.EscalationPolicy.SupportHours.Restrictions) == 0 {
						diags.AddAttributeError(supportHoursPath, "missing restrictions", "support_hours must specify at least one restriction")
					}
					if rule.TimeRestrictions != nil {
						diags.AddAttributeError(supportHoursPath, "conflicting time restriction configuration", "cannot combine the rule-level `time_restrictions` block with `support_hours` on an `escalation_policy` action in the same rule. Use one or the other.")
					}
				}
			}
		}

		if hasEscalationPolicyAction && !rule.EscalationPolicy.IsNull() {
			rootEscalationPolicyPath := root.AtName("escalation_policy")
			diags.AddAttributeError(rootEscalationPolicyPath, "conflicting escalation policy configuration", "cannot combine rule-level `escalation_policy` attribute with an `escalation_policy` action in the same rule. Use one or the other.")
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
									"trigger_workflow_automation": schema.SingleNestedBlock{
										Attributes: map[string]schema.Attribute{
											"handle": schema.StringAttribute{
												Optional:    true,
												Description: "The handle of the Workflow Automation to trigger.",
											},
										},
									},
									"escalation_policy": schema.SingleNestedBlock{
										Attributes: map[string]schema.Attribute{
											"policy_id": schema.StringAttribute{
												Optional:    true,
												Description: "Escalation policy ID.",
											},
											"ack_timeout_minutes": schema.Int64Attribute{
												Optional:    true,
												Description: "Number of minutes before an acknowledged page is re-triggered.",
												Validators:  []validator.Int64{int64validator.Between(30, 4320)},
											},
											"urgency": schema.StringAttribute{
												Optional:    true,
												Description: "Urgency for pages created via this action.",
												Validators:  []validator.String{stringvalidator.OneOf("high", "low", "dynamic")},
											},
										},
										Blocks: map[string]schema.Block{
											"support_hours": schema.SingleNestedBlock{
												Description: "Support hours during which the escalation policy will execute.",
												Attributes: map[string]schema.Attribute{
													"time_zone": schema.StringAttribute{
														Optional:    true,
														Description: "Specifies the time zone applicable to the restrictions, e.g. `America/New_York`.",
													},
												},
												Blocks: map[string]schema.Block{
													"restriction": schema.ListNestedBlock{
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

	clearUnknownPolicyActions(&resp)
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	state = *r.stateFromResponse(&resp, &state)

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

	clearUnknownPolicyActions(&resp)
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	state := r.stateFromResponse(&resp, &plan)
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

	clearUnknownPolicyActions(&resp)
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	state := r.stateFromResponse(&resp, &plan)

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

// clearUnknownPolicyActions removes UnparsedObject from routing rule actions of type "escalation_policy".
// The Go client does not yet know this type, so it lands in UnparsedObject and would cause
// CheckForUnparsed to fail. The escalation policy is read from relationships.Policy, not from
// the actions array, so it is safe to drop it here until the client is regenerated.
func clearUnknownPolicyActions(resp *datadogV2.TeamRoutingRules) {
	for i := range resp.Included {
		rr := resp.Included[i].RoutingRule
		if rr == nil || rr.Attributes == nil {
			continue
		}
		for j := range rr.Attributes.Actions {
			action := &rr.Attributes.Actions[j]
			if unparsed, ok := action.UnparsedObject.(map[string]interface{}); ok {
				if unparsed["type"] == "escalation_policy" {
					action.UnparsedObject = nil
				}
			}
		}
	}
}

// stateFromResponse projects the API response into Terraform state.
//
// For escalation policies, the API returns a dual view: legacy rule-level
// fields (policy_id, urgency) and an equivalent action entry. Terraform
// exposes these as two distinct shapes, so we use `prior` to pick which
// shape the user wrote and drop the other.
func (r *onCallTeamRoutingRulesResource) stateFromResponse(resp *datadogV2.TeamRoutingRules, prior *onCallTeamRoutingRulesModel) *onCallTeamRoutingRulesModel {
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

		var priorEscalationAction *escalationPolicyActionModel
		if prior != nil && i < len(prior.Rules) && prior.Rules[i] != nil {
			for _, a := range prior.Rules[i].Actions {
				if a.EscalationPolicy != nil {
					priorEscalationAction = a.EscalationPolicy
					break
				}
			}
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
		var responseEscalationAction *escalationPolicyActionModel
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
			} else if action.TriggerWorkflowAutomationAction != nil {
				stateActions = append(stateActions, &teamRuleActionModel{
					Workflow: &triggerWorkflowAutomationModel{
						Handle: types.StringValue(action.TriggerWorkflowAutomationAction.Handle),
					},
				})
			} else if action.RoutingRuleEscalationPolicyAction != nil {
				ep := action.RoutingRuleEscalationPolicyAction
				responseEscalationAction = &escalationPolicyActionModel{
					PolicyId:          types.StringValue(ep.PolicyId),
					AckTimeoutMinutes: types.Int64PointerValue(ep.AckTimeoutMinutes),
					Urgency:           types.StringPointerValue((*string)(ep.Urgency)),
				}
				if ep.SupportHours != nil {
					sh := &escalationPolicySupportHoursModel{
						TimeZone: types.StringValue(ep.SupportHours.TimeZone),
					}
					for _, r := range ep.SupportHours.Restrictions {
						sh.Restrictions = append(sh.Restrictions, &restrictionsModel{
							StartDay:  types.StringValue(string(r.GetStartDay())),
							StartTime: types.StringValue(r.GetStartTime()),
							EndDay:    types.StringValue(string(r.GetEndDay())),
							EndTime:   types.StringValue(r.GetEndTime()),
						})
					}
					responseEscalationAction.SupportHours = sh
				}
			}
		}

		ruleUrgency := types.StringPointerValue((*string)(attributes.Urgency))
		ruleEscalationPolicy := policyId
		if priorEscalationAction != nil {
			ep := responseEscalationAction
			if ep == nil {
				ep = priorEscalationAction
			}
			stateActions = append(stateActions, &teamRuleActionModel{
				EscalationPolicy: ep,
			})
			ruleUrgency = types.StringNull()
			ruleEscalationPolicy = types.StringNull()
		}

		state.Rules[i] = &teamRuleModel{
			Id:               types.StringValue(rule.Id),
			Query:            types.StringPointerValue(fullRule.GetAttributes().Query),
			Urgency:          ruleUrgency,
			EscalationPolicy: ruleEscalationPolicy,
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
			configured := 0
			if plannedAction.Teams != nil {
				configured++
			}
			if plannedAction.Slack != nil {
				configured++
			}
			if plannedAction.Workflow != nil {
				configured++
			}
			if plannedAction.EscalationPolicy != nil {
				configured++
			}
			if configured > 1 {
				diags.AddAttributeError(
					rulePath.AtName("action").AtListIndex(actionIndex),
					"action can only have one configuration",
					"only one of `send_slack_message`, `send_teams_message`, `trigger_workflow_automation`, `escalation_policy` is allowed per action. Consider adding a separate `action` block.")
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
			if plannedAction.Workflow != nil {
				action.TriggerWorkflowAutomationAction = datadogV2.NewTriggerWorkflowAutomationActionWithDefaults()
				action.TriggerWorkflowAutomationAction.Type = datadogV2.TRIGGERWORKFLOWAUTOMATIONACTIONTYPE_TRIGGER_WORKFLOW_AUTOMATION
				action.TriggerWorkflowAutomationAction.Handle = plannedAction.Workflow.Handle.ValueString()
			}
			if plannedAction.EscalationPolicy != nil {
				epAction := datadogV2.NewRoutingRuleEscalationPolicyActionWithDefaults()
				epAction.Type = datadogV2.ROUTINGRULEESCALATIONPOLICYACTIONTYPE_ESCALATION_POLICY
				epAction.PolicyId = plannedAction.EscalationPolicy.PolicyId.ValueString()
				epAction.AckTimeoutMinutes = plannedAction.EscalationPolicy.AckTimeoutMinutes.ValueInt64Pointer()
				epAction.Urgency = (*datadogV2.Urgency)(plannedAction.EscalationPolicy.Urgency.ValueStringPointer())
				if plannedAction.EscalationPolicy.SupportHours != nil {
					sh := datadogV2.NewRoutingRuleEscalationPolicyActionSupportHoursWithDefaults()
					sh.TimeZone = plannedAction.EscalationPolicy.SupportHours.TimeZone.ValueString()
					for _, plannedRestriction := range plannedAction.EscalationPolicy.SupportHours.Restrictions {
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
						sh.Restrictions = append(sh.Restrictions, restriction)
					}
					epAction.SupportHours = sh
				}
				action.RoutingRuleEscalationPolicyAction = epAction
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
