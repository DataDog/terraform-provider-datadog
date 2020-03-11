package datadog

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceDatadogServiceLevelObjective() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogServiceLevelObjectiveCreate,
		Read:   resourceDatadogServiceLevelObjectiveRead,
		Update: resourceDatadogServiceLevelObjectiveUpdate,
		Delete: resourceDatadogServiceLevelObjectiveDelete,
		Exists: resourceDatadogServiceLevelObjectiveExists,
		Importer: &schema.ResourceImporter{
			State: resourceDatadogServiceLevelObjectiveImport,
		},

		Schema: map[string]*schema.Schema{
			// Common
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				StateFunc: func(val interface{}) string {
					return strings.TrimSpace(val.(string))
				},
			},
			"tags": {
				// we use TypeSet to represent tags, paradoxically to be able to maintain them ordered;
				// we order them explicitly in the read/create/update methods of this resource and using
				// TypeSet makes Terraform ignore differences in order when creating a plan
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"thresholds": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"timeframe": {
							Type:     schema.TypeString,
							Required: true,
						},
						"target": {
							Type:             schema.TypeFloat,
							Required:         true,
							DiffSuppressFunc: suppressDataDogFloatIntDiff,
						},
						"target_display": {
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: suppressDataDogSLODisplayValueDiff,
						},
						"warning": {
							Type:             schema.TypeFloat,
							Optional:         true,
							DiffSuppressFunc: suppressDataDogFloatIntDiff,
						},
						"warning_display": {
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: suppressDataDogSLODisplayValueDiff,
						},
					},
				},
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: ValidateServiceLevelObjectiveTypeString,
			},

			// Metric-Based SLO
			"query": {
				// we use TypeList here because of https://github.com/hashicorp/terraform/issues/6215/
				Type:          schema.TypeList,
				MaxItems:      1,
				Optional:      true,
				ConflictsWith: []string{"monitor_ids", "groups"},
				Description:   "The metric query of good / total events",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"numerator": {
							Type:     schema.TypeString,
							Required: true,
							StateFunc: func(val interface{}) string {
								return strings.TrimSpace(val.(string))
							},
						},
						"denominator": {
							Type:     schema.TypeString,
							Required: true,
							StateFunc: func(val interface{}) string {
								return strings.TrimSpace(val.(string))
							},
						},
					},
				},
			},

			// Monitor-Based SLO
			"monitor_ids": {
				Type:          schema.TypeSet,
				Optional:      true,
				ConflictsWith: []string{"query"},
				Description:   "A static set of monitor IDs to use as part of the SLO",
				Elem:          &schema.Schema{Type: schema.TypeInt, MinItems: 1},
			},
			// NOTE: This feature was introduced but it never worked and then it was removed.
			// We didn't trigger a major release since it never worked. However, this may be introduced later again.
			// Keeping this here for now and we removed the related code.
			"monitor_search": {
				Type:     schema.TypeString,
				Optional: true,
				Removed:  "Feature is not yet supported",
				Computed: true,
			},
			"groups": {
				Type:          schema.TypeSet,
				Optional:      true,
				Description:   "A static set of groups to filter monitor-based SLOs",
				ConflictsWith: []string{"query"},
				Elem:          &schema.Schema{Type: schema.TypeString, MinItems: 1},
			},
		},
	}
}

// ValidateServiceLevelObjectiveTypeString is a ValidateFunc that ensures the SLO is of one of the supported types
func ValidateServiceLevelObjectiveTypeString(v interface{}, k string) (ws []string, errors []error) {
	switch v.(string) {
	case string(datadog.SERVICELEVELOBJECTIVETYPE_MONITOR):
		break
	case string(datadog.SERVICELEVELOBJECTIVETYPE_METRIC):
		break
	default:
		errors = append(errors, fmt.Errorf("invalid type %s specified for SLO", v.(string)))
	}
	return
}

func buildServiceLevelObjectiveStruct(d *schema.ResourceData) *datadog.ServiceLevelObjective {

	slo := &datadog.ServiceLevelObjective{
		Id: datadog.PtrString(d.Id()),
	}

	if attr, ok := d.GetOk("name"); ok {
		slo.SetName(attr.(string))
	}

	if attr, ok := d.GetOk("description"); ok {
		slo.SetDescription(*datadog.NewNullableString(datadog.PtrString(attr.(string))))
	}

	if attr, ok := d.GetOk("type"); ok {
		slo.SetType(datadog.ServiceLevelObjectiveType(attr.(string)))
	}

	if attr, ok := d.GetOk("tags"); ok {
		tags := make([]string, 0)
		for _, v := range attr.(*schema.Set).List() {
			tags = append(tags, v.(string))
		}
		if len(tags) > 0 {
			sort.Strings(tags)
			slo.SetTags(tags)
		}
	}

	if _, ok := d.GetOk("thresholds"); ok {
		numThresholds := d.Get("thresholds.#").(int)
		sloThresholds := make([]datadog.SLOThreshold, 0)
		for i := 0; i < numThresholds; i++ {
			prefix := fmt.Sprintf("thresholds.%d.", i)
			t := datadog.SLOThreshold{}

			if tf, ok := d.GetOk(prefix + "timeframe"); ok {
				t.SetTimeframe(datadog.SLOTimeframe(tf.(string)))
			}

			if targetValue, ok := d.GetOk(prefix + "target"); ok {
				if f, ok := floatOk(targetValue); ok {
					t.SetTarget(float64(f))
				}
			}

			if warningValue, ok := d.GetOk(prefix + "warning"); ok {
				if f, ok := floatOk(warningValue); ok {
					t.SetWarning(float64(f))
				}
			}

			if targetDisplayValue, ok := d.GetOk(prefix + "target_display"); ok {
				if s, ok := targetDisplayValue.(string); ok && strings.TrimSpace(s) != "" {
					t.SetTargetDisplay(strings.TrimSpace(targetDisplayValue.(string)))
				}
			}

			if warningDisplayValue, ok := d.GetOk(prefix + "warning_display"); ok {
				if s, ok := warningDisplayValue.(string); ok && strings.TrimSpace(s) != "" {
					t.SetWarningDisplay(strings.TrimSpace(warningDisplayValue.(string)))
				}
			}
			sloThresholds = append(sloThresholds, t)
		}
		if len(sloThresholds) > 0 {
			sort.Slice(sloThresholds, func(i, j int) bool {
				return convertTimeframeToInt(sloThresholds[i].GetTimeframe()) < convertTimeframeToInt(sloThresholds[j].GetTimeframe())
			})
			slo.SetThresholds(sloThresholds)
		}
	}

	switch string(slo.GetType()) {
	case string(datadog.SERVICELEVELOBJECTIVETYPE_MONITOR):
		// monitor type
		if attr, ok := d.GetOk("monitor_ids"); ok {
			mIds := make([]int64, 0)
			for _, v := range attr.(*schema.Set).List() {
				mIds = append(mIds, int64(v.(int)))
			}
			sort.Slice(mIds, func(i, j int) bool { return mIds[i] < mIds[j] })
			slo.SetMonitorIds(mIds)
		}
		//if attr, ok := d.GetOk("monitor_search"); ok {
		//	slo.SetMonitorSearch(attr.(string))
		//}
		if attr, ok := d.GetOk("groups"); ok {
			groups := make([]string, 0)
			for _, v := range attr.(*schema.Set).List() {
				groups = append(groups, v.(string))
			}
			sort.Strings(groups)
			slo.SetGroups(groups)
		}
	default:
		// metric type
		if attr, ok := d.GetOk("query"); ok {
			queries := make([]map[string]interface{}, 0)
			raw := attr.([]interface{})
			for _, rawQuery := range raw {
				if query, ok := rawQuery.(map[string]interface{}); ok {
					queries = append(queries, query)
				}
			}
			if len(queries) >= 1 {
				// only use the first defined query
				slo.SetQuery(datadog.ServiceLevelObjectiveQuery{
					Numerator:   queries[0]["numerator"].(string),
					Denominator: queries[0]["denominator"].(string),
				})
			}
		}
	}

	return slo
}

func floatOk(val interface{}) (float64, bool) {
	switch val.(type) {
	case float64:
		return val.(float64), true
	case *float64:
		return *(val.(*float64)), true
	case string:
		f, err := strconv.ParseFloat(val.(string), 64)
		if err == nil {
			return f, true
		}
	case *string:
		f, err := strconv.ParseFloat(*(val.(*string)), 64)
		if err == nil {
			return f, true
		}
	default:
		return 0, false
	}
	return 0, false
}

func resourceDatadogServiceLevelObjectiveCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	slo := buildServiceLevelObjectiveStruct(d)
	sloList, _, err := client.SLOApi.CreateSLO(auth).Body(*slo).Execute()
	if err != nil {
		return translateClientError(err, "error creating service level objective")
	}

	d.SetId(sloList.GetData()[0].GetId())

	return resourceDatadogServiceLevelObjectiveRead(d, meta)
}

func resourceDatadogServiceLevelObjectiveExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	// Exists - This is called to verify a resource still exists. It is called prior to Read,
	// and lowers the burden of Read to be able to assume the resource exists.
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	if _, _, err := client.SLOApi.GetSLO(auth, d.Id()).Execute(); err != nil {
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "not found") || strings.Contains(errStr, "no slo specified") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func resourceDatadogServiceLevelObjectiveRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	sloList, _, err := client.SLOApi.GetSLO(auth, d.Id()).Execute()
	if err != nil {
		return translateClientError(err, "error getting service level objective")
	}
	slo := sloList.GetData()
	d.Set("name", slo.GetName())
	d.Set("description", slo.GetDescription())
	d.Set("type", slo.GetType())

	thresholds := make([]map[string]interface{}, 0)
	for _, threshold := range slo.Thresholds {
		t := map[string]interface{}{
			"timeframe": threshold.GetTimeframe(),
			"target":    threshold.GetTarget(),
		}
		if warning, ok := threshold.GetWarningOk(); ok {
			t["warning"] = warning
		}
		if targetDisplay, ok := threshold.GetTargetDisplayOk(); ok {
			t["target_display"] = targetDisplay
		}
		if warningDisplay, ok := threshold.GetWarningDisplayOk(); ok {
			t["warning_display"] = warningDisplay
		}
		thresholds = append(thresholds, t)
	}
	sort.Slice(thresholds, func(i, j int) bool {
		return convertTimeframeToInt(thresholds[i]["timeframe"].(datadog.SLOTimeframe)) < convertTimeframeToInt(thresholds[j]["timeframe"].(datadog.SLOTimeframe))
	})
	d.Set("thresholds", thresholds)

	tags := make([]string, 0)
	for _, s := range slo.GetTags() {
		tags = append(tags, s)
	}
	sort.Strings(tags)
	d.Set("tags", tags)

	switch string(slo.GetType()) {
	case string(datadog.SERVICELEVELOBJECTIVETYPE_MONITOR):
		// monitor type
		if len(slo.GetMonitorIds()) > 0 {
			mIds := make([]int64, 0)
			for _, m := range slo.GetMonitorIds() {
				mIds = append(mIds, m)
			}
			sort.Slice(mIds, func(i, j int) bool { return mIds[i] < mIds[j] })
			d.Set("monitor_ids", mIds)
		}
		//if ms, ok := slo.GetMonitorSearchOk(); ok {
		//	d.Set("monitor_search", ms)
		//}
		groups := slo.GetGroups()
		sort.Strings(groups)
		d.Set("groups", groups)
	default:
		// metric type
		query := make(map[string]interface{})
		q := slo.GetQuery()
		query["numerator"] = q.GetNumerator()
		query["denominator"] = q.GetDenominator()
		d.Set("query", []map[string]interface{}{query})
	}

	return nil
}

func resourceDatadogServiceLevelObjectiveUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	slo := buildServiceLevelObjectiveStruct(d)

	if _, _, err := client.SLOApi.EditSLO(auth, slo.GetId()).Body(*slo).Execute(); err != nil {
		return translateClientError(err, "error updating service level objective")
	}

	return resourceDatadogServiceLevelObjectiveRead(d, meta)
}

func resourceDatadogServiceLevelObjectiveDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.DatadogClientV1
	auth := providerConf.Auth

	if _, _, err := client.SLOApi.DeleteSLO(auth, d.Id()).Execute(); err != nil {
		return translateClientError(err, "error deleting service level objective")
	}
	return nil
}

func resourceDatadogServiceLevelObjectiveImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceDatadogServiceLevelObjectiveRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

// Ignore any diff that results from the mix of *_display string values from the
// DataDog API.
func suppressDataDogSLODisplayValueDiff(k, old, new string, d *schema.ResourceData) bool {
	sloType := d.Get("type").(string)
	if sloType == string(datadog.SERVICELEVELOBJECTIVETYPE_MONITOR) {
		// always suppress monitor type, this is controlled via API.
		return true
	}

	// metric type otherwise
	if old == "" || new == "" {
		// always suppress if not specified
		return true
	}

	return suppressDataDogFloatIntDiff(k, old, new, d)
}

func convertTimeframeToInt(timeframe datadog.SLOTimeframe) int {
	switch timeframe {
	case datadog.SLOTIMEFRAME_SEVEN_DAYS:
		return 0
	case datadog.SLOTIMEFRAME_THIRTY_DAYS:
		return 1
	case datadog.SLOTIMEFRAME_NINETY_DAYS:
		return 2
	}
	return -1
}
