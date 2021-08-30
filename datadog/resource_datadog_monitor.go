package datadog

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Minimal interface between ResourceData and ResourceDiff so that we can use them interchangeably in buildMonitorStruct
type builtResource interface {
	Get(string) interface{}
	GetOk(string) (interface{}, bool)
}

var retryTimeout = time.Minute

func resourceDatadogMonitor() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog monitor resource. This can be used to create and manage Datadog monitors.",
		CreateContext: resourceDatadogMonitorCreate,
		ReadContext:   resourceDatadogMonitorRead,
		UpdateContext: resourceDatadogMonitorUpdate,
		DeleteContext: resourceDatadogMonitorDelete,
		CustomizeDiff: resourceDatadogMonitorCustomizeDiff,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of Datadog monitor.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"message": {
				Description: "A message to include with notifications for this monitor.\n\nEmail notifications can be sent to specific users by using the same `@username` notation as events.",
				Type:        schema.TypeString,
				Required:    true,
				StateFunc: func(val interface{}) string {
					return strings.TrimSpace(val.(string))
				},
			},
			"escalation_message": {
				Description: "A message to include with a re-notification. Supports the `@username` notification allowed elsewhere.",
				Type:        schema.TypeString,
				Optional:    true,
				StateFunc: func(val interface{}) string {
					return strings.TrimSpace(val.(string))
				},
			},
			"query": {
				Description: "The monitor query to notify on. Note this is not the same query you see in the UI and the syntax is different depending on the monitor type, please see the [API Reference](https://docs.datadoghq.com/api/v1/monitors/#create-a-monitor) for details. `terraform plan` will validate query contents unless `validate` is set to `false`.\n\n**Note:** APM latency data is now available as Distribution Metrics. Existing monitors have been migrated automatically but all terraformed monitors can still use the existing metrics. We strongly recommend updating monitor definitions to query the new metrics. To learn more, or to see examples of how to update your terraform definitions to utilize the new distribution metrics, see the [detailed doc](https://docs.datadoghq.com/tracing/guide/ddsketch_trace_metrics/).",
				Type:        schema.TypeString,
				Required:    true,
				StateFunc: func(val interface{}) string {
					return strings.TrimSpace(val.(string))
				},
			},
			"type": {
				Description:      "The type of the monitor. The mapping from these types to the types found in the Datadog Web UI can be found in the Datadog API [documentation page](https://docs.datadoghq.com/api/v1/monitors/#create-a-monitor). Note: The monitor type cannot be changed after a monitor is created.",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewMonitorTypeFromValue),
				// Datadog API quirk, see https://github.com/hashicorp/terraform/issues/13784
				DiffSuppressFunc: func(k, oldVal, newVal string, d *schema.ResourceData) bool {
					if (oldVal == "query alert" && newVal == "metric alert") ||
						(oldVal == "metric alert" && newVal == "query alert") {
						log.Printf("[DEBUG] Monitor '%s' got a '%s' response for an expected '%s' type. Suppressing change.", d.Get("name"), newVal, oldVal)
						return true
					}
					return newVal == oldVal
				},
			},
			"priority": {
				Description: "Integer from 1 (high) to 5 (low) indicating alert severity.",
				Type:        schema.TypeInt,
				Optional:    true,
			},

			// Options
			"monitor_thresholds": {
				Description: "Alert thresholds of the monitor.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ok": {
							Description:  "The monitor `OK` threshold. Must be a number.",
							Type:         schema.TypeString,
							ValidateFunc: validators.ValidateFloatString,
							Optional:     true,
						},
						"warning": {
							Description:  "The monitor `WARNING` threshold. Must be a number.",
							Type:         schema.TypeString,
							ValidateFunc: validators.ValidateFloatString,
							Optional:     true,
						},
						"critical": {
							Description:  "The monitor `CRITICAL` threshold. Must be a number.",
							Type:         schema.TypeString,
							ValidateFunc: validators.ValidateFloatString,
							Optional:     true,
						},
						"unknown": {
							Description:  "The monitor `UNKNOWN` threshold. Must be a number.",
							Type:         schema.TypeString,
							ValidateFunc: validators.ValidateFloatString,
							Optional:     true,
						},
						"warning_recovery": {
							Description:  "The monitor `WARNING` recovery threshold. Must be a number.",
							Type:         schema.TypeString,
							ValidateFunc: validators.ValidateFloatString,
							Optional:     true,
						},
						"critical_recovery": {
							Description:  "The monitor `CRITICAL` recovery threshold. Must be a number.",
							Type:         schema.TypeString,
							ValidateFunc: validators.ValidateFloatString,
							Optional:     true,
						},
					},
				},
				DiffSuppressFunc: suppressDataDogFloatIntDiff,
			},
			"monitor_threshold_windows": {
				Description: "A mapping containing `recovery_window` and `trigger_window` values, e.g. `last_15m` . Can only be used for, and are required for, anomaly monitors.",
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"recovery_window": {
							Description: "Describes how long an anomalous metric must be normal before the alert recovers.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"trigger_window": {
							Description: "Describes how long a metric must be anomalous before an alert triggers.",
							Type:        schema.TypeString,
							Optional:    true,
						},
					},
				},
			},
			"notify_no_data": {
				Description: "A boolean indicating whether this monitor will notify when data stops reporting. Defaults to `false`.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			// We only set new_group_delay in the monitor API payload if it is nonzero
			// because the SDKv2 terraform plugin API prevents unsetting new_group_delay
			// in updateMonitorState, so we can't reliably distinguish between new_group_delay
			// being unset (null) or set to zero.
			// Note that "new_group_delay overrides new_host_delay if it is set to a nonzero value"
			// refers to this terraform resource. In the API, setting new_group_delay
			// to any value, including zero, causes it to override new_host_delay.
			"new_group_delay": {
				Description: "Time (in seconds) to skip evaluations for new groups.\n\n`new_group_delay` overrides `new_host_delay` if it is set to a nonzero value.\n\nTo disable group delay for monitors grouped by host, `new_host_delay` must be set to zero due to the default value of `300` for that field (`new_group_delay` defaults to zero, so setting it to zero is not required).",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"new_host_delay": {
				// Removing the default requires removing the default in the API as well (possibly only for
				// terraform user agents)
				Description: "Time (in seconds) to allow a host to boot and applications to fully start before starting the evaluation of monitor results. Should be a non-negative integer. Defaults to `300` (this default will be removed in a major version release and `new_host_delay` will be removed entirely in a subsequent major version release).",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     300,
				Deprecated:  "Prefer using new_group_delay (except when setting `new_host_delay` to zero).",
			},
			"evaluation_delay": {
				Description: "(Only applies to metric alert) Time (in seconds) to delay evaluation, as a non-negative integer.\n\nFor example, if the value is set to `300` (5min), the `timeframe` is set to `last_5m` and the time is 7:00, the monitor will evaluate data from 6:50 to 6:55. This is useful for AWS CloudWatch and other backfilled metrics to ensure the monitor will always have data during evaluation.",
				Type:        schema.TypeInt,
				Computed:    true,
				Optional:    true,
			},
			"no_data_timeframe": {
				Description: "The number of minutes before a monitor will notify when data stops reporting. Provider defaults to 10 minutes.\n\nWe recommend at least 2x the monitor timeframe for metric alerts or 2 minutes for service checks.",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     10,
				DiffSuppressFunc: func(k, oldVal, newVal string, d *schema.ResourceData) bool {
					if !d.Get("notify_no_data").(bool) {
						if newVal != oldVal {
							log.Printf("[DEBUG] Ignore the no_data_timeframe change of monitor '%s' because notify_no_data is false.", d.Get("name"))
						}
						return true
					}
					return newVal == oldVal
				},
			},
			"renotify_interval": {
				Description: "The number of minutes after the last notification before a monitor will re-notify on the current status. It will only re-notify if it's not resolved.",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"notify_audit": {
				Description: "A boolean indicating whether tagged users will be notified on changes to this monitor. Defaults to `false`.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"timeout_h": {
				Description: "The number of hours of the monitor not reporting data before it will automatically resolve from a triggered state.",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"require_full_window": {
				Description: "A boolean indicating whether this monitor needs a full window of data before it's evaluated.\n\nWe highly recommend you set this to `false` for sparse metrics, otherwise some evaluations will be skipped. Default: `true` for `on average`, `at all times` and `in total` aggregation. `false` otherwise.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"locked": {
				Description:   "A boolean indicating whether changes to to this monitor should be restricted to the creator or admins. Defaults to `false`.",
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"restricted_roles"},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// if restricted_roles is defined, ignore locked
					if _, ok := d.GetOk("restricted_roles"); ok {
						return true
					}
					return false
				},
			},
			"restricted_roles": {
				// Uncomment when generally available
				// Description: "A list of role identifiers to associate with the monitor. Cannot be used with `locked`.",
				Type:          schema.TypeSet,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"locked"},
			},
			"include_tags": {
				Description: "A boolean indicating whether notifications from this monitor automatically insert its triggering tags into the title. Defaults to `true`.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"tags": {
				Description: "A list of tags to associate with your monitor. This can help you categorize and filter monitors in the manage monitors page of the UI. Note: it's not currently possible to filter by these tags when querying via the API",
				// we use TypeSet to represent tags, paradoxically to be able to maintain them ordered;
				// we order them explicitly in the read/create/update methods of this resource and using
				// TypeSet makes Terraform ignore differences in order when creating a plan
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"groupby_simple_monitor": {
				Description: "Whether or not to trigger one alert if any source breaches a threshold. This is only used by log monitors. Defaults to `false`.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			// since this is only useful for "log alert" type, we don't set a default value
			// if we did set it, it would be used for all types; we have to handle this manually
			// throughout the code
			"enable_logs_sample": {
				Description: "A boolean indicating whether or not to include a list of log values which triggered the alert. This is only used by log monitors. Defaults to `false`.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"force_delete": {
				Description: "A boolean indicating whether this monitor can be deleted even if it’s referenced by other resources (e.g. SLO, composite monitor).",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"validate": {
				Description: "If set to `false`, skip the validation call done during plan.",
				Type:        schema.TypeBool,
				Optional:    true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// This is never sent to the backend, so it should never generate a diff
					return true
				},
			},
		},
	}
}

func buildMonitorStruct(d builtResource) (*datadogV1.Monitor, *datadogV1.MonitorUpdateRequest) {

	var thresholds datadogV1.MonitorThresholds

	if r, ok := d.GetOk("monitor_thresholds.0.ok"); ok {
		v, _ := json.Number(r.(string)).Float64()
		thresholds.SetOk(v)
	}
	if r, ok := d.GetOk("monitor_thresholds.0.warning"); ok {
		v, _ := json.Number(r.(string)).Float64()
		thresholds.SetWarning(v)
	}
	if r, ok := d.GetOk("monitor_thresholds.0.unknown"); ok {
		v, _ := json.Number(r.(string)).Float64()
		thresholds.SetUnknown(v)
	}
	if r, ok := d.GetOk("monitor_thresholds.0.critical"); ok {
		v, _ := json.Number(r.(string)).Float64()
		thresholds.SetCritical(v)
	}
	if r, ok := d.GetOk("monitor_thresholds.0.warning_recovery"); ok {
		v, _ := json.Number(r.(string)).Float64()
		thresholds.SetWarningRecovery(v)
	}
	if r, ok := d.GetOk("monitor_thresholds.0.critical_recovery"); ok {
		v, _ := json.Number(r.(string)).Float64()
		thresholds.SetCriticalRecovery(v)
	}

	var thresholdWindows datadogV1.MonitorThresholdWindowOptions

	if r, ok := d.GetOk("monitor_threshold_windows.0.recovery_window"); ok {
		thresholdWindows.SetRecoveryWindow(r.(string))
	}

	if r, ok := d.GetOk("monitor_threshold_windows.0.trigger_window"); ok {
		thresholdWindows.SetTriggerWindow(r.(string))
	}

	o := datadogV1.MonitorOptions{}
	o.SetThresholds(thresholds)
	o.SetNotifyNoData(d.Get("notify_no_data").(bool))
	o.SetRequireFullWindow(d.Get("require_full_window").(bool))
	o.SetIncludeTags(d.Get("include_tags").(bool))

	if thresholdWindows.HasRecoveryWindow() || thresholdWindows.HasTriggerWindow() {
		o.SetThresholdWindows(thresholdWindows)
	}

	if attr, ok := d.GetOk("notify_no_data"); ok {
		o.SetNotifyNoData(attr.(bool))
	}
	if attr, ok := d.GetOk("new_group_delay"); ok {
		o.SetNewGroupDelay(int64(attr.(int)))
	}
	// Don't check with GetOk, doesn't work with 0 (we can't do the same for
	// new_group_delay because it would always override new_host_delay).
	o.SetNewHostDelay(int64(d.Get("new_host_delay").(int)))
	if attr, ok := d.GetOk("evaluation_delay"); ok {
		o.SetEvaluationDelay(int64(attr.(int)))
	}
	if attr, ok := d.GetOk("no_data_timeframe"); ok {
		o.SetNoDataTimeframe(int64(attr.(int)))
	}
	if attr, ok := d.GetOk("renotify_interval"); ok {
		o.SetRenotifyInterval(int64(attr.(int)))
	}
	if attr, ok := d.GetOk("notify_audit"); ok {
		o.SetNotifyAudit(attr.(bool))
	}
	if attr, ok := d.GetOk("timeout_h"); ok {
		o.SetTimeoutH(int64(attr.(int)))
	}
	if attr, ok := d.GetOk("escalation_message"); ok {
		o.SetEscalationMessage(attr.(string))
	}
	if attr, ok := d.GetOk("locked"); ok {
		o.SetLocked(attr.(bool))
	}

	monitorType := datadogV1.MonitorType(d.Get("type").(string))
	if monitorType == datadogV1.MONITORTYPE_LOG_ALERT {
		if attr, ok := d.GetOk("enable_logs_sample"); ok {
			o.SetEnableLogsSample(attr.(bool))
		} else {
			o.SetEnableLogsSample(false)
		}

		if attr, ok := d.GetOk("groupby_simple_monitor"); ok {
			o.SetGroupbySimpleMonitor(attr.(bool))
		}
	}

	m := datadogV1.NewMonitor(d.Get("query").(string), monitorType)
	m.SetName(d.Get("name").(string))
	m.SetMessage(d.Get("message").(string))
	m.SetPriority(int64(d.Get("priority").(int)))
	m.SetOptions(o)

	u := datadogV1.NewMonitorUpdateRequest()
	u.SetType(monitorType)
	u.SetQuery(d.Get("query").(string))
	u.SetName(d.Get("name").(string))
	u.SetMessage(d.Get("message").(string))
	u.SetPriority(int64(d.Get("priority").(int)))
	u.SetOptions(o)

	roles := make([]string, 0)
	if attr, ok := d.GetOk("restricted_roles"); ok {
		for _, r := range attr.(*schema.Set).List() {
			roles = append(roles, r.(string))
		}
		sort.Strings(roles)
		m.SetRestrictedRoles(roles)
		u.SetRestrictedRoles(roles)
	} else {
		m.SetRestrictedRoles(nil)
		u.SetRestrictedRoles(nil)
	}

	tags := make([]string, 0)
	if attr, ok := d.GetOk("tags"); ok {
		for _, s := range attr.(*schema.Set).List() {
			tags = append(tags, s.(string))
		}
		sort.Strings(tags)
	}
	m.SetTags(tags)
	u.SetTags(tags)

	return m, u
}

// Use CustomizeDiff to do monitor validation
func resourceDatadogMonitorCustomizeDiff(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	if _, ok := diff.GetOk("query"); !ok {
		// If "query" depends on other resources, we can't validate as the variables may not be interpolated yet.
		return nil
	}
	if _, ok := diff.GetOk("type"); !ok {
		// Same for type
		return nil
	}
	if validate, ok := diff.GetOkExists("validate"); ok && !validate.(bool) {
		// Explicitly skip validation
		return nil
	}
	m, _ := buildMonitorStruct(diff)

	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1
	return resource.RetryContext(ctx, retryTimeout, func() *resource.RetryError {
		_, httpresp, err := datadogClientV1.MonitorsApi.ValidateMonitor(authV1, *m)
		if err != nil {
			if httpresp != nil && (httpresp.StatusCode == 502 || httpresp.StatusCode == 504) {
				return resource.RetryableError(utils.TranslateClientError(err, httpresp, "error validating monitor, retrying"))
			}
			return resource.NonRetryableError(utils.TranslateClientError(err, httpresp, "error validating monitor"))
		}
		return nil
	})
}

func resourceDatadogMonitorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	m, _ := buildMonitorStruct(d)
	mCreated, httpResponse, err := datadogClientV1.MonitorsApi.CreateMonitor(authV1, *m)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating monitor")
	}
	if err := utils.CheckForUnparsed(m); err != nil {
		return diag.FromErr(err)
	}
	mCreatedID := strconv.FormatInt(mCreated.GetId(), 10)
	d.SetId(mCreatedID)

	return updateMonitorState(d, meta, &mCreated)
}

func updateMonitorState(d *schema.ResourceData, meta interface{}, m *datadogV1.Monitor) diag.Diagnostics {
	thresholds := make(map[string]string)

	if v, ok := m.Options.Thresholds.GetOkOk(); ok {
		thresholds["ok"] = fmt.Sprintf("%v", *v)
	}
	if v, ok := m.Options.Thresholds.GetWarningOk(); ok {
		thresholds["warning"] = fmt.Sprintf("%v", *v)
	}
	if v, ok := m.Options.Thresholds.GetCriticalOk(); ok {
		thresholds["critical"] = fmt.Sprintf("%v", *v)
	}
	if v, ok := m.Options.Thresholds.GetUnknownOk(); ok {
		thresholds["unknown"] = fmt.Sprintf("%v", *v)
	}
	if v, ok := m.Options.Thresholds.GetWarningRecoveryOk(); ok {
		thresholds["warning_recovery"] = fmt.Sprintf("%v", *v)
	}
	if v, ok := m.Options.Thresholds.GetCriticalRecoveryOk(); ok {
		thresholds["critical_recovery"] = fmt.Sprintf("%v", *v)
	}

	thresholdWindows := make(map[string]string)
	for k, v := range map[string]string{
		"recovery_window": m.Options.ThresholdWindows.GetRecoveryWindow(),
		"trigger_window":  m.Options.ThresholdWindows.GetTriggerWindow(),
	} {
		if v != "" {
			thresholdWindows[k] = v
		}
	}

	log.Printf("[DEBUG] monitor: %+v", m)
	if err := d.Set("name", m.GetName()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("message", m.GetMessage()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("query", m.GetQuery()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("type", m.GetType()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("priority", m.GetPriority()); err != nil {
		return diag.FromErr(err)
	}

	if len(thresholds) > 0 {
		if err := d.Set("monitor_thresholds", []interface{}{thresholds}); err != nil {
			return diag.FromErr(err)
		}
	}
	if len(thresholdWindows) > 0 {
		if err := d.Set("monitor_threshold_windows", []interface{}{thresholdWindows}); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set("new_group_delay", m.Options.GetNewGroupDelay()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("new_host_delay", m.Options.GetNewHostDelay()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("evaluation_delay", m.Options.GetEvaluationDelay()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("notify_no_data", m.Options.GetNotifyNoData()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("no_data_timeframe", m.Options.NoDataTimeframe.Get()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("renotify_interval", m.Options.GetRenotifyInterval()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("notify_audit", m.Options.GetNotifyAudit()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("timeout_h", m.Options.GetTimeoutH()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("escalation_message", m.Options.GetEscalationMessage()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("include_tags", m.Options.GetIncludeTags()); err != nil {
		return diag.FromErr(err)
	}

	var tags []string
	tags = append(tags, m.GetTags()...)
	sort.Strings(tags)
	if err := d.Set("tags", tags); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("require_full_window", m.Options.GetRequireFullWindow()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("locked", m.Options.GetLocked()); err != nil {
		return diag.FromErr(err)
	}
	// This helper function is defined in `resource_datadog_dashboard`
	restrictedRoles := buildTerraformRestrictedRoles(m.RestrictedRoles)
	if err := d.Set("restricted_roles", restrictedRoles); err != nil {
		return diag.FromErr(err)
	}

	if m.GetType() == datadogV1.MONITORTYPE_LOG_ALERT {
		if err := d.Set("enable_logs_sample", m.Options.GetEnableLogsSample()); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("groupby_simple_monitor", m.Options.GetGroupbySimpleMonitor()); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceDatadogMonitorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	i, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}
	var (
		m        datadogV1.Monitor
		httpresp *http.Response
	)
	if err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		m, httpresp, err = datadogClientV1.MonitorsApi.GetMonitor(authV1, i)
		if err != nil {
			if httpresp != nil {
				if httpresp.StatusCode == 404 {
					d.SetId("")
					return nil
				} else if httpresp.StatusCode == 502 {
					return resource.RetryableError(utils.TranslateClientError(err, httpresp, "error getting monitor, retrying"))
				}
			}
			return resource.NonRetryableError(utils.TranslateClientError(err, httpresp, "error getting monitor"))
		}
		if err := utils.CheckForUnparsed(m); err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	}); err != nil {
		return diag.FromErr(err)
	}

	if d.Id() == "" {
		return nil
	}

	return updateMonitorState(d, meta, &m)
}

func resourceDatadogMonitorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	_, m := buildMonitorStruct(d)
	i, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}

	m.Id = &i

	monitorResp, httpresp, err := datadogClientV1.MonitorsApi.UpdateMonitor(authV1, i, *m)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error updating monitor")
	}
	if err := utils.CheckForUnparsed(monitorResp); err != nil {
		return diag.FromErr(err)
	}

	return updateMonitorState(d, meta, &monitorResp)
}

func resourceDatadogMonitorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	i, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}

	var httpResponse *http.Response

	if d.Get("force_delete").(bool) {
		_, httpResponse, err = datadogClientV1.MonitorsApi.DeleteMonitor(authV1, i,
			*datadogV1.NewDeleteMonitorOptionalParameters().WithForce("true"))
	} else {
		_, httpResponse, err = datadogClientV1.MonitorsApi.DeleteMonitor(authV1, i)
	}

	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error deleting monitor")
	}

	return nil
}

// Ignore any diff that results from the mix of ints or floats returned from the
// DataDog API.
func suppressDataDogFloatIntDiff(_, old, new string, _ *schema.ResourceData) bool {
	oF, err := strconv.ParseFloat(old, 64)
	if err != nil {
		log.Printf("Error parsing float of old value (%s): %s", old, err)
		return false
	}

	nF, err := strconv.ParseFloat(new, 64)
	if err != nil {
		log.Printf("Error parsing float of new value (%s): %s", new, err)
		return false
	}

	// if the float values of these attributes are equivalent, ignore this
	// diff
	if oF == nF {
		return true
	}
	return false
}
