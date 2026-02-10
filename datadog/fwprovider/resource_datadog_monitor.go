package fwprovider

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"

	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/customtypes"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/fwutils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure        = &monitorResource{}
	_ resource.ResourceWithImportState      = &monitorResource{}
	_ resource.ResourceWithModifyPlan       = &monitorResource{}
	_ resource.ResourceWithConfigValidators = &monitorResource{}
)

var stringFloatValidator = stringvalidator.RegexMatches(
	regexp.MustCompile(`\d*(\.\d*)?`), "value must be a float")

func enumStrings[E ~string](vals []E) []string {
	res := make([]string, len(vals))
	for i, v := range vals {
		res[i] = string(v)
	}
	return res
}

type monitorResourceModel struct {
	ID                      types.String                     `tfsdk:"id"`
	Name                    types.String                     `tfsdk:"name"`
	Message                 customtypes.TrimSpaceStringValue `tfsdk:"message"`
	EscalationMessage       customtypes.TrimSpaceStringValue `tfsdk:"escalation_message"`
	Type                    customtypes.MonitorTypeValue     `tfsdk:"type"`
	Query                   customtypes.TrimSpaceStringValue `tfsdk:"query"`
	Priority                types.String                     `tfsdk:"priority"`
	Tags                    types.Set                        `tfsdk:"tags"`
	EffectiveTags           types.Set                        `tfsdk:"effective_tags"`
	NotifyNoData            types.Bool                       `tfsdk:"notify_no_data"`
	OnMissingData           types.String                     `tfsdk:"on_missing_data"`
	GroupRetentionDuration  types.String                     `tfsdk:"group_retention_duration"`
	NewGroupDelay           types.Int64                      `tfsdk:"new_group_delay"`
	NewHostDelay            types.Int64                      `tfsdk:"new_host_delay"`
	EvaluationDelay         types.Int64                      `tfsdk:"evaluation_delay"`
	NoDataTimeframe         types.Int64                      `tfsdk:"no_data_timeframe"`
	RenotifyInterval        types.Int64                      `tfsdk:"renotify_interval"`
	RenotifyOccurrences     types.Int64                      `tfsdk:"renotify_occurrences"`
	RenotifyStatuses        types.Set                        `tfsdk:"renotify_statuses"`
	NotifyAudit             types.Bool                       `tfsdk:"notify_audit"`
	TimeoutH                types.Int64                      `tfsdk:"timeout_h"`
	RequireFullWindow       types.Bool                       `tfsdk:"require_full_window"`
	Locked                  types.Bool                       `tfsdk:"locked"`
	RestrictedRoles         types.Set                        `tfsdk:"restricted_roles"`
	IncludeTags             types.Bool                       `tfsdk:"include_tags"`
	GroupbySimpleMonitor    types.Bool                       `tfsdk:"groupby_simple_monitor"`
	NotifyBy                types.Set                        `tfsdk:"notify_by"`
	EnableLogsSample        types.Bool                       `tfsdk:"enable_logs_sample"`
	EnableSamples           types.Bool                       `tfsdk:"enable_samples"`
	ForceDelete             types.Bool                       `tfsdk:"force_delete"`
	Validate                types.Bool                       `tfsdk:"validate"`
	NotificationPresetName  types.String                     `tfsdk:"notification_preset_name"`
	MonitorThresholds       []MonitorThreshold               `tfsdk:"monitor_thresholds"`
	MonitorThresholdWindows []MonitorThresholdWindow         `tfsdk:"monitor_threshold_windows"`
	SchedulingOptions       []SchedulingOption               `tfsdk:"scheduling_options"`
	Variables               []Variable                       `tfsdk:"variables"`
	DraftStatus             types.String                     `tfsdk:"draft_status"`
	MonitorAssets           []MonitorAsset                   `tfsdk:"assets"`
}

type MonitorAsset struct {
	Name         types.String `tfsdk:"name"`
	Url          types.String `tfsdk:"url"`
	Category     types.String `tfsdk:"category"`
	ResourceKey  types.String `tfsdk:"resource_key"`
	ResourceType types.String `tfsdk:"resource_type"`
}

type MonitorThreshold struct {
	Ok               customtypes.FloatStringValue `tfsdk:"ok"`
	Unknown          customtypes.FloatStringValue `tfsdk:"unknown"`
	Warning          customtypes.FloatStringValue `tfsdk:"warning"`
	WarningRecovery  customtypes.FloatStringValue `tfsdk:"warning_recovery"`
	Critical         customtypes.FloatStringValue `tfsdk:"critical"`
	CriticalRecovery customtypes.FloatStringValue `tfsdk:"critical_recovery"`
}

type MonitorThresholdWindow struct {
	RecoveryWindow types.String `tfsdk:"recovery_window"`
	TriggerWindow  types.String `tfsdk:"trigger_window"`
}

type SchedulingOption struct {
	EvaluationWindow []EvaluationWindow `tfsdk:"evaluation_window"`
	CustomSchedule   []CustomSchedule   `tfsdk:"custom_schedule"`
}

type EvaluationWindow struct {
	DayStarts   types.String `tfsdk:"day_starts"`
	MonthStarts types.Int32  `tfsdk:"month_starts"`
	HourStarts  types.Int32  `tfsdk:"hour_starts"`
	Timezone    types.String `tfsdk:"timezone"`
}

type CustomSchedule struct {
	Recurrence []Recurrence `tfsdk:"recurrence"`
}

type Recurrence struct {
	Rrule    types.String `tfsdk:"rrule"`
	Start    types.String `tfsdk:"start"`
	Timezone types.String `tfsdk:"timezone"`
}

type Variable struct {
	EventQuery       []EventQuery       `tfsdk:"event_query"`
	CloudCostQuery   []CloudCostQuery   `tfsdk:"cloud_cost_query"`
	DataQualityQuery []DataQualityQuery `tfsdk:"data_quality_query"`
}

type EventQuery struct {
	DataSource types.String `tfsdk:"data_source"`
	Indexes    types.List   `tfsdk:"indexes"`
	Name       types.String `tfsdk:"name"`
	Search     []Search     `tfsdk:"search"`
	Compute    []Compute    `tfsdk:"compute"`
	GroupBy    []GroupBy    `tfsdk:"group_by"`
}

type Search struct {
	Query types.String `tfsdk:"query"`
}

type Compute struct {
	Aggregation types.String `tfsdk:"aggregation"`
	Interval    types.Int64  `tfsdk:"interval"`
	Metric      types.String `tfsdk:"metric"`
}

type GroupBy struct {
	Facet types.String `tfsdk:"facet"`
	Limit types.Int64  `tfsdk:"limit"`
	Sort  []Sort       `tfsdk:"sort"`
}

type Sort struct {
	Aggregation types.String `tfsdk:"aggregation"`
	Metric      types.String `tfsdk:"metric"`
	Order       types.String `tfsdk:"order"`
}

type CloudCostQuery struct {
	DataSource types.String `tfsdk:"data_source"`
	Query      types.String `tfsdk:"query"`
	Aggregator types.String `tfsdk:"aggregator"`
	Name       types.String `tfsdk:"name"`
}

type DataQualityQuery struct {
	Name           types.String                `tfsdk:"name"`
	DataSource     types.String                `tfsdk:"data_source"`
	SchemaVersion  types.String                `tfsdk:"schema_version"`
	Measure        types.String                `tfsdk:"measure"`
	Filter         types.String                `tfsdk:"filter"`
	Scope          types.String                `tfsdk:"scope"`
	GroupBy        types.List                  `tfsdk:"group_by"`
	MonitorOptions []DataQualityMonitorOptions `tfsdk:"monitor_options"`
}

type DataQualityMonitorOptions struct {
	CustomSql         types.String `tfsdk:"custom_sql"`
	CustomWhere       types.String `tfsdk:"custom_where"`
	GroupByColumns    types.List   `tfsdk:"group_by_columns"`
	CrontabOverride   types.String `tfsdk:"crontab_override"`
	ModelTypeOverride types.String `tfsdk:"model_type_override"`
}

type monitorResource struct {
	Api         *datadogV1.MonitorsApi
	Auth        context.Context
	DefaultTags map[string]string
}

func NewMonitorResource() resource.Resource {
	return &monitorResource{}
}

func (r *monitorResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.Conflicting(
			frameworkPath.MatchRoot("notify_no_data"),
			frameworkPath.MatchRoot("on_missing_data"),
		),
		resourcevalidator.Conflicting(
			frameworkPath.MatchRoot("on_missing_data"),
			frameworkPath.MatchRoot("no_data_timeframe"),
		),
		resourcevalidator.Conflicting(
			frameworkPath.MatchRoot("locked"),
			frameworkPath.MatchRoot("restricted_roles"),
		),
		resourcevalidator.Conflicting(
			frameworkPath.MatchRoot("scheduling_options").AtAnyListIndex().
				AtName("evaluation_window").AtAnyListIndex().
				AtName("hour_starts"),
			frameworkPath.MatchRoot("scheduling_options").AtAnyListIndex().
				AtName("evaluation_window").AtAnyListIndex().
				AtName("day_starts"),
		),
		resourcevalidator.Conflicting(
			frameworkPath.MatchRoot("scheduling_options").AtAnyListIndex().
				AtName("evaluation_window").AtAnyListIndex().
				AtName("hour_starts"),
			frameworkPath.MatchRoot("scheduling_options").AtAnyListIndex().
				AtName("evaluation_window").AtAnyListIndex().
				AtName("month_starts"),
		),
		resourcevalidator.Conflicting(
			frameworkPath.MatchRoot("no_data_timeframe"),
			frameworkPath.MatchRoot("scheduling_options").AtAnyListIndex().
				AtName("custom_schedule"),
		),
		resourcevalidator.RequiredTogether(
			frameworkPath.MatchRoot("scheduling_options").AtAnyListIndex().
				AtName("evaluation_window").AtAnyListIndex().
				AtName("day_starts"),
			frameworkPath.MatchRoot("scheduling_options").AtAnyListIndex().
				AtName("evaluation_window").AtAnyListIndex().
				AtName("month_starts"),
		),
	}
}

func (r *monitorResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetMonitorsApiV1()
	r.Auth = providerData.Auth
	r.DefaultTags = providerData.DefaultTags
}

func (r *monitorResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "monitor"
}

func (r *monitorResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog monitor resource. This can be used to create and manage Datadog monitors.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Description: "Name of Datadog monitor.",
				Required:    true,
			},
			"message": schema.StringAttribute{
				Description: "A message to include with notifications for this monitor.\n\nEmail notifications can be sent to specific users by using the same `@username` notation as events.",
				Required:    true,
				CustomType:  customtypes.TrimSpaceStringType{},
			},
			"query": schema.StringAttribute{
				Description: "The monitor query to notify on. Note this is not the same query you see in the UI and the syntax is different depending on the monitor type, please see the [API Reference](https://docs.datadoghq.com/api/v1/monitors/#create-a-monitor) for details. `terraform plan` will validate query contents unless `validate` is set to `false`.\n\n**Note:** APM latency data is now available as Distribution Metrics. Existing monitors have been migrated automatically but all terraformed monitors can still use the existing metrics. We strongly recommend updating monitor definitions to query the new metrics. To learn more, or to see examples of how to update your terraform definitions to utilize the new distribution metrics, see the [detailed doc](https://docs.datadoghq.com/tracing/guide/ddsketch_trace_metrics/).",
				Required:    true,
				CustomType:  customtypes.TrimSpaceStringType{},
			},
			"type": schema.StringAttribute{
				Description: "The type of the monitor. The mapping from these types to the types found in the Datadog Web UI can be found in the Datadog API [documentation page](https://docs.datadoghq.com/api/v1/monitors/#create-a-monitor). Note: The monitor type cannot be changed after a monitor is created.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(r.getAllowTypes()...),
				},
				// Datadog API quirk, see https://github.com/hashicorp/terraform/issues/13784
				CustomType: customtypes.MonitorTypeType{},
				// Due to the API quirk mentioned above, will mute replace resource, when user tries to change type from
				// metric alert to query alert
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(func(_ context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
						oldType := req.StateValue.ValueString()
						newType := req.PlanValue.ValueString()
						if (oldType == "metric alert" && newType == "query alert") ||
							oldType == "query alert" && newType == "metric alert" {
							return
						}
						resp.RequiresReplace = true
					}, "", ""),
				},
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
				Optional:    true,
			},
			"escalation_message": schema.StringAttribute{
				Description: "A message to include with a re-notification. Supports the `@username` notification allowed elsewhere.",
				Optional:    true,
				CustomType:  customtypes.TrimSpaceStringType{},
			},
			"evaluation_delay": schema.Int64Attribute{
				Description: "(Only applies to metric alert) Time (in seconds) to delay evaluation, as a non-negative integer.\n\nFor example, if the value is set to `300` (5min), the `timeframe` is set to `last_5m` and the time is 7:00, the monitor will evaluate data from 6:50 to 6:55. This is useful for AWS CloudWatch and other backfilled metrics to ensure the monitor will always have data during evaluation.",
				Computed:    true,
				Optional:    true,
			},
			"force_delete": schema.BoolAttribute{
				Description: "A boolean indicating whether this monitor can be deleted even if it’s referenced by other resources (e.g. SLO, composite monitor).",
				Optional:    true,
			},
			"groupby_simple_monitor": schema.BoolAttribute{
				Description: "Whether or not to trigger one alert if any source breaches a threshold. This is only used by log monitors. Defaults to `false`.",
				Optional:    true,
				Computed:    true,
			},
			"group_retention_duration": schema.StringAttribute{
				Description: "The time span after which groups with missing data are dropped from the monitor state. The minimum value is one hour, and the maximum value is 72 hours. Example values are: 60m, 1h, and 2d. This option is only available for APM Trace Analytics, Audit Trail, CI, Error Tracking, Event, Logs, and RUM monitors.",
				Optional:    true,
			},
			"include_tags": schema.BoolAttribute{
				Description: "A boolean indicating whether notifications from this monitor automatically insert its triggering tags into the title.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"locked": schema.BoolAttribute{
				MarkdownDescription: "A boolean indicating whether changes to this monitor should be restricted to the creator or admins. Defaults to `false`.",
				Optional:            true,
				DeprecationMessage:  "Use `restricted_roles`.",
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
				MarkdownDescription: "**Deprecated**. See `new_group_delay`. Time (in seconds) to allow a host to boot and applications to fully start before starting the evaluation of monitor results. Should be a non-negative integer. This value is ignored for simple monitors and monitors not grouped by host. The only case when this should be used is to override the default and set `new_host_delay` to zero for monitors grouped by host.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(300),
				DeprecationMessage:  "Use `new_group_delay` except when setting `new_host_delay` to zero.",
			},
			"no_data_timeframe": schema.Int64Attribute{
				MarkdownDescription: "The number of minutes before a monitor will notify when data stops reporting.\n\nWe recommend at least 2x the monitor timeframe for metric alerts or 2 minutes for service checks.",
				Optional:            true,
			},
			"notification_preset_name": schema.StringAttribute{
				Description: "Toggles the display of additional content sent in the monitor notification.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf(r.getAllowMonitorOptionsNotificationPresets()...),
				},
			},
			"notify_audit": schema.BoolAttribute{
				MarkdownDescription: "A boolean indicating whether tagged users will be notified on changes to this monitor. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"notify_by": schema.SetAttribute{
				MarkdownDescription: "Controls what granularity a monitor alerts on. Only available for monitors with groupings. For instance, a monitor grouped by `cluster`, `namespace`, and `pod` can be configured to only notify on each new `cluster` violating the alert conditions by setting `notify_by` to `['cluster']`. Tags mentioned in `notify_by` must be a subset of the grouping tags in the query. For example, a query grouped by `cluster` and `namespace` cannot notify on `region`. Setting `notify_by` to `[*]` configures the monitor to notify as a simple-alert.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"notify_no_data": schema.BoolAttribute{
				Description: "A boolean indicating whether this monitor will notify when data stops reporting.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"on_missing_data": schema.StringAttribute{
				MarkdownDescription: "Controls how groups or monitors are treated if an evaluation does not return any data points. The default option results in different behavior depending on the monitor query type. For monitors using `Count` queries, an empty monitor evaluation is treated as 0 and is compared to the threshold conditions. For monitors using any query type other than `Count`, for example `Gauge`, `Measure`, or `Rate`, the monitor shows the last known status. This option is not available for Service Check, Composite, or SLO monitors. Valid values are: `show_no_data`, `show_and_notify_no_data`, `resolve`, and `default`.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(r.getAllowOnMissingData()...),
				},
			},
			"priority": schema.StringAttribute{
				Description: "Integer from 1 (high) to 5 (low) indicating alert severity.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("0", "1", "2", "3", "4", "5"),
				},
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
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(stringvalidator.OneOf(r.getAllowRenotifyStatus()...)),
				},
			},
			"require_full_window": schema.BoolAttribute{
				Description: "A boolean indicating whether this monitor needs a full window of data before it's evaluated. Datadog strongly recommends you set this to `false` for sparse metrics, otherwise some evaluations may be skipped. If there's a custom_schedule set, `require_full_window` must be false and will be ignored.",
				Optional:    true,
			},
			"restricted_roles": schema.SetAttribute{
				Description:        "A list of unique role identifiers to define which roles are allowed to edit the monitor. Editing a monitor includes any updates to the monitor configuration, monitor deletion, and muting of the monitor for any amount of time. Roles unique identifiers can be pulled from the [Roles API](https://docs.datadoghq.com/api/latest/roles/#list-roles) in the `data.id` field.\n",
				DeprecationMessage: "Use `datadog_restriction_policy` resource to manage permission",
				Optional:           true,
				Computed:           true,
				ElementType:        types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"tags": schema.SetAttribute{
				Description: "A list of tags to associate with your monitor. This can help you categorize and filter monitors in the manage monitors page of the UI. Note: it's not currently possible to filter by these tags when querying via the API",
				// we use TypeSet to represent tags, paradoxically to be able to maintain them ordered;
				// we order them explicitly in the read/create/update methods of this resource and using
				// TypeSet makes Terraform ignore differences in order when creating a plan
				Optional:    true,
				ElementType: types.StringType,
			},
			"effective_tags": schema.SetAttribute{
				Description: "A list of tags to associate with your monitor, including those inherited from the provider's `default_tags` configuration. This can help you categorize and filter monitors in the manage monitors page of the UI.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"timeout_h": schema.Int64Attribute{
				Description: "The number of hours of the monitor not reporting data before it automatically resolves from a triggered state. The minimum allowed value is 0 hours. The maximum allowed value is 24 hours.",
				Optional:    true,
			},
			"validate": schema.BoolAttribute{
				Description: "If set to `false`, skip the validation call done during plan.",
				Optional:    true,
			},
			"draft_status": schema.StringAttribute{
				Description: "Indicates whether the monitor is in a draft or published state. When set to `draft`, the monitor appears as Draft and does not send notifications. When set to `published`, the monitor is active, and it evaluates conditions and sends notifications as configured.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(string(datadogV1.MONITORDRAFTSTATUS_PUBLISHED)),
				Validators: []validator.String{
					stringvalidator.OneOf(r.getDraftStatusTypes()...),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"assets": schema.ListNestedBlock{
				Description: "List of monitor assets (e.g., runbooks, dashboards, workflows) tied to this monitor.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "Name for the monitor asset.",
						},
						"url": schema.StringAttribute{
							Required:    true,
							Description: "URL for the asset.",
						},
						"category": schema.StringAttribute{
							Required:    true,
							Description: "Type of asset the entity represents on a monitor.",
							Validators: []validator.String{
								stringvalidator.OneOf(r.getAllowMonitorAssetCategory()...),
							},
						},
						"resource_key": schema.StringAttribute{
							Optional:    true,
							Description: "Identifier of the internal Datadog resource that this asset represents.",
						},
						"resource_type": schema.StringAttribute{
							Optional:    true,
							Description: "Type of internal Datadog resource associated with a monitor asset.",
							Validators: []validator.String{
								stringvalidator.OneOf(r.getAllowMonitorAssetResourceType()...),
							},
						},
					},
				},
			},
			"monitor_thresholds": schema.ListNestedBlock{
				Description: "Alert thresholds of the monitor.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"ok": schema.StringAttribute{
							Description: "The monitor `OK` threshold. Only supported in monitor type `service check`. Must be a number.",
							Validators: []validator.String{
								stringFloatValidator,
							},
							Optional:   true,
							CustomType: customtypes.FloatStringType{},
						},
						"warning": schema.StringAttribute{
							Description: "The monitor `WARNING` threshold. Must be a number.",
							Validators: []validator.String{
								stringFloatValidator,
							},
							Optional:   true,
							CustomType: customtypes.FloatStringType{},
						},
						"critical": schema.StringAttribute{
							Description: "The monitor `CRITICAL` threshold. Must be a number.",
							Validators: []validator.String{
								stringFloatValidator,
							},
							Optional:   true,
							CustomType: customtypes.FloatStringType{},
						},
						"unknown": schema.StringAttribute{
							Description: "The monitor `UNKNOWN` threshold. Only supported in monitor type `service check`. Must be a number.",
							Validators: []validator.String{
								stringFloatValidator,
							},
							Optional:   true,
							CustomType: customtypes.FloatStringType{},
						},
						"warning_recovery": schema.StringAttribute{
							Description: "The monitor `WARNING` recovery threshold. Must be a number.",
							Validators: []validator.String{
								stringFloatValidator,
							},
							Optional:   true,
							CustomType: customtypes.FloatStringType{},
						},
						"critical_recovery": schema.StringAttribute{
							Description: "The monitor `CRITICAL` recovery threshold. Must be a number.",
							Validators: []validator.String{
								stringFloatValidator,
							},
							Optional:   true,
							CustomType: customtypes.FloatStringType{},
						},
					},
				},
			},
			"monitor_threshold_windows": schema.ListNestedBlock{
				Description: "A mapping containing `recovery_window` and `trigger_window` values, e.g. `last_15m` . Can only be used for, and are required for, anomaly monitors.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
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
			},
			"scheduling_options": schema.ListNestedBlock{
				Description: "Configuration options for scheduling.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"evaluation_window": schema.ListNestedBlock{
							MarkdownDescription: "Configuration options for the evaluation window. If `hour_starts` is set, no other fields may be set. Otherwise, `day_starts` and `month_starts` must be set together.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"day_starts": schema.StringAttribute{
										Optional:    true,
										Description: "The time of the day at which a one day cumulative evaluation window starts. Must be defined in UTC time in `HH:mm` format.",
										Validators: []validator.String{
											stringvalidator.RegexMatches(regexp.MustCompile(`^\d{2}:\d{2}$`), "must be HH:mm"),
										},
									},
									"month_starts": schema.Int32Attribute{
										Optional:    true,
										Description: "The day of the month at which a one month cumulative evaluation window starts. Must be a value of 1.",
									},
									"hour_starts": schema.Int32Attribute{
										Optional:    true,
										Description: "The minute of the hour at which a one hour cumulative evaluation window starts. Must be between 0 and 59.",
										Validators: []validator.Int32{
											int32validator.Between(0, 59),
										},
									},
									"timezone": schema.StringAttribute{
										Optional:    true,
										Description: "The timezone for the cumulative evaluation window start time.",
									},
								},
							},
						},
						"custom_schedule": schema.ListNestedBlock{
							MarkdownDescription: "Configuration options for the custom schedules. If `start` is omitted, the monitor creation time will be used.",
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"recurrence": schema.ListNestedBlock{
										Description: "A list of recurrence definitions. Length must be 1.",
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"rrule": schema.StringAttribute{
													Description: "Must be a valid `rrule`. See API docs for supported fields",
													Required:    true,
												},
												"start": schema.StringAttribute{
													MarkdownDescription: "Time to start recurrence cycle. Similar to DTSTART. Expected format 'YYYY-MM-DDThh:mm:ss'",
													Optional:            true,
													Validators: []validator.String{
														stringvalidator.RegexMatches(
															regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}$`),
															"must be YYYY-MM-DDThh:mm:ss",
														),
													},
												},
												"timezone": schema.StringAttribute{
													MarkdownDescription: "'tz database' format. Example: `America/New_York` or `UTC`",
													Required:            true,
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
			"variables": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"event_query": schema.ListNestedBlock{
							Description: "A timeseries formula and functions events query.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"data_source": schema.StringAttribute{
										Required:    true,
										Description: "The data source for event platform-based queries.",
										Validators: []validator.String{
											stringvalidator.OneOf(r.getAllowEventQueryDataSource()...),
										},
									},
									"indexes": schema.ListAttribute{
										Optional:    true,
										ElementType: types.StringType,
										Description: "An array of index names to query in the stream.",
									},
									"name": schema.StringAttribute{
										Required:    true,
										Description: "The name of query for use in formulas.",
									},
								},
								Blocks: map[string]schema.Block{
									"search": schema.ListNestedBlock{
										Description: "The search options.",
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"query": schema.StringAttribute{
													Required:    true,
													Description: "The events search string.",
												},
											},
										},
									},
									"compute": schema.ListNestedBlock{
										Description: "The compute options.",
										Validators: []validator.List{
											listvalidator.IsRequired(),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"aggregation": schema.StringAttribute{
													Required:    true,
													Description: "The aggregation methods for event platform queries.",
													Validators: []validator.String{
														stringvalidator.OneOf(r.getAllowEventQueryAggregation()...),
													},
												},
												"interval": schema.Int64Attribute{
													Optional:    true,
													Description: "A time interval in milliseconds.",
												},
												"metric": schema.StringAttribute{
													Optional:    true,
													Description: "The measurable attribute to compute.",
												},
											},
										},
									},
									"group_by": schema.ListNestedBlock{
										Description: "Group by options.",
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"facet": schema.StringAttribute{
													Required:    true,
													Description: "The event facet.",
												},
												"limit": schema.Int64Attribute{
													Optional:    true,
													Description: "The number of groups to return.",
												},
											},
											Blocks: map[string]schema.Block{
												"sort": schema.ListNestedBlock{
													Description: "The options for sorting group by results.",
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"aggregation": schema.StringAttribute{
																Required:    true,
																Description: "The aggregation methods for the event platform queries.",
																Validators: []validator.String{
																	stringvalidator.OneOf(r.getAllowEventQueryAggregation()...),
																},
															},
															"metric": schema.StringAttribute{
																Optional:    true,
																Description: "The metric used for sorting group by results.",
															},
															"order": schema.StringAttribute{
																Optional:    true,
																Description: "Direction of sort.",
																Validators: []validator.String{
																	stringvalidator.OneOf(r.getAllowEventQueryOrder()...),
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
						"cloud_cost_query": schema.ListNestedBlock{
							Description: "The Cloud Cost query using formulas and functions.",
							Validators: []validator.List{
								listvalidator.SizeAtMost(5),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"data_source": schema.StringAttribute{
										Required:    true,
										Description: "The data source for cloud cost queries.",
										Validators: []validator.String{
											stringvalidator.OneOf(r.getAllowCloudCostDataSource()...),
										},
									},
									"query": schema.StringAttribute{
										Required:    true,
										Description: "The cloud cost query definition.",
									},
									"aggregator": schema.StringAttribute{
										Optional:    true,
										Description: "The aggregation methods available for cloud cost queries.",
										Validators: []validator.String{
											stringvalidator.OneOf(r.getAllowCloudCostAggregator()...),
										},
									},
									"name": schema.StringAttribute{
										Required:    true,
										Description: "The name of the query for use in formulas.",
									},
								},
							},
						},
						"data_quality_query": schema.ListNestedBlock{
							Description: "The Data Quality query using formulas and functions.",
							Validators: []validator.List{
								listvalidator.SizeAtMost(5),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Required:    true,
										Description: "The name of the query for use in formulas.",
									},
									"data_source": schema.StringAttribute{
										Required:    true,
										Description: "The data source for data quality queries. Valid value is `data_quality_metrics`.",
										Validators: []validator.String{
											stringvalidator.OneOf(r.getAllowDataQualityDataSource()...),
										},
									},
									"schema_version": schema.StringAttribute{
										Optional:    true,
										Description: "Schema version for the data quality query.",
									},
									"measure": schema.StringAttribute{
										Required:    true,
										Description: "The measure to query. Common values include `bytes`, `cardinality`, `custom`, `freshness`, `max`, `mean`, `min`, `nullness`, `percent_negative`, `percent_zero`, `row_count`, `stddev`, `sum`, `uniqueness`. Additional values may be supported.",
									},
									"filter": schema.StringAttribute{
										Required:    true,
										Description: "Filter expression used to match on data entities. Uses AAstra query syntax.",
									},
									"scope": schema.StringAttribute{
										Optional:    true,
										Description: "Optional scoping expression to further filter metrics.",
									},
									"group_by": schema.ListAttribute{
										Optional:    true,
										ElementType: types.StringType,
										Description: "Optional grouping fields for aggregation.",
									},
								},
								Blocks: map[string]schema.Block{
									"monitor_options": schema.ListNestedBlock{
										Description: "Monitor configuration options for data quality queries.",
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"custom_sql": schema.StringAttribute{
													Optional:    true,
													Description: "Custom SQL query for the monitor.",
												},
												"custom_where": schema.StringAttribute{
													Optional:    true,
													Description: "Custom WHERE clause for the query.",
												},
												"group_by_columns": schema.ListAttribute{
													Optional:    true,
													ElementType: types.StringType,
													Description: "Columns to group results by.",
												},
												"crontab_override": schema.StringAttribute{
													Optional:    true,
													Description: "Crontab expression to override the default schedule.",
												},
												"model_type_override": schema.StringAttribute{
													Optional:    true,
													Description: "Override for the model type. Valid values are `freshness`, `percentage`, `any`.",
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

func (r *monitorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), req, resp)
}

func (r *monitorResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state monitorResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, diags := r.getMonitorId(&state)
	if diags.HasError() {
		return
	}
	resp, httpResp, err := r.Api.GetMonitor(r.Auth, *id, *datadogV1.NewGetMonitorOptionalParameters().WithWithAssets(true))
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving Monitor"))
		return
	}

	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *monitorResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state monitorResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	monitorBody, _, diags := r.buildMonitorStruct(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateMonitor(r.Auth, *monitorBody)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving Monitor"))
		return
	}
	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *monitorResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state monitorResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, diags := r.getMonitorId(&state)
	if diags.HasError() {
		return
	}
	_, updateRequestBody, diags := r.buildMonitorStruct(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateMonitor(r.Auth, *id, *updateRequestBody)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving Monitor"))
		return
	}
	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *monitorResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state monitorResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, diags := r.getMonitorId(&state)
	if diags.HasError() {
		return
	}
	var httpResp *http.Response
	var err error
	if !state.ForceDelete.IsNull() && state.ForceDelete.ValueBool() {
		_, httpResp, err = r.Api.DeleteMonitor(r.Auth, *id, *datadogV1.NewDeleteMonitorOptionalParameters().WithForce("true"))
	} else {
		_, httpResp, err = r.Api.DeleteMonitor(r.Auth, *id)
	}
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting monitor"))
		return
	}
}

func (r *monitorResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// If the plan is null (resource is being destroyed) or no state exists yet, return early
	// as there's nothing to modify
	if req.Plan.Raw.IsNull() {
		return
	}
	var plan, state monitorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if !req.State.Raw.IsNull() {
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}
	if !plan.Validate.IsNull() && !plan.Validate.ValueBool() {
		// Explicitly skip validation
		return
	}
	combinedTags, diags := fwutils.CombineTags(ctx, plan.Tags, r.DefaultTags)
	if diags.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, frameworkPath.Root("effective_tags"), combinedTags)...)
	m, _, diags := r.buildMonitorStruct(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	isCreation := state.ID.IsNull()
	var httpresp *http.Response
	var err error
	log.Printf("[DEBUG] monitor/validate m=%#v", m)
	if !isCreation {
		id, diags := r.getMonitorId(&state)
		if diags.HasError() {
			return
		}
		_, httpresp, err = r.Api.ValidateExistingMonitor(r.Auth, *id, *m)
	} else {
		_, httpresp, err = r.Api.ValidateMonitor(r.Auth, *m)
	}
	if err != nil {
		if httpresp != nil && (httpresp.StatusCode == 502 || httpresp.StatusCode == 504) {
			resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error validating monitor, retrying"))
		}
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error validating monitor"))
	}
}

func (r *monitorResource) buildMonitorStruct(ctx context.Context, state *monitorResourceModel) (*datadogV1.Monitor, *datadogV1.MonitorUpdateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	if !state.NotifyNoData.ValueBool() && !state.NoDataTimeframe.IsNull() {
		diags.AddAttributeError(frameworkPath.Root("no_data_timeframe"), "`notify_no_data` has to be set to true with `no_data_timeframe`.", "")
		return nil, nil, diags
	}
	message := strings.TrimSpace(state.Message.ValueString())
	query := strings.TrimSpace(state.Query.ValueString())

	monitorType := datadogV1.MonitorType(state.Type.ValueString())

	m := datadogV1.NewMonitor(query, monitorType)
	m.SetName(state.Name.ValueString())
	m.SetMessage(message)

	u := datadogV1.NewMonitorUpdateRequest()
	u.SetType(monitorType)
	u.SetQuery(query)
	u.SetName(state.Name.ValueString())
	u.SetMessage(message)

	if state.DraftStatus.ValueString() != "" {
		m.SetDraftStatus(datadogV1.MonitorDraftStatus(state.DraftStatus.ValueString()))
		u.SetDraftStatus(datadogV1.MonitorDraftStatus(state.DraftStatus.ValueString()))
	}

	if v := r.parseInt(state.Priority); v != nil {
		m.SetPriority(*v)
		u.SetPriority(*v)
	} else {
		m.SetPriorityNil()
		u.SetPriorityNil()
	}
	fwutils.SetOptStringList(state.EffectiveTags, m.SetTags, ctx)
	fwutils.SetOptStringList(state.EffectiveTags, u.SetTags, ctx)
	fwutils.SetOptStringList(state.RestrictedRoles, m.SetRestrictedRoles, ctx)
	fwutils.SetOptStringList(state.RestrictedRoles, u.SetRestrictedRoles, ctx)
	// This handles an edge case where an empty array produce a 400 error,
	// so converting it to nil in the request.
	if restrictedRoles, ok := m.GetRestrictedRolesOk(); ok && len(*restrictedRoles) == 0 {
		m.SetRestrictedRolesNil()
		u.SetRestrictedRolesNil()
	}

	// Assets
	if assets := r.buildAssetsStruct(ctx, state.MonitorAssets); len(assets) > 0 {
		m.SetAssets(assets)
		u.SetAssets(assets)
	}

	monitorOptions := datadogV1.MonitorOptions{}
	if !state.EscalationMessage.IsNull() {
		escalationMessage := strings.TrimSpace(state.EscalationMessage.ValueString())
		monitorOptions.SetEscalationMessage(escalationMessage)
	}
	if !state.OnMissingData.IsNull() {
		monitorOptions.SetOnMissingData(datadogV1.OnMissingDataOption(state.OnMissingData.ValueString()))
	}
	if !state.RenotifyStatuses.IsNull() {
		var renotifyStatusesStr []string
		diags.Append(state.RenotifyStatuses.ElementsAs(ctx, &renotifyStatusesStr, false)...)
		renotifyStatuses := make([]datadogV1.MonitorRenotifyStatusType, 0)
		for _, str := range renotifyStatusesStr {
			renotifyStatuses = append(renotifyStatuses, datadogV1.MonitorRenotifyStatusType(str))
		}
		monitorOptions.SetRenotifyStatuses(renotifyStatuses)
	}
	if !state.NotificationPresetName.IsNull() {
		monitorOptions.SetNotificationPresetName(datadogV1.MonitorOptionsNotificationPresets(state.NotificationPresetName.ValueString()))
	}
	fwutils.SetOptBool(state.RequireFullWindow, monitorOptions.SetRequireFullWindow)
	fwutils.SetOptInt64(state.NoDataTimeframe, monitorOptions.SetNoDataTimeframe)
	fwutils.SetOptStringList(state.NotifyBy, monitorOptions.SetNotifyBy, ctx)
	fwutils.SetOptBool(state.NotifyNoData, monitorOptions.SetNotifyNoData)
	fwutils.SetOptString(state.GroupRetentionDuration, monitorOptions.SetGroupRetentionDuration)
	fwutils.SetOptInt64(state.NewGroupDelay, monitorOptions.SetNewGroupDelay)
	fwutils.SetOptInt64(state.NewHostDelay, monitorOptions.SetNewHostDelay)
	fwutils.SetOptInt64(state.EvaluationDelay, monitorOptions.SetEvaluationDelay)
	fwutils.SetOptInt64(state.RenotifyInterval, monitorOptions.SetRenotifyInterval)
	fwutils.SetOptInt64(state.RenotifyOccurrences, monitorOptions.SetRenotifyOccurrences)
	fwutils.SetOptBool(state.NotifyAudit, monitorOptions.SetNotifyAudit)
	fwutils.SetOptInt64(state.TimeoutH, monitorOptions.SetTimeoutH)
	fwutils.SetOptBool(state.IncludeTags, monitorOptions.SetIncludeTags)
	fwutils.SetOptBool(state.GroupbySimpleMonitor, monitorOptions.SetGroupbySimpleMonitor)
	fwutils.SetOptBool(state.EnableLogsSample, monitorOptions.SetEnableLogsSample)
	fwutils.SetOptBool(state.EnableSamples, monitorOptions.SetEnableSamples)
	fwutils.SetOptBool(state.Locked, monitorOptions.SetLocked)

	if state.MonitorThresholds != nil {
		thresholdObj := state.MonitorThresholds[0]
		var thresholds = datadogV1.MonitorThresholds{}
		if v := r.parseFloat(thresholdObj.Ok); v != nil {
			thresholds.SetOk(*v)
		}
		if v := r.parseFloat(thresholdObj.Unknown); v != nil {
			thresholds.SetUnknown(*v)
		}
		if v := r.parseFloat(thresholdObj.Critical); v != nil {
			thresholds.SetCritical(*v)
		}
		if v := r.parseFloat(thresholdObj.CriticalRecovery); v != nil {
			thresholds.SetCriticalRecovery(*v)
		}
		if v := r.parseFloat(thresholdObj.Warning); v != nil {
			thresholds.SetWarning(*v)
		}
		if v := r.parseFloat(thresholdObj.WarningRecovery); v != nil {
			thresholds.SetWarningRecovery(*v)
		}
		monitorOptions.SetThresholds(thresholds)
	}
	if state.MonitorThresholdWindows != nil {
		thresholdWindow := state.MonitorThresholdWindows[0]
		thresholdWindowOptions := datadogV1.MonitorThresholdWindowOptions{}
		fwutils.SetOptString(thresholdWindow.RecoveryWindow, thresholdWindowOptions.SetRecoveryWindow)
		fwutils.SetOptString(thresholdWindow.TriggerWindow, thresholdWindowOptions.SetTriggerWindow)
		monitorOptions.SetThresholdWindows(thresholdWindowOptions)
	}
	if schedulingOptionStruct := r.buildSchedulingOptionsStruct(ctx, state.SchedulingOptions); schedulingOptionStruct != nil {
		monitorOptions.SetSchedulingOptions(*schedulingOptionStruct)
	}
	if variableStruct := r.buildVariablesStruct(ctx, state.Variables); variableStruct != nil {
		monitorOptions.SetVariables(variableStruct)
	}
	m.SetOptions(monitorOptions)
	u.SetOptions(monitorOptions)

	return m, u, diags
}

func (r *monitorResource) buildSchedulingOptionsStruct(ctx context.Context, schedulingOptions []SchedulingOption) *datadogV1.MonitorOptionsSchedulingOptions {
	if schedulingOptions == nil || len(schedulingOptions) == 0 {
		return nil
	}
	schedulingOptionsReq := datadogV1.MonitorOptionsSchedulingOptions{}
	schedulingOption := schedulingOptions[0]
	if evalWindows := schedulingOption.EvaluationWindow; len(evalWindows) > 0 {
		evaluationWindowReq := datadogV1.MonitorOptionsSchedulingOptionsEvaluationWindow{}
		evalWindow := evalWindows[0]
		fwutils.SetOptString(evalWindow.DayStarts, evaluationWindowReq.SetDayStarts)
		fwutils.SetOptInt32(evalWindow.HourStarts, evaluationWindowReq.SetHourStarts)
		fwutils.SetOptInt32(evalWindow.MonthStarts, evaluationWindowReq.SetMonthStarts)
		fwutils.SetOptString(evalWindow.Timezone, evaluationWindowReq.SetTimezone)
		schedulingOptionsReq.SetEvaluationWindow(evaluationWindowReq)
	}
	if customSchedules := schedulingOption.CustomSchedule; len(customSchedules) > 0 {
		recurrencesReq := []datadogV1.MonitorOptionsCustomScheduleRecurrence{}
		customSchedule := customSchedules[0]
		for _, recurrence := range customSchedule.Recurrence {
			recurrenceReq := datadogV1.MonitorOptionsCustomScheduleRecurrence{}
			fwutils.SetOptString(recurrence.Rrule, recurrenceReq.SetRrule)
			fwutils.SetOptString(recurrence.Start, recurrenceReq.SetStart)
			fwutils.SetOptString(recurrence.Timezone, recurrenceReq.SetTimezone)
			recurrencesReq = append(recurrencesReq, recurrenceReq)
		}
		schedulingOptionsReq.SetCustomSchedule(datadogV1.MonitorOptionsCustomSchedule{
			Recurrences: recurrencesReq,
		})
	}
	if schedulingOptionsReq.EvaluationWindow == nil && schedulingOptionsReq.CustomSchedule == nil {
		return nil
	}
	return &schedulingOptionsReq
}

func (r *monitorResource) buildVariablesStruct(ctx context.Context, variables []Variable) []datadogV1.MonitorFormulaAndFunctionQueryDefinition {
	if variables == nil || len(variables) == 0 {
		return nil
	}
	variablesReq := []datadogV1.MonitorFormulaAndFunctionQueryDefinition{}
	// we always have zero or one `variables`
	variable := variables[0]
	if eventQReq := r.buildEventQueryStruct(ctx, variable.EventQuery); len(eventQReq) > 0 {
		variablesReq = append(variablesReq, eventQReq...)
	}
	if cloudCostReq := r.buildCloudCostQueryStruct(variable.CloudCostQuery); len(cloudCostReq) > 0 {
		variablesReq = append(variablesReq, cloudCostReq...)
	}
	if dataQualityReq := r.buildDataQualityQueryStruct(ctx, variable.DataQualityQuery); len(dataQualityReq) > 0 {
		variablesReq = append(variablesReq, dataQualityReq...)
	}
	return variablesReq
}

func (r *monitorResource) buildEventQueryStruct(ctx context.Context, eventQs []EventQuery) []datadogV1.MonitorFormulaAndFunctionQueryDefinition {
	if eventQs == nil || len(eventQs) == 0 {
		return nil
	}
	variablesReq := []datadogV1.MonitorFormulaAndFunctionQueryDefinition{}
	for _, eventQ := range eventQs {
		variableReq := datadogV1.MonitorFormulaAndFunctionQueryDefinition{}
		eventQueryReq := datadogV1.MonitorFormulaAndFunctionEventQueryDefinition{}
		fwutils.SetOptString(eventQ.Name, eventQueryReq.SetName)
		fwutils.SetOptStringList(eventQ.Indexes, eventQueryReq.SetIndexes, ctx)
		if !eventQ.DataSource.IsNull() {
			eventQueryReq.SetDataSource(datadogV1.MonitorFormulaAndFunctionEventsDataSource(eventQ.DataSource.ValueString()))
		}
		if search := eventQ.Search; search != nil {
			searchReq := datadogV1.MonitorFormulaAndFunctionEventQueryDefinitionSearch{}
			fwutils.SetOptString(search[0].Query, searchReq.SetQuery)
			eventQueryReq.SetSearch(searchReq)
		}
		if computes := eventQ.Compute; len(computes) > 0 {
			computeReq := datadogV1.MonitorFormulaAndFunctionEventQueryDefinitionCompute{}
			compute := computes[0]
			fwutils.SetOptInt64(compute.Interval, computeReq.SetInterval)
			fwutils.SetOptString(compute.Metric, computeReq.SetMetric)
			if !compute.Aggregation.IsNull() {
				computeReq.SetAggregation(datadogV1.MonitorFormulaAndFunctionEventAggregation(compute.Aggregation.ValueString()))
			}
			eventQueryReq.SetCompute(computeReq)
		}
		if groupBys := eventQ.GroupBy; len(groupBys) > 0 {
			groupBysReq := []datadogV1.MonitorFormulaAndFunctionEventQueryGroupBy{}
			for _, groupBy := range groupBys {
				groupByReq := datadogV1.MonitorFormulaAndFunctionEventQueryGroupBy{}
				fwutils.SetOptString(groupBy.Facet, groupByReq.SetFacet)
				fwutils.SetOptInt64(groupBy.Limit, groupByReq.SetLimit)
				if sortList := groupBy.Sort; len(sortList) > 0 {
					sortReq := datadogV1.MonitorFormulaAndFunctionEventQueryGroupBySort{}
					sort := sortList[0]
					fwutils.SetOptString(sort.Metric, sortReq.SetMetric)
					if !sort.Aggregation.IsNull() {
						sortReq.SetAggregation(datadogV1.MonitorFormulaAndFunctionEventAggregation(sort.Aggregation.ValueString()))
					}
					if !sort.Order.IsNull() {
						sortReq.SetOrder(datadogV1.QuerySortOrder(sort.Order.ValueString()))
					}
					groupByReq.SetSort(sortReq)
				}
				groupBysReq = append(groupBysReq, groupByReq)
			}
			eventQueryReq.SetGroupBy(groupBysReq)
		}
		variableReq.MonitorFormulaAndFunctionEventQueryDefinition = &eventQueryReq
		variablesReq = append(variablesReq, variableReq)
	}
	return variablesReq
}

func (r *monitorResource) buildCloudCostQueryStruct(cloudCostQs []CloudCostQuery) []datadogV1.MonitorFormulaAndFunctionQueryDefinition {
	if cloudCostQs == nil || len(cloudCostQs) == 0 {
		return nil
	}
	variablesReq := []datadogV1.MonitorFormulaAndFunctionQueryDefinition{}
	for _, cloudCostQ := range cloudCostQs {
		variableReq := datadogV1.MonitorFormulaAndFunctionQueryDefinition{}
		cloudCostQueryReq := datadogV1.MonitorFormulaAndFunctionCostQueryDefinition{}
		fwutils.SetOptString(cloudCostQ.Query, cloudCostQueryReq.SetQuery)
		fwutils.SetOptString(cloudCostQ.Name, cloudCostQueryReq.SetName)
		if !cloudCostQ.DataSource.IsNull() {
			cloudCostQueryReq.SetDataSource(datadogV1.MonitorFormulaAndFunctionCostDataSource(cloudCostQ.DataSource.ValueString()))
		}
		if !cloudCostQ.Aggregator.IsNull() {
			cloudCostQueryReq.SetAggregator(datadogV1.MonitorFormulaAndFunctionCostAggregator(cloudCostQ.Aggregator.ValueString()))
		}
		variableReq.MonitorFormulaAndFunctionCostQueryDefinition = &cloudCostQueryReq
		variablesReq = append(variablesReq, variableReq)
	}
	return variablesReq
}

func (r *monitorResource) buildDataQualityQueryStruct(ctx context.Context, dataQualityQs []DataQualityQuery) []datadogV1.MonitorFormulaAndFunctionQueryDefinition {
	if dataQualityQs == nil || len(dataQualityQs) == 0 {
		return nil
	}
	variablesReq := []datadogV1.MonitorFormulaAndFunctionQueryDefinition{}
	for _, dataQualityQ := range dataQualityQs {
		variableReq := datadogV1.MonitorFormulaAndFunctionQueryDefinition{}
		dataQualityQueryReq := datadogV1.MonitorFormulaAndFunctionDataQualityQueryDefinition{}
		fwutils.SetOptString(dataQualityQ.Name, dataQualityQueryReq.SetName)
		fwutils.SetOptString(dataQualityQ.Filter, dataQualityQueryReq.SetFilter)
		fwutils.SetOptString(dataQualityQ.SchemaVersion, dataQualityQueryReq.SetSchemaVersion)
		fwutils.SetOptString(dataQualityQ.Scope, dataQualityQueryReq.SetScope)
		if !dataQualityQ.DataSource.IsNull() {
			dataQualityQueryReq.SetDataSource(datadogV1.MonitorFormulaAndFunctionDataQualityDataSource(dataQualityQ.DataSource.ValueString()))
		}
		if !dataQualityQ.Measure.IsNull() {
			dataQualityQueryReq.SetMeasure(dataQualityQ.Measure.ValueString())
		}
		// Group by
		fwutils.SetOptStringList(dataQualityQ.GroupBy, dataQualityQueryReq.SetGroupBy, ctx)
		// Monitor options
		if monitorOpts := dataQualityQ.MonitorOptions; len(monitorOpts) > 0 {
			monitorOptsReq := datadogV1.MonitorFormulaAndFunctionDataQualityMonitorOptions{}
			opt := monitorOpts[0]
			fwutils.SetOptString(opt.CustomSql, monitorOptsReq.SetCustomSql)
			fwutils.SetOptString(opt.CustomWhere, monitorOptsReq.SetCustomWhere)
			fwutils.SetOptString(opt.CrontabOverride, monitorOptsReq.SetCrontabOverride)
			fwutils.SetOptStringList(opt.GroupByColumns, monitorOptsReq.SetGroupByColumns, ctx)
			if !opt.ModelTypeOverride.IsNull() {
				monitorOptsReq.SetModelTypeOverride(datadogV1.MonitorFormulaAndFunctionDataQualityModelTypeOverride(opt.ModelTypeOverride.ValueString()))
			}
			dataQualityQueryReq.SetMonitorOptions(monitorOptsReq)
		}
		variableReq.MonitorFormulaAndFunctionDataQualityQueryDefinition = &dataQualityQueryReq
		variablesReq = append(variablesReq, variableReq)
	}
	return variablesReq
}

func (r *monitorResource) buildAssetsStruct(ctx context.Context, tfAssets []MonitorAsset) []datadogV1.MonitorAsset {
	if len(tfAssets) == 0 {
		return nil
	}
	assets := make([]datadogV1.MonitorAsset, 0, len(tfAssets))
	for _, a := range tfAssets {
		category := datadogV1.MonitorAssetCategory(a.Category.ValueString())
		asset := datadogV1.NewMonitorAsset(category, a.Name.ValueString(), a.Url.ValueString())
		if !a.ResourceKey.IsNull() {
			asset.SetResourceKey(a.ResourceKey.ValueString())
		}
		if !a.ResourceType.IsNull() {
			asset.SetResourceType(datadogV1.MonitorAssetResourceType(a.ResourceType.ValueString()))
		}
		assets = append(assets, *asset)
	}
	return assets
}

func (r *monitorResource) updateState(ctx context.Context, state *monitorResourceModel, m *datadogV1.Monitor) {
	if id, ok := m.GetIdOk(); ok && id != nil {
		state.ID = types.StringValue(strconv.FormatInt(*id, 10))
	}
	state.Name = fwutils.ToTerraformStr(m.GetNameOk())

	if message, ok := m.GetMessageOk(); ok && message != nil {
		state.Message = customtypes.TrimSpaceStringValue{
			StringValue: types.StringValue(*message),
		}
	}

	if mType, ok := m.GetTypeOk(); ok && mType != nil {
		state.Type = customtypes.MonitorTypeValue{
			StringValue: types.StringValue(string(*mType)),
		}
	}

	if query, ok := m.GetQueryOk(); ok && query != nil {
		state.Query = customtypes.TrimSpaceStringValue{
			StringValue: types.StringValue(*query),
		}
	}

	if draftStatus, ok := m.GetDraftStatusOk(); ok && draftStatus != nil {
		state.DraftStatus = types.StringValue(string(*draftStatus))
	}

	if priority, ok := m.GetPriorityOk(); ok && priority != nil {
		state.Priority = types.StringValue(strconv.FormatInt(*priority, 10))
	}
	state.EffectiveTags = fwutils.ToTerraformSetString(ctx, m.GetTagsOk)
	if restrictedRoles, ok := m.GetRestrictedRolesOk(); ok && restrictedRoles == nil {
		state.RestrictedRoles = types.SetValueMust(types.StringType, []attr.Value{})
	} else {
		state.RestrictedRoles = fwutils.ToTerraformSetString(ctx, m.GetRestrictedRolesOk)
	}

	if escalationMessage, ok := m.Options.GetEscalationMessageOk(); ok && escalationMessage != nil {
		state.EscalationMessage = customtypes.TrimSpaceStringValue{
			StringValue: types.StringValue(*escalationMessage),
		}
	}
	if onMissingData, ok := m.Options.GetOnMissingDataOk(); ok && onMissingData != nil {
		state.OnMissingData = types.StringValue(string(*onMissingData))
	}
	if renotifyStatuses, ok := m.Options.GetRenotifyStatusesOk(); ok && renotifyStatuses != nil {
		state.RenotifyStatuses, _ = types.SetValueFrom(ctx, types.StringType, renotifyStatuses)
	}
	if notificationPresetName, ok := m.Options.GetNotificationPresetNameOk(); ok && notificationPresetName != nil {
		state.NotificationPresetName = types.StringValue(string(*notificationPresetName))
	}

	state.RequireFullWindow = fwutils.ToTerraformBool(m.Options.GetRequireFullWindowOk())
	state.NoDataTimeframe = fwutils.ToTerraformInt64(m.Options.GetNoDataTimeframeOk())
	state.NotifyNoData = fwutils.ToTerraformBool(m.Options.GetNotifyNoDataOk())
	state.GroupRetentionDuration = fwutils.ToTerraformStr(m.Options.GetGroupRetentionDurationOk())
	state.NewGroupDelay = fwutils.ToTerraformInt64(m.Options.GetNewGroupDelayOk())
	state.NewHostDelay = fwutils.ToTerraformInt64(m.Options.GetNewHostDelayOk())
	state.EvaluationDelay = fwutils.ToTerraformInt64(m.Options.GetEvaluationDelayOk())
	state.RenotifyInterval = fwutils.ToTerraformInt64(m.Options.GetRenotifyIntervalOk())
	state.RenotifyOccurrences = fwutils.ToTerraformInt64(m.Options.GetRenotifyOccurrencesOk())
	state.NotifyAudit = fwutils.ToTerraformBool(m.Options.GetNotifyAuditOk())
	state.TimeoutH = fwutils.ToTerraformInt64(m.Options.GetTimeoutHOk())
	state.IncludeTags = fwutils.ToTerraformBool(m.Options.GetIncludeTagsOk())
	state.GroupbySimpleMonitor = fwutils.ToTerraformBool(m.Options.GetGroupbySimpleMonitorOk())
	state.NotifyBy = fwutils.ToTerraformSetString(ctx, m.Options.GetNotifyByOk)
	state.EnableLogsSample = fwutils.ToTerraformBool(m.Options.GetEnableLogsSampleOk())
	state.Locked = fwutils.ToTerraformBool(m.Options.GetLockedOk())

	r.updateAssetsState(ctx, state, m)

	if monitorThresholds, ok := m.Options.GetThresholdsOk(); ok && monitorThresholds != nil {
		state.MonitorThresholds = []MonitorThreshold{{
			Ok:               r.buildFloatStringValue(fwutils.ToTerraformStr(monitorThresholds.GetOkOk())),
			Unknown:          r.buildFloatStringValue(fwutils.ToTerraformStr(monitorThresholds.GetUnknownOk())),
			Warning:          r.buildFloatStringValue(fwutils.ToTerraformStr(monitorThresholds.GetWarningOk())),
			WarningRecovery:  r.buildFloatStringValue(fwutils.ToTerraformStr(monitorThresholds.GetWarningRecoveryOk())),
			Critical:         r.buildFloatStringValue(fwutils.ToTerraformStr(monitorThresholds.GetCriticalOk())),
			CriticalRecovery: r.buildFloatStringValue(fwutils.ToTerraformStr(monitorThresholds.GetCriticalRecoveryOk())),
		}}
	}
	if thresholdWindow, ok := m.Options.GetThresholdWindowsOk(); ok && thresholdWindow != nil {
		state.MonitorThresholdWindows = []MonitorThresholdWindow{{
			RecoveryWindow: fwutils.ToTerraformStr(thresholdWindow.GetRecoveryWindowOk()),
			TriggerWindow:  fwutils.ToTerraformStr(thresholdWindow.GetTriggerWindowOk()),
		}}
	}
	r.updateSchedulingOptionState(state, m.Options)
	r.updateVariablesState(ctx, state, m.Options)
}

func (r *monitorResource) updateAssetsState(ctx context.Context, state *monitorResourceModel, m *datadogV1.Monitor) {
	// Assets -> state
	if assets, ok := m.GetAssetsOk(); ok && assets != nil {
		tfAssets := make([]MonitorAsset, 0, len(*assets))
		for _, a := range *assets {
			tfAsset := MonitorAsset{
				Name:     fwutils.ToTerraformStr(a.GetNameOk()),
				Url:      fwutils.ToTerraformStr(a.GetUrlOk()),
				Category: types.StringValue(string(a.GetCategory())),
			}
			if rk, ok := a.GetResourceKeyOk(); ok && rk != nil {
				tfAsset.ResourceKey = types.StringValue(*rk)
			} else {
				tfAsset.ResourceKey = types.StringNull()
			}
			if rt, ok := a.GetResourceTypeOk(); ok && rt != nil {
				tfAsset.ResourceType = types.StringValue(string(*rt))
			} else {
				tfAsset.ResourceType = types.StringNull()
			}
			tfAssets = append(tfAssets, tfAsset)
		}
		state.MonitorAssets = tfAssets
	}
}

func (r *monitorResource) updateSchedulingOptionState(state *monitorResourceModel, mOptions *datadogV1.MonitorOptions) {
	schedulingOptions, ok := mOptions.GetSchedulingOptionsOk()
	if !ok || schedulingOptions == nil {
		return
	}
	schedulingOptionState := SchedulingOption{}
	if evalWindow, ok := schedulingOptions.GetEvaluationWindowOk(); ok && evalWindow != nil &&
		(evalWindow.DayStarts != nil || evalWindow.MonthStarts != nil || evalWindow.HourStarts != nil) {
		schedulingOptionState.EvaluationWindow = []EvaluationWindow{{
			DayStarts:   fwutils.ToTerraformStr(evalWindow.GetDayStartsOk()),
			MonthStarts: fwutils.ToTerraformInt32(evalWindow.GetMonthStartsOk()),
			HourStarts:  fwutils.ToTerraformInt32(evalWindow.GetHourStartsOk()),
			Timezone:    fwutils.ToTerraformStr(evalWindow.GetTimezoneOk()),
		}}
	}
	if customSchedule, ok := schedulingOptions.GetCustomScheduleOk(); ok && customSchedule != nil && customSchedule.GetRecurrences() != nil &&
		(customSchedule.GetRecurrences()[0].Rrule != nil || customSchedule.GetRecurrences()[0].Start != nil || customSchedule.GetRecurrences()[0].Timezone != nil) {
		recurrence := customSchedule.GetRecurrences()[0]
		schedulingOptionState.CustomSchedule = []CustomSchedule{{
			Recurrence: []Recurrence{{
				Rrule:    fwutils.ToTerraformStr(recurrence.GetRruleOk()),
				Start:    fwutils.ToTerraformStr(recurrence.GetStartOk()),
				Timezone: fwutils.ToTerraformStr(recurrence.GetTimezoneOk()),
			}},
		}}
	}
	state.SchedulingOptions = []SchedulingOption{schedulingOptionState}
}

func (r *monitorResource) updateVariablesState(ctx context.Context, state *monitorResourceModel, mOptions *datadogV1.MonitorOptions) {
	variables, ok := mOptions.GetVariablesOk()
	if !ok || variables == nil || len(*variables) == 0 {
		return
	}
	eventQueryStates := []EventQuery{}
	CloudCostQueryStates := []CloudCostQuery{}
	DataQualityQueryStates := []DataQualityQuery{}

	for _, v := range *variables {
		if eventQState := r.buildEventQueryState(ctx, v.MonitorFormulaAndFunctionEventQueryDefinition); eventQState != nil {
			eventQueryStates = append(eventQueryStates, *eventQState)
		}
		if costQState := r.buildCloudCostQueryState(v.MonitorFormulaAndFunctionCostQueryDefinition); costQState != nil {
			CloudCostQueryStates = append(CloudCostQueryStates, *costQState)
		}
		if dataQualityQState := r.buildDataQualityQueryState(ctx, v.MonitorFormulaAndFunctionDataQualityQueryDefinition); dataQualityQState != nil {
			DataQualityQueryStates = append(DataQualityQueryStates, *dataQualityQState)
		}
	}
	state.Variables = []Variable{{
		EventQuery:       eventQueryStates,
		CloudCostQuery:   CloudCostQueryStates,
		DataQualityQuery: DataQualityQueryStates,
	}}
}

func (r *monitorResource) buildEventQueryState(ctx context.Context, eventQ *datadogV1.MonitorFormulaAndFunctionEventQueryDefinition) *EventQuery {
	if eventQ == nil {
		return nil
	}
	eventQueryState := EventQuery{
		Name: fwutils.ToTerraformStr(eventQ.GetNameOk()),
	}
	if dataSource, ok := eventQ.GetDataSourceOk(); ok && dataSource != nil {
		eventQueryState.DataSource = types.StringValue(string(*dataSource))
	}
	if indexes, ok := eventQ.GetIndexesOk(); ok && indexes != nil {
		eventQueryState.Indexes, _ = types.ListValueFrom(ctx, types.StringType, indexes)
	}
	if search, ok := eventQ.GetSearchOk(); ok && search != nil {
		eventQueryState.Search = []Search{{
			Query: types.StringValue(search.Query),
		}}
	}
	if compute, ok := eventQ.GetComputeOk(); ok && compute != nil {
		eventQueryState.Compute = []Compute{{
			Aggregation: types.StringValue(string(compute.Aggregation)),
			Interval:    fwutils.ToTerraformInt64(compute.GetIntervalOk()),
			Metric:      fwutils.ToTerraformStr(compute.GetMetricOk()),
		}}
	}
	if groupBys, ok := eventQ.GetGroupByOk(); ok && groupBys != nil {
		groupBysState := []GroupBy{}
		for _, groupBy := range *groupBys {
			groupByState := GroupBy{
				Facet: fwutils.ToTerraformStr(groupBy.GetFacetOk()),
				Limit: fwutils.ToTerraformInt64(groupBy.GetLimitOk()),
			}
			if sort, ok := groupBy.GetSortOk(); ok && sort != nil {
				sortState := Sort{
					Aggregation: types.StringValue(string(sort.Aggregation)),
					Metric:      fwutils.ToTerraformStr(sort.GetMetricOk()),
				}
				if order, ok := sort.GetOrderOk(); ok && order != nil {
					sortState.Order = types.StringValue(string(*sort.Order))
				}
				groupByState.Sort = []Sort{sortState}
			}
			groupBysState = append(groupBysState, groupByState)
		}
		eventQueryState.GroupBy = groupBysState
	}
	return &eventQueryState
}

func (r *monitorResource) buildCloudCostQueryState(cloudCostQ *datadogV1.MonitorFormulaAndFunctionCostQueryDefinition) *CloudCostQuery {
	if cloudCostQ == nil {
		return nil
	}
	cloudCostQueryState := CloudCostQuery{
		DataSource: types.StringValue(string(cloudCostQ.DataSource)),
		Query:      fwutils.ToTerraformStr(cloudCostQ.GetQueryOk()),
		Name:       fwutils.ToTerraformStr(cloudCostQ.GetNameOk()),
	}
	if aggregator, ok := cloudCostQ.GetAggregatorOk(); ok && aggregator != nil {
		cloudCostQueryState.Aggregator = types.StringValue(string(*cloudCostQ.Aggregator))
	}
	return &cloudCostQueryState
}

func (r *monitorResource) buildDataQualityQueryState(ctx context.Context, dataQualityQ *datadogV1.MonitorFormulaAndFunctionDataQualityQueryDefinition) *DataQualityQuery {
	if dataQualityQ == nil {
		return nil
	}
	dataQualityQueryState := DataQualityQuery{
		Name:          fwutils.ToTerraformStr(dataQualityQ.GetNameOk()),
		Filter:        fwutils.ToTerraformStr(dataQualityQ.GetFilterOk()),
		SchemaVersion: fwutils.ToTerraformStr(dataQualityQ.GetSchemaVersionOk()),
		Scope:         fwutils.ToTerraformStr(dataQualityQ.GetScopeOk()),
	}
	if dataSource, ok := dataQualityQ.GetDataSourceOk(); ok && dataSource != nil {
		dataQualityQueryState.DataSource = types.StringValue(string(*dataSource))
	}
	if measure, ok := dataQualityQ.GetMeasureOk(); ok && measure != nil {
		dataQualityQueryState.Measure = types.StringValue(string(*measure))
	}
	if groupBy, ok := dataQualityQ.GetGroupByOk(); ok && groupBy != nil {
		dataQualityQueryState.GroupBy, _ = types.ListValueFrom(ctx, types.StringType, groupBy)
	}
	// Monitor options
	if monitorOpts, ok := dataQualityQ.GetMonitorOptionsOk(); ok && monitorOpts != nil {
		monitorOptsState := DataQualityMonitorOptions{
			CustomSql:       fwutils.ToTerraformStr(monitorOpts.GetCustomSqlOk()),
			CustomWhere:     fwutils.ToTerraformStr(monitorOpts.GetCustomWhereOk()),
			CrontabOverride: fwutils.ToTerraformStr(monitorOpts.GetCrontabOverrideOk()),
		}
		if groupByCols, ok := monitorOpts.GetGroupByColumnsOk(); ok && groupByCols != nil {
			monitorOptsState.GroupByColumns, _ = types.ListValueFrom(ctx, types.StringType, groupByCols)
		}
		if modelTypeOverride, ok := monitorOpts.GetModelTypeOverrideOk(); ok && modelTypeOverride != nil {
			monitorOptsState.ModelTypeOverride = types.StringValue(string(*modelTypeOverride))
		}
		dataQualityQueryState.MonitorOptions = []DataQualityMonitorOptions{monitorOptsState}
	}
	return &dataQualityQueryState
}

func (r *monitorResource) getMonitorId(state *monitorResourceModel) (*int64, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	stateId := state.ID.ValueString()
	id, err := strconv.ParseInt(stateId, 10, 64)
	if err != nil {
		diags.Append(utils.FrameworkErrorDiag(err, "error on monitor id"))
		return nil, diags
	}
	return &id, diags
}

func (r *monitorResource) getAllowTypes() []string {
	return enumStrings((*datadogV1.MonitorType)(nil).GetAllowedValues())
}

func (r *monitorResource) getDraftStatusTypes() []string {
	return enumStrings((*datadogV1.MonitorDraftStatus)(nil).GetAllowedValues())
}

func (r *monitorResource) getAllowRenotifyStatus() []string {
	return enumStrings((*datadogV1.MonitorRenotifyStatusType)(nil).GetAllowedValues())
}

func (r *monitorResource) getAllowOnMissingData() []string {
	return enumStrings((*datadogV1.OnMissingDataOption)(nil).GetAllowedValues())
}

func (r *monitorResource) getAllowMonitorOptionsNotificationPresets() []string {
	return enumStrings((*datadogV1.MonitorOptionsNotificationPresets)(nil).GetAllowedValues())
}

func (r *monitorResource) getAllowEventQueryDataSource() []string {
	return enumStrings((*datadogV1.MonitorFormulaAndFunctionEventsDataSource)(nil).GetAllowedValues())
}

func (r *monitorResource) getAllowEventQueryAggregation() []string {
	return enumStrings((*datadogV1.MonitorFormulaAndFunctionEventAggregation)(nil).GetAllowedValues())
}

func (r *monitorResource) getAllowEventQueryOrder() []string {
	return enumStrings((*datadogV1.QuerySortOrder)(nil).GetAllowedValues())
}

func (r *monitorResource) getAllowCloudCostDataSource() []string {
	return enumStrings((*datadogV1.MonitorFormulaAndFunctionCostDataSource)(nil).GetAllowedValues())
}

func (r *monitorResource) getAllowCloudCostAggregator() []string {
	return enumStrings((*datadogV1.MonitorFormulaAndFunctionCostAggregator)(nil).GetAllowedValues())
}

func (r *monitorResource) getAllowDataQualityDataSource() []string {
	return enumStrings((*datadogV1.MonitorFormulaAndFunctionDataQualityDataSource)(nil).GetAllowedValues())
}

func (r *monitorResource) getAllowMonitorAssetCategory() []string {
	return enumStrings((*datadogV1.MonitorAssetCategory)(nil).GetAllowedValues())
}

func (r *monitorResource) getAllowMonitorAssetResourceType() []string {
	return enumStrings((*datadogV1.MonitorAssetResourceType)(nil).GetAllowedValues())
}

func (r *monitorResource) parseInt(v types.String) *int64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	result, err := strconv.ParseInt(v.ValueString(), 10, 64)
	if err != nil {
		return nil
	}
	return &result
}

func (r *monitorResource) parseFloat(v customtypes.FloatStringValue) *float64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	result, err := json.Number(v.ValueString()).Float64()
	if err != nil {
		return nil
	}
	return &result
}

func (r *monitorResource) buildFloatStringValue(v types.String) customtypes.FloatStringValue {
	return customtypes.FloatStringValue{
		StringValue: v,
	}
}
