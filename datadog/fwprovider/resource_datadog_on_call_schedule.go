package fwprovider

import (
	"context"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
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
	_ resource.ResourceWithConfigure   = &onCallScheduleResource{}
	_ resource.ResourceWithImportState = &onCallScheduleResource{}
)

type onCallScheduleResource struct {
	Api  *datadogV2.OnCallApi
	Auth context.Context
}

type onCallScheduleModel struct {
	ID       types.String       `tfsdk:"id"`
	Name     types.String       `tfsdk:"name"`
	TimeZone types.String       `tfsdk:"time_zone"`
	Teams    []*onCallTeamModel `tfsdk:"teams"`
	Tags     types.List         `tfsdk:"tags"`
	Layers   []*layersModel     `tfsdk:"layers"`
}

type onCallTeamModel struct {
	Id types.String `tfsdk:"id"`
}

type layersModel struct {
	Id                   types.String                    `tfsdk:"id"`
	EffectiveDate        customtypes.BackwardRFC3339Date `tfsdk:"effective_date"`
	AppliedEffectiveDate customtypes.BackwardRFC3339Date `tfsdk:"applied_effective_date"`
	EndDate              types.String                    `tfsdk:"end_date"`
	Name                 types.String                    `tfsdk:"name"`
	RotationStart        types.String                    `tfsdk:"rotation_start"`
	Members              []*membersModel                 `tfsdk:"members"`
	Restrictions         []*restrictionsModel            `tfsdk:"restrictions"`
	Interval             *intervalModel                  `tfsdk:"interval"`
}
type membersModel struct {
	User *userModel `tfsdk:"user"`
}
type userModel struct {
	Id types.String `tfsdk:"id"`
}

type restrictionsModel struct {
	EndDay    types.String `tfsdk:"end_day"`
	EndTime   types.String `tfsdk:"end_time"`
	StartDay  types.String `tfsdk:"start_day"`
	StartTime types.String `tfsdk:"start_time"`
}

type intervalModel struct {
	Days    types.Int32 `tfsdk:"days"`
	Seconds types.Int64 `tfsdk:"seconds"`
}

func NewOnCallScheduleResource() resource.Resource {
	return &onCallScheduleResource{}
}

func (r *onCallScheduleResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetOnCallApiV2()
	r.Auth = providerData.Auth
}

func (r *onCallScheduleResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "on_call_schedule"
}

func (r *onCallScheduleResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog OnCallSchedule resource. This can be used to create and manage Datadog on_call_schedule.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "A human-readable name for the new schedule.",
			},
			"time_zone": schema.StringAttribute{
				Required:    true,
				Description: "The time zone in which the schedule is defined.",
			},
			"tags": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				Description: "A list of tags for categorizing or filtering the schedule.",
				ElementType: types.StringType,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"id": utils.ResourceIDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"teams": schema.ListNestedBlock{
				Description: "A list of team ids associated with the schedule.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Optional:    true,
							Description: "The ID of the team.",
						},
					},
				},
			},
			"layers": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of this layer.",
						},
						"effective_date": schema.StringAttribute{
							CustomType:  customtypes.BackwardRFC3339DateType{},
							Required:    true,
							Description: "The date/time when this layer should become active (in ISO 8601).",
							Validators:  []validator.String{validators.TimeFormatValidator(time.RFC3339)},
						},
						"applied_effective_date": schema.StringAttribute{
							CustomType:  customtypes.BackwardRFC3339DateType{},
							Computed:    true,
							Description: "The date/time when this layer becomes active (in ISO 8601).",
							Validators:  []validator.String{validators.TimeFormatValidator(time.RFC3339)},
						},
						"end_date": schema.StringAttribute{
							Optional:    true,
							Description: "The date/time after which this layer no longer applies (in ISO 8601).",
							Validators:  []validator.String{validators.TimeFormatValidator(time.RFC3339)},
						},
						"name": schema.StringAttribute{
							Required:    true,
							Description: "The name of this layer.",
						},
						"rotation_start": schema.StringAttribute{
							Optional:    true,
							Description: "The date/time when the rotation for this layer starts (in ISO 8601).",
							Validators:  []validator.String{validators.TimeFormatValidator(time.RFC3339)},
						},
					},
					Blocks: map[string]schema.Block{
						"members": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"user": schema.SingleNestedBlock{
										Description: "The user assigned to this member. Can be omitted for empty members.",
										Attributes: map[string]schema.Attribute{
											"id": schema.StringAttribute{
												// Member can be empty, so we need it to allow optional
												Optional:    true,
												Description: "The user's ID. Can be omitted for empty members.",
											},
										},
									},
								},
							},
						},
						"restrictions": schema.ListNestedBlock{
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
						"interval": schema.SingleNestedBlock{
							Attributes: map[string]schema.Attribute{
								"days": schema.Int32Attribute{
									Optional:    true,
									Computed:    true,
									Description: "The number of full days in each rotation period.",
									Default:     int32default.StaticInt32(int32(0)),
								},
								"seconds": schema.Int64Attribute{
									Optional:    true,
									Computed:    true,
									Description: "For intervals that are not expressible in whole days, this will be added to `days`.",
									Default:     int64default.StaticInt64(int64(0)),
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *onCallScheduleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *onCallScheduleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state onCallScheduleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	id := state.ID.ValueString()

	include := "layers,layers.members,layers.members.user"
	resp, httpResp, err := r.Api.GetOnCallSchedule(r.Auth, id, datadogV2.GetOnCallScheduleOptionalParameters{
		Include: &include,
	})
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving OnCallSchedule"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	state = *r.newState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *onCallScheduleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan onCallScheduleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildOnCallScheduleRequestBody(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	include := "layers,layers.members,layers.members.user"
	resp, _, err := r.Api.CreateOnCallSchedule(r.Auth, *body, datadogV2.CreateOnCallScheduleOptionalParameters{
		Include: &include,
	})
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving OnCallSchedule"))
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

func (r *onCallScheduleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan onCallScheduleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := plan.ID.ValueString()

	if id == "" {
		response.Diagnostics.AddError("id is required", "id is required")
		return
	}

	// Create a map of existing layer names to their IDs
	// This is based on the assumption that the layer name is unique
	// The name is used as a proxy to identify the layer to be updated
	// If two layer names are the same, fail the update
	// we could also don't try to update the given layer but just recreate it on name collision
	var previousState onCallScheduleModel
	request.State.Get(ctx, &previousState)
	existingLayerIdByName := make(map[string]string)
	for _, layer := range previousState.Layers {
		if layer.Name.ValueString() == "" {
			response.Diagnostics.AddError("layer name is required", "layer name is required")
			return
		}
		if existingLayerIdByName[layer.Name.ValueString()] != "" {
			response.Diagnostics.AddError("layer name is not unique", "layer name is not unique")
			return
		}
		existingLayerIdByName[layer.Name.ValueString()] = layer.Id.ValueString()
	}

	body, diags := r.buildOnCallScheduleUpdateRequestBody(ctx, &plan, existingLayerIdByName)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	include := "layers,layers.members,layers.members.user"
	resp, _, err := r.Api.UpdateOnCallSchedule(r.Auth, id, *body, datadogV2.UpdateOnCallScheduleOptionalParameters{
		Include: &include,
	})
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving OnCallSchedule"))
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

func (r *onCallScheduleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state onCallScheduleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	if id == "" {
		response.Diagnostics.AddError("id is required", "id is required")
		return
	}

	httpResp, err := r.Api.DeleteOnCallSchedule(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting on_call_schedule"))
		return
	}
}

func (r *onCallScheduleResource) newState(ctx context.Context, plan *onCallScheduleModel, resp *datadogV2.Schedule) *onCallScheduleModel {
	state := &onCallScheduleModel{}
	state.ID = types.StringValue(resp.Data.GetId())

	// Update the layers with the response.
	// We could avoid this step, this is just a way to confirm our API is working correctly
	// returning the requested changes and to have the correct state.

	data := resp.GetData()
	attributes := data.GetAttributes()

	if name, ok := attributes.GetNameOk(); ok {
		state.Name = types.StringValue(*name)
	}

	if timeZone, ok := attributes.GetTimeZoneOk(); ok {
		state.TimeZone = types.StringValue(*timeZone)
	}

	state.Tags, _ = types.ListValueFrom(ctx, types.StringType, attributes.Tags)

	var teams []string
	for _, team := range data.GetRelationships().Teams.GetData() {
		teams = append(teams, team.GetId())
	}
	state.Teams = make([]*onCallTeamModel, len(teams))
	for i, team := range teams {
		state.Teams[i] = &onCallTeamModel{Id: types.StringValue(team)}
	}

	membersByID := make(map[string]*datadogV2.ScheduleMember)
	usersByID := make(map[string]*datadogV2.ScheduleUser)
	layersByID := make(map[string]*datadogV2.Layer)

	// Update layers with their IDs from the API response
	included := resp.GetIncluded()
	for _, item := range included {
		if item.ScheduleMember != nil {
			membersByID[item.ScheduleMember.GetId()] = item.ScheduleMember
		}
		if item.ScheduleUser != nil {
			usersByID[item.ScheduleUser.GetId()] = item.ScheduleUser
		}
		if item.Layer != nil {
			layersByID[item.Layer.GetId()] = item.Layer
		}
	}

	state.Layers = make([]*layersModel, len(data.Relationships.Layers.Data))
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

func newLayerModel(layer *datadogV2.Layer, membersByID map[string]*datadogV2.ScheduleMember, layerExistingEffectiveDate customtypes.BackwardRFC3339Date) *layersModel {
	membersData := layer.GetRelationships().Members.GetData()
	members := make([]*membersModel, len(membersData))
	for j, member := range membersData {
		includedMember := membersByID[member.GetId()]
		// Always create the user, we can't omit it/keep it nil
		// As empty block are not working on some old terraform versions
		// See https://github.com/hashicorp/terraform/pull/32463
		// Therefore `members{}` is not valid terraform, we require `members.user{}` [ie user with a null id]
		user := &userModel{Id: types.StringNull()}
		userId := includedMember.GetRelationships().User.GetData().Id
		if userId != "" {
			user = &userModel{Id: types.StringValue(userId)}
		}
		members[j] = &membersModel{User: user}
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

	return &layersModel{
		Id:                   types.StringValue(layer.GetId()),
		AppliedEffectiveDate: appliedEffectiveDate,
		EffectiveDate:        effectiveDate,
		EndDate:              endDateStringValue,
		Name:                 types.StringValue(layer.Attributes.GetName()),
		RotationStart:        types.StringValue(formatTime(layer.Attributes.GetRotationStart())),
		Members:              members,
		Restrictions:         restrictionsModels,
		Interval:             &intervalModel{Days: types.Int32Value(int32(interval.GetDays())), Seconds: types.Int64Value(interval.GetSeconds())},
	}
}

func formatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

func (r *onCallScheduleResource) buildOnCallScheduleRequestBody(ctx context.Context, state *onCallScheduleModel) (*datadogV2.ScheduleCreateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	req := &datadogV2.ScheduleCreateRequest{}

	relationships := buildCreateRelationships(state.Teams)

	attributes := datadogV2.NewScheduleCreateRequestDataAttributesWithDefaults()

	attributes.SetName(state.Name.ValueString())
	if !state.TimeZone.IsNull() {
		attributes.SetTimeZone(state.TimeZone.ValueString())
	}

	tags := make([]string, 0)
	if !state.Tags.IsNull() {
		diags.Append(state.Tags.ElementsAs(ctx, &tags, false)...)
	}
	attributes.SetTags(tags)

	var layers []datadogV2.ScheduleCreateRequestDataAttributesLayersItems
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

		layersDDItem := datadogV2.NewScheduleCreateRequestDataAttributesLayersItems(
			effectiveDate,
			datadogV2.ScheduleCreateRequestDataAttributesLayersItemsInterval{
				Days:    layersTFItem.Interval.Days.ValueInt32Pointer(),
				Seconds: layersTFItem.Interval.Seconds.ValueInt64Pointer(),
			},
			[]datadogV2.ScheduleCreateRequestDataAttributesLayersItemsMembersItems{},
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

		if layersTFItem.Members == nil {
			diags.AddError("members is required", "members is required")
			return nil, diags
		}
		var members []datadogV2.ScheduleCreateRequestDataAttributesLayersItemsMembersItems
		for _, membersTFItem := range layersTFItem.Members {
			membersDDItem := datadogV2.NewScheduleCreateRequestDataAttributesLayersItemsMembersItems()

			if !membersTFItem.User.Id.IsNull() {
				membersDDItem.User = &datadogV2.ScheduleCreateRequestDataAttributesLayersItemsMembersItemsUser{
					Id: membersTFItem.User.Id.ValueStringPointer(),
				}
			}

			members = append(members, *membersDDItem)
		}
		layersDDItem.SetMembers(members)

		if layersTFItem.Restrictions != nil {
			var restrictions []datadogV2.ScheduleCreateRequestDataAttributesLayersItemsRestrictionsItems
			for _, restrictionsTFItem := range layersTFItem.Restrictions {
				restrictionsDDItem := datadogV2.NewScheduleCreateRequestDataAttributesLayersItemsRestrictionsItems()

				if !restrictionsTFItem.EndDay.IsNull() {
					restrictionsDDItem.SetEndDay(datadogV2.ScheduleCreateRequestDataAttributesLayersItemsRestrictionsItemsEndDay(restrictionsTFItem.EndDay.ValueString()))
				}
				if !restrictionsTFItem.EndTime.IsNull() {
					restrictionsDDItem.SetEndTime(restrictionsTFItem.EndTime.ValueString())
				}
				if !restrictionsTFItem.StartDay.IsNull() {
					restrictionsDDItem.SetStartDay(datadogV2.ScheduleCreateRequestDataAttributesLayersItemsRestrictionsItemsStartDay(restrictionsTFItem.StartDay.ValueString()))
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

	req = datadogV2.NewScheduleCreateRequest(
		datadogV2.ScheduleCreateRequestData{
			Type:          datadogV2.SCHEDULECREATEREQUESTDATATYPE_SCHEDULES,
			Attributes:    *attributes,
			Relationships: relationships,
		},
	)

	return req, diags
}

func parseTime(timeString string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, timeString)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func (r *onCallScheduleResource) buildOnCallScheduleUpdateRequestBody(
	ctx context.Context,
	plan *onCallScheduleModel,
	existingLayerIdByName map[string]string,
) (*datadogV2.ScheduleUpdateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	req := &datadogV2.ScheduleUpdateRequest{}
	attributes := datadogV2.NewScheduleUpdateRequestDataAttributesWithDefaults()
	relationships := buildUpdateRelationships(plan.Teams)

	if plan.ID.IsNull() {
		diags.AddError("id is required", "id is required")
		return nil, diags
	}

	if !plan.Name.IsNull() {
		attributes.SetName(plan.Name.ValueString())
	}
	if !plan.TimeZone.IsNull() {
		attributes.SetTimeZone(plan.TimeZone.ValueString())
	}

	if !plan.Tags.IsNull() {
		var tags []string
		diags.Append(plan.Tags.ElementsAs(ctx, &tags, false)...)
		attributes.SetTags(tags)
	}

	if plan.Layers != nil {
		var layers []datadogV2.ScheduleUpdateRequestDataAttributesLayersItems
		for _, layersTFItem := range plan.Layers {
			layersDDItem := datadogV2.NewScheduleUpdateRequestDataAttributesLayersItemsWithDefaults()

			layerName := layersTFItem.Name.ValueString()

			if layerName == "" {
				diags.AddError("layer name is required", "layer name is required")
				return nil, diags
			}
			layersDDItem.SetName(layerName)

			if id, ok := existingLayerIdByName[layerName]; ok {
				layersDDItem.SetId(id)
			}

			effectiveDate, err := parseTime(layersTFItem.EffectiveDate.ValueString())
			if err != nil {
				diags.AddError("error parsing effective_date", err.Error())
				return nil, diags
			}
			layersDDItem.SetEffectiveDate(effectiveDate)

			if !layersTFItem.EndDate.IsNull() {
				endDate, err := parseTime(layersTFItem.EndDate.ValueString())
				if err != nil {
					diags.AddError("error parsing end_date", err.Error())
					return nil, diags
				}
				layersDDItem.SetEndDate(endDate)
			}

			rotationStart, err := parseTime(layersTFItem.RotationStart.ValueString())
			if err != nil {
				diags.AddError("error parsing rotation_start", err.Error())
				return nil, diags
			}
			layersDDItem.SetRotationStart(rotationStart)

			if layersTFItem.Members == nil {
				diags.AddError("members is required", "members is required")
				return nil, diags
			}

			var members []datadogV2.ScheduleUpdateRequestDataAttributesLayersItemsMembersItems
			for _, membersTFItem := range layersTFItem.Members {
				membersDDItem := datadogV2.NewScheduleUpdateRequestDataAttributesLayersItemsMembersItems()

				if !membersTFItem.User.Id.IsNull() {
					membersDDItem.User = &datadogV2.ScheduleUpdateRequestDataAttributesLayersItemsMembersItemsUser{
						Id: membersTFItem.User.Id.ValueStringPointer(),
					}
				}
				members = append(members, *membersDDItem)
			}
			layersDDItem.SetMembers(members)

			if layersTFItem.Restrictions != nil {
				var restrictions []datadogV2.ScheduleUpdateRequestDataAttributesLayersItemsRestrictionsItems
				for _, restrictionsTFItem := range layersTFItem.Restrictions {
					restrictionsDDItem := datadogV2.NewScheduleUpdateRequestDataAttributesLayersItemsRestrictionsItems()

					if !restrictionsTFItem.EndDay.IsNull() {
						restrictionsDDItem.SetEndDay(datadogV2.ScheduleUpdateRequestDataAttributesLayersItemsRestrictionsItemsEndDay(restrictionsTFItem.EndDay.ValueString()))
					}
					if !restrictionsTFItem.EndTime.IsNull() {
						restrictionsDDItem.SetEndTime(restrictionsTFItem.EndTime.ValueString())
					}
					if !restrictionsTFItem.StartDay.IsNull() {
						restrictionsDDItem.SetStartDay(datadogV2.ScheduleUpdateRequestDataAttributesLayersItemsRestrictionsItemsStartDay(restrictionsTFItem.StartDay.ValueString()))
					}
					if !restrictionsTFItem.StartTime.IsNull() {
						restrictionsDDItem.SetStartTime(restrictionsTFItem.StartTime.ValueString())
					}
					restrictions = append(restrictions, *restrictionsDDItem)
				}
				layersDDItem.SetRestrictions(restrictions)
			}

			if layersTFItem.Interval != nil {
				var interval datadogV2.ScheduleUpdateRequestDataAttributesLayersItemsInterval

				if !layersTFItem.Interval.Days.IsNull() {
					interval.SetDays(layersTFItem.Interval.Days.ValueInt32())
				}
				if !layersTFItem.Interval.Seconds.IsNull() {
					interval.SetSeconds(layersTFItem.Interval.Seconds.ValueInt64())
				}
				layersDDItem.SetInterval(interval)
			}

			layers = append(layers, *layersDDItem)
		}
		attributes.SetLayers(layers)
	}

	req = datadogV2.NewScheduleUpdateRequest(
		datadogV2.ScheduleUpdateRequestData{
			Id:            plan.ID.ValueString(),
			Type:          datadogV2.SCHEDULEUPDATEREQUESTDATATYPE_SCHEDULES,
			Attributes:    *attributes,
			Relationships: relationships,
		},
	)

	return req, diags
}

func buildCreateRelationships(plannedTeams []*onCallTeamModel) *datadogV2.ScheduleCreateRequestDataRelationships {
	var relationships *datadogV2.ScheduleCreateRequestDataRelationships
	plannedTeamsIds := make([]string, len(plannedTeams))
	for i, team := range plannedTeams {
		plannedTeamsIds[i] = team.Id.ValueString()
	}

	teamRelationships := make([]datadogV2.ScheduleCreateRequestDataRelationshipsTeamsDataItems, len(plannedTeamsIds))

	for t, teamId := range plannedTeamsIds {
		item := datadogV2.NewScheduleCreateRequestDataRelationshipsTeamsDataItemsWithDefaults()
		item.SetId(teamId)
		teamRelationships[t] = *item
	}
	if len(teamRelationships) > 0 {
		relationships = &datadogV2.ScheduleCreateRequestDataRelationships{
			Teams: &datadogV2.ScheduleCreateRequestDataRelationshipsTeams{
				Data: teamRelationships,
			},
		}
	}
	return relationships
}

func buildUpdateRelationships(plannedTeams []*onCallTeamModel) *datadogV2.ScheduleUpdateRequestDataRelationships {
	var relationships *datadogV2.ScheduleUpdateRequestDataRelationships

	plannedTeamsIds := make([]string, len(plannedTeams))
	for i, team := range plannedTeams {
		plannedTeamsIds[i] = team.Id.ValueString()
	}

	teamRelationships := make([]datadogV2.ScheduleUpdateRequestDataRelationshipsTeamsDataItems, len(plannedTeamsIds))

	for t, teamId := range plannedTeamsIds {
		item := datadogV2.NewScheduleUpdateRequestDataRelationshipsTeamsDataItemsWithDefaults()
		item.SetId(teamId)
		teamRelationships[t] = *item
	}
	if len(teamRelationships) > 0 {
		relationships = &datadogV2.ScheduleUpdateRequestDataRelationships{
			Teams: &datadogV2.ScheduleUpdateRequestDataRelationshipsTeams{
				Data: teamRelationships,
			},
		}
	}
	return relationships
}
