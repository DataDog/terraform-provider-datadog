package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &onCallEscalationPolicyResource{}
	_ resource.ResourceWithImportState = &onCallEscalationPolicyResource{}
)

type onCallEscalationPolicyResource struct {
	Api  *datadogV2.OnCallApi
	Auth context.Context
}

type onCallEscalationPolicyModel struct {
	ID                     types.String           `tfsdk:"id"`
	Name                   types.String           `tfsdk:"name"`
	Retries                types.Int64            `tfsdk:"retries"`
	Teams                  types.List             `tfsdk:"teams"`
	Steps                  []*escalationStepModel `tfsdk:"step"`
	ResolvePageOnPolicyEnd types.Bool             `tfsdk:"resolve_page_on_policy_end"`
}

type escalationStepModel struct {
	Id            types.String             `tfsdk:"id"`
	EscalateAfter types.Int64              `tfsdk:"escalate_after_seconds"`
	Assignment    types.String             `tfsdk:"assignment"`
	Targets       []*escalationTargetModel `tfsdk:"target"`
}
type escalationTargetModel struct {
	Schedule         types.String `tfsdk:"schedule"`
	SchedulePosition types.String `tfsdk:"position"`
	Team             types.String `tfsdk:"team"`
	User             types.String `tfsdk:"user"`
}

func NewOnCallEscalationPolicyResource() resource.Resource {
	return &onCallEscalationPolicyResource{}
}

func (r *onCallEscalationPolicyResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetOnCallApiV2()
	r.Auth = providerData.Auth
}

func (r *onCallEscalationPolicyResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "on_call_escalation_policy"
}

func (r *onCallEscalationPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog On-Call escalation policy resource. This can be used to create and manage Datadog On-Call escalation policies.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "A human-readable name for the escalation policy.",
			},
			"retries": schema.Int64Attribute{
				Optional:    true,
				Required:    false,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Description: "If set, policy will be retried this many times after the final step. Must be in the range 0-10.",
				Validators:  []validator.Int64{int64validator.Between(0, 10)},
			},
			"resolve_page_on_policy_end": schema.BoolAttribute{
				Optional:    true,
				Required:    false,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "If true, pages will be automatically resolved if unacknowledged after the final step.",
			},
			"teams": schema.ListAttribute{
				Description: "A list of team ids associated with the escalation policy.",
				Optional:    true,
				Required:    false,
				Computed:    true,
				ElementType: types.StringType,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"id": utils.ResourceIDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"step": schema.ListNestedBlock{
				Description: "List of steps for the escalation policy.",
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeBetween(1, 10),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of this step.",
						},
						"assignment": schema.StringAttribute{
							Required:    false,
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString("default"),
							Description: "Specifies how this escalation step will assign targets. Can be `default` (page all targets at once) or `round-robin`.",
							Validators:  []validator.String{stringvalidator.OneOf("default", "round-robin")},
						},
						"escalate_after_seconds": schema.Int64Attribute{
							Required:    true,
							Description: "Defines how many seconds to wait before escalating to the next step.",
							Validators:  []validator.Int64{int64validator.Between(60, 36000)},
						},
					},
					Blocks: map[string]schema.Block{
						"target": schema.ListNestedBlock{
							Description: "List of targets for the step.",
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeBetween(1, 10),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"user": schema.StringAttribute{
										Optional:    true,
										Description: "Targeted user ID.",
									},
									"schedule": schema.StringAttribute{
										Optional:    true,
										Description: "Targeted schedule ID.",
									},
									"position": schema.StringAttribute{
										Optional:    true,
										Description: "For schedule targets, specifies which on-call user to page. Valid values: `current` (default), `previous`, `next`.",
										Validators:  []validator.String{stringvalidator.OneOf("current", "previous", "next")},
									},
									"team": schema.StringAttribute{
										Optional:    true,
										Description: "Targeted team ID.",
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

func (r *onCallEscalationPolicyResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *onCallEscalationPolicyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state onCallEscalationPolicyModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	id := state.ID.ValueString()

	include := "steps.targets"
	resp, httpResp, err := r.Api.GetOnCallEscalationPolicy(r.Auth, id, datadogV2.GetOnCallEscalationPolicyOptionalParameters{
		Include: &include,
	})
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving OnCallEscalationPolicy"))
		return
	}

	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	state = r.stateFromResponse(&resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *onCallEscalationPolicyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan onCallEscalationPolicyModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildOnCallEscalationPolicyRequestBody(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	include := "steps.targets"
	resp, _, err := r.Api.CreateOnCallEscalationPolicy(r.Auth, *body, datadogV2.CreateOnCallEscalationPolicyOptionalParameters{
		Include: &include,
	})
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating OnCallEscalationPolicy"))
		return
	}

	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}
	stepsById := map[string]*datadogV2.EscalationPolicyStep{}

	for _, item := range resp.GetIncluded() {
		if item.EscalationPolicyStep != nil && item.EscalationPolicyStep.Id != nil {
			stepsById[*item.EscalationPolicyStep.Id] = item.EscalationPolicyStep
		}
	}

	state := r.stateFromResponse(&resp)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *onCallEscalationPolicyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan onCallEscalationPolicyModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := plan.ID.ValueString()

	if id == "" {
		response.Diagnostics.AddError("id is required", "id is required")
		return
	}

	body, diags := r.buildOnCallEscalationPolicyUpdateRequestBody(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	include := "steps.targets"
	resp, _, err := r.Api.UpdateOnCallEscalationPolicy(r.Auth, id, *body, datadogV2.UpdateOnCallEscalationPolicyOptionalParameters{
		Include: &include,
	})
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating OnCallEscalationPolicy"))
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

func (r *onCallEscalationPolicyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state onCallEscalationPolicyModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	if id == "" {
		response.Diagnostics.AddError("id is required", "id is required")
		return
	}

	httpResp, err := r.Api.DeleteOnCallEscalationPolicy(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting on_call_schedule"))
		return
	}
}

func (r *onCallEscalationPolicyResource) stateFromResponse(resp *datadogV2.EscalationPolicy) onCallEscalationPolicyModel {
	state := onCallEscalationPolicyModel{}
	state.ID = types.StringValue(resp.Data.GetId())

	data := resp.GetData()
	attributes := data.GetAttributes()

	if name, ok := attributes.GetNameOk(); ok {
		state.Name = types.StringValue(*name)
	}

	if retries, ok := attributes.GetRetriesOk(); ok {
		state.Retries = types.Int64Value(*retries)
	}

	if resolveOnEnd, ok := attributes.GetResolvePageOnPolicyEndOk(); ok {
		state.ResolvePageOnPolicyEnd = types.BoolValue(*resolveOnEnd)
	}

	teams := []attr.Value{}
	for _, team := range data.GetRelationships().Teams.GetData() {
		teams = append(teams, types.StringValue(team.GetId()))
	}

	state.Teams, _ = types.ListValue(types.StringType, teams)

	stepsById := map[string]*datadogV2.EscalationPolicyStep{}

	for _, item := range resp.GetIncluded() {
		if item.EscalationPolicyStep != nil && item.EscalationPolicyStep.Id != nil {
			stepsById[*item.EscalationPolicyStep.Id] = item.EscalationPolicyStep
		}
	}
	// Build map of configured schedules from included
	configuredSchedulesById := map[string]*datadogV2.ConfiguredSchedule{}
	for _, item := range resp.GetIncluded() {
		if item.ConfiguredSchedule != nil {
			configuredSchedulesById[item.ConfiguredSchedule.GetId()] = item.ConfiguredSchedule
		}
	}

	for _, step := range resp.Data.Relationships.Steps.Data {
		fullStep := stepsById[step.Id]
		targets := []*escalationTargetModel{}
		for _, target := range fullStep.Relationships.Targets.Data {
			stateTarget := &escalationTargetModel{}
			if target.UserTarget != nil {
				stateTarget.User = types.StringValue(target.UserTarget.Id)
			} else if target.ScheduleTarget != nil {
				stateTarget.Schedule = types.StringValue(target.ScheduleTarget.Id)
			} else if target.ConfiguredScheduleTarget != nil {
				// Lookup the full configured schedule from included
				if configSched, ok := configuredSchedulesById[target.ConfiguredScheduleTarget.Id]; ok {
					stateTarget.Schedule = types.StringValue(configSched.Relationships.Schedule.Data.Id)
					if pos, ok := configSched.Attributes.GetPositionOk(); ok {
						stateTarget.SchedulePosition = types.StringValue(string(*pos))
					}
				}
			} else if target.TeamTarget != nil {
				stateTarget.Team = types.StringValue(target.TeamTarget.Id)
			}

			targets = append(targets, stateTarget)
		}
		state.Steps = append(state.Steps, &escalationStepModel{
			Id:            types.StringValue(step.Id),
			EscalateAfter: types.Int64PointerValue(fullStep.Attributes.EscalateAfterSeconds),
			Assignment:    types.StringPointerValue((*string)(fullStep.Attributes.Assignment)),
			Targets:       targets,
		})
	}

	return state
}

func (r *onCallEscalationPolicyResource) buildOnCallEscalationPolicyRequestBody(ctx context.Context, state *onCallEscalationPolicyModel) (*datadogV2.EscalationPolicyCreateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	req := &datadogV2.EscalationPolicyCreateRequest{}

	var teams []string
	diags.Append(state.Teams.ElementsAs(ctx, &teams, false)...)

	var relationships *datadogV2.EscalationPolicyCreateRequestDataRelationships

	teamRelationships := make([]datadogV2.DataRelationshipsTeamsDataItems, len(teams))

	for t, teamId := range teams {
		item := datadogV2.NewDataRelationshipsTeamsDataItemsWithDefaults()
		item.SetId(teamId)
		teamRelationships[t] = *item
	}

	if len(teamRelationships) > 0 {
		relationships = &datadogV2.EscalationPolicyCreateRequestDataRelationships{
			Teams: &datadogV2.DataRelationshipsTeams{
				Data: teamRelationships,
			},
		}
	}

	attributes := datadogV2.NewEscalationPolicyCreateRequestDataAttributesWithDefaults()

	attributes.SetName(state.Name.ValueString())

	if !state.Retries.IsNull() {
		attributes.SetRetries(state.Retries.ValueInt64())
	}

	if !state.ResolvePageOnPolicyEnd.IsNull() {
		attributes.SetResolvePageOnPolicyEnd(state.ResolvePageOnPolicyEnd.ValueBool())
	}

	var steps []datadogV2.EscalationPolicyCreateRequestDataAttributesStepsItems
	for _, plannedStep := range state.Steps {
		step := datadogV2.NewEscalationPolicyCreateRequestDataAttributesStepsItemsWithDefaults()

		plannedAssignment := plannedStep.Assignment.ValueString()
		assignment := datadogV2.ESCALATIONPOLICYSTEPATTRIBUTESASSIGNMENT_DEFAULT
		switch plannedAssignment {
		case "default":
			// default already set
		case "":
			// default already set
		case "round-robin":
			assignment = datadogV2.ESCALATIONPOLICYSTEPATTRIBUTESASSIGNMENT_ROUND_ROBIN
		default:
			diags.AddError("assignment is invalid", "assignment must be either 'round-robin' or 'default'")
		}

		step.SetAssignment(assignment)

		if !plannedStep.EscalateAfter.IsNull() {
			step.SetEscalateAfterSeconds(plannedStep.EscalateAfter.ValueInt64())
		}

		targets := buildTargetsFromPlan(plannedStep.Targets, &diags)
		if diags.HasError() {
			return nil, diags
		}
		step.SetTargets(targets)
		steps = append(steps, *step)

	}
	attributes.SetSteps(steps)

	req = datadogV2.NewEscalationPolicyCreateRequest(
		datadogV2.EscalationPolicyCreateRequestData{
			Type:          datadogV2.ESCALATIONPOLICYCREATEREQUESTDATATYPE_POLICIES,
			Attributes:    *attributes,
			Relationships: relationships,
		},
	)

	return req, diags
}

func (r *onCallEscalationPolicyResource) buildOnCallEscalationPolicyUpdateRequestBody(
	ctx context.Context,
	plan *onCallEscalationPolicyModel,
) (*datadogV2.EscalationPolicyUpdateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	req := &datadogV2.EscalationPolicyUpdateRequest{}
	attributes := datadogV2.NewEscalationPolicyUpdateRequestDataAttributesWithDefaults()
	var teams []string
	diags.Append(plan.Teams.ElementsAs(ctx, &teams, false)...)

	var relationships *datadogV2.EscalationPolicyUpdateRequestDataRelationships

	teamRelationships := make([]datadogV2.DataRelationshipsTeamsDataItems, len(teams))

	for t, teamId := range teams {
		item := datadogV2.NewDataRelationshipsTeamsDataItemsWithDefaults()
		item.SetId(teamId)
		teamRelationships[t] = *item
	}

	if len(teamRelationships) > 0 {
		relationships = &datadogV2.EscalationPolicyUpdateRequestDataRelationships{
			Teams: &datadogV2.DataRelationshipsTeams{
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

	if !plan.Retries.IsNull() {
		attributes.SetRetries(plan.Retries.ValueInt64())
	}

	if !plan.ResolvePageOnPolicyEnd.IsNull() {
		attributes.SetResolvePageOnPolicyEnd(plan.ResolvePageOnPolicyEnd.ValueBool())
	}

	if plan.Steps != nil {
		var steps []datadogV2.EscalationPolicyUpdateRequestDataAttributesStepsItems
		for _, plannedStep := range plan.Steps {
			step := datadogV2.NewEscalationPolicyUpdateRequestDataAttributesStepsItemsWithDefaults()

			plannedAssignment := plannedStep.Assignment.ValueString()
			assignment := datadogV2.ESCALATIONPOLICYSTEPATTRIBUTESASSIGNMENT_DEFAULT
			switch plannedAssignment {
			case "default":
				// default already set
			case "":
				// default already set
			case "round-robin":
				assignment = datadogV2.ESCALATIONPOLICYSTEPATTRIBUTESASSIGNMENT_ROUND_ROBIN
			default:
				diags.AddError("assignment is invalid", "assignment must be either 'round-robin' or 'default'")
			}

			step.SetAssignment(assignment)

			if !plannedStep.EscalateAfter.IsNull() {
				step.SetEscalateAfterSeconds(plannedStep.EscalateAfter.ValueInt64())
			}

			targets := buildTargetsFromPlan(plannedStep.Targets, &diags)
			if diags.HasError() {
				return nil, diags
			}
			step.SetTargets(targets)
			steps = append(steps, *step)
		}

		attributes.SetSteps(steps)
	}

	req = datadogV2.NewEscalationPolicyUpdateRequest(
		datadogV2.EscalationPolicyUpdateRequestData{
			Id:            plan.ID.ValueString(),
			Type:          datadogV2.ESCALATIONPOLICYUPDATEREQUESTDATATYPE_POLICIES,
			Attributes:    *attributes,
			Relationships: relationships,
		},
	)

	return req, diags
}

func ptrValue[T any](t T) *T {
	return &t
}

func createTarget(plannedTarget *escalationTargetModel, diags *diag.Diagnostics) *datadogV2.EscalationPolicyStepTarget {
	// Validate position is only set with schedule
	if !plannedTarget.SchedulePosition.IsNull() && plannedTarget.Schedule.IsNull() {
		diags.AddError("invalid target", "`position` can only be set when `schedule` is specified")
		return nil
	}

	var subTargets []datadogV2.EscalationPolicyStepTarget

	if !plannedTarget.User.IsNull() {
		subTargets = append(subTargets, datadogV2.EscalationPolicyStepTarget{
			Id:   plannedTarget.User.ValueStringPointer(),
			Type: ptrValue(datadogV2.ESCALATIONPOLICYSTEPTARGETTYPE_USERS),
		})
	}

	if !plannedTarget.Schedule.IsNull() {
		target := datadogV2.EscalationPolicyStepTarget{
			Id:   plannedTarget.Schedule.ValueStringPointer(),
			Type: ptrValue(datadogV2.ESCALATIONPOLICYSTEPTARGETTYPE_SCHEDULES),
		}
		if !plannedTarget.SchedulePosition.IsNull() {
			position := datadogV2.ScheduleTargetPosition(plannedTarget.SchedulePosition.ValueString())
			target.Config = &datadogV2.EscalationPolicyStepTargetConfig{
				Schedule: &datadogV2.EscalationPolicyStepTargetConfigSchedule{
					Position: &position,
				},
			}
		}
		subTargets = append(subTargets, target)
	}

	if !plannedTarget.Team.IsNull() {
		subTargets = append(subTargets, datadogV2.EscalationPolicyStepTarget{
			Id:   plannedTarget.Team.ValueStringPointer(),
			Type: ptrValue(datadogV2.ESCALATIONPOLICYSTEPTARGETTYPE_TEAMS),
		})
	}

	if len(subTargets) != 1 {
		diags.AddError("invalid target", "target must specify exactly one of `user`, `schedule` or `team`")
		return nil
	}

	return &subTargets[0]
}

func buildTargetsFromPlan(plannedTargets []*escalationTargetModel, diags *diag.Diagnostics) []datadogV2.EscalationPolicyStepTarget {
	var targets []datadogV2.EscalationPolicyStepTarget

	for _, plannedTarget := range plannedTargets {
		target := createTarget(plannedTarget, diags)
		if diags.HasError() {
			return nil
		}
		targets = append(targets, *target)
	}

	return targets
}
