package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
	TimeZone     types.String         `tfsdk:"timezone"`
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
			"id": utils.ResourceIDAttribute(),
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
							Required:    false,
							Description: "Defines the query or condition that triggers this routing rule.",
						},
						"query": schema.StringAttribute{
							Required:    false,
							Description: "Defines the query or condition that triggers this routing rule.",
						},
						"escalation_policy": schema.StringAttribute{
							Required:    false,
							Description: "ID of the policy to be applied when this routing rule matches.",
						},
					},
					Blocks: map[string]schema.Block{
						"time_restriction": schema.SingleNestedBlock{
							Description: "Holds time zone information and a list of time restrictions for a routing rule.",
							Attributes: map[string]schema.Attribute{
								"time_zone": schema.StringAttribute{
									Required:    true,
									Description: "Specifies the time zone applicable to the restrictions.",
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
												Description: "The weekday when the restriction period ends (Monday through Sunday).",
											},
											"end_time": schema.StringAttribute{
												Optional:    true,
												Description: "The time of day when the restriction ends (hh:mm:ss).",
											},
											"start_day": schema.StringAttribute{
												Optional:    true,
												Validators:  []validator.String{validators.NewEnumValidator[validator.String](datadogV2.NewWeekdayFromValue)},
												Description: "The weekday when the restriction period starts (Monday through Sunday).",
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
												Optional:    false,
												Description: "Slack channel ID.",
											},
											"workspace": schema.StringAttribute{
												Optional:    false,
												Description: "Slack workspace ID.",
											},
										},
									},
									"send_teams_message": schema.SingleNestedBlock{
										Attributes: map[string]schema.Attribute{
											"channel": schema.StringAttribute{
												Optional:    false,
												Description: "Teams channel ID.",
											},
											"tenant": schema.StringAttribute{
												Optional:    false,
												Description: "Teams tenant ID.",
											},
											"team": schema.StringAttribute{
												Optional:    false,
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

	include := "steps.targets"
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

	stepsById := map[string]*datadogV2.RoutingRule{}

	for _, item := range resp.GetIncluded() {
		if item.RoutingRule != nil && item.RoutingRule.Id != nil {
			stepsById[*item.RoutingRule.Id] = item.RoutingRule
		}
	}

	state := r.stateFromResponse(&resp)

	state.Rules = make([]*teamRuleModel, len(resp.Data.Relationships.Rules.Data))
	// We use the index to match the layer with the plan.Layers
	// As we expect the server to return the layers in the same order as the plan.Layers
	for i, rule := range resp.Data.Relationships.Rules.Data {
		fullRule := stepsById[rule.Id]
		policyId := types.StringNull()
		if fullRule.Relationships.Policy != nil {
			policyId = types.StringValue(fullRule.Relationships.Policy.Data.Id)
		}
		var stateRestrictions *teamTimeRestrictionsModel
		if fullRule.Attributes.TimeRestriction != nil {
			stateRestrictions = &teamTimeRestrictionsModel{
				TimeZone: types.StringValue(fullRule.Attributes.TimeRestriction.TimeZone),
			}
			for _, restriction := range fullRule.Attributes.TimeRestriction.Restrictions {
				stateRestrictions.Restrictions = append(stateRestrictions.Restrictions, &restrictionsModel{
					EndDay:    types.StringValue(string(restriction.GetEndDay())),
					EndTime:   types.StringValue(restriction.GetEndTime()),
					StartDay:  types.StringValue(string(restriction.GetStartDay())),
					StartTime: types.StringValue(restriction.GetStartTime()),
				})
			}
		}
		stateActions := []*teamRuleActionModel{}
		for _, action := range fullRule.Attributes.Actions {
			if action.SlackAction != nil {
				stateActions = append(stateActions, &teamRuleActionModel{
					Slack: &slackMessageModel{
						Workspace: types.StringValue(action.SlackAction.Workspace),
						Channel:   types.StringValue(action.SlackAction.Channel),
					},
				})
			} else if action.TeamsAction != nil {
				stateActions = append(stateActions, &teamRuleActionModel{
					Teams: &teamsMessageModel{
						Tenant:  types.StringValue(action.TeamsAction.Tenant),
						Team:    types.StringValue(action.TeamsAction.Team),
						Channel: types.StringValue(action.TeamsAction.Channel),
					},
				})
			}
		}

		state.Rules[i] = &teamRuleModel{
			Id:               types.StringValue(rule.Id),
			Query:            types.StringPointerValue(fullRule.Attributes.Query),
			Urgency:          types.StringPointerValue((*string)(fullRule.Attributes.Urgency)),
			EscalationPolicy: policyId,
			TimeRestrictions: stateRestrictions,
			Actions:          stateActions,
		}
	}

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *onCallTeamRoutingRulesResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan onCallTeamRoutingRulesModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
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

	//state := r.newState(ctx, &plan, &resp)
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

	stepsById := map[string]*datadogV2.RoutingRule{}

	for _, item := range resp.GetIncluded() {
		if item.RoutingRule != nil && item.RoutingRule.Id != nil {
			stepsById[*item.RoutingRule.Id] = item.RoutingRule
		}
	}

	state.Rules = make([]*teamRuleModel, len(resp.Data.Relationships.Rules.Data))
	for i, rule := range resp.Data.Relationships.Rules.Data {
		fullRule := stepsById[rule.Id]
		policyId := types.StringNull()
		if fullRule.Relationships.Policy != nil {
			policyId = types.StringValue(fullRule.Relationships.Policy.Data.Id)
		}
		var stateRestrictions *teamTimeRestrictionsModel
		if fullRule.Attributes.TimeRestriction != nil {
			stateRestrictions = &teamTimeRestrictionsModel{
				TimeZone: types.StringValue(fullRule.Attributes.TimeRestriction.TimeZone),
			}
			for _, restriction := range fullRule.Attributes.TimeRestriction.Restrictions {
				stateRestrictions.Restrictions = append(stateRestrictions.Restrictions, &restrictionsModel{
					EndDay:    types.StringValue(string(restriction.GetEndDay())),
					EndTime:   types.StringValue(restriction.GetEndTime()),
					StartDay:  types.StringValue(string(restriction.GetStartDay())),
					StartTime: types.StringValue(restriction.GetStartTime()),
				})
			}
		}
		stateActions := []*teamRuleActionModel{}
		for _, action := range fullRule.Attributes.Actions {
			if action.SlackAction != nil {
				stateActions = append(stateActions, &teamRuleActionModel{
					Slack: &slackMessageModel{
						Workspace: types.StringValue(action.SlackAction.Workspace),
						Channel:   types.StringValue(action.SlackAction.Channel),
					},
				})
			} else if action.TeamsAction != nil {
				stateActions = append(stateActions, &teamRuleActionModel{
					Teams: &teamsMessageModel{
						Tenant:  types.StringValue(action.TeamsAction.Tenant),
						Team:    types.StringValue(action.TeamsAction.Team),
						Channel: types.StringValue(action.TeamsAction.Channel),
					},
				})
			}
		}

		state.Rules[i] = &teamRuleModel{
			Id:               types.StringValue(rule.Id),
			Query:            types.StringPointerValue(fullRule.Attributes.Query),
			Urgency:          types.StringPointerValue((*string)(fullRule.Attributes.Urgency)),
			EscalationPolicy: policyId,
			TimeRestrictions: stateRestrictions,
			Actions:          stateActions,
		}
	}
	return state
}

//func (r *onCallTeamRoutingRulesResource) newState(ctx context.Context, plan *onCallTeamRoutingRulesModel, resp *datadogV2.TeamRoutingRules) *onCallTeamRoutingRulesModel {
//	state := &onCallTeamRoutingRulesModel{}
//	state.ID = types.StringValue(resp.Data.GetId())
//
//	rulesById := map[string]*datadogV2.RoutingRule{}
//
//	for _, item := range resp.GetIncluded() {
//		if item.RoutingRule != nil && item.RoutingRule.Id != nil {
//			rulesById[*item.RoutingRule.Id] = item.RoutingRule
//		}
//	}
//
//	for _, ruleRef := range resp.Data.Relationships.Rules.Data {
//		rule := rulesById[ruleRef.Id]
//		attrs := rule.GetAttributes()
//		var ep *string
//		if rule.GetRelationships().Policy != nil {
//			ep = &rule.Relationships.Policy.Data.Id
//		}
//		state.Rules = append(state.Rules, &teamRuleModel{
//			Id:               types.StringValue(ruleRef.Id),
//			Query:            types.StringPointerValue(attrs.Query),
//			Urgency:          types.StringPointerValue((*string)(attrs.Urgency)),
//			EscalationPolicy: types.StringPointerValue(ep),
//			TimeRestrictions: nil,
//			Actions:          nil,
//		})
//	}
//
//	//
//	//state.Layers = make([]*escalationStepModel, len(data.Relationships.Layers.Data))
//	//// We use the index to match the layer with the plan.Layers
//	//// As we expect the server to return the layers in the same order as the plan.Layers
//	//for i, layer := range data.Relationships.Layers.Data {
//	//	var layerExistingEffectiveDate customtypes.BackwardRFC3339Date
//	//	if i < len(plan.Layers) {
//	//		layerExistingEffectiveDate = customtypes.NewBackwardRFC3339Date(plan.Layers[i].EffectiveDate.ValueString())
//	//	}
//	//	state.Layers[i] = newLayerModel(layersByID[layer.GetId()], membersByID, layerExistingEffectiveDate)
//	//}
//	return state
//}

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
				action.TeamsAction = datadogV2.NewTeamsActionWithDefaults()
				action.TeamsAction.Team = plannedAction.Teams.Team.ValueString()
				action.TeamsAction.Tenant = plannedAction.Teams.Tenant.ValueString()
				action.TeamsAction.Channel = plannedAction.Teams.Channel.ValueString()
			}
			if plannedAction.Slack != nil {
				action.SlackAction = datadogV2.NewSlackActionWithDefaults()
				action.SlackAction.Channel = plannedAction.Slack.Channel.ValueString()
				action.SlackAction.Workspace = plannedAction.Slack.Workspace.ValueString()
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
