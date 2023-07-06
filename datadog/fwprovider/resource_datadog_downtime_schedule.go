package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &DowntimeScheduleResource{}
	_ resource.ResourceWithImportState = &DowntimeScheduleResource{}
)

type DowntimeScheduleResource struct {
	Api  *datadogV2.DowntimesApi
	Auth context.Context
}

type DowntimeScheduleModel struct {
	ID                            types.String            `tfsdk:"id"`
	DisplayTimezone               types.String            `tfsdk:"display_timezone"`
	Message                       types.String            `tfsdk:"message"`
	MuteFirstRecoveryNotification types.Bool              `tfsdk:"mute_first_recovery_notification"`
	Scope                         types.String            `tfsdk:"scope"`
	NotifyEndStates               types.List              `tfsdk:"notify_end_states"`
	NotifyEndTypes                types.List              `tfsdk:"notify_end_types"`
	MonitorIdentifier             *MonitorIdentifierModel `tfsdk:"monitor_identifier"`
	Schedule                      *ScheduleModel          `tfsdk:"schedule"`
}

type MonitorIdentifierModel struct {
	DowntimeMonitorIdentifierId   *DowntimeMonitorIdentifierIdModel   `tfsdk:"downtime_monitor_identifier_id"`
	DowntimeMonitorIdentifierTags *DowntimeMonitorIdentifierTagsModel `tfsdk:"downtime_monitor_identifier_tags"`
}
type DowntimeMonitorIdentifierIdModel struct {
	MonitorId types.Int64 `tfsdk:"monitor_id"`
}
type DowntimeMonitorIdentifierTagsModel struct {
	MonitorTags types.List `tfsdk:"monitor_tags"`
}

type ScheduleModel struct {
	DowntimeScheduleRecurrencesCreateRequest   *DowntimeScheduleRecurrencesCreateRequestModel   `tfsdk:"downtime_schedule_recurrences_create_request"`
	DowntimeScheduleOneTimeCreateUpdateRequest *DowntimeScheduleOneTimeCreateUpdateRequestModel `tfsdk:"downtime_schedule_one_time_create_update_request"`
}
type DowntimeScheduleRecurrencesCreateRequestModel struct {
	Timezone    types.String        `tfsdk:"timezone"`
	Recurrences []*RecurrencesModel `tfsdk:"recurrences"`
}
type RecurrencesModel struct {
	Duration types.String `tfsdk:"duration"`
	Rrule    types.String `tfsdk:"rrule"`
	Start    types.String `tfsdk:"start"`
}

type DowntimeScheduleOneTimeCreateUpdateRequestModel struct {
	End   types.String `tfsdk:"end"`
	Start types.String `tfsdk:"start"`
}

func NewDowntimeScheduleResource() resource.Resource {
	return &DowntimeScheduleResource{}
}

func (r *DowntimeScheduleResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetDowntimesApiV2()
	r.Auth = providerData.Auth
}

func (r *DowntimeScheduleResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "downtime_schedule"
}

func (r *DowntimeScheduleResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog DowntimeSchedule resource. This can be used to create and manage Datadog downtime_schedule.",
		Attributes: map[string]schema.Attribute{
			"display_timezone": schema.StringAttribute{
				Optional:    true,
				Description: "The timezone in which to display the downtime's start and end times in Datadog applications. This is not used as an offset for scheduling.",
			},
			"message": schema.StringAttribute{
				Optional:    true,
				Description: "A message to include with notifications for this downtime. Email notifications can be sent to specific users by using the same `@username` notation as events.",
			},
			"mute_first_recovery_notification": schema.BoolAttribute{
				Optional:    true,
				Description: "If the first recovery notification during a downtime should be muted.",
			},
			"scope": schema.StringAttribute{
				Optional:    true,
				Description: "The scope to which the downtime applies. Must follow the [common search syntax](https://docs.datadoghq.com/logs/explorer/search_syntax/).",
			},
			"notify_end_states": schema.ListAttribute{
				Optional:    true,
				Description: "States that will trigger a monitor notification when the `notify_end_types` action occurs.",
				ElementType: types.StringType,
			},
			"notify_end_types": schema.ListAttribute{
				Optional:    true,
				Description: "Actions that will trigger a monitor notification if the downtime is in the `notify_end_types` state.",
				ElementType: types.StringType,
			},
			"id": utils.ResourceIDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"monitor_identifier": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{},
				Blocks: map[string]schema.Block{
					"downtime_monitor_identifier_id": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"monitor_id": schema.Int64Attribute{
								Optional:    true,
								Description: "ID of the monitor to prevent notifications.",
							},
						},
					},
					"downtime_monitor_identifier_tags": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"monitor_tags": schema.ListAttribute{
								Optional:    true,
								Description: "A list of monitor tags. For example, tags that are applied directly to monitors, not tags that are used in monitor queries (which are filtered by the scope parameter), to which the downtime applies. The resulting downtime applies to monitors that match **all** provided monitor tags. Setting `monitor_tags` to `[*]` configures the downtime to mute all monitors for the given scope.",
								ElementType: types.StringType,
							},
						},
					},
				},
			},
			"schedule": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{},
				Blocks: map[string]schema.Block{
					"downtime_schedule_recurrences_create_request": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"timezone": schema.StringAttribute{
								Optional:    true,
								Description: "The timezone in which to schedule the downtime.",
							},
						},
						Blocks: map[string]schema.Block{
							"recurrences": schema.ListNestedBlock{
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"duration": schema.StringAttribute{
											Optional:    true,
											Description: "The length of the downtime. Must begin with an integer and end with one of 'm', 'h', d', or 'w'.",
										},
										"rrule": schema.StringAttribute{
											Optional:    true,
											Description: "The `RRULE` standard for defining recurring events. For example, to have a recurring event on the first day of each month, set the type to `rrule` and set the `FREQ` to `MONTHLY` and `BYMONTHDAY` to `1`. Most common `rrule` options from the [iCalendar Spec](https://tools.ietf.org/html/rfc5545) are supported.  **Note**: Attributes specifying the duration in `RRULE` are not supported (for example, `DTSTART`, `DTEND`, `DURATION`). More examples available in this [downtime guide](https://docs.datadoghq.com/monitors/guide/suppress-alert-with-downtimes/?tab=api).",
										},
										"start": schema.StringAttribute{
											Optional:    true,
											Description: "ISO-8601 Datetime to start the downtime. Must not include a UTC offset. If not provided, the downtime starts the moment it is created.",
										},
									},
								},
							},
						},
					},
					"downtime_schedule_one_time_create_update_request": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"end": schema.StringAttribute{
								Optional:    true,
								Description: "ISO-8601 Datetime to end the downtime. Must include a UTC offset of zero. If not provided, the downtime starts the moment it is created.",
							},
							"start": schema.StringAttribute{
								Optional:    true,
								Description: "ISO-8601 Datetime to start the downtime. Must include a UTC offset of zero. If not provided, the downtime starts the moment it is created.",
							},
						},
					},
				},
			},
		},
	}
}

func (r *DowntimeScheduleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *DowntimeScheduleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state DowntimeScheduleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	include := state.Include.ValueString()
	resp, httpResp, err := r.Api.GetDowntime(r.Auth, id, include)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving DowntimeSchedule"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *DowntimeScheduleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state DowntimeScheduleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildDowntimeScheduleRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateDowntime(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving DowntimeSchedule"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *DowntimeScheduleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state DowntimeScheduleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	body, diags := r.buildDowntimeScheduleUpdateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateDowntime(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving DowntimeSchedule"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *DowntimeScheduleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state DowntimeScheduleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	httpResp, err := r.Api.CancelDowntime(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting downtime_schedule"))
		return
	}
}

func (r *DowntimeScheduleResource) updateState(ctx context.Context, state *DowntimeScheduleModel, resp *datadogV2.DowntimeResponse) {
	state.ID = types.StringValue(resp.Data.GetId())

	data := resp.GetData()
	attributes := data.GetAttributes()

	if createdAt, ok := attributes.GetCreatedAtOk(); ok {
		state.CreatedAt = types.StringValue(*createdAt)
	}

	if displayTimezone, ok := attributes.GetDisplayTimezoneOk(); ok {
		state.DisplayTimezone = types.StringValue(*displayTimezone)
	}

	if message, ok := attributes.GetMessageOk(); ok {
		state.Message = types.StringValue(*message)
	}

	if modifiedAt, ok := attributes.GetModifiedAtOk(); ok {
		state.ModifiedAt = types.StringValue(*modifiedAt)
	}

	if muteFirstRecoveryNotification, ok := attributes.GetMuteFirstRecoveryNotificationOk(); ok {
		state.MuteFirstRecoveryNotification = types.BoolValue(*muteFirstRecoveryNotification)
	}

	state.Scope = types.StringValue(attributes.GetScope())

	if status, ok := attributes.GetStatusOk(); ok {
		state.Status = types.StringValue(*status)
	}

	if notifyEndStates, ok := attributes.GetNotifyEndStatesOk(); ok && len(*notifyEndStates) > 0 {
		state.NotifyEndStates, _ = types.ListValueFrom(ctx, types.StringType, *notifyEndStates)
	}

	if notifyEndTypes, ok := attributes.GetNotifyEndTypesOk(); ok && len(*notifyEndTypes) > 0 {
		state.NotifyEndTypes, _ = types.ListValueFrom(ctx, types.StringType, *notifyEndTypes)
	}

	monitorIdentifierTf := MonitorIdentifierModel{}
	if attributes.MonitorIdentifier.DowntimeMonitorIdentifierId != nil {
		downtimeMonitorIdentifierIdTf := DowntimeMonitorIdentifierIdModel{}
		if monitorId, ok := attributes.MonitorIdentifier.DowntimeMonitorIdentifierId.GetMonitorIdOk(); ok {
			downtimeMonitorIdentifierIdTf.MonitorId = types.Int64Value(*monitorId)
		}

		monitorIdentifierTf.DowntimeMonitorIdentifierId = &downtimeMonitorIdentifierIdTf
	}
	if attributes.MonitorIdentifier.DowntimeMonitorIdentifierTags != nil {
		downtimeMonitorIdentifierTagsTf := DowntimeMonitorIdentifierTagsModel{}
		if monitorTags, ok := attributes.MonitorIdentifier.DowntimeMonitorIdentifierTags.GetMonitorTagsOk(); ok && len(*monitorTags) > 0 {
			downtimeMonitorIdentifierTagsTf.MonitorTags, _ = types.ListValueFrom(ctx, types.StringType, *monitorTags)
		}

		monitorIdentifierTf.DowntimeMonitorIdentifierTags = &downtimeMonitorIdentifierTagsTf
	}

	if schedule, ok := attributes.GetScheduleOk(); ok {

		scheduleTf := ScheduleModel{}
		if schedule.DowntimeScheduleRecurrencesResponse != nil {
			downtimeScheduleRecurrencesResponseTf := DowntimeScheduleRecurrencesResponseModel{}
			if currentDowntime, ok := schedule.DowntimeScheduleRecurrencesResponse.GetCurrentDowntimeOk(); ok {

				currentDowntimeTf := CurrentDowntimeModel{}
				if end, ok := currentDowntime.GetEndOk(); ok {
					currentDowntimeTf.End = types.StringValue(*end)
				}
				if start, ok := currentDowntime.GetStartOk(); ok {
					currentDowntimeTf.Start = types.StringValue(*start)
				}

				downtimeScheduleRecurrencesResponseTf.CurrentDowntime = &currentDowntimeTf
			}
			if recurrences, ok := schedule.DowntimeScheduleRecurrencesResponse.GetRecurrencesOk(); ok && len(*recurrences) > 0 {
				downtimeScheduleRecurrencesResponseTf.Recurrences = []*RecurrencesModel{}
				for _, recurrencesDd := range *recurrences {
					recurrencesTfItem := RecurrencesModel{}
					if duration, ok := recurrencesDd.GetDurationOk(); ok {
						recurrencesTfItem.Duration = types.StringValue(*duration)
					}
					if rrule, ok := recurrencesDd.GetRruleOk(); ok {
						recurrencesTfItem.Rrule = types.StringValue(*rrule)
					}
					if start, ok := recurrencesDd.GetStartOk(); ok {
						recurrencesTfItem.Start = types.StringValue(*start)
					}

					downtimeScheduleRecurrencesResponseTf.Recurrences = append(downtimeScheduleRecurrencesResponseTf.Recurrences, &recurrencesTfItem)
				}
			}
			if timezone, ok := schedule.DowntimeScheduleRecurrencesResponse.GetTimezoneOk(); ok {
				downtimeScheduleRecurrencesResponseTf.Timezone = types.StringValue(*timezone)
			}

			scheduleTf.DowntimeScheduleRecurrencesResponse = &downtimeScheduleRecurrencesResponseTf
		}
		if schedule.DowntimeScheduleOneTimeResponse != nil {
			downtimeScheduleOneTimeResponseTf := DowntimeScheduleOneTimeResponseModel{}
			if end, ok := schedule.DowntimeScheduleOneTimeResponse.GetEndOk(); ok {
				downtimeScheduleOneTimeResponseTf.End = types.StringValue(*end)
			}
			if start, ok := schedule.DowntimeScheduleOneTimeResponse.GetStartOk(); ok {
				downtimeScheduleOneTimeResponseTf.Start = types.StringValue(*start)
			}

			scheduleTf.DowntimeScheduleOneTimeResponse = &downtimeScheduleOneTimeResponseTf
		}
	}
}

func (r *DowntimeScheduleResource) buildDowntimeScheduleRequestBody(ctx context.Context, state *DowntimeScheduleModel) (*datadogV2.DowntimeCreateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewDowntimeCreateRequestAttributesWithDefaults()

	if !state.DisplayTimezone.IsNull() {
		attributes.SetDisplayTimezone(state.DisplayTimezone.ValueString())
	}
	if !state.Message.IsNull() {
		attributes.SetMessage(state.Message.ValueString())
	}
	if !state.MuteFirstRecoveryNotification.IsNull() {
		attributes.SetMuteFirstRecoveryNotification(state.MuteFirstRecoveryNotification.ValueBool())
	}
	attributes.SetScope(state.Scope.ValueString())

	if !state.NotifyEndStates.IsNull() {
		var notifyEndStates []datadogV2.DowntimeNotifyEndStateTypes
		diags.Append(state.NotifyEndStates.ElementsAs(ctx, &notifyEndStates, false)...)
		attributes.SetNotifyEndStates(notifyEndStates)
	}

	if !state.NotifyEndTypes.IsNull() {
		var notifyEndTypes []datadogV2.DowntimeNotifyEndStateActions
		diags.Append(state.NotifyEndTypes.ElementsAs(ctx, &notifyEndTypes, false)...)
		attributes.SetNotifyEndTypes(notifyEndTypes)
	}

	var monitorIdentifier datadogV2.DowntimeMonitorIdentifier

	if state.MonitorIdentifier.DowntimeMonitorIdentifierId != nil {
		var downtimeMonitorIdentifierId datadogV2.DowntimeMonitorIdentifierId

		downtimeMonitorIdentifierId.SetMonitorId(state.MonitorIdentifier.DowntimeMonitorIdentifierId.MonitorId.ValueInt64())

		monitorIdentifier.DowntimeMonitorIdentifierId = &downtimeMonitorIdentifierId
	}

	if state.MonitorIdentifier.DowntimeMonitorIdentifierTags != nil {
		var downtimeMonitorIdentifierTags datadogV2.DowntimeMonitorIdentifierTags

		var monitorTags []string
		diags.Append(state.MonitorIdentifier.DowntimeMonitorIdentifierTags.MonitorTags.ElementsAs(ctx, &monitorTags, false)...)
		downtimeMonitorIdentifierTags.SetMonitorTags(monitorTags)

		monitorIdentifier.DowntimeMonitorIdentifierTags = &downtimeMonitorIdentifierTags
	}

	attributes.MonitorIdentifier = monitorIdentifier

	if state.Schedule != nil {
		var schedule datadogV2.DowntimeScheduleCreateRequest

		if state.Schedule.DowntimeScheduleRecurrencesCreateRequest != nil {
			var downtimeScheduleRecurrencesCreateRequest datadogV2.DowntimeScheduleRecurrencesCreateRequest

			if !state.Schedule.DowntimeScheduleRecurrencesCreateRequest.Timezone.IsNull() {
				downtimeScheduleRecurrencesCreateRequest.SetTimezone(state.Schedule.DowntimeScheduleRecurrencesCreateRequest.Timezone.ValueString())
			}

			if state.Schedule.DowntimeScheduleRecurrencesCreateRequest.Recurrences != nil {
				var recurrences []datadogV2.DowntimeScheduleRecurrenceCreateUpdateRequest
				for _, recurrencesTFItem := range state.Schedule.DowntimeScheduleRecurrencesCreateRequest.Recurrences {
					recurrencesDDItem := datadogV2.NewDowntimeScheduleRecurrenceCreateUpdateRequest()

					recurrencesDDItem.SetDuration(recurrencesTFItem.Duration.ValueString())
					recurrencesDDItem.SetRrule(recurrencesTFItem.Rrule.ValueString())
					if !recurrencesTFItem.Start.IsNull() {
						recurrencesDDItem.SetStart(recurrencesTFItem.Start.ValueString())
					}
				}
				downtimeScheduleRecurrencesCreateRequest.SetRecurrences(recurrences)
			}

			schedule.DowntimeScheduleRecurrencesCreateRequest = &downtimeScheduleRecurrencesCreateRequest
		}

		if state.Schedule.DowntimeScheduleOneTimeCreateUpdateRequest != nil {
			var downtimeScheduleOneTimeCreateUpdateRequest datadogV2.DowntimeScheduleOneTimeCreateUpdateRequest

			if !state.Schedule.DowntimeScheduleOneTimeCreateUpdateRequest.End.IsNull() {
				downtimeScheduleOneTimeCreateUpdateRequest.SetEnd(state.Schedule.DowntimeScheduleOneTimeCreateUpdateRequest.End.ValueString())
			}
			if !state.Schedule.DowntimeScheduleOneTimeCreateUpdateRequest.Start.IsNull() {
				downtimeScheduleOneTimeCreateUpdateRequest.SetStart(state.Schedule.DowntimeScheduleOneTimeCreateUpdateRequest.Start.ValueString())
			}

			schedule.DowntimeScheduleOneTimeCreateUpdateRequest = &downtimeScheduleOneTimeCreateUpdateRequest
		}

		attributes.Schedule = &schedule
	}

	req := datadogV2.NewDowntimeCreateRequestWithDefaults()
	req.Data = *datadogV2.NewDowntimeCreateRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}

func (r *DowntimeScheduleResource) buildDowntimeScheduleUpdateRequestBody(ctx context.Context, state *DowntimeScheduleModel) (*datadogV2.DowntimeUpdateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewDowntimeUpdateRequestAttributesWithDefaults()

	if !state.DisplayTimezone.IsNull() {
		attributes.SetDisplayTimezone(state.DisplayTimezone.ValueString())
	}
	if !state.Message.IsNull() {
		attributes.SetMessage(state.Message.ValueString())
	}
	if !state.MuteFirstRecoveryNotification.IsNull() {
		attributes.SetMuteFirstRecoveryNotification(state.MuteFirstRecoveryNotification.ValueBool())
	}
	if !state.Scope.IsNull() {
		attributes.SetScope(state.Scope.ValueString())
	}

	if !state.NotifyEndStates.IsNull() {
		var notifyEndStates []datadogV2.DowntimeNotifyEndStateTypes
		diags.Append(state.NotifyEndStates.ElementsAs(ctx, &notifyEndStates, false)...)
		attributes.SetNotifyEndStates(notifyEndStates)
	}

	if !state.NotifyEndTypes.IsNull() {
		var notifyEndTypes []datadogV2.DowntimeNotifyEndStateActions
		diags.Append(state.NotifyEndTypes.ElementsAs(ctx, &notifyEndTypes, false)...)
		attributes.SetNotifyEndTypes(notifyEndTypes)
	}

	if state.MonitorIdentifier != nil {
		var monitorIdentifier datadogV2.DowntimeMonitorIdentifier

		if state.MonitorIdentifier.DowntimeMonitorIdentifierId != nil {
			var downtimeMonitorIdentifierId datadogV2.DowntimeMonitorIdentifierId

			downtimeMonitorIdentifierId.SetMonitorId(state.MonitorIdentifier.DowntimeMonitorIdentifierId.MonitorId.ValueInt64())

			monitorIdentifier.DowntimeMonitorIdentifierId = &downtimeMonitorIdentifierId
		}

		if state.MonitorIdentifier.DowntimeMonitorIdentifierTags != nil {
			var downtimeMonitorIdentifierTags datadogV2.DowntimeMonitorIdentifierTags

			var monitorTags []string
			diags.Append(state.MonitorIdentifier.DowntimeMonitorIdentifierTags.MonitorTags.ElementsAs(ctx, &monitorTags, false)...)
			downtimeMonitorIdentifierTags.SetMonitorTags(monitorTags)

			monitorIdentifier.DowntimeMonitorIdentifierTags = &downtimeMonitorIdentifierTags
		}

		attributes.MonitorIdentifier = &monitorIdentifier
	}

	if state.Schedule != nil {
		var schedule datadogV2.DowntimeScheduleUpdateRequest

		if state.Schedule.DowntimeScheduleRecurrencesUpdateRequest != nil {
			var downtimeScheduleRecurrencesUpdateRequest datadogV2.DowntimeScheduleRecurrencesUpdateRequest

			if !state.Schedule.DowntimeScheduleRecurrencesUpdateRequest.Timezone.IsNull() {
				downtimeScheduleRecurrencesUpdateRequest.SetTimezone(state.Schedule.DowntimeScheduleRecurrencesUpdateRequest.Timezone.ValueString())
			}

			if state.Schedule.DowntimeScheduleRecurrencesUpdateRequest.Recurrences != nil {
				var recurrences []datadogV2.DowntimeScheduleRecurrenceCreateUpdateRequest
				for _, recurrencesTFItem := range state.Schedule.DowntimeScheduleRecurrencesUpdateRequest.Recurrences {
					recurrencesDDItem := datadogV2.NewDowntimeScheduleRecurrenceCreateUpdateRequest()

					recurrencesDDItem.SetDuration(recurrencesTFItem.Duration.ValueString())
					recurrencesDDItem.SetRrule(recurrencesTFItem.Rrule.ValueString())
					if !recurrencesTFItem.Start.IsNull() {
						recurrencesDDItem.SetStart(recurrencesTFItem.Start.ValueString())
					}
				}
				downtimeScheduleRecurrencesUpdateRequest.SetRecurrences(recurrences)
			}

			schedule.DowntimeScheduleRecurrencesUpdateRequest = &downtimeScheduleRecurrencesUpdateRequest
		}

		if state.Schedule.DowntimeScheduleOneTimeCreateUpdateRequest != nil {
			var downtimeScheduleOneTimeCreateUpdateRequest datadogV2.DowntimeScheduleOneTimeCreateUpdateRequest

			if !state.Schedule.DowntimeScheduleOneTimeCreateUpdateRequest.End.IsNull() {
				downtimeScheduleOneTimeCreateUpdateRequest.SetEnd(state.Schedule.DowntimeScheduleOneTimeCreateUpdateRequest.End.ValueString())
			}
			if !state.Schedule.DowntimeScheduleOneTimeCreateUpdateRequest.Start.IsNull() {
				downtimeScheduleOneTimeCreateUpdateRequest.SetStart(state.Schedule.DowntimeScheduleOneTimeCreateUpdateRequest.Start.ValueString())
			}

			schedule.DowntimeScheduleOneTimeCreateUpdateRequest = &downtimeScheduleOneTimeCreateUpdateRequest
		}

		attributes.Schedule = &schedule
	}

	req := datadogV2.NewDowntimeUpdateRequestWithDefaults()
	req.Data = *datadogV2.NewDowntimeUpdateRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}
