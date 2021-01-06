package datadog

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceDatadogServiceLevelObjective() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog service level objective resource. This can be used to create and manage Datadog service level objectives.",
		Create:        resourceDatadogServiceLevelObjectiveCreate,
		Read:          resourceDatadogServiceLevelObjectiveRead,
		Update:        resourceDatadogServiceLevelObjectiveUpdate,
		Delete:        resourceDatadogServiceLevelObjectiveDelete,
		CustomizeDiff: resourceDatadogServiceLevelObjectiveCustomizeDiff,
		Importer: &schema.ResourceImporter{
			State: resourceDatadogServiceLevelObjectiveImport,
		},

		Schema: map[string]*schema.Schema{
			// Common
			"name": {
				Description: "Name of Datadog service level objective",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description:      "A description of this service level objective.",
				Type:             schema.TypeString,
				Optional:         true,
				StateFunc:        trimStateValue,
				DiffSuppressFunc: diffTrimmedValues,
			},
			"tags": {
				// we use TypeSet to represent tags, paradoxically to be able to maintain them ordered;
				// we order them explicitly in the read/create/update methods of this resource and using
				// TypeSet makes Terraform ignore differences in order when creating a plan
				Type:        schema.TypeSet,
				Description: "A list of tags to associate with your service level objective. This can help you categorize and filter service level objectives in the service level objectives page of the UI. Note: it's not currently possible to filter by these tags when querying via the API",
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"thresholds": {
				Description: "A list of thresholds and targets that define the service level objectives from the provided SLIs.",
				Type:        schema.TypeList,
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"timeframe": {
							Description:  "The time frame for the objective. The mapping from these types to the types found in the Datadog Web UI can be found in the Datadog API documentation page. Available options to choose from are: `7d`, `30d`, `90d`.",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateEnumValue(datadogV1.NewSLOTimeframeFromValue),
						},
						"target": {
							Description:      "The objective's target in`[0,100]`.",
							Type:             schema.TypeFloat,
							Required:         true,
							DiffSuppressFunc: suppressDataDogFloatIntDiff,
						},
						"target_display": {
							Description:      "A string representation of the target that indicates its precision. It uses trailing zeros to show significant decimal places (e.g. `98.00`).",
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: suppressDataDogSLODisplayValueDiff,
						},
						"warning": {
							Description:      "The objective's warning value in `[0,100]`. This must be greater than the target value.",
							Type:             schema.TypeFloat,
							Optional:         true,
							DiffSuppressFunc: suppressDataDogFloatIntDiff,
						},
						"warning_display": {
							Description:      "A string representation of the warning target (see the description of the target_display field for details).",
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: suppressDataDogSLODisplayValueDiff,
						},
					},
				},
			},
			"type": {
				Description:  "The type of the service level objective. The mapping from these types to the types found in the Datadog Web UI can be found in the Datadog API [documentation page](https://docs.datadoghq.com/api/v1/service-level-objectives/#create-a-slo-object). Available options to choose from are: `metric` and `monitor`.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateEnumValue(datadogV1.NewSLOTypeFromValue),
			},
			"force_delete": {
				Description: "A boolean indicating whether this monitor can be deleted even if it’s referenced by other resources (e.g. dashboards).",
				Type:        schema.TypeBool,
				Optional:    true,
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
							Description:      "The sum of all the `good` events.",
							Type:             schema.TypeString,
							Required:         true,
							StateFunc:        trimStateValue,
							DiffSuppressFunc: diffTrimmedValues,
						},
						"denominator": {
							Description:      "The sum of the `total` events.",
							Type:             schema.TypeString,
							Required:         true,
							StateFunc:        trimStateValue,
							DiffSuppressFunc: diffTrimmedValues,
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
			"validate": {
				Description: "Whether or not to validate the SLO.",
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

// ValidateServiceLevelObjectiveTypeString is a ValidateFunc that ensures the SLO is of one of the supported types

// ValidateServiceLevelObjectiveTypeString is a ValidateFunc that ensures the SLO is of one of the supported types

// Use CustomizeDiff to do monitor validation
func resourceDatadogServiceLevelObjectiveCustomizeDiff(diff *schema.ResourceDiff, meta interface{}) error {
	if validate, ok := diff.GetOkExists("validate"); ok && !validate.(bool) {
		// Explicitly skip validation
		log.Printf("[DEBUG] Validate is %v, skipping validation", validate.(bool))
		return nil
	}

	if val, ok := diff.GetOk("type"); ok && (val != string(datadogV1.SLOTYPE_MONITOR)) {
		// If the SLO is not a Monitor type, skip the validation
		log.Printf("[DEBUG] SLO type is: %v, skipping validation", val)
		return nil
	}

	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1
	if attr, ok := diff.GetOk("monitor_ids"); ok {
		for _, v := range attr.(*schema.Set).List() {
			// Check that each monitor being added to the SLO exists
			if _, _, err := datadogClientV1.MonitorsApi.GetMonitor(authV1, int64(v.(int))).Execute(); err != nil {
				return translateClientError(err, "error finding monitor to add to SLO")
			}
		}
	}

	return nil
}

func buildServiceLevelObjectiveStructs(d *schema.ResourceData) (*datadogV1.ServiceLevelObjective, *datadogV1.ServiceLevelObjectiveRequest) {

	slo := datadogV1.NewServiceLevelObjectiveWithDefaults()
	slo.SetName(d.Get("name").(string))
	slo.SetType(datadogV1.SLOType(d.Get("type").(string)))
	slo.SetId(d.Id())

	slor := datadogV1.NewServiceLevelObjectiveRequestWithDefaults()
	slor.SetName(d.Get("name").(string))
	slor.SetType(datadogV1.SLOType(d.Get("type").(string)))
	slor.SetId(d.Id())

	if attr, ok := d.GetOk("description"); ok {
		slo.SetDescription(attr.(string))
		slor.SetDescription(attr.(string))
	}

	switch slo.GetType() {
	case datadogV1.SLOTYPE_MONITOR:
		// monitor type
		if attr, ok := d.GetOk("monitor_ids"); ok {
			s := make([]int64, 0)
			for _, v := range attr.(*schema.Set).List() {
				s = append(s, int64(v.(int)))
			}
			slo.SetMonitorIds(s)
			slor.SetMonitorIds(s)
		}
		if attr, ok := d.GetOk("groups"); ok {
			s := make([]string, 0)
			for _, v := range attr.(*schema.Set).List() {
				s = append(s, v.(string))
			}
			slo.SetGroups(s)
			slor.SetGroups(s)
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
				slo.SetQuery(*datadogV1.NewServiceLevelObjectiveQuery(
					queries[0]["denominator"].(string),
					queries[0]["numerator"].(string)))
				slor.SetQuery(*datadogV1.NewServiceLevelObjectiveQuery(
					queries[0]["denominator"].(string),
					queries[0]["numerator"].(string)))
			}
		}
	}

	if attr, ok := d.GetOk("tags"); ok {
		s := make([]string, 0)
		for _, v := range attr.(*schema.Set).List() {
			s = append(s, v.(string))
		}
		slo.SetTags(s)
		slor.SetTags(s)
	}

	if _, ok := d.GetOk("thresholds"); ok {
		numThresholds := d.Get("thresholds.#").(int)
		sloThresholds := make([]datadogV1.SLOThreshold, 0)
		for i := 0; i < numThresholds; i++ {
			prefix := fmt.Sprintf("thresholds.%d.", i)
			t := datadogV1.NewSLOThresholdWithDefaults()

			if tf, ok := d.GetOk(prefix + "timeframe"); ok {
				t.SetTimeframe(datadogV1.SLOTimeframe(tf.(string)))
			}

			if targetValue, ok := d.GetOk(prefix + "target"); ok {
				if f, ok := floatOk(targetValue); ok {
					t.SetTarget(f)
				}
			}

			if warningValue, ok := d.GetOk(prefix + "warning"); ok {
				if f, ok := floatOk(warningValue); ok {
					t.SetWarning(f)
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
			sloThresholds = append(sloThresholds, *t)
		}
		if len(sloThresholds) > 0 {
			slo.SetThresholds(sloThresholds)
			slor.SetThresholds(sloThresholds)
		}
	}

	return slo, slor
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
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	_, slor := buildServiceLevelObjectiveStructs(d)
	sloResp, _, err := datadogClientV1.ServiceLevelObjectivesApi.CreateSLO(authV1).Body(*slor).Execute()
	if err != nil {
		return translateClientError(err, "error creating service level objective")
	}

	slo := &sloResp.GetData()[0]
	d.SetId(slo.GetId())

	return resourceDatadogServiceLevelObjectiveRead(d, meta)
}

func resourceDatadogServiceLevelObjectiveRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	sloResp, httpresp, err := datadogClientV1.ServiceLevelObjectivesApi.GetSLO(authV1, d.Id()).Execute()
	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return translateClientError(err, "error getting service level objective")
	}
	slo := sloResp.GetData()

	thresholds := make([]map[string]interface{}, 0)
	for _, threshold := range slo.GetThresholds() {
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

	tags := make([]string, 0)
	for _, s := range slo.GetTags() {
		tags = append(tags, s)
	}

	d.Set("name", slo.GetName())
	d.Set("description", slo.GetDescription())
	d.Set("type", slo.GetType())
	d.Set("tags", tags)
	d.Set("thresholds", thresholds)
	switch slo.GetType() {
	case datadogV1.SLOTYPE_MONITOR:
		// monitor type
		if len(slo.GetMonitorIds()) > 0 {
			d.Set("monitor_ids", slo.GetMonitorIds())
		}
		d.Set("groups", slo.GetGroups())
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
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1
	slo, _ := buildServiceLevelObjectiveStructs(d)

	if _, _, err := datadogClientV1.ServiceLevelObjectivesApi.UpdateSLO(authV1, d.Id()).Body(*slo).Execute(); err != nil {
		return translateClientError(err, "error updating service level objective")
	}

	return resourceDatadogServiceLevelObjectiveRead(d, meta)
}

func resourceDatadogServiceLevelObjectiveDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1
	var err error

	if d.Get("force_delete").(bool) {
		_, _, err = datadogClientV1.ServiceLevelObjectivesApi.DeleteSLO(authV1, d.Id()).Force("true").Execute()
	} else {
		_, _, err = datadogClientV1.ServiceLevelObjectivesApi.DeleteSLO(authV1, d.Id()).Execute()
	}
	if err != nil {
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
	sloType := d.Get("type")
	if sloType == datadogV1.SLOTYPE_MONITOR {
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

func trimStateValue(val interface{}) string {
	return strings.TrimSpace(val.(string))
}

// Ignore any diff for trimmed state values.
func diffTrimmedValues(k, old, new string, d *schema.ResourceData) bool {
	return strings.TrimSpace(old) == strings.TrimSpace(new)
}
