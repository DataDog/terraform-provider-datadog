package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var (
	_ resource.ResourceWithConfigure   = &onCallTeamRulesResource{}
	_ resource.ResourceWithImportState = &onCallTeamRulesResource{}
)

type onCallTeamRulesResource struct {
	Api  *datadogV2.OnCallApi
	Auth context.Context
}

//type TeamProcessingRules struct {
//	TeamUUID string           `jsonapi:"primary,team_routing_rules" json:"id"`
//	Rules    []ProcessingRule `jsonapi:"relationship" json:"rules"`
//}
//
//type TeamProcessingRulesRequest struct {
//	TeamUUID string                  `jsonapi:"primary,team_routing_rules" json:"id"`
//	Rules    []ProcessingRuleRequest `jsonapi:"attribute" json:"rules"`
//}
//
//type RuleTimeRestriction struct {
//	TimeZone     string            `json:"time_zone"`
//	Restrictions []TimeRestriction `json:"restrictions"`
//}
//
//type ProcessingRule struct {
//	ID               string                    `jsonapi:"primary,team_routing_rules" json:"id"`
//	Query            string                    `jsonapi:"attribute" json:"query"`
//	EscalationPolicy EscalationPolicyReference `jsonapi:"relationship" json:"policy"`
//	Urgency          string                    `jsonapi:"attribute" json:"urgency" openapi:"enum=low|high|dynamic"`
//	TimeRestriction  *RuleTimeRestriction      `jsonapi:"attribute" json:"time_restriction"`
//	Actions          []ProcessingRuleAction    `jsonapi:"attribute" json:"actions"`
//}

type onCallTeamRulesModel struct {
	ID    types.String     `tfsdk:"id"`
	Rules []*teamRuleModel `tfsdk:"rule"`
}

type teamRuleModel struct {
	Id               types.String           `tfsdk:"id"`
	Query            types.String           `tfsdk:"query"`
	Urgency          types.String           `tfsdk:"urgency"`
	EscalationPolicy types.String           `tfsdk:"escalation_policy"`
	Assignment       types.String           `tfsdk:"assignment"`
	Restrictions     []*restrictionsModel   `tfsdk:"restriction"`
	Actions          []*teamRuleActionModel `tfsdk:"action"`
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

func NewOnCallTeamRulesResource() resource.Resource {
	return &onCallTeamRulesResource{}
}

func (r *onCallTeamRulesResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetOnCallApiV2()
	r.Auth = providerData.Auth
}

func (r *onCallTeamRulesResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "on_call_team_rules"
}

func (r *onCallTeamRulesResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog On-Call escalation policy resource. This can be used to create and manage Datadog On-Call escalation policies.",
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
												Validators:  []validator.String{validators.NewEnumValidator[validator.String](datadogV2.NewLayerAttributesRestrictionsItemsEndDayFromValue)},
												Description: "The weekday when the restriction period ends (Monday through Sunday).",
											},
											"end_time": schema.StringAttribute{
												Optional:    true,
												Description: "The time of day when the restriction ends (hh:mm:ss).",
											},
											"start_day": schema.StringAttribute{
												Optional:    true,
												Validators:  []validator.String{validators.NewEnumValidator[validator.String](datadogV2.NewLayerAttributesRestrictionsItemsStartDayFromValue)},
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

func (r *onCallTeamRulesResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *onCallTeamRulesResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state onCallTeamRulesModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	id := state.ID.ValueString()

	include := "steps.targets"
	resp, httpResp, err := r.Api.GetOnCallTeamRules(r.Auth, id, datadogV2.GetOnCallTeamRulesOptionalParameters{
		Include: &include,
	})
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving OnCallTeamRules"))
		return
	}
	// TODO: restore once client is updated
	//if err := utils.CheckForUnparsed(resp); err != nil {
	//	response.Diagnostics.AddError("response contains unparsed object", err.Error())
	//	return
	//}

	state = *r.newState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *onCallTeamRulesResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan onCallTeamRulesModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildOnCallTeamRulesRequestBody(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	include := "steps.targets"
	resp, _, err := r.Api.CreateOnCallTeamRules(r.Auth, *body, datadogV2.CreateOnCallTeamRulesOptionalParameters{
		Include: &include,
	})
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating OnCallTeamRules"))
		return
	}

	// TODO: restore once client is updated
	//if err := utils.CheckForUnparsed(resp); err != nil {
	//	response.Diagnostics.AddError("response contains unparsed object", err.Error())
	//	return
	//}
	stepsById := map[string]*datadogV2.TeamRulesStep{}

	for _, item := range resp.GetIncluded() {
		if item.TeamRulesStep != nil && item.TeamRulesStep.Id != nil {
			stepsById[*item.TeamRulesStep.Id] = item.TeamRulesStep
		}
	}

	state := r.newState(ctx, &plan, &resp)

	state.Steps = make([]*escalationStepModel, len(resp.Data.Relationships.Steps.Data))
	// We use the index to match the layer with the plan.Layers
	// As we expect the server to return the layers in the same order as the plan.Layers
	for i, step := range resp.Data.Relationships.Steps.Data {
		fullStep := stepsById[step.Id]
		state.Steps[i] = &escalationStepModel{
			Id:            types.StringValue(step.Id),
			EscalateAfter: types.Int64PointerValue(fullStep.Attributes.EscalateAfterSeconds),
			Assignment:    types.StringPointerValue((*string)(fullStep.Attributes.Assignment)),
			Targets:       nil,
		}
	}

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *onCallTeamRulesResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan onCallTeamRulesModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := plan.ID.ValueString()

	if id == "" {
		response.Diagnostics.AddError("id is required", "id is required")
		return
	}

	body, diags := r.buildOnCallTeamRulesUpdateRequestBody(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	include := "steps.targets"
	resp, _, err := r.Api.UpdateOnCallTeamRules(r.Auth, id, *body, datadogV2.UpdateOnCallTeamRulesOptionalParameters{
		Include: &include,
	})
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating OnCallTeamRules"))
		return
	}
	// TODO: restore once client is updated
	//if err := utils.CheckForUnparsed(resp); err != nil {
	//	response.Diagnostics.AddError("response contains unparsedObject", err.Error())
	//	return
	//}
	state := r.newState(ctx, &plan, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *onCallTeamRulesResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state onCallTeamRulesModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	if id == "" {
		response.Diagnostics.AddError("id is required", "id is required")
		return
	}

	httpResp, err := r.Api.DeleteOnCallTeamRules(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting on_call_schedule"))
		return
	}
}

func (r *onCallTeamRulesResource) newState(ctx context.Context, plan *onCallTeamRulesModel, resp *datadogV2.TeamRules) *onCallTeamRulesModel {
	state := &onCallTeamRulesModel{}
	state.ID = types.StringValue(resp.Data.GetId())

	data := resp.GetData()
	attributes := data.GetAttributes()

	if name, ok := attributes.GetNameOk(); ok {
		state.Name = types.StringValue(*name)
	}

	attributes.GetDescriptionOk()

	if retries, ok := attributes.GetRetriesOk(); ok {
		state.Retries = types.Int64Value(*retries)
	}

	if resolveOnEnd, ok := attributes.GetResolvePageOnPolicyEndOk(); ok {
		state.ResolvePageOnPolicyEnd = types.BoolValue(*resolveOnEnd)
	}

	teams := make([]string, len(data.GetRelationships().Teams.GetData()))
	for i, team := range data.GetRelationships().Teams.GetData() {
		teams[i] = team.GetId()
	}
	state.Teams, _ = types.ListValueFrom(ctx, types.StringType, teams)

	stepsById := map[string]*datadogV2.TeamRulesStep{}

	for _, item := range resp.GetIncluded() {
		if item.TeamRulesStep != nil && item.TeamRulesStep.Id != nil {
			stepsById[*item.TeamRulesStep.Id] = item.TeamRulesStep
		}
	}
	//
	//state.Layers = make([]*escalationStepModel, len(data.Relationships.Layers.Data))
	//// We use the index to match the layer with the plan.Layers
	//// As we expect the server to return the layers in the same order as the plan.Layers
	//for i, layer := range data.Relationships.Layers.Data {
	//	var layerExistingEffectiveDate customtypes.BackwardRFC3339Date
	//	if i < len(plan.Layers) {
	//		layerExistingEffectiveDate = customtypes.NewBackwardRFC3339Date(plan.Layers[i].EffectiveDate.ValueString())
	//	}
	//	state.Layers[i] = newLayerModel(layersByID[layer.GetId()], membersByID, layerExistingEffectiveDate)
	//}
	return state
}

//
//func newLayerModel(layer *datadogV2.Layer, membersByID map[string]*datadogV2.TeamRulesMember, layerExistingEffectiveDate customtypes.BackwardRFC3339Date) *escalationStepModel {
//	membersData := layer.GetRelationships().Members.GetData()
//	memberIds := make([]types.String, len(membersData))
//	for j, member := range membersData {
//		includedMember := membersByID[member.GetId()]
//		userId := includedMember.GetRelationships().User.GetData().Id
//		if userId != "" {
//			memberIds[j] = types.StringValue(userId)
//		} else {
//			memberIds[j] = types.StringNull()
//		}
//	}
//	restrictions := layer.GetAttributes().Restrictions
//	restrictionsModels := make([]*restrictionsModel, len(restrictions))
//	for j, restriction := range restrictions {
//		restrictionsModels[j] = &restrictionsModel{
//			EndDay:    types.StringValue(string(restriction.GetEndDay())),
//			EndTime:   types.StringValue(restriction.GetEndTime()),
//			StartDay:  types.StringValue(string(restriction.GetStartDay())),
//			StartTime: types.StringValue(restriction.GetStartTime()),
//		}
//	}
//	interval := layer.GetAttributes().Interval
//
//	var endDateStringValue types.String
//	endDate := layer.GetAttributes().EndDate
//	if endDate != nil && !endDate.IsZero() {
//		endDateStringValue = types.StringValue(formatTime(*endDate))
//	} else {
//		endDateStringValue = types.StringNull()
//	}
//
//	appliedEffectiveDate := customtypes.NewBackwardRFC3339Date(formatTime(layer.Attributes.GetEffectiveDate()))
//
//	// If the effective date is not set, use the applied effective date
//	// Otherwise, use the existing effective date
//	// The effective date is irrelevant to the state (applied_effective_date is), we just need to be sure to make it
//	// consistent for Terraform.
//	var effectiveDate customtypes.BackwardRFC3339Date
//	if layerExistingEffectiveDate.ValueString() != "" {
//		effectiveDate = customtypes.NewBackwardRFC3339Date(layerExistingEffectiveDate.ValueString())
//	} else {
//		effectiveDate = appliedEffectiveDate
//	}
//
//	return &escalationStepModel{
//		Id:                   types.StringValue(layer.GetId()),
//		AppliedEffectiveDate: appliedEffectiveDate,
//		EffectiveDate:        effectiveDate,
//		EndDate:              endDateStringValue,
//		Name:                 types.StringValue(layer.Attributes.GetName()),
//		RotationStart:        types.StringValue(formatTime(layer.Attributes.GetRotationStart())),
//		Users:                memberIds,
//		Restrictions:         restrictionsModels,
//		Interval:             &intervalModel{Days: types.Int32Value(int32(interval.GetDays())), Seconds: types.Int64Value(interval.GetSeconds())},
//	}
//}

func (r *onCallTeamRulesResource) buildOnCallTeamRulesRequestBody(ctx context.Context, state *onCallTeamRulesModel) (*datadogV2.TeamRulesCreateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	req := &datadogV2.TeamRulesCreateRequest{}

	var teams []string
	diags.Append(state.Teams.ElementsAs(ctx, &teams, false)...)

	var relationships *datadogV2.TeamRulesCreateRequestDataRelationships

	teamRelationships := make([]datadogV2.TeamRulesCreateRequestDataRelationshipsTeamsDataItems, len(teams))

	for t, teamId := range teams {
		item := datadogV2.NewTeamRulesCreateRequestDataRelationshipsTeamsDataItemsWithDefaults()
		item.SetId(teamId)
		teamRelationships[t] = *item
	}

	if len(teamRelationships) > 0 {
		relationships = &datadogV2.TeamRulesCreateRequestDataRelationships{
			Teams: &datadogV2.TeamRulesCreateRequestDataRelationshipsTeams{
				Data: teamRelationships,
			},
		}
	}

	attributes := datadogV2.NewTeamRulesCreateRequestDataAttributesWithDefaults()

	attributes.SetName(state.Name.ValueString())

	if !state.Retries.IsNull() {
		attributes.SetRetries(state.Retries.ValueInt64())
	}

	if !state.ResolvePageOnPolicyEnd.IsNull() {
		attributes.SetResolvePageOnPolicyEnd(state.ResolvePageOnPolicyEnd.ValueBool())
	}

	var steps []datadogV2.TeamRulesCreateRequestDataAttributesStepsItems
	for _, plannedStep := range state.Steps {
		step := datadogV2.NewTeamRulesCreateRequestDataAttributesStepsItemsWithDefaults()

		plannedAssignment := plannedStep.Assignment.ValueString()
		assignment := datadogV2.ESCALATIONPOLICYCREATEREQUESTDATAATTRIBUTESSTEPSITEMSASSIGNMENT_DEFAULT
		switch plannedAssignment {
		case "default":
			// default already set
		case "":
			// default already set
		case "round-robin":
			assignment = datadogV2.ESCALATIONPOLICYCREATEREQUESTDATAATTRIBUTESSTEPSITEMSASSIGNMENT_DEFAULT
		default:
			diags.AddError("assignment is invalid", "assignment must be either 'round-robin' or 'default'")
		}

		step.SetAssignment(assignment)

		if !plannedStep.EscalateAfter.IsNull() {
			step.SetEscalateAfterSeconds(plannedStep.EscalateAfter.ValueInt64())
		}

		var targets []datadogV2.TeamRulesCreateRequestDataAttributesStepsItemsTargetsItems

		for _, plannedTarget := range plannedStep.Targets {
			assignedFields := 0
			if !plannedTarget.User.IsNull() {
				targets = append(targets, datadogV2.TeamRulesCreateRequestDataAttributesStepsItemsTargetsItems{
					Id:   ptrValue(plannedTarget.User.ValueString()),
					Type: ptrValue(datadogV2.ESCALATIONPOLICYCREATEREQUESTDATAATTRIBUTESSTEPSITEMSTARGETSITEMSTYPE_USERS),
				})
				assignedFields += 1
			}

			if !plannedTarget.Schedule.IsNull() {
				targets = append(targets, datadogV2.TeamRulesCreateRequestDataAttributesStepsItemsTargetsItems{
					Id:   ptrValue(plannedTarget.Schedule.ValueString()),
					Type: ptrValue(datadogV2.ESCALATIONPOLICYCREATEREQUESTDATAATTRIBUTESSTEPSITEMSTARGETSITEMSTYPE_SCHEDULES),
				})
				assignedFields += 1
			}

			if !plannedTarget.Team.IsNull() {
				targets = append(targets, datadogV2.TeamRulesCreateRequestDataAttributesStepsItemsTargetsItems{
					Id:   ptrValue(plannedTarget.Team.ValueString()),
					Type: ptrValue(datadogV2.ESCALATIONPOLICYCREATEREQUESTDATAATTRIBUTESSTEPSITEMSTARGETSITEMSTYPE_TEAMS),
				})
				assignedFields += 1
			}

			if assignedFields != 1 {
				diags.AddError("invalid target", "target must specify one of `user`, `schedule` or `team`")
				return nil, diags
			}
		}

		step.SetTargets(targets)
		steps = append(steps, *step)

	}
	attributes.SetSteps(steps)

	req = datadogV2.NewTeamRulesCreateRequest(
		datadogV2.TeamRulesCreateRequestData{
			Type:          datadogV2.ESCALATIONPOLICYCREATEREQUESTDATATYPE_POLICIES,
			Attributes:    *attributes,
			Relationships: relationships,
		},
	)

	return req, diags
}

func stringValue(str string) *string {
	return &str
}

func ptrValue[T any](str T) *T {
	return &str
}

func (r *onCallTeamRulesResource) buildOnCallTeamRulesUpdateRequestBody(
	ctx context.Context,
	plan *onCallTeamRulesModel,
) (*datadogV2.TeamRulesUpdateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	req := &datadogV2.TeamRulesUpdateRequest{}
	attributes := datadogV2.NewTeamRulesUpdateRequestDataAttributesWithDefaults()
	var teams []string
	diags.Append(plan.Teams.ElementsAs(ctx, &teams, false)...)

	var relationships *datadogV2.TeamRulesUpdateRequestDataRelationships

	teamRelationships := make([]datadogV2.TeamRulesUpdateRequestDataRelationshipsTeamsDataItems, len(teams))

	for t, teamId := range teams {
		item := datadogV2.NewTeamRulesUpdateRequestDataRelationshipsTeamsDataItemsWithDefaults()
		item.SetId(teamId)
		teamRelationships[t] = *item
	}

	if len(teamRelationships) > 0 {
		relationships = &datadogV2.TeamRulesUpdateRequestDataRelationships{
			Teams: &datadogV2.TeamRulesUpdateRequestDataRelationshipsTeams{
				Data: teamRelationships,
			},
		}
	}

	if plan.ID.IsNull() {
		diags.AddError("id is required", "id is required")
		return nil, diags
	}

	if !plan.Name.IsNull() {
		attributes.SetName(plan.Name.ValueString())
	}

	if !plan.ResolvePageOnPolicyEnd.IsNull() {
		attributes.SetResolvePageOnPolicyEnd(plan.ResolvePageOnPolicyEnd.ValueBool())
	}

	if plan.Steps != nil {
		var steps []datadogV2.TeamRulesUpdateRequestDataAttributesStepsItems
		for _, plannedStep := range plan.Steps {
			step := datadogV2.NewTeamRulesUpdateRequestDataAttributesStepsItemsWithDefaults()

			plannedAssignment := plannedStep.Assignment.ValueString()
			assignment := datadogV2.ESCALATIONPOLICYUPDATEREQUESTDATAATTRIBUTESSTEPSITEMSASSIGNMENT_DEFAULT
			switch plannedAssignment {
			case "default":
				// default already set
			case "":
				// default already set
			case "round-robin":
				assignment = datadogV2.ESCALATIONPOLICYUPDATEREQUESTDATAATTRIBUTESSTEPSITEMSASSIGNMENT_DEFAULT
			default:
				diags.AddError("assignment is invalid", "assignment must be either 'round-robin' or 'default'")
			}

			step.SetAssignment(assignment)

			if !plannedStep.EscalateAfter.IsNull() {
				step.SetEscalateAfterSeconds(plannedStep.EscalateAfter.ValueInt64())
			}

			var targets []datadogV2.TeamRulesUpdateRequestDataAttributesStepsItemsTargetsItems

			for _, plannedTarget := range plannedStep.Targets {
				assignedFields := 0
				if !plannedTarget.User.IsNull() {
					targets = append(targets, datadogV2.TeamRulesUpdateRequestDataAttributesStepsItemsTargetsItems{
						Id:   ptrValue(plannedTarget.User.ValueString()),
						Type: ptrValue(datadogV2.ESCALATIONPOLICYUPDATEREQUESTDATAATTRIBUTESSTEPSITEMSTARGETSITEMSTYPE_USERS),
					})
					assignedFields += 1
				}

				if !plannedTarget.Schedule.IsNull() {
					targets = append(targets, datadogV2.TeamRulesUpdateRequestDataAttributesStepsItemsTargetsItems{
						Id:   ptrValue(plannedTarget.Schedule.ValueString()),
						Type: ptrValue(datadogV2.ESCALATIONPOLICYUPDATEREQUESTDATAATTRIBUTESSTEPSITEMSTARGETSITEMSTYPE_SCHEDULES),
					})
					assignedFields += 1
				}

				if !plannedTarget.Team.IsNull() {
					targets = append(targets, datadogV2.TeamRulesUpdateRequestDataAttributesStepsItemsTargetsItems{
						Id:   ptrValue(plannedTarget.Team.ValueString()),
						Type: ptrValue(datadogV2.ESCALATIONPOLICYUPDATEREQUESTDATAATTRIBUTESSTEPSITEMSTARGETSITEMSTYPE_TEAMS),
					})
					assignedFields += 1
				}

				if assignedFields != 1 {
					diags.AddError("invalid target", "target must specify one of `user`, `schedule` or `team`")
					return nil, diags
				}
			}

			step.SetTargets(targets)
			steps = append(steps, *step)
		}

		attributes.SetSteps(steps)
	}

	req = datadogV2.NewTeamRulesUpdateRequest(
		datadogV2.TeamRulesUpdateRequestData{
			Id:            plan.ID.ValueString(),
			Type:          datadogV2.ESCALATIONPOLICYUPDATEREQUESTDATATYPE_POLICIES,
			Attributes:    *attributes,
			Relationships: relationships,
		},
	)

	return req, diags
}
