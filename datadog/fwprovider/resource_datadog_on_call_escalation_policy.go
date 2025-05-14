package fwprovider

import (
	"context"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/customtypes"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var (
	_ resource.ResourceWithConfigure   = &onCallEscalationPolicyResource{}
	_ resource.ResourceWithImportState = &onCallEscalationPolicyResource{}
)

type onCallEscalationPolicyResource struct {
	Api  *datadogV2.OnCallApi
	Auth context.Context
}

//type EscalationPolicy struct {
//	ID                     string                       `jsonapi:"primary,policies" json:"id"`
//	Name                   string                       `jsonapi:"attribute" json:"name" validate:"required,min=1"`
//	Description            string                       `jsonapi:"attribute" json:"description"`
//	Teams                  []oncallcommon.TeamReference `jsonapi:"relationship" json:"teams"`
//	Retries                int                          `jsonapi:"attribute" json:"retries" validate:"gte=0,lte=10"`
//	Steps                  []EscalationPolicyStep       `jsonapi:"relationship" json:"steps" validate:"required,min=1,max=10,dive"`
//	ResolvePageOnPolicyEnd bool                         `jsonapi:"attribute" json:"resolve_page_on_policy_end"`
//}
//
//type EscalationPolicyStep struct {
//	ID            string                    `jsonapi:"primary,steps" json:"id"`
//	EscalateAfter schedules.DurationSeconds `jsonapi:"attribute" json:"escalate_after_seconds" validate:"gte=60,lte=3600"`
//	Assignment    AssignmentType            `jsonapi:"attribute" json:"assignment" validate:"omitempty,oneof=default round-robin" openapi:"enum=default|round-robin"`
//	Targets       []EscalationTarget        `jsonapi:"relationship" json:"targets" jsonschema:"EscalationTarget"`
//}

type onCallEscalationPolicyModel struct {
	ID                     types.String           `tfsdk:"id"`
	Name                   types.String           `tfsdk:"name"`
	Retries                types.Int64            `tfsdk:"retries"`
	Teams                  types.List             `tfsdk:"teams"`
	Steps                  []*escalationStepModel `tfsdk:"steps"`
	ResolvePageOnPolicyEnd types.Bool             `tfsdk:"resolve_page_on_policy_end"`
}

type escalationStepModel struct {
	Id            types.String             `tfsdk:"id"`
	EscalateAfter types.Int64              `tfsdk:"escalate_after_seconds"`
	Assignment    types.String             `tfsdk:"assignment"`
	Targets       []*escalationTargetModel `tfsdk:"targets"`
}
type escalationTargetModel struct {
	Schedule types.String `tfsdk:"schedule"`
	Team     types.String `tfsdk:"schedule"`
	User     types.String `tfsdk:"user"`
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
			"retries": schema.StringAttribute{
				Optional:    true,
				Required:    false,
				Description: "If set, policy will be retried this many times after the final step. Must be in the range 0-10.",
			},
			"resolve_page_on_policy_end": schema.StringAttribute{
				Optional:    true,
				Required:    false,
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
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of this step.",
						},
						"assignment": schema.StringAttribute{
							Required:    true,
							Description: "Specifies how this escalation step will assign targets. Can be `default` (page all targets at once) or `round-robin`.",
						},
						"escalate_after_seconds": schema.Int32Attribute{
							Optional:    true,
							Required:    false,
							Description: "Defines how many seconds to wait before escalating to the next step.",
							Validators:  []validator.Int32{int32validator.Between(0, 36000)},
						},
					},
					Blocks: map[string]schema.Block{
						"target": schema.ListNestedBlock{
							Description: "List of targets for the step.",
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

	state = *r.newState(ctx, &state, &resp)

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
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	state := r.newState(ctx, &plan, &resp)

	// Save data into Terraform state
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
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	state := r.newState(ctx, &plan, &resp)

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

func (r *onCallEscalationPolicyResource) newState(ctx context.Context, plan *onCallEscalationPolicyModel, resp *datadogV2.EscalationPolicy) *onCallEscalationPolicyModel {
	state := &onCallEscalationPolicyModel{}
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

	membersByID := make(map[string]*datadogV2.EscalationPolicyMember)
	usersByID := make(map[string]*datadogV2.EscalationPolicyUser)
	layersByID := make(map[string]*datadogV2.Layer)

	// Update steps with their IDs from the API response
	included := resp.GetIncluded()
	for _, item := range included {
		item.EscalationPolicyStep != nil{}
		if item.EscalationPolicyMember != nil {
			membersByID[item.EscalationPolicyMember.GetId()] = item.EscalationPolicyMember
		}
		if item.EscalationPolicyUser != nil {
			usersByID[item.EscalationPolicyUser.GetId()] = item.EscalationPolicyUser
		}
		if item.Layer != nil {
			layersByID[item.Layer.GetId()] = item.Layer
		}
	}

	state.Layers = make([]*escalationStepModel, len(data.Relationships.Layers.Data))
	// We use the index to match the layer with the plan.Layers
	// As we expect the server to return the layers in the same order as the plan.Layers
	for i, layer := range data.Relationships.Layers.Data {
		var layerExistingEffectiveDate customtypes.BackwardRFC3339Date
		if i < len(plan.Layers) {
			layerExistingEffectiveDate = customtypes.NewBackwardRFC3339Date(plan.Layers[i].EffectiveDate.ValueString())
		}
		state.Layers[i] = newLayerModel(layersByID[layer.GetId()], membersByID, layerExistingEffectiveDate)
	}
	return state
}

func newLayerModel(layer *datadogV2.Layer, membersByID map[string]*datadogV2.EscalationPolicyMember, layerExistingEffectiveDate customtypes.BackwardRFC3339Date) *escalationStepModel {
	membersData := layer.GetRelationships().Members.GetData()
	memberIds := make([]types.String, len(membersData))
	for j, member := range membersData {
		includedMember := membersByID[member.GetId()]
		userId := includedMember.GetRelationships().User.GetData().Id
		if userId != "" {
			memberIds[j] = types.StringValue(userId)
		} else {
			memberIds[j] = types.StringNull()
		}
	}
	restrictions := layer.GetAttributes().Restrictions
	restrictionsModels := make([]*restrictionsModel, len(restrictions))
	for j, restriction := range restrictions {
		restrictionsModels[j] = &restrictionsModel{
			EndDay:    types.StringValue(string(restriction.GetEndDay())),
			EndTime:   types.StringValue(restriction.GetEndTime()),
			StartDay:  types.StringValue(string(restriction.GetStartDay())),
			StartTime: types.StringValue(restriction.GetStartTime()),
		}
	}
	interval := layer.GetAttributes().Interval

	var endDateStringValue types.String
	endDate := layer.GetAttributes().EndDate
	if endDate != nil && !endDate.IsZero() {
		endDateStringValue = types.StringValue(formatTime(*endDate))
	} else {
		endDateStringValue = types.StringNull()
	}

	appliedEffectiveDate := customtypes.NewBackwardRFC3339Date(formatTime(layer.Attributes.GetEffectiveDate()))

	// If the effective date is not set, use the applied effective date
	// Otherwise, use the existing effective date
	// The effective date is irrelevant to the state (applied_effective_date is), we just need to be sure to make it
	// consistent for Terraform.
	var effectiveDate customtypes.BackwardRFC3339Date
	if layerExistingEffectiveDate.ValueString() != "" {
		effectiveDate = customtypes.NewBackwardRFC3339Date(layerExistingEffectiveDate.ValueString())
	} else {
		effectiveDate = appliedEffectiveDate
	}

	return &escalationStepModel{
		Id:                   types.StringValue(layer.GetId()),
		AppliedEffectiveDate: appliedEffectiveDate,
		EffectiveDate:        effectiveDate,
		EndDate:              endDateStringValue,
		Name:                 types.StringValue(layer.Attributes.GetName()),
		RotationStart:        types.StringValue(formatTime(layer.Attributes.GetRotationStart())),
		Users:                memberIds,
		Restrictions:         restrictionsModels,
		Interval:             &intervalModel{Days: types.Int32Value(int32(interval.GetDays())), Seconds: types.Int64Value(interval.GetSeconds())},
	}
}

func (r *onCallEscalationPolicyResource) buildOnCallEscalationPolicyRequestBody(ctx context.Context, state *onCallEscalationPolicyModel) (*datadogV2.EscalationPolicyCreateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	req := &datadogV2.EscalationPolicyCreateRequest{}

	var teams []string
	diags.Append(state.Teams.ElementsAs(ctx, &teams, false)...)

	relationships := buildCreateRelationships(teams)

	attributes := datadogV2.NewEscalationPolicyCreateRequestDataAttributesWithDefaults()

	attributes.SetName(state.Name.ValueString())
	if !state.TimeZone.IsNull() {
		attributes.SetTimeZone(state.TimeZone.ValueString())
	}

	var layers []datadogV2.EscalationPolicyCreateRequestDataAttributesLayersItems
	for _, layersTFItem := range state.Layers {
		effectiveDate, err := parseTime(layersTFItem.EffectiveDate.ValueString())
		if err != nil {
			diags.AddError("error parsing effective_date", err.Error())
			return nil, diags
		}

		rotationStart, err := parseTime(layersTFItem.RotationStart.ValueString())
		if err != nil {
			diags.AddError("error parsing rotation_start", err.Error())
			return nil, diags
		}

		layersDDItem := datadogV2.NewEscalationPolicyCreateRequestDataAttributesLayersItems(
			effectiveDate,
			datadogV2.EscalationPolicyCreateRequestDataAttributesLayersItemsInterval{
				Days:    layersTFItem.Interval.Days.ValueInt32Pointer(),
				Seconds: layersTFItem.Interval.Seconds.ValueInt64Pointer(),
			},
			[]datadogV2.EscalationPolicyCreateRequestDataAttributesLayersItemsMembersItems{},
			layersTFItem.Name.ValueString(),
			rotationStart,
		)

		if !layersTFItem.EndDate.IsNull() {
			endDate, err := parseTime(layersTFItem.EndDate.ValueString())
			if err != nil {
				diags.AddError("error parsing end_date", err.Error())
				return nil, diags
			}
			layersDDItem.SetEndDate(endDate)
		}

		var members []datadogV2.EscalationPolicyCreateRequestDataAttributesLayersItemsMembersItems
		for _, memberId := range layersTFItem.Users {
			membersDDItem := datadogV2.NewEscalationPolicyCreateRequestDataAttributesLayersItemsMembersItems()

			if !memberId.IsNull() {
				userId := memberId.ValueString()
				if userId == "" {
					diags.AddError("user_id can't be empty, either set the user_id to a valid user, set the user_id to null or omit the field", "user_id can't be empty, either set the user_id to a valid user or to null")
					return nil, diags
				}
				membersDDItem.User = &datadogV2.EscalationPolicyCreateRequestDataAttributesLayersItemsMembersItemsUser{
					Id: &userId,
				}
			}

			members = append(members, *membersDDItem)
		}
		layersDDItem.SetMembers(members)

		if layersTFItem.Restrictions != nil {
			var restrictions []datadogV2.EscalationPolicyCreateRequestDataAttributesLayersItemsRestrictionsItems
			for _, restrictionsTFItem := range layersTFItem.Restrictions {
				restrictionsDDItem := datadogV2.NewEscalationPolicyCreateRequestDataAttributesLayersItemsRestrictionsItems()

				if !restrictionsTFItem.EndDay.IsNull() {
					restrictionsDDItem.SetEndDay(datadogV2.EscalationPolicyCreateRequestDataAttributesLayersItemsRestrictionsItemsEndDay(restrictionsTFItem.EndDay.ValueString()))
				}
				if !restrictionsTFItem.EndTime.IsNull() {
					restrictionsDDItem.SetEndTime(restrictionsTFItem.EndTime.ValueString())
				}
				if !restrictionsTFItem.StartDay.IsNull() {
					restrictionsDDItem.SetStartDay(datadogV2.EscalationPolicyCreateRequestDataAttributesLayersItemsRestrictionsItemsStartDay(restrictionsTFItem.StartDay.ValueString()))
				}
				if !restrictionsTFItem.StartTime.IsNull() {
					restrictionsDDItem.SetStartTime(restrictionsTFItem.StartTime.ValueString())
				}
				restrictions = append(restrictions, *restrictionsDDItem)
			}
			layersDDItem.SetRestrictions(restrictions)
		}

		layers = append(layers, *layersDDItem)
	}
	attributes.SetLayers(layers)

	req = datadogV2.NewEscalationPolicyCreateRequest(
		datadogV2.EscalationPolicyCreateRequestData{
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

func (r *onCallEscalationPolicyResource) buildOnCallEscalationPolicyUpdateRequestBody(
	ctx context.Context,
	plan *onCallEscalationPolicyModel,
) (*datadogV2.EscalationPolicyUpdateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	req := &datadogV2.EscalationPolicyUpdateRequest{}
	attributes := datadogV2.NewEscalationPolicyUpdateRequestDataAttributesWithDefaults()
	var teams []string
	diags.Append(plan.Teams.ElementsAs(ctx, &teams, false)...)
	relationships := buildUpdateRelationships(teams)

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
		var steps []datadogV2.EscalationPolicyUpdateRequestDataAttributesStepsItems
		for _, plannedStep := range plan.Steps {
			step := datadogV2.NewEscalationPolicyUpdateRequestDataAttributesStepsItemsWithDefaults()

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

			var targets []datadogV2.EscalationPolicyUpdateRequestDataAttributesStepsItemsTargetsItems

			for _, plannedTarget := range plannedStep.Targets {
				assignedFields := 0
				if !plannedTarget.User.IsNull() {
					targets = append(targets, datadogV2.EscalationPolicyUpdateRequestDataAttributesStepsItemsTargetsItems{
						Id:   ptrValue(plannedTarget.User.ValueString()),
						Type: ptrValue(datadogV2.ESCALATIONPOLICYUPDATEREQUESTDATAATTRIBUTESSTEPSITEMSTARGETSITEMSTYPE_USERS),
					})
					assignedFields += 1
				}

				if !plannedTarget.Schedule.IsNull() {
					targets = append(targets, datadogV2.EscalationPolicyUpdateRequestDataAttributesStepsItemsTargetsItems{
						Id:   ptrValue(plannedTarget.Schedule.ValueString()),
						Type: ptrValue(datadogV2.ESCALATIONPOLICYUPDATEREQUESTDATAATTRIBUTESSTEPSITEMSTARGETSITEMSTYPE_SCHEDULES),
					})
					assignedFields += 1
				}

				if !plannedTarget.Team.IsNull() {
					targets = append(targets, datadogV2.EscalationPolicyUpdateRequestDataAttributesStepsItemsTargetsItems{
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

func buildCreateRelationships(plannedTeams []string) *datadogV2.EscalationPolicyCreateRequestDataRelationships {
	var relationships *datadogV2.EscalationPolicyCreateRequestDataRelationships

	teamRelationships := make([]datadogV2.EscalationPolicyCreateRequestDataRelationshipsTeamsDataItems, len(plannedTeams))

	for t, teamId := range plannedTeams {
		item := datadogV2.NewEscalationPolicyCreateRequestDataRelationshipsTeamsDataItemsWithDefaults()
		item.SetId(teamId)
		teamRelationships[t] = *item
	}

	if len(teamRelationships) > 0 {
		relationships = &datadogV2.EscalationPolicyCreateRequestDataRelationships{
			Teams: &datadogV2.EscalationPolicyCreateRequestDataRelationshipsTeams{
				Data: teamRelationships,
			},
		}
	}
	return relationships
}

func buildUpdateRelationships(plannedTeams []string) *datadogV2.EscalationPolicyUpdateRequestDataRelationships {
	var relationships *datadogV2.EscalationPolicyUpdateRequestDataRelationships

	teamRelationships := make([]datadogV2.EscalationPolicyUpdateRequestDataRelationshipsTeamsDataItems, len(plannedTeams))

	for t, teamId := range plannedTeams {
		item := datadogV2.NewEscalationPolicyUpdateRequestDataRelationshipsTeamsDataItemsWithDefaults()
		item.SetId(teamId)
		teamRelationships[t] = *item
	}

	if len(teamRelationships) > 0 {
		relationships = &datadogV2.EscalationPolicyUpdateRequestDataRelationships{
			Teams: &datadogV2.EscalationPolicyUpdateRequestDataRelationshipsTeams{
				Data: teamRelationships,
			},
		}
	}
	return relationships
}
