package datadog

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	datadogCommunity "github.com/zorkian/go-datadog-api"
)

const logAlertMonitorType = "log alert"

func resourceDatadogMonitor() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogMonitorCreate,
		Read:   resourceDatadogMonitorRead,
		Update: resourceDatadogMonitorUpdate,
		Delete: resourceDatadogMonitorDelete,
		Exists: resourceDatadogMonitorExists,
		Importer: &schema.ResourceImporter{
			State: resourceDatadogMonitorImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"message": {
				Type:     schema.TypeString,
				Required: true,
				StateFunc: func(val interface{}) string {
					return strings.TrimSpace(val.(string))
				},
			},
			"escalation_message": {
				Type:     schema.TypeString,
				Optional: true,
				StateFunc: func(val interface{}) string {
					return strings.TrimSpace(val.(string))
				},
			},
			"query": {
				Type:     schema.TypeString,
				Required: true,
				StateFunc: func(val interface{}) string {
					return strings.TrimSpace(val.(string))
				},
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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

			// Options
			"thresholds": {
				Type:     schema.TypeMap,
				Optional: true,
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
			"threshold_windows": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"recovery_window": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"trigger_window": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"notify_no_data": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"new_host_delay": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  300,
			},
			"evaluation_delay": {
				Type:     schema.TypeInt,
				Computed: true,
				Optional: true,
			},
			"no_data_timeframe": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  10,
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
				Type:     schema.TypeInt,
				Optional: true,
			},
			"notify_audit": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"timeout_h": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"require_full_window": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"locked": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"silenced": {
				Type:       schema.TypeMap,
				Optional:   true,
				Elem:       schema.TypeInt,
				Deprecated: "use Downtime Resource instead",
			},
			"include_tags": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"tags": {
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
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func buildMonitorStruct(d *schema.ResourceData) (*datadog.Monitor, error) {

	var thresholds datadog.MonitorThresholds

	if r, ok := d.GetOk("thresholds.ok"); ok {
		v, err := strconv.ParseFloat(r.(string), 64)
		if err != nil {
			return nil, err
		}
		thresholds.SetOk(*datadog.NewNullableFloat64(datadog.PtrFloat64(v)))
	}
	if r, ok := d.GetOk("thresholds.warning"); ok {
		v, err := strconv.ParseFloat(r.(string), 64)
		if err != nil {
			return nil, err
		}
		thresholds.SetWarning(*datadog.NewNullableFloat64(datadog.PtrFloat64(v)))
	}
	if r, ok := d.GetOk("thresholds.unknown"); ok {
		v, err := strconv.ParseFloat(r.(string), 64)
		if err != nil {
			return nil, err
		}
		thresholds.SetUnknown(*datadog.NewNullableFloat64(datadog.PtrFloat64(v)))
	}
	if r, ok := d.GetOk("thresholds.critical"); ok {
		v, err := strconv.ParseFloat(r.(string), 64)
		if err != nil {
			return nil, err
		}
		thresholds.SetCritical(v)
	}
	if r, ok := d.GetOk("thresholds.warning_recovery"); ok {
		v, err := strconv.ParseFloat(r.(string), 64)
		if err != nil {
			return nil, err
		}
		thresholds.SetWarningRecovery(*datadog.NewNullableFloat64(datadog.PtrFloat64(v)))
	}
	if r, ok := d.GetOk("thresholds.critical_recovery"); ok {
		v, err := strconv.ParseFloat(r.(string), 64)
		if err != nil {
			return nil, err
		}
		thresholds.SetCriticalRecovery(*datadog.NewNullableFloat64(datadog.PtrFloat64(v)))
	}

	var thresholdWindows datadog.MonitorThresholdWindowOptions

	if r, ok := d.GetOk("threshold_windows.recovery_window"); ok {
		thresholdWindows.SetRecoveryWindow(*datadog.NewNullableString(datadog.PtrString(r.(string))))
	}

	if r, ok := d.GetOk("threshold_windows.trigger_window"); ok {
		thresholdWindows.SetTriggerWindow(*datadog.NewNullableString(datadog.PtrString(r.(string))))
	}

	o := datadog.MonitorOptions{
		Thresholds:        &thresholds,
		NotifyNoData:      datadog.PtrBool(d.Get("notify_no_data").(bool)),
		RequireFullWindow: datadog.PtrBool(d.Get("require_full_window").(bool)),
		IncludeTags:       datadog.PtrBool(d.Get("include_tags").(bool)),
	}
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
	if attr, ok := d.GetOk("new_host_delay"); ok {
		o.SetNewHostDelay(*datadog.NewNullableInt64(datadog.PtrInt64(int64(attr.(int)))))
	}
	if attr, ok := d.GetOk("evaluation_delay"); ok {
		o.SetEvaluationDelay(*datadog.NewNullableInt64(datadog.PtrInt64(int64(attr.(int)))))
	}
	if attr, ok := d.GetOk("no_data_timeframe"); ok {
		o.SetNoDataTimeframe(*datadog.NewNullableInt64(datadog.PtrInt64(int64(attr.(int)))))
	}
	if attr, ok := d.GetOk("renotify_interval"); ok {
		o.SetRenotifyInterval(*datadog.NewNullableInt64(datadog.PtrInt64(int64(attr.(int)))))
	}
	if attr, ok := d.GetOk("notify_audit"); ok {
		o.SetNotifyAudit(attr.(bool))
	}
	if attr, ok := d.GetOk("timeout_h"); ok {
		o.SetTimeoutH(*datadog.NewNullableInt64(datadog.PtrInt64(int64(attr.(int)))))
	}
	if attr, ok := d.GetOk("escalation_message"); ok {
		o.SetEscalationMessage(attr.(string))
	}
	if attr, ok := d.GetOk("locked"); ok {
		o.SetLocked(attr.(bool))
	}

	m := datadog.Monitor{
		Type:    datadog.MonitorType(d.Get("type").(string)).Ptr(),
		Query:   datadog.PtrString(d.Get("query").(string)),
		Name:    datadog.PtrString(d.Get("name").(string)),
		Message: datadog.PtrString(d.Get("message").(string)),
		Options: &o,
	}

	if m.GetType() == logAlertMonitorType {
		if attr, ok := d.GetOk("enable_logs_sample"); ok {
			o.SetEnableLogsSample(attr.(bool))
		} else {
			o.SetEnableLogsSample(false)
		}
	}

	tags := []string{}
	if attr, ok := d.GetOk("tags"); ok {
		for _, s := range attr.(*schema.Set).List() {
			tags = append(tags, s.(string))
		}
		sort.Strings(tags)
	}
	m.SetTags(tags)

	return &m, nil
}

func resourceDatadogMonitorExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	// Exists - This is called to verify a resource still exists. It is called prior to Read,
	// and lowers the burden of Read to be able to assume the resource exists.
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	i, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return false, err
	}

	if _, _, err = client.MonitorsApi.GetMonitor(auth, i).Execute(); err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return false, nil
		}
		return false, err
	}

	return true, nil
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
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	ms, err := buildMonitorStruct(d)
	if err != nil {
		return err
	}
	m, _, err := client.MonitorsApi.CreateMonitor(auth).Body(*ms).Execute()
	if err != nil {
		return translateClientError(err, "error updating monitor")
	}

	d.SetId(strconv.FormatInt(m.GetId(), 10))

	return resourceDatadogMonitorRead(d, meta)
}

func resourceDatadogMonitorRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}

	m, _, err := client.MonitorsApi.GetMonitor(auth, id).Execute()
	if err != nil {
		return err
	}
	thresholds := make(map[string]string)

	for thresholdKey, thresholdGetFunc := range map[string]func() (datadog.NullableFloat64, bool){
		"ok":                m.Options.Thresholds.GetOkOk,
		"warning":           m.Options.Thresholds.GetWarningOk,
		"unknown":           m.Options.Thresholds.GetUnknownOk,
		"warning_recovery":  m.Options.Thresholds.GetWarningRecoveryOk,
		"critical_recovery": m.Options.Thresholds.GetCriticalRecoveryOk,
	} {
		thresholdValue, ok := thresholdGetFunc()
		if ok && thresholdValue.Get() != nil {
			thresholds[thresholdKey] = strconv.FormatFloat(*thresholdValue.Get(), 'g', -1, 64)
		}
	}
	if criticalValue, ok := m.Options.Thresholds.GetCriticalOk(); ok {
		thresholds["critical"] = strconv.FormatFloat(criticalValue, 'g', -1, 64)
	}

	thresholdWindows := make(map[string]string)
	for thresholdWindowsKey, thresholdWindowsGetFunc := range map[string]func() (datadog.NullableString, bool){
		"recovery_window": m.Options.ThresholdWindows.GetRecoveryWindowOk,
		"trigger_window":  m.Options.ThresholdWindows.GetTriggerWindowOk,
	} {
		thresholdWindowsValue, ok := thresholdWindowsGetFunc()
		if ok && thresholdWindowsValue.Get() != nil {
			thresholdWindows[thresholdWindowsKey] = string(*thresholdWindowsValue.Get())
		}
	}

	tags := []string{}
	for _, s := range m.GetTags() {
		tags = append(tags, s)
	}
	sort.Strings(tags)

	log.Printf("[DEBUG] monitor: %+v", m)
	d.Set("name", m.GetName())
	d.Set("message", m.GetMessage())
	d.Set("query", m.GetQuery())
	d.Set("type", m.GetType())

	d.Set("thresholds", thresholds)
	d.Set("threshold_windows", thresholdWindows)

	if newHostDelay, ok := m.Options.GetNewHostDelayOk(); ok && newHostDelay.Get() != nil {
		d.Set("new_host_delay", newHostDelay.Get())
	}
	if evaluationDelay, ok := m.Options.GetEvaluationDelayOk(); ok && evaluationDelay.Get() != nil {
		d.Set("evaluation_delay", evaluationDelay.Get())
	}
	d.Set("notify_no_data", m.Options.GetNotifyNoData())
	if noDataTimeframe, ok := m.Options.GetNoDataTimeframeOk(); ok && noDataTimeframe.Get() != nil {
		d.Set("no_data_timeframe", noDataTimeframe.Get())
	}
	if renotifyInterval, ok := m.Options.GetRenotifyIntervalOk(); ok && renotifyInterval.Get() != nil {
		d.Set("renotify_interval", renotifyInterval.Get())
	}
	d.Set("notify_audit", m.Options.GetNotifyAudit())
	if timeoutH, ok := m.Options.GetTimeoutHOk(); ok && timeoutH.Get() != nil {
		d.Set("timeout_h", timeoutH.Get())
	}
	d.Set("escalation_message", m.Options.GetEscalationMessage())
	d.Set("include_tags", m.Options.GetIncludeTags())
	d.Set("tags", tags)
	d.Set("require_full_window", m.Options.GetRequireFullWindow()) // TODO Is this one of those options that we neeed to check?
	d.Set("locked", m.Options.GetLocked())

	if m.GetType() == logAlertMonitorType {
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
		if v.(int) < int(time.Now().Unix()) && v.(int) != 0 && v.(int) != -1 {
			// sync the state with whats in the config so its ignored
			apiSilenced[k] = int64(v.(int))
		}
	}
	d.Set("silenced", apiSilenced)

	return nil
}

func resourceDatadogMonitorUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	m := &datadog.Monitor{}

	m, err := buildMonitorStruct(d)
	if err != nil {
		return fmt.Errorf("Error parsing parsing monitor definition: %v", err)
	}

	id, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}
	m.Id = datadog.PtrInt64(id)

	silenced := false
	configuredSilenced := map[string]int{}
	if attr, ok := d.GetOk("silenced"); ok {
		// TODO: this is not very defensive, test if we can fail non int input
		s := make(map[string]int64)
		for k, v := range attr.(map[string]interface{}) {
			s[k] = int64(v.(int))
			configuredSilenced[k] = v.(int)
		}
		silenced = true
	}

	if _, _, err = client.MonitorsApi.EditMonitor(auth, m.GetId()).Body(*m).Execute(); err != nil {
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
		retval = providerConf.CommunityClient.UnmuteMonitorScopes(int(*m.Id), &datadogCommunity.UnmuteMonitorScopes{AllScopes: datadog.PtrBool(true)})
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
		for _, scope := range unmutedScopes {
			providerConf.CommunityClient.UnmuteMonitorScopes(int(*m.Id), &datadogCommunity.UnmuteMonitorScopes{Scope: &scope})
		}
	}

	return resourceDatadogMonitorRead(d, meta)
}

func resourceDatadogMonitorDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	i, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return err
	}

	if _, _, err = client.MonitorsApi.DeleteMonitor(auth, i).Execute(); err != nil {
		return err
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
func suppressDataDogFloatIntDiff(k, old, new string, d *schema.ResourceData) bool {
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
