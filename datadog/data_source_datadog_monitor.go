package datadog

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDatadogMonitor() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about an existing monitor for use in other resources.",
		ReadContext: dataSourceDatadogMonitorRead,
		Schema: map[string]*schema.Schema{
			"name_filter": {
				Description: "A monitor name to limit the search.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"tags_filter": {
				Description: "A list of tags to limit the search. This filters on the monitor scope.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"monitor_tags_filter": {
				Description: "A list of monitor tags to limit the search. This filters on the tags set on the monitor itself.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},

			// Computed values
			"name": {
				Description: "Name of the monitor",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"message": {
				Description: "Message included with notifications for this monitor",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"escalation_message": {
				Description: "Message included with a re-notification for this monitor.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"query": {
				Description: "Query of the monitor.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"type": {
				Description: "Type of the monitor.",
				Type:        schema.TypeString,
				Computed:    true,
			},

			// Options
			"monitor_thresholds": {
				Description: "Alert thresholds of the monitor.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ok": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"warning": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"critical": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"unknown": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"warning_recovery": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"critical_recovery": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"monitor_threshold_windows": {
				Description: "Mapping containing `recovery_window` and `trigger_window` values, e.g. `last_15m`. This is only used by anomaly monitors.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"recovery_window": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"trigger_window": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"notify_no_data": {
				Description: "Whether or not this monitor notifies when data stops reporting.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"new_group_delay": {
				Description: "Time (in seconds) to skip evaluations for new groups.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"new_host_delay": {
				Description: "Time (in seconds) allowing a host to boot and applications to fully start before starting the evaluation of monitor results.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"evaluation_delay": {
				Description: "Time (in seconds) for which evaluation is delayed. This is only used by metric monitors.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"no_data_timeframe": {
				Description: "The number of minutes before the monitor notifies when data stops reporting.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"renotify_interval": {
				Description: "The number of minutes after the last notification before the monitor re-notifies on the current status.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"notify_audit": {
				Description: "Whether or not tagged users are notified on changes to the monitor.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"timeout_h": {
				Description: "Number of hours of the monitor not reporting data before it automatically resolves from a triggered state.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"require_full_window": {
				Description: "Whether or not the monitor needs a full window of data before it is evaluated.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"locked": {
				Description: "Whether or not changes to the monitor are restricted to the creator or admins.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"restricted_roles": {
				// Uncomment when generally available
				// Description: "A list of role identifiers to associate with the monitor. Cannot be used with `locked`.",
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"include_tags": {
				Description: "Whether or not notifications from the monitor automatically inserts its triggering tags into the title.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"tags": {
				Description: "List of tags associated with the monitor.",
				// we use TypeSet to represent tags, paradoxically to be able to maintain them ordered;
				// we order them explicitly in the read/create/update methods of this resource and using
				// TypeSet makes Terraform ignore differences in order when creating a plan
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"enable_logs_sample": {
				Description: "Whether or not a list of log values which triggered the alert is included. This is only used by log monitors.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"groupby_simple_monitor": {
				Description: "Whether or not to trigger one alert if any source breaches a threshold.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
		},
	}
}

func dataSourceDatadogMonitorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	optionalParams := datadogV1.NewListMonitorsOptionalParameters()
	if v, ok := d.GetOk("name_filter"); ok {
		optionalParams = optionalParams.WithName(v.(string))
	}
	if v, ok := d.GetOk("tags_filter"); ok {
		optionalParams = optionalParams.WithTags(strings.Join(expandStringList(v.([]interface{})), ","))
	}
	if v, ok := d.GetOk("monitor_tags_filter"); ok {
		optionalParams = optionalParams.WithMonitorTags(strings.Join(expandStringList(v.([]interface{})), ","))
	}

	monitors, httpresp, err := datadogClientV1.MonitorsApi.ListMonitors(authV1, *optionalParams)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error querying monitors")
	}
	if len(monitors) > 1 {
		return diag.Errorf("your query returned more than one result, please try a more specific search criteria")
	}
	if len(monitors) == 0 {
		return diag.Errorf("your query returned no result, please try a less specific search criteria")
	}

	m := monitors[0]
	if err := utils.CheckForUnparsed(m); err != nil {
		return diag.FromErr(err)
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
	tags = append(tags, m.GetTags()...)
	sort.Strings(tags)

	var restricted_roles []string
	restricted_roles = append(restricted_roles, m.GetRestrictedRoles()...)
	sort.Strings(restricted_roles)

	d.SetId(strconv.FormatInt(m.GetId(), 10))
	d.Set("name", m.GetName())
	d.Set("message", m.GetMessage())
	d.Set("query", m.GetQuery())
	d.Set("type", m.GetType())

	if err := d.Set("monitor_thresholds", []interface{}{thresholds}); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("monitor_threshold_windows", []interface{}{thresholdWindows}); err != nil {
		return diag.FromErr(err)
	}

	d.Set("new_group_delay", m.Options.GetNewGroupDelay())
	d.Set("new_host_delay", m.Options.GetNewHostDelay())
	d.Set("evaluation_delay", m.Options.GetEvaluationDelay())
	d.Set("notify_no_data", m.Options.GetNotifyNoData())
	d.Set("no_data_timeframe", m.Options.GetNoDataTimeframe())
	d.Set("renotify_interval", m.Options.GetRenotifyInterval())
	d.Set("notify_audit", m.Options.GetNotifyAudit())
	d.Set("timeout_h", m.Options.GetTimeoutH())
	d.Set("escalation_message", m.Options.GetEscalationMessage())
	d.Set("include_tags", m.Options.GetIncludeTags())
	d.Set("tags", tags)
	d.Set("require_full_window", m.Options.GetRequireFullWindow()) // TODO Is this one of those options that we neeed to check?
	d.Set("locked", m.Options.GetLocked())
	d.Set("restricted_roles", restricted_roles)
	d.Set("groupby_simple_monitor", m.Options.GetGroupbySimpleMonitor())

	if m.GetType() == datadogV1.MONITORTYPE_LOG_ALERT {
		d.Set("enable_logs_sample", m.Options.GetEnableLogsSample())
	}

	return nil
}

func expandStringList(configured []interface{}) []string {
	vs := make([]string, 0, len(configured))
	for _, v := range configured {
		val, ok := v.(string)
		if ok && val != "" {
			vs = append(vs, v.(string))
		}
	}
	return vs
}
