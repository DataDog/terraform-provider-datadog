package datadog

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// Minimal interface between ResourceData and ResourceDiff so that we can use them interchangeably in buildMonitorStruct
type BuiltResource interface {
	Get(string) interface{}
	GetOk(string) (interface{}, bool)
}

var retryTimeout = time.Minute

func resourceDatadogMonitor() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog monitor resource. This can be used to create and manage Datadog monitors.",
		Create:        resourceDatadogMonitorCreate,
		Read:          resourceDatadogMonitorRead,
		Update:        resourceDatadogMonitorUpdate,
		Delete:        resourceDatadogMonitorDelete,
		CustomizeDiff: resourceDatadogMonitorCustomizeDiff,
		Importer: &schema.ResourceImporter{
			State: resourceDatadogMonitorImport,
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
				Description: "The monitor query to notify on. Note this is not the same query you see in the UI and the syntax is different depending on the monitor type, please see the [API Reference](https://docs.datadoghq.com/api/v1/monitors/#create-a-monitor) for details. `terraform plan` will validate query contents unless `validate` is set to `false`.",
				Type:        schema.TypeString,
				Required:    true,
				StateFunc: func(val interface{}) string {
					return strings.TrimSpace(val.(string))
				},
			},
			"type": {
				Description:  "The type of the monitor. The mapping from these types to the types found in the Datadog Web UI can be found in the Datadog API [documentation page](https://docs.datadoghq.com/api/v1/monitors/#create-a-monitor). Note: The monitor type cannot be changed after a monitor is created.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateEnumValue(datadogV1.NewMonitorTypeFromValue),
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
			"thresholds": {
				Description:   "Alert thresholds of the monitor.",
				Deprecated:    "Define `monitor_thresholds` list with one element instead.",
				ConflictsWith: []string{"monitor_thresholds"},
				Type:          schema.TypeMap,
				Optional:      true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ok": {
							Type:     schema.TypeFloat,
							Optional: true,
						},
						"warning": {
							Type:     schema.TypeFloat,
							Optional: true,
						},
						"critical": {
							Type:     schema.TypeFloat,
							Optional: true,
						},
						"unknown": {
							Type:     schema.TypeFloat,
							Optional: true,
						},
						"warning_recovery": {
							Type:     schema.TypeFloat,
							Optional: true,
						},
						"critical_recovery": {
							Type:     schema.TypeFloat,
							Optional: true,
						},
					},
				},
				DiffSuppressFunc: suppressDataDogFloatIntDiff,
			},
			"monitor_thresholds": {
				Description:   "Alert thresholds of the monitor.",
				Type:          schema.TypeList,
				ConflictsWith: []string{"thresholds"},
				MaxItems:      1,
				Optional:      true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ok": {
							Description:  "The monitor `OK` threshold. Must be a number.",
							Type:         schema.TypeString,
							ValidateFunc: validateFloatString,
							Optional:     true,
						},
						"warning": {
							Description:  "The monitor `WARNING` threshold. Must be a number.",
							Type:         schema.TypeString,
							ValidateFunc: validateFloatString,
							Optional:     true,
						},
						"critical": {
							Description:  "The monitor `CRITICAL` recovery threshold. Must be a number.",
							Type:         schema.TypeString,
							ValidateFunc: validateFloatString,
							Optional:     true,
						},
						"unknown": {
							Description:  "The monitor `UNKNOWN` threshold. Must be a number.",
							Type:         schema.TypeString,
							ValidateFunc: validateFloatString,
							Optional:     true,
						},
						"warning_recovery": {
							Description:  "The monitor `WARNING` recovery threshold. Must be a number.",
							Type:         schema.TypeString,
							ValidateFunc: validateFloatString,
							Optional:     true,
						},
						"critical_recovery": {
							Description:  "The monitor `CRITICAL` recovery threshold. Must be a number.",
							Type:         schema.TypeString,
							ValidateFunc: validateFloatString,
							Optional:     true,
						},
					},
				},
				DiffSuppressFunc: suppressDataDogFloatIntDiff,
			},
			"threshold_windows": {
				Description:   "A mapping containing `recovery_window` and `trigger_window` values, e.g. `last_15m`. Can only be used for, and are required for, anomaly monitors.",
				Deprecated:    "Define `monitor_threshold_windows` list with one element instead.",
				ConflictsWith: []string{"monitor_threshold_windows"},
				Type:          schema.TypeMap,
				Optional:      true,
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
			"monitor_threshold_windows": {
				Description:   "A mapping containing `recovery_window` and `trigger_window` values, e.g. `last_15m` . Can only be used for, and are required for, anomaly monitors.",
				Type:          schema.TypeList,
				ConflictsWith: []string{"threshold_windows"},
				MaxItems:      1,
				Optional:      true,
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
			"new_host_delay": {
				Description: "Time (in seconds) to allow a host to boot and applications to fully start before starting the evaluation of monitor results. Should be a non negative integer. Defaults to `300`.",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     300,
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
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// if locked is defined, ignore restricted_roles
					if _, ok := d.GetOk("locked"); ok {
						return true
					}
					return false
				},
			},
			"silenced": {
				Description: "Each scope will be muted until the given POSIX timestamp or forever if the value is `0`. Use `-1` if you want to unmute the scope. Deprecated: the silenced parameter is being deprecated in favor of the downtime resource. This will be removed in the next major version of the Terraform Provider.",
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        schema.TypeInt,
				Deprecated:  "Use the Downtime resource instead.",
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

func buildMonitorStruct(d BuiltResource) (*datadogV1.Monitor, *datadogV1.MonitorUpdateRequest) {

	var thresholds datadogV1.MonitorThresholds

	if r, ok := d.GetOk("monitor_thresholds.0.ok"); ok {
		v, _ := json.Number(r.(string)).Float64()
		thresholds.SetOk(v)
	} else if r, ok := d.GetOk("thresholds.ok"); ok {
		v, _ := json.Number(r.(string)).Float64()
		thresholds.SetOk(v)
	}
	if r, ok := d.GetOk("monitor_thresholds.0.warning"); ok {
		v, _ := json.Number(r.(string)).Float64()
		thresholds.SetWarning(v)
	} else if r, ok := d.GetOk("thresholds.warning"); ok {
		v, _ := json.Number(r.(string)).Float64()
		thresholds.SetWarning(v)
	}
	if r, ok := d.GetOk("monitor_thresholds.0.unknown"); ok {
		v, _ := json.Number(r.(string)).Float64()
		thresholds.SetUnknown(v)
	} else if r, ok := d.GetOk("thresholds.unknown"); ok {
		v, _ := json.Number(r.(string)).Float64()
		thresholds.SetUnknown(v)
	}
	if r, ok := d.GetOk("monitor_thresholds.0.critical"); ok {
		v, _ := json.Number(r.(string)).Float64()
		thresholds.SetCritical(v)
	} else if r, ok := d.GetOk("thresholds.critical"); ok {
		v, _ := json.Number(r.(string)).Float64()
		thresholds.SetCritical(v)
	}
	if r, ok := d.GetOk("monitor_thresholds.0.warning_recovery"); ok {
		v, _ := json.Number(r.(string)).Float64()
		thresholds.SetWarningRecovery(v)
	} else if r, ok := d.GetOk("thresholds.warning_recovery"); ok {
		v, _ := json.Number(r.(string)).Float64()
		thresholds.SetWarningRecovery(v)
	}
	if r, ok := d.GetOk("monitor_thresholds.0.critical_recovery"); ok {
		v, _ := json.Number(r.(string)).Float64()
		thresholds.SetCriticalRecovery(v)
	} else if r, ok := d.GetOk("thresholds.critical_recovery"); ok {
		v, _ := json.Number(r.(string)).Float64()
		thresholds.SetCriticalRecovery(v)
	}

	var thresholdWindows datadogV1.MonitorThresholdWindowOptions

	if r, ok := d.GetOk("monitor_threshold_windows.0.recovery_window"); ok {
		thresholdWindows.SetRecoveryWindow(r.(string))
	} else if r, ok := d.GetOk("threshold_windows.recovery_window"); ok {
		thresholdWindows.SetRecoveryWindow(r.(string))
	}

	if r, ok := d.GetOk("monitor_threshold_windows.0.trigger_window"); ok {
		thresholdWindows.SetTriggerWindow(r.(string))
	} else if r, ok := d.GetOk("threshold_windows.trigger_window"); ok {
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

	if attr, ok := d.GetOk("silenced"); ok {
		s := make(map[string]int64)
		// TODO: this is not very defensive, test if we can fail on non int input
		for k, v := range attr.(map[string]interface{}) {
			s[k] = int64(v.(int))
		}
		o.Silenced = &s
	}
	if attr, ok := d.GetOk("notify_no_data"); ok {
		o.SetNotifyNoData(attr.(bool))
	}
	// Don't check with GetOk, doesn't work with 0
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
	}

	m := datadogV1.NewMonitor()
	m.SetType(monitorType)
	m.SetQuery(d.Get("query").(string))
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
		// don't pass an empty array, it's not accepted
		m.SetRestrictedRoles(roles)
		u.SetRestrictedRoles(roles)
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
func resourceDatadogMonitorCustomizeDiff(diff *schema.ResourceDiff, meta interface{}) error {
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
	return resource.Retry(retryTimeout, func() *resource.RetryError {
		_, httpresp, err := datadogClientV1.MonitorsApi.ValidateMonitor(authV1).Body(*m).Execute()
		if err != nil {
			if httpresp != nil && httpresp.StatusCode == 502 {
				return resource.RetryableError(translateClientError(err, "error validating monitor, retrying"))
			}
			return resource.NonRetryableError(translateClientError(err, "error validating monitor"))
		}
		return nil
	})
}

func getUnmutedScopes(d *schema.ResourceData) []string {
	var unmuteScopes []string

	if attr, ok := d.GetOk("silenced"); ok {
		for k, v := range attr.(map[string]interface{}) {
			if v.(int) == -1 {
				unmuteScopes = append(unmuteScopes, k)
			}
		}
		log.Printf("[DEBUG] Unmute Scopes are: %v", unmuteScopes)
	}
	return unmuteScopes
}

func resourceDatadogMonitorCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	m, _ := buildMonitorStruct(d)
	mCreated, _, err := datadogClientV1.MonitorsApi.CreateMonitor(authV1).Body(*m).Execute()
	if err != nil {
		return translateClientError(err, "error creating monitor")
	}
	mCreatedId := strconv.FormatInt(mCreated.GetId(), 10)
	d.SetId(mCreatedId)

	return resourceDatadogMonitorRead(d, meta)
}

func resourceDatadogMonitorRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	i, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}
	var (
		m        datadogV1.Monitor
		httpresp *http.Response
	)
	if err = resource.Retry(d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		m, httpresp, err = datadogClientV1.MonitorsApi.GetMonitor(authV1, i).Execute()
		if err != nil {
			if httpresp != nil {
				if httpresp.StatusCode == 404 {
					d.SetId("")
					return nil
				} else if httpresp.StatusCode == 502 {
					return resource.RetryableError(translateClientError(err, "error getting monitor, retrying"))
				}
			}
			return resource.NonRetryableError(translateClientError(err, "error getting monitor"))
		}
		return nil
	}); err != nil {
		return err
	}

	if d.Id() == "" {
		return nil
	}

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

	var tags []string
	for _, s := range m.GetTags() {
		tags = append(tags, s)
	}
	sort.Strings(tags)

	log.Printf("[DEBUG] monitor: %+v", m)
	d.Set("name", m.GetName())
	d.Set("message", m.GetMessage())
	d.Set("query", m.GetQuery())
	d.Set("type", m.GetType())
	d.Set("priority", m.GetPriority())
	d.Set("restricted_roles", m.GetRestrictedRoles())

	// Set to deprecated field if that's what is used in the config, otherwise, set in the new field
	if _, ok := d.GetOk("thresholds"); ok {
		d.Set("thresholds", thresholds)
	} else if len(thresholds) > 0 { // Only set if there are values in the map to avoid diff
		if err := d.Set("monitor_thresholds", []interface{}{thresholds}); err != nil {
			return err
		}
	}
	// Set to deprecated field if that's what is used in the config, otherwise, set in the new field
	if _, ok := d.GetOk("threshold_windows"); ok {
		d.Set("threshold_windows", thresholdWindows)
	} else if len(thresholdWindows) > 0 { // Only set if there are values in the map to avoid diff
		if err := d.Set("monitor_threshold_windows", []interface{}{thresholdWindows}); err != nil {
			return err
		}
	}

	d.Set("new_host_delay", m.Options.GetNewHostDelay())
	d.Set("evaluation_delay", m.Options.GetEvaluationDelay())
	d.Set("notify_no_data", m.Options.GetNotifyNoData())
	d.Set("no_data_timeframe", m.Options.NoDataTimeframe.Get())
	d.Set("renotify_interval", m.Options.GetRenotifyInterval())
	d.Set("notify_audit", m.Options.GetNotifyAudit())
	d.Set("timeout_h", m.Options.GetTimeoutH())
	d.Set("escalation_message", m.Options.GetEscalationMessage())
	d.Set("include_tags", m.Options.GetIncludeTags())
	d.Set("tags", tags)
	d.Set("require_full_window", m.Options.GetRequireFullWindow()) // TODO Is this one of those options that we neeed to check?
	d.Set("locked", m.Options.GetLocked())

	if m.GetType() == datadogV1.MONITORTYPE_LOG_ALERT {
		d.Set("enable_logs_sample", m.Options.GetEnableLogsSample())
	}

	// The Datadog API doesn't return old timestamps or support a special value for unmuting scopes
	// So we provide this functionality by saving values to the state
	apiSilenced := m.Options.GetSilenced()
	configSilenced := d.Get("silenced").(map[string]interface{})

	for _, scope := range getUnmutedScopes(d) {
		if _, ok := apiSilenced[scope]; !ok {
			apiSilenced[scope] = -1
		}
	}

	// Ignore any timestamps in the past that aren't -1 or 0
	for k, v := range configSilenced {
		if v.(int) < int(providerConf.Now().Unix()) && v.(int) != 0 && v.(int) != -1 {
			// sync the state with whats in the config so its ignored
			apiSilenced[k] = int64(v.(int))
		}
	}
	d.Set("silenced", apiSilenced)

	return nil
}

func resourceDatadogMonitorUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	_, m := buildMonitorStruct(d)
	i, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}

	m.Id = &i

	silenced := false
	configuredSilenced := map[string]int{}
	if attr, ok := d.GetOk("silenced"); ok {
		// TODO: this is not very defensive, test if we can fail non int input
		s := make(map[string]int)
		for k, v := range attr.(map[string]interface{}) {
			s[k] = v.(int)
			configuredSilenced[k] = v.(int)
		}
		silenced = true
	}

	monitorResp, _, err := datadogClientV1.MonitorsApi.UpdateMonitor(authV1, i).Body(*m).Execute()
	if err != nil {
		return translateClientError(err, "error updating monitor")
	}

	var retval error
	if retval = resourceDatadogMonitorRead(d, meta); retval != nil {
		return retval
	}

	// if the silenced section was removed from the config, we unmute it via the API
	// The API wouldn't automatically unmute the monitor if the config is just missing
	// else we check what other silenced scopes were added from API response in the
	// "read" above and add them to "unmutedScopes" to be explicitly unmuted (because
	// they're "drift")
	unmutedScopes := getUnmutedScopes(d)
	if newSilenced, ok := d.GetOk("silenced"); ok && !silenced {
		// Because the Update method had a payload object which is not the same as the return result,
		// we need to set this attribute from one to the other.
		m.Options.SetSilenced(monitorResp.Options.GetSilenced())
		mSilenced := m.Options.GetSilenced()
		for k, _ := range mSilenced {
			// Since the Datadog GO client doesn't support unmuting on all scopes, loop over GetSilenced() and set the
			// end timestamp to time.Now().Unix()
			mSilenced[k] = providerConf.Now().Unix()
		}
		monitorResp, _, err = datadogClientV1.MonitorsApi.UpdateMonitor(authV1, i).Body(*m).Execute()
		if err != nil {
			return translateClientError(err, "error updating monitor")
		}
		d.Set("silenced", map[string]int{})
	} else {
		for scope := range newSilenced.(map[string]interface{}) {
			if _, ok := configuredSilenced[scope]; !ok {
				unmutedScopes = append(unmutedScopes, scope)
			}
		}
	}

	// Similarly, if the silenced attribute is -1, lets unmute those scopes
	if len(unmutedScopes) != 0 {
		// Because the Update method had a payload object which is not the same as the return result,
		// we need to set this attribute from one to the other.
		m.Options.SetSilenced(monitorResp.Options.GetSilenced())
		silencedList := m.Options.GetSilenced()
		for _, scope := range unmutedScopes {
			if _, ok := silencedList[scope]; ok {
				delete(silencedList, scope)
			}
		}
		if _, _, err = datadogClientV1.MonitorsApi.UpdateMonitor(authV1, i).Body(*m).Execute(); err != nil {
			return translateClientError(err, "error updating monitor")
		}
	}

	return resourceDatadogMonitorRead(d, meta)
}

func resourceDatadogMonitorDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	i, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}

	if d.Get("force_delete").(bool) {
		_, _, err = datadogClientV1.MonitorsApi.DeleteMonitor(authV1, i).Force("true").Execute()
	} else {
		_, _, err = datadogClientV1.MonitorsApi.DeleteMonitor(authV1, i).Execute()
	}

	if err != nil {
		return translateClientError(err, "error deleting monitor")
	}

	return nil
}

func resourceDatadogMonitorImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceDatadogMonitorRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

// Ignore any diff that results from the mix of ints or floats returned from the
// DataDog API.
func suppressDataDogFloatIntDiff(k, old string, new string, d *schema.ResourceData) bool {
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
