package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	// frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &monitorResource{}
	_ resource.ResourceWithImportState = &monitorResource{}
	_ resource.ResourceWithModifyPlan  = &monitorResource{}
)

type monitorResourceModel struct {
	ID                      types.String              `tfsdk:"id"`
	Name                    types.String              `tfsdk:"name"`
	Message                 types.String              `tfsdk:"message"`
	EscalationMessage       types.String              `tfsdk:"escalation_message"`
	Type                    types.String              `tfsdk:"type"`
	Query                   types.String              `tfsdk:"query"`
	Priority                types.Int64               `tfsdk:"priority"`
	NotifyNoData            types.Bool                `tfsdk:"notify_no_data"`
	OnMissingData           types.String              `tfsdk:"on_missing_data"`
	GroupRetentionDuration  types.String              `tfsdk:"group_retention_duration"`
	NewGroupDelay           types.Int64               `tfsdk:"new_group_delay"`
	NewHostDelay            types.Int64               `tfsdk:"new_host_delay"`
	EvaluationDelay         types.Int64               `tfsdk:"evaluation_delay"`
	NoDataTimeframe         types.Int64               `tfsdk:"no_data_timeframe"`
	RenotifyInterval        types.Int64               `tfsdk:"renotify_interval"`
	RenotifyOccurrences     types.Int64               `tfsdk:"renotify_occurrences"`
	RenotifyStatuses        types.Set                 `tfsdk:"renotify_statuses"`
	NotifyAudit             types.Bool                `tfsdk:"notify_audit"`
	TimeoutH                types.Int64               `tfsdk:"timeout_h"`
	RequireFullWindow       types.Bool                `tfsdk:"require_full_window"`
	Locked                  types.Bool                `tfsdk:"locked"`
	RestrictedRoles         types.Set                 `tfsdk:"restricted_roles"`
	IncludeTags             types.Bool                `tfsdk:"include_tags"`
	Tags                    types.Set                 `tfsdk:"tags"`
	GroupbySimpleMonitor    types.Bool                `tfsdk:"groupby_simple_monitor"`
	NotifyBy                types.Set                 `tfsdk:"notify_by"`
	EnableLogsSample        types.Bool                `tfsdk:"enable_logs_sample"`
	EnableSamples           types.Bool                `tfsdk:"enable_samples"`
	ForceDelete             types.Bool                `tfsdk:"force_delete"`
	Validate                types.Bool                `tfsdk:"validate"`
	NotificationPresetName  types.String              `tfsdk:"notification_preset_name"`
	SchedulingOptions       *MonitorSchedulingOptions `tfsdk:"scheduling_options"`
	MonitorThresholds       *MonitorThresholds        `tfsdk:"monitor_thresholds"`
	MonitorThresholdWindows *MonitorThresholdWindows  `tfsdk:"monitor_threshold_windows"`
}

type MonitorSchedulingOptions struct {
}

type MonitorThresholds struct {
	Ok               types.String `tfsdk:"ok"`
	Unknonw          types.String `tfsdk:"unknown"`
	Warning          types.String `tfsdk:"warning"`
	WarningRecovery  types.String `tfsdk:"warning_recovery"`
	Critical         types.String `tfsdk:"critical"`
	CriticalRecovery types.String `tfsdk:"critical_recovery"`
}

type MonitorThresholdWindows struct {
}

type monitorResource struct {
	Api  *datadogV1.MonitorsApi
	Auth context.Context
}

func NewMonitorResource() resource.Resource {
	return &monitorResource{}
}

func (r *monitorResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetMonitorsApiV1()
	r.Auth = providerData.Auth
}

func (r *monitorResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "monitor"
}

func (r *monitorResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog monitor resource. This can be used to create and manage Datadog monitors.",
		Attributes: map[string]schema.Attribute{
			// Resource ID
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Description: "Name of Datadog monitor.",
				Required:    true,
			},
			"message": schema.StringAttribute{
				Description: "A message to include with notifications for this monitor.\n\nEmail notifications can be sent to specific users by using the same `@username` notation as events.",
				Required:    true,
				// StateFunc: func(val interface{}) string {
				// 	return strings.TrimSpace(val.(string))
				// },
			},
			"escalation_message": schema.StringAttribute{
				Description: "A message to include with a re-notification. Supports the `@username` notification allowed elsewhere.",
				Optional:    true,
				// StateFunc: func(val interface{}) string {
				// 	return strings.TrimSpace(val.(string))
				// },
			},
			"query": schema.StringAttribute{
				Description: "The monitor query to notify on. Note this is not the same query you see in the UI and the syntax is different depending on the monitor type, please see the [API Reference](https://docs.datadoghq.com/api/v1/monitors/#create-a-monitor) for details. `terraform plan` will validate query contents unless `validate` is set to `false`.\n\n**Note:** APM latency data is now available as Distribution Metrics. Existing monitors have been migrated automatically but all terraformed monitors can still use the existing metrics. We strongly recommend updating monitor definitions to query the new metrics. To learn more, or to see examples of how to update your terraform definitions to utilize the new distribution metrics, see the [detailed doc](https://docs.datadoghq.com/tracing/guide/ddsketch_trace_metrics/).",
				Required:    true,
				// StateFunc: func(val interface{}) string {
				// 	return strings.TrimSpace(val.(string))
				// },
			},
			"type": schema.StringAttribute{
				Description: "The type of the monitor. The mapping from these types to the types found in the Datadog Web UI can be found in the Datadog API [documentation page](https://docs.datadoghq.com/api/v1/monitors/#create-a-monitor). Note: The monitor type cannot be changed after a monitor is created.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				// ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewMonitorTypeFromValue),
				// // Datadog API quirk, see https://github.com/hashicorp/terraform/issues/13784
				// DiffSuppressFunc: func(k, oldVal, newVal string, d *schema.ResourceData) bool {
				// 	if (oldVal == "query alert" && newVal == "metric alert") ||
				// 		(oldVal == "metric alert" && newVal == "query alert") {
				// 		log.Printf("[DEBUG] Monitor '%s' got a '%s' response for an expected '%s' type. Suppressing change.", d.Get("name"), newVal, oldVal)
				// 		return true
				// 	}
				// 	return newVal == oldVal
				// },
			},
			"priority": schema.Int64Attribute{
				Description: "Integer from 1 (high) to 5 (low) indicating alert severity.",
				Optional:    true,
				// todo validate?
			},
			"notify_no_data": schema.BoolAttribute{
				Description: "A boolean indicating whether this monitor will notify when data stops reporting.",
				Optional:    true,
				// Default:     booldefault.StaticBool(false), // TODO change from sdkv2, this is just reflecing the API, should be able to remove this
				// ConflictsWith: []string{"on_missing_data"},
			},
			"on_missing_data": schema.StringAttribute{
				Description: "Controls how groups or monitors are treated if an evaluation does not return any data points. The default option results in different behavior depending on the monitor query type. For monitors using `Count` queries, an empty monitor evaluation is treated as 0 and is compared to the threshold conditions. For monitors using any query type other than `Count`, for example `Gauge`, `Measure`, or `Rate`, the monitor shows the last known status. This option is only available for APM Trace Analytics, Audit Trail, CI, Error Tracking, Event, Logs, and RUM monitors. Valid values are: `show_no_data`, `show_and_notify_no_data`, `resolve`, and `default`.",
				Optional:    true,
				// ConflictsWith: []string{"notify_no_data", "no_data_timeframe"},
			},
			"group_retention_duration": schema.StringAttribute{
				Description: "The time span after which groups with missing data are dropped from the monitor state. The minimum value is one hour, and the maximum value is 72 hours. Example values are: 60m, 1h, and 2d. This option is only available for APM Trace Analytics, Audit Trail, CI, Error Tracking, Event, Logs, and RUM monitors.",
				Optional:    true,
			},
			// We only set new_group_delay in the monitor API payload if it is nonzero
			// because the SDKv2 terraform plugin API prevents unsetting new_group_delay
			// in updateMonitorState, so we can't reliably distinguish between new_group_delay
			// being unset (null) or set to zero.
			// Note that "new_group_delay overrides new_host_delay if it is set to a nonzero value"
			// refers to this terraform resource. In the API, setting new_group_delay
			// to any value, including zero, causes it to override new_host_delay.
			"new_group_delay": schema.Int64Attribute{
				Description: "The time (in seconds) to skip evaluations for new groups.\n\n`new_group_delay` overrides `new_host_delay` if it is set to a nonzero value.",
				Optional:    true,
			},
			"new_host_delay": schema.Int64Attribute{
				// Removing the default requires removing the default in the API as well (possibly only for
				// terraform user agents)
				Description:        "**Deprecated**. See `new_group_delay`. Time (in seconds) to allow a host to boot and applications to fully start before starting the evaluation of monitor results. Should be a non-negative integer. This value is ignored for simple monitors and monitors not grouped by host. The only case when this should be used is to override the default and set `new_host_delay` to zero for monitors grouped by host.",
				Optional:           true,
				Computed:           true, // TODO change from sdk v2 required to use default
				Default:            int64default.StaticInt64(300),
				DeprecationMessage: "Use `new_group_delay` except when setting `new_host_delay` to zero.",
			},
			"evaluation_delay": schema.Int64Attribute{
				Description: "(Only applies to metric alert) Time (in seconds) to delay evaluation, as a non-negative integer.\n\nFor example, if the value is set to `300` (5min), the `timeframe` is set to `last_5m` and the time is 7:00, the monitor will evaluate data from 6:50 to 6:55. This is useful for AWS CloudWatch and other backfilled metrics to ensure the monitor will always have data during evaluation.",
				Optional:    true,
			},
			"no_data_timeframe": schema.Int64Attribute{
				Description: "The number of minutes before a monitor will notify when data stops reporting.\n\nWe recommend at least 2x the monitor timeframe for metric alerts or 2 minutes for service checks.",
				Optional:    true,
				Computed:    true, // TODO change from sdk v2 required to use default
				Default:     int64default.StaticInt64(10),
				// DiffSuppressFunc: func(k, oldVal, newVal string, d *schema.ResourceData) bool {
				// 	if !d.Get("notify_no_data").(bool) {
				// 		if newVal != oldVal {
				// 			log.Printf("[DEBUG] Ignore the no_data_timeframe change of monitor '%s' because notify_no_data is false.", d.Get("name"))
				// 		}
				// 		return true
				// 	}
				// 	return newVal == oldVal
				// },
				// ConflictsWith: []string{"on_missing_data"},
			},
			"renotify_interval": schema.Int64Attribute{
				Description: "The number of minutes after the last notification before a monitor will re-notify on the current status. It will only re-notify if it's not resolved.",
				Optional:    true,
			},
			"renotify_occurrences": schema.Int64Attribute{
				Description: "The number of re-notification messages that should be sent on the current status.",
				Optional:    true,
			},
			"renotify_statuses": schema.SetAttribute{
				Description: "The types of statuses for which re-notification messages should be sent.",
				ElementType: types.StringType,
				// ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewMonitorRenotifyStatusTypeFromValue),
				Optional: true,
			},
			"notify_audit": schema.BoolAttribute{
				Description: "A boolean indicating whether tagged users will be notified on changes to this monitor. Defaults to `false`.",
				Optional:    true,
			},
			"timeout_h": schema.Int64Attribute{
				Description: "The number of hours of the monitor not reporting data before it automatically resolves from a triggered state. The minimum allowed value is 0 hours. The maximum allowed value is 24 hours.",
				Optional:    true,
			},
			"require_full_window": schema.BoolAttribute{
				Description: "A boolean indicating whether this monitor needs a full window of data before it's evaluated. Datadog strongly recommends you set this to `false` for sparse metrics, otherwise some evaluations may be skipped. If there's a custom_schedule set, `require_full_window` must be false and will be ignored.",
				Optional:    true,
				Computed:    true,                         // TODO change from sdk v2 required to use default
				Default:     booldefault.StaticBool(true), // TODO update this so that is true only if its not set for backward compatibility?
				// DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
				// 	if attr, ok := d.GetOk("scheduling_options"); ok {
				// 		scheduling_options_list := attr.([]interface{})
				// 		if scheduling_options_map, ok := scheduling_options_list[0].(map[string]interface{}); ok {
				// 			custom_schedule_map, custom_schedule_found := scheduling_options_map["custom_schedule"].([]interface{})
				// 			if custom_schedule_found && len(custom_schedule_map) > 0 {
				// 				return true
				// 			}
				// 		}
				// 	}
				// 	return false
				// },
			},
			"locked": schema.BoolAttribute{
				Description:        "A boolean indicating whether changes to this monitor should be restricted to the creator or admins. Defaults to `false`.",
				Optional:           true,
				DeprecationMessage: "Use `restricted_roles`.",
				// ConflictsWith: []string{"restricted_roles"},
				// DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
				// 	// if restricted_roles is defined, ignore locked
				// 	if _, ok := d.GetOk("restricted_roles"); ok {
				// 		return true
				// 	}
				// 	return false
				// },
			},
			"restricted_roles": schema.SetAttribute{
				Description: "A list of unique role identifiers to define which roles are allowed to edit the monitor. Editing a monitor includes any updates to the monitor configuration, monitor deletion, and muting of the monitor for any amount of time. Roles unique identifiers can be pulled from the [Roles API](https://docs.datadoghq.com/api/latest/roles/#list-roles) in the `data.id` field.",
				Optional:    true,
				ElementType: types.StringType,
				// ConflictsWith: []string{"locked"},
			},
			"include_tags": schema.BoolAttribute{
				Description: "A boolean indicating whether notifications from this monitor automatically insert its triggering tags into the title.",
				Optional:    true,
				Computed:    true, // TODO change from sdk v2 required to use default
				Default:     booldefault.StaticBool(true),
			},
			"tags": schema.SetAttribute{
				Description: "A list of tags to associate with your monitor. This can help you categorize and filter monitors in the manage monitors page of the UI. Note: it's not currently possible to filter by these tags when querying via the API",
				// we use TypeSet to represent tags, paradoxically to be able to maintain them ordered;
				// we order them explicitly in the read/create/update methods of this resource and using
				// TypeSet makes Terraform ignore differences in order when creating a plan
				Optional: true,
				// Computed:    true, // diff from sdkv2
				ElementType: types.StringType,
			},
			"groupby_simple_monitor": schema.BoolAttribute{
				Description: "Whether or not to trigger one alert if any source breaches a threshold. This is only used by log monitors. Defaults to `false`.",
				Optional:    true,
			},
			"notify_by": schema.SetAttribute{
				Description: "Controls what granularity a monitor alerts on. Only available for monitors with groupings. For instance, a monitor grouped by `cluster`, `namespace`, and `pod` can be configured to only notify on each new `cluster` violating the alert conditions by setting `notify_by` to `['cluster']`. Tags mentioned in `notify_by` must be a subset of the grouping tags in the query. For example, a query grouped by `cluster` and `namespace` cannot notify on `region`. Setting `notify_by` to `[*]` configures the monitor to notify as a simple-alert.",
				Optional:    true,
				ElementType: types.StringType,
			},
			// since this is only useful for "log alert" type, we don't set a default value
			// if we did set it, it would be used for all types; we have to handle this manually
			// throughout the code
			"enable_logs_sample": schema.BoolAttribute{
				Description: "A boolean indicating whether or not to include a list of log values which triggered the alert. This is only used by log monitors. Defaults to `false`.",
				Optional:    true,
			},
			"enable_samples": schema.BoolAttribute{
				Description: "Whether or not a list of samples which triggered the alert is included. This is only used by CI Test and Pipeline monitors.",
				Optional:    true, // diff from sdkv2
				// Computed:    true,
			},
			"force_delete": schema.BoolAttribute{
				Description: "A boolean indicating whether this monitor can be deleted even if it’s referenced by other resources (e.g. SLO, composite monitor).",
				Optional:    true,
			},
			"validate": schema.BoolAttribute{
				Description: "If set to `false`, skip the validation call done during plan.",
				Optional:    true,
				// DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
				// 	// This is never sent to the backend, so it should never generate a diff
				// 	return true
				// },
			},
			"notification_preset_name": schema.StringAttribute{
				Description: "Toggles the display of additional content sent in the monitor notification.",
				Optional:    true,
				// ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewMonitorOptionsNotificationPresetsFromValue),
			},
		},
		Blocks: map[string]schema.Block{
			"monitor_thresholds": schema.SingleNestedBlock{
				Description: "Alert thresholds of the monitor.",
				Attributes: map[string]schema.Attribute{
					"ok": schema.StringAttribute{
						Description: "The monitor `OK` threshold. Only supported in monitor type `service check`. Must be a number.",
						// ValidateFunc: validators.ValidateFloatString,
						Optional: true,
						// DiffSuppressFunc: func(k, oldVal, newVal string, d *schema.ResourceData) bool {
						// 	monitorType := d.Get("type").(string)
						// 	return monitorType != string(datadogV1.MONITORTYPE_SERVICE_CHECK)
						// },
					},
					"warning": schema.StringAttribute{
						Description: "The monitor `WARNING` threshold. Must be a number.",
						// ValidateFunc: validators.ValidateFloatString,
						Optional: true,
					},
					"critical": schema.StringAttribute{
						Description: "The monitor `CRITICAL` threshold. Must be a number.",
						// ValidateFunc: validators.ValidateFloatString,
						Optional: true,
					},
					"unknown": schema.StringAttribute{
						Description: "The monitor `UNKNOWN` threshold. Only supported in monitor type `service check`. Must be a number.",
						// ValidateFunc: validators.ValidateFloatString,
						Optional: true,
						// DiffSuppressFunc: func(k, oldVal, newVal string, d *schema.ResourceData) bool {
						// 	monitorType := d.Get("type").(string)
						// 	return monitorType != string(datadogV1.MONITORTYPE_SERVICE_CHECK)
						// },
					},
					"warning_recovery": schema.StringAttribute{
						Description: "The monitor `WARNING` recovery threshold. Must be a number.",
						// ValidateFunc: validators.ValidateFloatString,
						Optional: true,
					},
					"critical_recovery": schema.StringAttribute{
						Description: "The monitor `CRITICAL` recovery threshold. Must be a number.",
						// ValidateFunc: validators.ValidateFloatString,
						Optional: true,
					},
				},
				// DiffSuppressFunc: suppressDataDogFloatIntDiff,
			},
			"monitor_threshold_windows": schema.SingleNestedBlock{
				Description: "A mapping containing `recovery_window` and `trigger_window` values, e.g. `last_15m` . Can only be used for, and are required for, anomaly monitors.",
				Attributes: map[string]schema.Attribute{
					"recovery_window": schema.StringAttribute{
						Description: "Describes how long an anomalous metric must be normal before the alert recovers.",
						Optional:    true,
					},
					"trigger_window": schema.StringAttribute{
						Description: "Describes how long a metric must be anomalous before an alert triggers.",
						Optional:    true,
					},
				},
			},
			// TODO "variables": getMonitorFormulaQuerySchema(),
			"scheduling_options": schema.SingleNestedBlock{
				Description: "Configuration options for scheduling.",
				Blocks: map[string]schema.Block{
					"evaluation_window": schema.SingleNestedBlock{
						Description: "Configuration options for the evaluation window. If `hour_starts` is set, no other fields may be set. Otherwise, `day_starts` and `month_starts` must be set together.",
						Attributes: map[string]schema.Attribute{
							"day_starts": schema.StringAttribute{
								Description: "The time of the day at which a one day cumulative evaluation window starts. Must be defined in UTC time in `HH:mm` format.",
								Optional:    true,
							},
							"month_starts": schema.Int64Attribute{
								Description: "The day of the month at which a one month cumulative evaluation window starts. Must be a value of 1.",
								Optional:    true,
							},
							"hour_starts": schema.Int64Attribute{
								Description: "The minute of the hour at which a one hour cumulative evaluation window starts. Must be between 0 and 59.",
								Optional:    true,
							},
						},
					},
					"custom_schedule": schema.SingleNestedBlock{
						Description: "Configuration options for the custom schedules. If `start` is omitted, the monitor creation time will be used.",
						Blocks: map[string]schema.Block{
							"recurrence": schema.SingleNestedBlock{
								Description: "A list of recurrence definitions. Length must be 1.",
								// todo Required:    true,
								Attributes: map[string]schema.Attribute{
									"rrule": schema.StringAttribute{
										Description: "Must be a valid `rrule`. See API docs for supported fields",
										// TODO Required:    true,
										Optional: true,
									},
									"start": schema.StringAttribute{
										Description: "Time to start recurrence cycle. Similar to DTSTART. Expected format 'YYYY-MM-DDThh:mm:ss'",
										Optional:    true,
									},
									"timezone": schema.StringAttribute{
										Description: "'tz database' format. Example: `America/New_York` or `UTC`",
										// TODO Required:    true,
										Optional: true,
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

func (r *monitorResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state monitorResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	m, diags := r.buildMonitorCreateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateMonitor(r.Auth, *m)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving DowntimeSchedule"))
		return
	}
	diags = r.updateState(&state, &resp)
	response.Diagnostics.Append(diags...)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)

}

func (r *monitorResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state monitorResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	stateId := state.ID.ValueString()
	id, err := strconv.ParseInt(stateId, 10, 64)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error on monitor id"))
		return
	}
	resp, httpResp, err := r.Api.GetMonitor(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving Monitor"))
		return
	}

	diags := r.updateState(&state, &resp)
	response.Diagnostics.Append(diags...)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *monitorResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state monitorResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	stateId := state.ID.ValueString()
	id, err := strconv.ParseInt(stateId, 10, 64)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error on monitor id"))
		return
	}

	body, diags := r.buildMonitorUpdateRequestBody(&state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateMonitor(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving DowntimeSchedule"))
		return
	}
	diags = r.updateState(&state, &resp)
	response.Diagnostics.Append(diags...)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *monitorResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state monitorResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	stateId := state.ID.ValueString()
	id, err := strconv.ParseInt(stateId, 10, 64)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error on monitor id"))
		return
	}

	_, httpResp, err := r.Api.DeleteMonitor(r.Auth, id) // todo use force
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting downtime_schedule"))
		return
	}
}

func (r *monitorResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	// response.Diagnostics.AddWarning(
	// 	"Deprecated",
	// 	"The import functionality for datadog_api_key resources is deprecated and will be removed in a future release with prior notice. Securely store your API keys using a secret management system or use the datadog_api_key resource to create and manage new API keys.",
	// )
	// resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *monitorResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Check if the resource is being destroyed.
	if req.Plan.Raw.IsNull() {
		return
	}

	var state, config, plan monitorResourceModel
	var isCreation bool

	if req.State.Raw.IsNull() {
		isCreation = true
	} else {
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	}

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	var httpresp *http.Response
	var err error

	if !isCreation {
		m, diags := r.buildMonitorCreateRequestBody(ctx, &plan)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		var id int64
		stateId := state.ID.ValueString()
		id, err = strconv.ParseInt(stateId, 10, 64)
		if err != nil {
			resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error on monitor id"))
			return
		}

		_, httpresp, err = r.Api.ValidateExistingMonitor(r.Auth, id, *m)
	} else {
		m, diags := r.buildMonitorCreateRequestBody(ctx, &plan)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		log.Printf("[DEBUG] monitor/validate m=%#v", m)
		_, httpresp, err = r.Api.ValidateMonitor(r.Auth, *m)
	}
	if err != nil {
		if httpresp != nil && (httpresp.StatusCode == 502 || httpresp.StatusCode == 504) {
			resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error validating monitor, retrying"))
		}
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error validating monitor"))
		return
	}

	j := make(map[string]interface{})
	// err = utils.GetMetadataFromJSON([]byte(stepParamsElement.(string)), &validation)
	err = json.NewDecoder(httpresp.Body).Decode(&j)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error on validation decode"))
		return
	}
	if meta, ok := j["meta"].(map[string]interface{}); ok {
		log.Printf("[DEBUG] monitor/validate meta=%+v", meta)
		if quality_issues, ok := meta["quality_issues"].([]interface{}); ok && len(quality_issues) > 0 {
			resp.Diagnostics.AddWarning(
				"Monitor Quality Issue",
				"Found the following quality issues: "+fmt.Sprintf("%v", quality_issues),
			)
		}
	}
}

func (r *monitorResource) buildMonitorCreateRequestBody(ctx context.Context, state *monitorResourceModel) (*datadogV1.Monitor, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	monitorType := datadogV1.MonitorType(state.Type.ValueString())
	m := datadogV1.NewMonitor(state.Query.ValueString(), monitorType)
	m.SetName(state.Name.ValueString())
	m.SetMessage(state.Message.ValueString())
	m.SetPriority(state.Priority.ValueInt64())

	var tags []string
	if !state.Tags.IsNull() {
		diags.Append(state.Tags.ElementsAs(ctx, &tags, false)...)
	}
	m.SetTags(tags)

	return m, diags
}

func (r *monitorResource) buildMonitorUpdateRequestBody(state *monitorResourceModel) (*datadogV1.MonitorUpdateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	u := datadogV1.NewMonitorUpdateRequest()

	monitorType := datadogV1.MonitorType(state.Type.ValueString())
	u.SetType(monitorType)
	u.SetQuery(state.Query.ValueString())
	u.SetName(state.Name.ValueString())
	u.SetMessage(state.Message.ValueString())
	u.SetPriority(state.Priority.ValueInt64())
	// u.SetOptions(o)

	return u, diags
}

func (r *monitorResource) updateState(state *monitorResourceModel, m *datadogV1.Monitor) diag.Diagnostics {
	diags := diag.Diagnostics{}

	if id, ok := m.GetIdOk(); ok && id != nil {
		state.ID = types.StringValue(strconv.FormatInt(*id, 10))
	}

	if name, ok := m.GetNameOk(); ok && name != nil {
		state.Name = types.StringValue(*name)
	}

	if message, ok := m.GetMessageOk(); ok && message != nil {
		state.Message = types.StringValue(*message)
	}

	if mType, ok := m.GetTypeOk(); ok && mType != nil {
		state.Type = types.StringValue(string(*mType))
	}

	if query, ok := m.GetQueryOk(); ok && query != nil {
		state.Query = types.StringValue(*query)
	}

	if priority, ok := m.GetPriorityOk(); ok && priority != nil {
		state.Priority = types.Int64Value(*priority)
	}

	if escalationMessage, ok := m.Options.GetEscalationMessageOk(); ok && escalationMessage != nil {
		state.EscalationMessage = types.StringValue(*escalationMessage)
	}

	if onMissingData, ok := m.Options.GetOnMissingDataOk(); ok && onMissingData != nil {
		state.OnMissingData = types.StringValue(string(*onMissingData))
	}

	if groupRetentionDuration, ok := m.Options.GetGroupRetentionDurationOk(); ok && groupRetentionDuration != nil {
		state.GroupRetentionDuration = types.StringValue(*groupRetentionDuration)
	}

	if notificationPresetName, ok := m.Options.GetNotificationPresetNameOk(); ok && notificationPresetName != nil {
		state.NotificationPresetName = types.StringValue(string(*notificationPresetName))
	}

	if evaluationDelay, ok := m.Options.GetEvaluationDelayOk(); ok && evaluationDelay != nil {
		state.EvaluationDelay = types.Int64Value(*evaluationDelay)
	}

	// NotifyNoData           types.Bool   `tfsdk:"notify_no_data"`
	// NewGroupDelay          types.Int64  `tfsdk:"new_group_delay"`
	// NewHostDelay           types.Int64  `tfsdk:"new_host_delay"`
	// NoDataTimeframe        types.Int64  `tfsdk:"no_data_timeframe"`
	// RenotifyInterval       types.Int64  `tfsdk:"renotify_interval"`
	// RenotifyOccurrences    types.Int64  `tfsdk:"renotify_occurrences"`
	// RenotifyStatuses       types.Set    `tfsdk:"renotify_statuses"`
	// NotifyAudit            types.Bool   `tfsdk:"notify_audit"`
	// TimeoutH               types.Int64  `tfsdk:"timeout_h"`
	// RequireFullWindow      types.Bool   `tfsdk:"require_full_window"`
	// Locked                 types.Bool   `tfsdk:"locked"`
	// RestrictedRoles        types.Set    `tfsdk:"restricted_roles"`
	// IncludeTags            types.Bool   `tfsdk:"include_tags"`
	// Tags                   types.Set    `tfsdk:"tags"`
	// GroupbySimpleMonitor   types.Bool   `tfsdk:"groupby_simple_monitor"`
	// NotifyBy               types.Set    `tfsdk:"notify_by"`
	// EnableLogsSample       types.Bool   `tfsdk:"enable_logs_sample"`
	// EnableSamples          types.Bool   `tfsdk:"enable_samples"`
	// ForceDelete            types.Bool   `tfsdk:"force_delete"`
	// Validate               types.Bool   `tfsdk:"validate"`

	return diags
}
