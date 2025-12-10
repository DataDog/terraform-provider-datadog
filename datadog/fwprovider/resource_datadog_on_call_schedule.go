package fwprovider

import (
	"context"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	ID       types.String   `tfsdk:"id"`
	Name     types.String   `tfsdk:"name"`
	TimeZone types.String   `tfsdk:"time_zone"`
	Teams    types.List     `tfsdk:"teams"`
	Layers   []*layersModel `tfsdk:"layer"`
}

type layersModel struct {
	Id            types.String         `tfsdk:"id"`
	EffectiveDate timetypes.RFC3339    `tfsdk:"effective_date"`
	EndDate       timetypes.RFC3339    `tfsdk:"end_date"`
	Name          types.String         `tfsdk:"name"`
	TimeZone      types.String         `tfsdk:"time_zone"`
	RotationStart timetypes.RFC3339    `tfsdk:"rotation_start"`
	Users         []types.String       `tfsdk:"users"`
	Restrictions  []*restrictionsModel `tfsdk:"restriction"`
	Interval      *intervalModel       `tfsdk:"interval"`
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

func (m *onCallScheduleModel) Validate() diag.Diagnostics {
	diags := diag.Diagnostics{}

	for i, layer := range m.Layers {
		root := frameworkPath.Root("layer").AtListIndex(i)

		if layer.Interval == nil {
			diags.AddAttributeError(root.AtName("interval"), "missing interval", "schedules must specify an interval")
		} else {
			if layer.Interval.Seconds.IsNull() && layer.Interval.Days.IsNull() {
				diags.AddAttributeError(root.AtName("interval"), "missing interval", "interval must specify at least one of `days` or `seconds`")
			}
			if layer.Interval.Days.ValueInt32() < 0 {
				diags.AddAttributeError(root.AtName("interval").AtName("days"), "invalid value", "days must be a positive integer")
			}
			if layer.Interval.Seconds.ValueInt64() < 0 {
				diags.AddAttributeError(root.AtName("interval").AtName("seconds"), "invalid value", "seconds must be a positive integer")
			}
		}
	}

	return diags
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
		Description: "Provides a Datadog On-Call schedule resource. This can be used to create and manage Datadog On-Call schedules.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "A human-readable name for the new schedule.",
			},
			"time_zone": schema.StringAttribute{
				Required:    true,
				Description: "The time zone in which the schedule is defined.",
			},
			"teams": schema.ListAttribute{
				Description: "A list of team ids associated with the schedule.",
				Optional:    true,
				Required:    false,
				Computed:    true,
				ElementType: types.StringType,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
			},
			"id": utils.ResourceIDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"layer": schema.ListNestedBlock{
				Description: "List of layers for the schedule.",
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of this layer.",
						},
						"effective_date": schema.StringAttribute{
							CustomType:  timetypes.RFC3339Type{},
							Required:    true,
							Description: "The date/time when this layer should become active (in ISO 8601).",
							Validators:  []validator.String{validators.TimeFormatValidator(time.RFC3339)},
						},
						"end_date": schema.StringAttribute{
							CustomType:  timetypes.RFC3339Type{},
							Optional:    true,
							Description: "The date/time after which this layer no longer applies (in ISO 8601).",
							Validators:  []validator.String{validators.TimeFormatValidator(time.RFC3339)},
						},
						"name": schema.StringAttribute{
							Required:    true,
							Description: "The name of this layer. Should be unique within the schedule.",
						},
						"rotation_start": schema.StringAttribute{
							CustomType:  timetypes.RFC3339Type{},
							Required:    true,
							Description: "The date/time when the rotation for this layer starts (in ISO 8601).",
							Validators:  []validator.String{validators.TimeFormatValidator(time.RFC3339)},
						},
						"time_zone": schema.StringAttribute{
							Optional:    true,
							Description: "The time zone for this layer.",
							Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
						},
						"users": schema.ListAttribute{
							Required:    true,
							Description: "List of user IDs for the layer. Can either be a valid user id or null",
							ElementType: types.StringType,
							Validators:  []validator.List{listvalidator.SizeAtLeast(1)},
						},
					},
					Blocks: map[string]schema.Block{
						"restriction": schema.ListNestedBlock{
							Description: "List of restrictions for the layer.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"end_day": schema.StringAttribute{
										Validators:  []validator.String{validators.NewEnumValidator[validator.String](datadogV2.NewWeekdayFromValue)},
										Required:    true,
										Description: "The weekday when the restriction period ends.",
									},
									"end_time": schema.StringAttribute{
										Required:    true,
										Description: "The time of day when the restriction ends (hh:mm:ss).",
									},
									"start_day": schema.StringAttribute{
										Validators:  []validator.String{validators.NewEnumValidator[validator.String](datadogV2.NewWeekdayFromValue)},
										Required:    true,
										Description: "The weekday when the restriction period starts.",
									},
									"start_time": schema.StringAttribute{
										Required:    true,
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

	include := "layers,layers.members.user"
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

	newState := *r.newState(ctx, &resp, &state)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

func (r *onCallScheduleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan onCallScheduleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(plan.Validate()...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildOnCallScheduleRequestBody(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	include := "layers,layers.members.user"
	resp, _, err := r.Api.CreateOnCallSchedule(r.Auth, *body, datadogV2.CreateOnCallScheduleOptionalParameters{
		Include: &include,
	})
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating OnCallSchedule"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	state := r.newState(ctx, &resp, nil)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *onCallScheduleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan onCallScheduleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(plan.Validate()...)
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

	include := "layers,layers.members.user"
	resp, _, err := r.Api.UpdateOnCallSchedule(r.Auth, id, *body, datadogV2.UpdateOnCallScheduleOptionalParameters{
		Include: &include,
	})
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating OnCallSchedule"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	state := r.newState(ctx, &resp, &previousState)

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

func (r *onCallScheduleResource) newState(ctx context.Context, resp *datadogV2.Schedule, currentState *onCallScheduleModel) *onCallScheduleModel {
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

	teams := make([]string, len(data.GetRelationships().Teams.GetData()))
	for i, team := range data.GetRelationships().Teams.GetData() {
		teams[i] = team.GetId()
	}
	state.Teams, _ = types.ListValueFrom(ctx, types.StringType, teams)

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
		var currentLayer *layersModel
		if currentState != nil && currentState.Layers != nil && i < len(currentState.Layers) {
			currentLayer = currentState.Layers[i]
		}
		state.Layers[i] = newLayerModel(layersByID[layer.GetId()], membersByID, currentLayer)
	}
	return state
}

func newLayerModel(layer *datadogV2.Layer, membersByID map[string]*datadogV2.ScheduleMember, currentLayer *layersModel) *layersModel {
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

	currentEndDate := timetypes.NewRFC3339Null()
	if currentLayer != nil && !currentLayer.EndDate.IsNull() {
		currentEndDate = currentLayer.EndDate
	}
	endDate := keepCurrentTimezone(currentEndDate, layer.Attributes.GetEndDate())

	currentEffectiveDate := timetypes.NewRFC3339Null()
	if currentLayer != nil && !currentLayer.EffectiveDate.IsNull() {
		currentEffectiveDate = currentLayer.EffectiveDate
	}
	effectiveDate := keepCurrentTimezone(currentEffectiveDate, layer.Attributes.GetEffectiveDate())

	currentRotationStart := timetypes.NewRFC3339Null()
	if currentLayer != nil && !currentLayer.RotationStart.IsNull() {
		currentRotationStart = currentLayer.RotationStart
	}
	rotationStart := keepCurrentTimezone(currentRotationStart, layer.Attributes.GetRotationStart())

	timeZone, ok := layer.Attributes.GetTimeZoneOk()
	timeZoneValue := types.StringNull()
	if ok {
		timeZoneValue = types.StringValue(*timeZone)
	}

	return &layersModel{
		Id:            types.StringValue(layer.GetId()),
		EffectiveDate: effectiveDate,
		EndDate:       endDate,
		Name:          types.StringValue(layer.Attributes.GetName()),
		TimeZone:      timeZoneValue,
		RotationStart: rotationStart,
		Users:         memberIds,
		Restrictions:  restrictionsModels,
		Interval:      &intervalModel{Days: types.Int32Value(interval.GetDays()), Seconds: types.Int64Value(interval.GetSeconds())},
	}
}

// The timetypes.RFC3339 ensures that whatever the timezone returned by the server is,
// we do not change it if the time is actually the same.
// We still need to ensure to keep the time in the same timezone as one defined in the plan.
func keepCurrentTimezone(currentTime timetypes.RFC3339, newTime time.Time) timetypes.RFC3339 {
	if newTime.IsZero() {
		return timetypes.NewRFC3339Null()
	}
	newTimeRfc := timetypes.NewRFC3339TimeValue(newTime)
	if currentTime.IsNull() {
		return newTimeRfc
	}

	currentTimeRfc, _ := currentTime.ValueRFC3339Time()

	if currentTimeRfc.Unix() == newTime.Unix() {
		return currentTime
	}
	return newTimeRfc

}

func (r *onCallScheduleResource) buildOnCallScheduleRequestBody(ctx context.Context, state *onCallScheduleModel) (*datadogV2.ScheduleCreateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	req := &datadogV2.ScheduleCreateRequest{}

	var teams []string
	diags.Append(state.Teams.ElementsAs(ctx, &teams, false)...)

	relationships := buildCreateRelationships(teams)

	attributes := datadogV2.NewScheduleCreateRequestDataAttributesWithDefaults()

	attributes.SetName(state.Name.ValueString())
	if !state.TimeZone.IsNull() {
		attributes.SetTimeZone(state.TimeZone.ValueString())
	}

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
			datadogV2.LayerAttributesInterval{},
			[]datadogV2.ScheduleRequestDataAttributesLayersItemsMembersItems{},
			layersTFItem.Name.ValueString(),
			rotationStart,
		)

		if layersTFItem.Interval != nil {
			layersDDItem.Interval.Days = layersTFItem.Interval.Days.ValueInt32Pointer()
			layersDDItem.Interval.Seconds = layersTFItem.Interval.Seconds.ValueInt64Pointer()
		}

		if !layersTFItem.TimeZone.IsNull() {
			layersDDItem.TimeZone = layersTFItem.TimeZone.ValueStringPointer()
		}

		if !layersTFItem.EndDate.IsNull() {
			endDate, err := parseTime(layersTFItem.EndDate.ValueString())
			if err != nil {
				diags.AddError("error parsing end_date", err.Error())
				return nil, diags
			}
			layersDDItem.SetEndDate(endDate)
		}

		var members []datadogV2.ScheduleRequestDataAttributesLayersItemsMembersItems
		for _, memberId := range layersTFItem.Users {
			membersDDItem := datadogV2.NewScheduleRequestDataAttributesLayersItemsMembersItems()

			if !memberId.IsNull() {
				userId := memberId.ValueString()
				if userId == "" {
					diags.AddError("user_id can't be empty, either set the user_id to a valid user, set the user_id to null or omit the field", "user_id can't be empty, either set the user_id to a valid user or to null")
					return nil, diags
				}
				membersDDItem.User = &datadogV2.ScheduleRequestDataAttributesLayersItemsMembersItemsUser{
					Id: &userId,
				}
			}

			members = append(members, *membersDDItem)
		}
		layersDDItem.SetMembers(members)

		if layersTFItem.Restrictions != nil {
			var restrictions []datadogV2.TimeRestriction
			for _, restrictionsTFItem := range layersTFItem.Restrictions {
				restrictionsDDItem := datadogV2.NewTimeRestriction()

				if !restrictionsTFItem.EndDay.IsNull() {
					restrictionsDDItem.SetEndDay(datadogV2.Weekday(restrictionsTFItem.EndDay.ValueString()))
				}
				if !restrictionsTFItem.EndTime.IsNull() {
					restrictionsDDItem.SetEndTime(restrictionsTFItem.EndTime.ValueString())
				}
				if !restrictionsTFItem.StartDay.IsNull() {
					restrictionsDDItem.SetStartDay(datadogV2.Weekday(restrictionsTFItem.StartDay.ValueString()))
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
	if !plan.TimeZone.IsNull() {
		attributes.SetTimeZone(plan.TimeZone.ValueString())
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

			var members []datadogV2.ScheduleRequestDataAttributesLayersItemsMembersItems
			for _, memberId := range layersTFItem.Users {
				membersDDItem := datadogV2.NewScheduleRequestDataAttributesLayersItemsMembersItems()

				if !memberId.IsNull() {
					userId := memberId.ValueString()
					if userId == "" {
						diags.AddError("user_id can't be empty, either set the user_id to a valid user, set the user_id to null or omit the field", "user_id can't be empty, either set the user_id to a valid user or to null")
						return nil, diags
					}
					membersDDItem.User = &datadogV2.ScheduleRequestDataAttributesLayersItemsMembersItemsUser{
						Id: &userId,
					}
				}
				members = append(members, *membersDDItem)
			}
			layersDDItem.SetMembers(members)

			if layersTFItem.Restrictions != nil {
				var restrictions []datadogV2.TimeRestriction
				for _, restrictionsTFItem := range layersTFItem.Restrictions {
					restrictionsDDItem := datadogV2.NewTimeRestriction()

					if !restrictionsTFItem.EndDay.IsNull() {
						restrictionsDDItem.SetEndDay(datadogV2.Weekday(restrictionsTFItem.EndDay.ValueString()))
					}
					if !restrictionsTFItem.EndTime.IsNull() {
						restrictionsDDItem.SetEndTime(restrictionsTFItem.EndTime.ValueString())
					}
					if !restrictionsTFItem.StartDay.IsNull() {
						restrictionsDDItem.SetStartDay(datadogV2.Weekday(restrictionsTFItem.StartDay.ValueString()))
					}
					if !restrictionsTFItem.StartTime.IsNull() {
						restrictionsDDItem.SetStartTime(restrictionsTFItem.StartTime.ValueString())
					}
					restrictions = append(restrictions, *restrictionsDDItem)
				}
				layersDDItem.SetRestrictions(restrictions)
			}

			if layersTFItem.Interval != nil {
				var interval datadogV2.LayerAttributesInterval

				if !layersTFItem.Interval.Days.IsNull() {
					interval.SetDays(layersTFItem.Interval.Days.ValueInt32())
				}
				if !layersTFItem.Interval.Seconds.IsNull() {
					interval.SetSeconds(layersTFItem.Interval.Seconds.ValueInt64())
				}
				layersDDItem.SetInterval(interval)
			}

			if !layersTFItem.TimeZone.IsNull() {
				layersDDItem.TimeZone = layersTFItem.TimeZone.ValueStringPointer()
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

func buildCreateRelationships(plannedTeams []string) *datadogV2.ScheduleCreateRequestDataRelationships {
	var relationships *datadogV2.ScheduleCreateRequestDataRelationships

	teamRelationships := make([]datadogV2.DataRelationshipsTeamsDataItems, len(plannedTeams))

	for t, teamId := range plannedTeams {
		item := datadogV2.NewDataRelationshipsTeamsDataItemsWithDefaults()
		item.SetId(teamId)
		teamRelationships[t] = *item
	}

	if len(teamRelationships) > 0 {
		relationships = &datadogV2.ScheduleCreateRequestDataRelationships{
			Teams: &datadogV2.DataRelationshipsTeams{
				Data: teamRelationships,
			},
		}
	}
	return relationships
}

func buildUpdateRelationships(plannedTeams []string) *datadogV2.ScheduleUpdateRequestDataRelationships {
	var relationships *datadogV2.ScheduleUpdateRequestDataRelationships

	teamRelationships := make([]datadogV2.DataRelationshipsTeamsDataItems, len(plannedTeams))

	for t, teamId := range plannedTeams {
		item := datadogV2.NewDataRelationshipsTeamsDataItemsWithDefaults()
		item.SetId(teamId)
		teamRelationships[t] = *item
	}

	if len(teamRelationships) > 0 {
		relationships = &datadogV2.ScheduleUpdateRequestDataRelationships{
			Teams: &datadogV2.DataRelationshipsTeams{
				Data: teamRelationships,
			},
		}
	}
	return relationships
}
