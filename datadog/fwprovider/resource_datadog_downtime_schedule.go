package fwprovider

import (
	"context"
	"time"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/planmodifiers"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	ID                                 types.String                        `tfsdk:"id"`
	DisplayTimezone                    types.String                        `tfsdk:"display_timezone"`
	Message                            types.String                        `tfsdk:"message"`
	MuteFirstRecoveryNotification      types.Bool                          `tfsdk:"mute_first_recovery_notification"`
	Scope                              types.String                        `tfsdk:"scope"`
	NotifyEndStates                    types.Set                           `tfsdk:"notify_end_states"`
	NotifyEndTypes                     types.Set                           `tfsdk:"notify_end_types"`
	MonitorIdentifier                  *MonitorIdentifierModel             `tfsdk:"monitor_identifier"`
	DowntimeScheduleRecurrenceSchedule *DowntimeScheduleRecurrenceSchedule `tfsdk:"recurring_schedule"`
	DowntimeScheduleOneTimeSchedule    *DowntimeScheduleOneTimeSchedule    `tfsdk:"one_time_schedule"`
}

type MonitorIdentifierModel struct {
	DowntimeMonitorIdentifierId   types.Int64 `tfsdk:"monitor_id"`
	DowntimeMonitorIdentifierTags types.Set   `tfsdk:"monitor_tags"`
}

type DowntimeScheduleRecurrenceSchedule struct {
	Timezone    types.String        `tfsdk:"timezone"`
	Recurrences []*RecurrencesModel `tfsdk:"recurrence"`
}
type RecurrencesModel struct {
	Duration types.String `tfsdk:"duration"`
	Rrule    types.String `tfsdk:"rrule"`
	Start    types.String `tfsdk:"start"`
}

type DowntimeScheduleOneTimeSchedule struct {
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
		Description: "Provides a Datadog DowntimeSchedule resource. This can be used to create and manage Datadog downtimes.",
		Attributes: map[string]schema.Attribute{
			"display_timezone": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The timezone in which to display the downtime's start and end times in Datadog applications. This is not used as an offset for scheduling.",
			},
			"message": schema.StringAttribute{
				Optional:    true,
				Description: "A message to include with notifications for this downtime. Email notifications can be sent to specific users by using the same `@username` notation as events.",
			},
			"mute_first_recovery_notification": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "If the first recovery notification during a downtime should be muted.",
			},
			"scope": schema.StringAttribute{
				Required:    true,
				Description: "The scope to which the downtime applies. Must follow the [common search syntax](https://docs.datadoghq.com/logs/explorer/search_syntax/).",
			},
			"notify_end_states": schema.SetAttribute{
				Optional:    true,
				Computed:    true,
				Description: "States that will trigger a monitor notification when the `notify_end_types` action occurs.",
				ElementType: types.StringType,
			},
			"notify_end_types": schema.SetAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Actions that will trigger a monitor notification if the downtime is in the `notify_end_types` state.",
				ElementType: types.StringType,
			},
			"id": utils.ResourceIDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"monitor_identifier": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"monitor_id": schema.Int64Attribute{
						Optional:    true,
						Description: "ID of the monitor to prevent notifications.",
						Validators:  []validator.Int64{int64validator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("monitor_tags"))},
					},
					"monitor_tags": schema.SetAttribute{
						Optional:    true,
						Description: "A list of monitor tags. For example, tags that are applied directly to monitors, not tags that are used in monitor queries (which are filtered by the scope parameter), to which the downtime applies. The resulting downtime applies to monitors that match **all** provided monitor tags. Setting `monitor_tags` to `[*]` configures the downtime to mute all monitors for the given scope.",
						ElementType: types.StringType,
					},
				},
				Validators: []validator.Object{
					objectvalidator.IsRequired(),
				},
			},
			"one_time_schedule": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"end": schema.StringAttribute{
						Optional:    true,
						Description: "ISO-8601 Datetime to end the downtime. Must include a UTC offset of zero. If not provided, the downtime never ends.",
						Validators:  []validator.String{validators.TimeFormatValidator("2006-01-02T15:04:05Z")},
					},
					"start": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "ISO-8601 Datetime to start the downtime. Must include a UTC offset of zero. If not provided, the downtime starts the moment it is created.",
						Validators:  []validator.String{validators.TimeFormatValidator("2006-01-02T15:04:05Z")},
					},
				},
				Validators:    []validator.Object{objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("recurring_schedule"))},
				PlanModifiers: []planmodifier.Object{planmodifiers.RemoveBlockModifier()},
			},
			"recurring_schedule": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"timezone": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "The timezone in which to schedule the downtime.",
					},
				},
				Blocks: map[string]schema.Block{
					"recurrence": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"duration": schema.StringAttribute{
									Required:    true,
									Description: "The length of the downtime. Must begin with an integer and end with one of 'm', 'h', d', or 'w'.",
								},
								"rrule": schema.StringAttribute{
									Required:    true,
									Description: "The `RRULE` standard for defining recurring events. For example, to have a recurring event on the first day of each month, set the type to `rrule` and set the `FREQ` to `MONTHLY` and `BYMONTHDAY` to `1`. Most common `rrule` options from the [iCalendar Spec](https://tools.ietf.org/html/rfc5545) are supported.  **Note**: Attributes specifying the duration in `RRULE` are not supported (for example, `DTSTART`, `DTEND`, `DURATION`). More examples available in this [downtime guide](https://docs.datadoghq.com/monitors/guide/suppress-alert-with-downtimes/?tab=api).",
								},
								"start": schema.StringAttribute{
									Optional:    true,
									Computed:    true,
									Description: "ISO-8601 Datetime to start the downtime. Must not include a UTC offset. If not provided, the downtime starts the moment it is created.",
								},
							},
						},
					},
				},
				Validators:    []validator.Object{objectvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("one_time_schedule"))},
				PlanModifiers: []planmodifier.Object{planmodifiers.RemoveBlockModifier()},
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
	resp, httpResp, err := r.Api.GetDowntime(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving DowntimeSchedule"))
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

	body, diags := r.buildDowntimeScheduleCreateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateDowntime(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving DowntimeSchedule"))
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

	if displayTimezone, ok := attributes.GetDisplayTimezoneOk(); ok && displayTimezone != nil {
		state.DisplayTimezone = types.StringValue(*displayTimezone)
	}

	if message, ok := attributes.GetMessageOk(); ok && message != nil {
		state.Message = types.StringValue(*message)
	}

	if muteFirstRecoveryNotification, ok := attributes.GetMuteFirstRecoveryNotificationOk(); ok {
		state.MuteFirstRecoveryNotification = types.BoolValue(*muteFirstRecoveryNotification)
	}

	state.Scope = types.StringValue(attributes.GetScope())

	if notifyEndStates, ok := attributes.GetNotifyEndStatesOk(); ok && notifyEndStates != nil {
		state.NotifyEndStates, _ = types.SetValueFrom(ctx, types.StringType, *notifyEndStates)
	}

	if notifyEndTypes, ok := attributes.GetNotifyEndTypesOk(); ok && notifyEndTypes != nil {
		state.NotifyEndTypes, _ = types.SetValueFrom(ctx, types.StringType, *notifyEndTypes)
	}

	state.MonitorIdentifier = &MonitorIdentifierModel{}
	if attributes.MonitorIdentifier.DowntimeMonitorIdentifierId != nil {
		state.MonitorIdentifier.DowntimeMonitorIdentifierId = types.Int64Value(attributes.MonitorIdentifier.DowntimeMonitorIdentifierId.MonitorId)
		state.MonitorIdentifier.DowntimeMonitorIdentifierTags = types.SetNull(types.StringType)
	}
	if attributes.MonitorIdentifier.DowntimeMonitorIdentifierTags != nil {
		monitorTags := attributes.MonitorIdentifier.DowntimeMonitorIdentifierTags.MonitorTags
		state.MonitorIdentifier.DowntimeMonitorIdentifierTags, _ = types.SetValueFrom(ctx, types.StringType, monitorTags)
	}

	if schedule, ok := attributes.GetScheduleOk(); ok {

		if schedule.DowntimeScheduleRecurrencesResponse != nil {
			downtimeScheduleRecurrencesResponseTf := DowntimeScheduleRecurrenceSchedule{}
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

			state.DowntimeScheduleRecurrenceSchedule = &downtimeScheduleRecurrencesResponseTf
		}
		if schedule.DowntimeScheduleOneTimeResponse != nil {
			downtimeScheduleOneTimeResponseTf := DowntimeScheduleOneTimeSchedule{}
			if end, ok := schedule.DowntimeScheduleOneTimeResponse.GetEndOk(); ok && end != nil {
				downtimeScheduleOneTimeResponseTf.End = types.StringValue((*end).Format("2006-01-02T15:04:05Z"))
			}
			if start, ok := schedule.DowntimeScheduleOneTimeResponse.GetStartOk(); ok {
				downtimeScheduleOneTimeResponseTf.Start = types.StringValue((*start).Format("2006-01-02T15:04:05Z"))
			}

			state.DowntimeScheduleOneTimeSchedule = &downtimeScheduleOneTimeResponseTf
		}
	}
}

func (r *DowntimeScheduleResource) buildDowntimeScheduleCreateRequestBody(ctx context.Context, state *DowntimeScheduleModel) (*datadogV2.DowntimeCreateRequest, diag.Diagnostics) {
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

	if !state.NotifyEndStates.IsUnknown() {
		var notifyEndStates []datadogV2.DowntimeNotifyEndStateTypes
		diags.Append(state.NotifyEndStates.ElementsAs(ctx, &notifyEndStates, false)...)
		attributes.SetNotifyEndStates(notifyEndStates)
	}

	if !state.NotifyEndTypes.IsUnknown() {
		var notifyEndTypes []datadogV2.DowntimeNotifyEndStateActions
		diags.Append(state.NotifyEndTypes.ElementsAs(ctx, &notifyEndTypes, false)...)
		attributes.SetNotifyEndTypes(notifyEndTypes)
	}

	var monitorIdentifier datadogV2.DowntimeMonitorIdentifier

	if !state.MonitorIdentifier.DowntimeMonitorIdentifierId.IsNull() {
		var downtimeMonitorIdentifierId datadogV2.DowntimeMonitorIdentifierId

		downtimeMonitorIdentifierId.SetMonitorId(state.MonitorIdentifier.DowntimeMonitorIdentifierId.ValueInt64())

		monitorIdentifier.DowntimeMonitorIdentifierId = &downtimeMonitorIdentifierId
	} else if !state.MonitorIdentifier.DowntimeMonitorIdentifierTags.IsNull() {
		var downtimeMonitorIdentifierTags datadogV2.DowntimeMonitorIdentifierTags

		var monitorTags []string
		diags.Append(state.MonitorIdentifier.DowntimeMonitorIdentifierTags.ElementsAs(ctx, &monitorTags, false)...)
		downtimeMonitorIdentifierTags.SetMonitorTags(monitorTags)

		monitorIdentifier.DowntimeMonitorIdentifierTags = &downtimeMonitorIdentifierTags
	} else {
		diags.AddError("monitor_identifier.monitor_id or monitor_identifier.monitor_tags must be set", "")
	}

	attributes.MonitorIdentifier = monitorIdentifier

	var schedule datadogV2.DowntimeScheduleCreateRequest

	if state.DowntimeScheduleRecurrenceSchedule != nil {
		if len(state.DowntimeScheduleRecurrenceSchedule.Recurrences) == 0 {
			diags.AddError("Must provide one or more recurrence definitions", "")
		} else {
			var DowntimeScheduleRecurrenceSchedule datadogV2.DowntimeScheduleRecurrencesCreateRequest
			if !state.DowntimeScheduleRecurrenceSchedule.Timezone.IsNull() {
				DowntimeScheduleRecurrenceSchedule.SetTimezone(state.DowntimeScheduleRecurrenceSchedule.Timezone.ValueString())
			}

			if state.DowntimeScheduleRecurrenceSchedule.Recurrences != nil {
				var recurrences []datadogV2.DowntimeScheduleRecurrenceCreateUpdateRequest
				for _, recurrencesTFItem := range state.DowntimeScheduleRecurrenceSchedule.Recurrences {
					recurrencesDDItem := datadogV2.NewDowntimeScheduleRecurrenceCreateUpdateRequest(recurrencesTFItem.Start.ValueString(), recurrencesTFItem.Rrule.ValueString())

					recurrencesDDItem.SetDuration(recurrencesTFItem.Duration.ValueString())
					recurrencesDDItem.SetRrule(recurrencesTFItem.Rrule.ValueString())
					if !recurrencesTFItem.Start.IsUnknown() {
						recurrencesDDItem.SetStart(recurrencesTFItem.Start.ValueString())
					}
					recurrences = append(recurrences, *recurrencesDDItem)
				}
				DowntimeScheduleRecurrenceSchedule.SetRecurrences(recurrences)
			}
			schedule.DowntimeScheduleRecurrencesCreateRequest = &DowntimeScheduleRecurrenceSchedule
		}

	} else if state.DowntimeScheduleOneTimeSchedule != nil {
		var DowntimeScheduleOneTimeSchedule datadogV2.DowntimeScheduleOneTimeCreateUpdateRequest

		if !state.DowntimeScheduleOneTimeSchedule.End.IsUnknown() {
			if state.DowntimeScheduleOneTimeSchedule.End.IsNull() {
				DowntimeScheduleOneTimeSchedule.SetEndNil()
			} else {
				end, _ := time.Parse(time.RFC3339, state.DowntimeScheduleOneTimeSchedule.End.ValueString())
				DowntimeScheduleOneTimeSchedule.SetEnd(end)
			}
		}
		if !state.DowntimeScheduleOneTimeSchedule.Start.IsUnknown() {
			if state.DowntimeScheduleOneTimeSchedule.Start.IsNull() {
				DowntimeScheduleOneTimeSchedule.SetStartNil()
			} else {
				start, _ := time.Parse(time.RFC3339, state.DowntimeScheduleOneTimeSchedule.Start.ValueString())
				DowntimeScheduleOneTimeSchedule.SetStart(start)
			}
		}

		schedule.DowntimeScheduleOneTimeCreateUpdateRequest = &DowntimeScheduleOneTimeSchedule
	} else {
		diags.AddError("one_time_schedule or recurring_schedule must be set", "")
	}

	attributes.Schedule = &schedule

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
	} else {
		attributes.SetDisplayTimezoneNil()
	}

	if !state.Message.IsNull() {
		attributes.SetMessage(state.Message.ValueString())
	} else {
		attributes.SetMessageNil()
	}

	if !state.MuteFirstRecoveryNotification.IsNull() {
		attributes.SetMuteFirstRecoveryNotification(state.MuteFirstRecoveryNotification.ValueBool())
	} else {
		attributes.SetMuteFirstRecoveryNotification(false)
	}

	if !state.Scope.IsNull() {
		attributes.SetScope(state.Scope.ValueString())
	}

	if !state.NotifyEndStates.IsUnknown() {
		var notifyEndStates []datadogV2.DowntimeNotifyEndStateTypes
		diags.Append(state.NotifyEndStates.ElementsAs(ctx, &notifyEndStates, false)...)
		attributes.SetNotifyEndStates(notifyEndStates)
	}

	if !state.NotifyEndTypes.IsUnknown() {
		var notifyEndTypes []datadogV2.DowntimeNotifyEndStateActions
		diags.Append(state.NotifyEndTypes.ElementsAs(ctx, &notifyEndTypes, false)...)
		attributes.SetNotifyEndTypes(notifyEndTypes)
	}

	var monitorIdentifier datadogV2.DowntimeMonitorIdentifier

	if !state.MonitorIdentifier.DowntimeMonitorIdentifierId.IsNull() {
		var downtimeMonitorIdentifierId datadogV2.DowntimeMonitorIdentifierId

		downtimeMonitorIdentifierId.SetMonitorId(state.MonitorIdentifier.DowntimeMonitorIdentifierId.ValueInt64())

		monitorIdentifier.DowntimeMonitorIdentifierId = &downtimeMonitorIdentifierId
	} else if !state.MonitorIdentifier.DowntimeMonitorIdentifierTags.IsNull() {
		var downtimeMonitorIdentifierTags datadogV2.DowntimeMonitorIdentifierTags

		var monitorTags []string
		diags.Append(state.MonitorIdentifier.DowntimeMonitorIdentifierTags.ElementsAs(ctx, &monitorTags, false)...)
		downtimeMonitorIdentifierTags.SetMonitorTags(monitorTags)

		monitorIdentifier.DowntimeMonitorIdentifierTags = &downtimeMonitorIdentifierTags
	} else {
		diags.AddError("monitor_identifier.monitor_id or monitor_identifier.monitor_tags must be set", "")
	}

	attributes.MonitorIdentifier = &monitorIdentifier

	var schedule datadogV2.DowntimeScheduleUpdateRequest

	if state.DowntimeScheduleRecurrenceSchedule != nil {
		if len(state.DowntimeScheduleRecurrenceSchedule.Recurrences) == 0 {
			diags.AddError("Must provide one or more recurrence definitions", "")
		} else {
			var DowntimeScheduleRecurrenceSchedule datadogV2.DowntimeScheduleRecurrencesUpdateRequest

			if !state.DowntimeScheduleRecurrenceSchedule.Timezone.IsNull() {
				DowntimeScheduleRecurrenceSchedule.SetTimezone(state.DowntimeScheduleRecurrenceSchedule.Timezone.ValueString())
			}

			if state.DowntimeScheduleRecurrenceSchedule.Recurrences != nil {
				var recurrences []datadogV2.DowntimeScheduleRecurrenceCreateUpdateRequest
				for _, recurrencesTFItem := range state.DowntimeScheduleRecurrenceSchedule.Recurrences {
					recurrencesDDItem := datadogV2.NewDowntimeScheduleRecurrenceCreateUpdateRequest(recurrencesTFItem.Start.ValueString(), recurrencesTFItem.Rrule.ValueString())

					recurrencesDDItem.SetDuration(recurrencesTFItem.Duration.ValueString())
					recurrencesDDItem.SetRrule(recurrencesTFItem.Rrule.ValueString())
					if recurrencesTFItem.Start.IsUnknown() || recurrencesTFItem.Start.IsNull() {
						recurrencesDDItem.SetStartNil()
					} else {
						recurrencesDDItem.SetStart(recurrencesTFItem.Start.ValueString())
					}
					recurrences = append(recurrences, *recurrencesDDItem)
				}
				DowntimeScheduleRecurrenceSchedule.SetRecurrences(recurrences)
			}
			schedule.DowntimeScheduleRecurrencesUpdateRequest = &DowntimeScheduleRecurrenceSchedule
		}

	} else if state.DowntimeScheduleOneTimeSchedule != nil {
		var DowntimeScheduleOneTimeSchedule datadogV2.DowntimeScheduleOneTimeCreateUpdateRequest

		if state.DowntimeScheduleOneTimeSchedule.End.IsUnknown() || state.DowntimeScheduleOneTimeSchedule.End.IsNull() {
			DowntimeScheduleOneTimeSchedule.SetEndNil()
		} else {
			end, _ := time.Parse(time.RFC3339, state.DowntimeScheduleOneTimeSchedule.End.ValueString())
			DowntimeScheduleOneTimeSchedule.SetEnd(end)
		}

		if state.DowntimeScheduleOneTimeSchedule.Start.IsUnknown() || state.DowntimeScheduleOneTimeSchedule.Start.IsNull() {
			DowntimeScheduleOneTimeSchedule.SetStartNil()
		} else {
			start, _ := time.Parse(time.RFC3339, state.DowntimeScheduleOneTimeSchedule.Start.ValueString())
			DowntimeScheduleOneTimeSchedule.SetStart(start)
		}
		schedule.DowntimeScheduleOneTimeCreateUpdateRequest = &DowntimeScheduleOneTimeSchedule
	} else {
		diags.AddError("one_time_schedule or recurring_schedule must be set", "")
	}

	attributes.Schedule = &schedule

	req := datadogV2.NewDowntimeUpdateRequestWithDefaults()
	req.Data = *datadogV2.NewDowntimeUpdateRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}
