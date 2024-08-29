package datadog

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func getTimeseriesQuerySchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "A timeseries query, containing named data-source-specific queries and a formula involving the named queries.",
		MaxItems:    1,
		Required:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"formula": {
					Type:        schema.TypeList,
					Description: "A list that contains exactly one formula, as only a single formula may be used to define a timeseries query for a time-slice SLO.",
					Required:    true,
					MinItems:    1,
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"formula_expression": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "The formula string, which is an expression involving named queries.",
							},
						},
					},
				},
				"query": {
					Type:        schema.TypeList,
					Description: "A list of data-source-specific queries that are in the formula.",
					Required:    true,
					MinItems:    1,
					Elem: &schema.Resource{
						// Note this purposefully mirrors "metric_query" defined in resource_datadog_dashboard.go in the `getFormulaQuerySchema()` function.
						// One difference is that we don't support the "aggregator" field here, as it's not supported by the SLO API.
						// We may support "event_query" in the future, but for now we only support "metric_query".
						Schema: map[string]*schema.Schema{
							"metric_query": {
								Type:        schema.TypeList,
								Optional:    true,
								MaxItems:    1,
								Description: "A timeseries formula and functions metrics query.",
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"data_source": {
											Type:        schema.TypeString,
											Optional:    true,
											Default:     "metrics",
											Description: "The data source for metrics queries.",
										},
										"query": {
											Type:        schema.TypeString,
											Required:    true,
											Description: "The metrics query definition.",
										},
										"name": {
											Type:        schema.TypeString,
											Required:    true,
											Description: "The name of the query for use in formulas.",
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

func resourceDatadogServiceLevelObjective() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog service level objective resource. This can be used to create and manage Datadog service level objectives.",
		CreateContext: resourceDatadogServiceLevelObjectiveCreate,
		ReadContext:   resourceDatadogServiceLevelObjectiveRead,
		UpdateContext: resourceDatadogServiceLevelObjectiveUpdate,
		DeleteContext: resourceDatadogServiceLevelObjectiveDelete,
		CustomizeDiff: resourceDatadogServiceLevelObjectiveCustomizeDiff,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
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
					Elem: &schema.Schema{
						Type: schema.TypeString,
						StateFunc: func(val any) string {
							return utils.NormalizeTag(val.(string))
						},
					},
				},
				"thresholds": {
					Description: "A list of thresholds and targets that define the service level objectives from the provided SLIs.",
					Type:        schema.TypeList,
					Required:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"timeframe": {
								Description:      "The time frame for the objective. The mapping from these types to the types found in the Datadog Web UI can be found in the Datadog API documentation page.",
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSLOTimeframeFromValue),
							},
							"target": {
								Description:      "The objective's target in `(0,100)`.",
								Type:             schema.TypeFloat,
								Required:         true,
								DiffSuppressFunc: suppressDataDogFloatIntDiff,
							},
							"target_display": {
								Description: "A string representation of the target that indicates its precision. It uses trailing zeros to show significant decimal places (e.g. `98.00`).",
								Type:        schema.TypeString,
								Computed:    true,
							},
							"warning": {
								Description:      "The objective's warning value in `(0,100)`. This must be greater than the target value.",
								Type:             schema.TypeFloat,
								Optional:         true,
								DiffSuppressFunc: suppressDataDogFloatIntDiff,
							},
							"warning_display": {
								Description: "A string representation of the warning target (see the description of the target_display field for details).",
								Type:        schema.TypeString,
								Computed:    true,
							},
						},
					},
				},
				"target_threshold": {
					Description:      "The objective's target in `(0,100)`. This must match the corresponding thresholds of the primary time frame.",
					Type:             schema.TypeFloat,
					Optional:         true,
					Computed:         true,
					DiffSuppressFunc: suppressDataDogFloatIntDiff,
				},
				"warning_threshold": {
					Description:      "The objective's warning value in `(0,100)`. This must be greater than the target value and match the corresponding thresholds of the primary time frame.",
					Type:             schema.TypeFloat,
					Optional:         true,
					Computed:         true,
					DiffSuppressFunc: suppressDataDogFloatIntDiff,
				},
				"timeframe": {
					Description:      "The primary time frame for the objective. The mapping from these types to the types found in the Datadog Web UI can be found in the Datadog API documentation page.",
					Type:             schema.TypeString,
					Optional:         true,
					Computed:         true,
					ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSLOTimeframeFromValue),
				},
				"type": {
					Description:      "The type of the service level objective. The mapping from these types to the types found in the Datadog Web UI can be found in the Datadog API [documentation page](https://docs.datadoghq.com/api/v1/service-level-objectives/#create-a-slo-object).",
					Type:             schema.TypeString,
					Required:         true,
					ForceNew:         true,
					ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSLOTypeFromValue),
				},
				"force_delete": {
					Description: "A boolean indicating whether this monitor can be deleted even if it's referenced by other resources (for example, dashboards).",
					Type:        schema.TypeBool,
					Optional:    true,
				},

				// Metric-Based SLO
				"query": {
					// we use TypeList here because of https://github.com/hashicorp/terraform/issues/6215/
					Type:          schema.TypeList,
					MaxItems:      1,
					Optional:      true,
					ConflictsWith: []string{"monitor_ids", "sli_specification", "groups"},
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
					ConflictsWith: []string{"query", "sli_specification"},
					Description:   "A static set of monitor IDs to use as part of the SLO",
					Elem:          &schema.Schema{Type: schema.TypeInt, MinItems: 1},
				},

				// Time-Slice SLO
				"sli_specification": {
					Type:          schema.TypeList,
					MinItems:      1,
					MaxItems:      1,
					Optional:      true,
					ConflictsWith: []string{"query", "monitor_ids", "groups"},
					Description:   "A map of SLI specifications to use as part of the SLO.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"time_slice": {
								Type:        schema.TypeList,
								MinItems:    1,
								MaxItems:    1,
								Required:    true,
								Description: "The time slice condition, composed of 3 parts: 1. The timeseries query, 2. The comparator, and 3. The threshold. Optionally, a fourth part, the query interval, can be provided.",
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"comparator": {
											Type:             schema.TypeString,
											Required:         true,
											Description:      "The comparator used to compare the SLI value to the threshold.",
											ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSLOTimeSliceComparatorFromValue),
										},
										"threshold": {
											Type:        schema.TypeFloat,
											Required:    true,
											Description: "The threshold value to which each SLI value will be compared.",
										},
										"query": getTimeseriesQuerySchema(),
										"query_interval_seconds": {
											Type:             schema.TypeInt,
											Optional:         true,
											Default:          300,
											Description:      "The interval used when querying data, which defines the size of a time slice.",
											ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSLOTimeSliceIntervalFromValue),
										},
									},
								},
							},
						},
					},
				},

				"groups": {
					Type:          schema.TypeSet,
					Optional:      true,
					Description:   "A static set of groups to filter monitor-based SLOs",
					ConflictsWith: []string{"query"},
					Elem:          &schema.Schema{Type: schema.TypeString, MinItems: 1},
				},
				"validate": {
					Description: "Whether or not to validate the SLO. It checks if monitors added to a monitor SLO already exist.",
					Type:        schema.TypeBool,
					Optional:    true,
					DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
						// This is never sent to the backend, so it should never generate a diff
						return true
					},
				},
			}
		},
	}
}

// ValidateServiceLevelObjectiveTypeString is a ValidateFunc that ensures the SLO is of one of the supported types

// Use CustomizeDiff to do monitor validation
func resourceDatadogServiceLevelObjectiveCustomizeDiff(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
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
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	if attr, ok := diff.GetOk("monitor_ids"); ok {
		for _, v := range attr.(*schema.Set).List() {
			// Check that each monitor being added to the SLO exists
			if _, httpResponse, err := apiInstances.GetMonitorsApiV1().GetMonitor(auth, int64(v.(int))); err != nil {
				return utils.TranslateClientError(err, httpResponse, "error finding monitor to add to SLO")
			}
		}
	}

	return nil
}

func buildSLOTimeSliceQueryStruct(d []interface{}) *datadogV1.SLOTimeSliceQuery {
	// only use the first defined query
	ret := datadogV1.NewSLOTimeSliceQueryWithDefaults()
	ret.Formulas = make([]datadogV1.SLOFormula, 0)
	ret.Queries = make([]datadogV1.SLODataSourceQueryDefinition, 0)
	if (len(d)) > 0 {
		if raw, ok := d[0].(map[string]interface{}); ok {
			if rawFormulas, ok := raw["formula"].([]interface{}); ok {
				for _, rawFormulaEl := range rawFormulas {
					if rawFormula, ok := rawFormulaEl.(map[string]interface{}); ok {
						if formula, ok := rawFormula["formula_expression"].(string); ok {
							ret.Formulas = append(ret.Formulas, *datadogV1.NewSLOFormula(formula))
						}
					}
				}
			}
			if rawQueries, ok := raw["query"].([]interface{}); ok {
				for _, rawQueryEl := range rawQueries {
					rawQuery := rawQueryEl.(map[string]interface{})
					rawMetricQueries := rawQuery["metric_query"].([]interface{})
					if len(rawMetricQueries) >= 1 {
						if rawMetricQuery, ok := rawMetricQueries[0].(map[string]interface{}); ok {
							name := rawMetricQuery["name"].(string)
							query := rawMetricQuery["query"].(string)
							rawDataSource := rawMetricQuery["data_source"].(string)
							dataSource, _ := datadogV1.NewFormulaAndFunctionMetricDataSourceFromValue(rawDataSource)
							ret.Queries = append(ret.Queries,
								datadogV1.FormulaAndFunctionMetricQueryDefinitionAsSLODataSourceQueryDefinition(
									datadogV1.NewFormulaAndFunctionMetricQueryDefinition(*dataSource, name, query)))
						}
					}
				}
			}
		}
	}
	return ret
}

func buildServiceLevelObjectiveStructs(d *schema.ResourceData) (*datadogV1.ServiceLevelObjective, *datadogV1.ServiceLevelObjectiveRequest, error) {

	slo := datadogV1.NewServiceLevelObjectiveWithDefaults()
	slo.SetName(d.Get("name").(string))
	slo.SetType(datadogV1.SLOType(d.Get("type").(string)))
	slo.SetId(d.Id())

	slor := datadogV1.NewServiceLevelObjectiveRequestWithDefaults()
	slor.SetName(d.Get("name").(string))
	slor.SetType(datadogV1.SLOType(d.Get("type").(string)))

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
		} else {
			return nil, nil, fmt.Errorf("monitor_ids is required for monitor SLOs")
		}
		if attr, ok := d.GetOk("groups"); ok {
			s := make([]string, 0)
			for _, v := range attr.(*schema.Set).List() {
				s = append(s, v.(string))
			}
			slo.SetGroups(s)
			slor.SetGroups(s)
		}
	case datadogV1.SLOTYPE_TIME_SLICE:
		var sliSpec datadogV1.SLOSliSpec
		if attr, ok := d.GetOk("sli_specification"); ok {
			raw := attr.([]interface{})
			if len(raw) >= 1 {
				rawSliSpec := raw[0].(map[string]interface{})
				if rawTimeSliceSpec, ok := rawSliSpec["time_slice"]; ok {
					sliSpec.SLOTimeSliceSpec = datadogV1.NewSLOTimeSliceSpecWithDefaults()
					rawTimeSliceConds := rawTimeSliceSpec.([]interface{})
					if len(rawTimeSliceConds) >= 1 {
						rawTimeSliceCond := rawTimeSliceConds[0].(map[string]interface{})
						if rawTimeSliceQuery, ok := rawTimeSliceCond["query"].([]interface{}); ok {
							sliSpec.SLOTimeSliceSpec.TimeSlice.SetQuery(*buildSLOTimeSliceQueryStruct(rawTimeSliceQuery))
						}
						if comparator, ok := rawTimeSliceCond["comparator"].(string); ok {
							sliSpec.SLOTimeSliceSpec.TimeSlice.SetComparator(datadogV1.SLOTimeSliceComparator(comparator))
						}
						if threshold, ok := rawTimeSliceCond["threshold"].(float64); ok {
							sliSpec.SLOTimeSliceSpec.TimeSlice.SetThreshold(threshold)
						}
						if queryInterval, ok := rawTimeSliceCond["query_interval_seconds"].(int); ok {
							// Terraform doesn't have a way to represent an optional int, and so we
							// will get a 0 value if the user doesn't specify a query_interval_seconds.
							if queryInterval != 0 {
								sliSpec.SLOTimeSliceSpec.TimeSlice.SetQueryIntervalSeconds(datadogV1.SLOTimeSliceInterval(queryInterval))
							}
						}
					}
				}
			}
		} else {
			return nil, nil, fmt.Errorf("sli_specification is required for time slice SLOs")
		}
		slo.SetSliSpecification(sliSpec)
		slor.SetSliSpecification(sliSpec)
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
		} else {
			return nil, nil, fmt.Errorf("query is required for metric SLOs")
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

			sloThresholds = append(sloThresholds, *t)
		}
		if len(sloThresholds) > 0 {
			slo.SetThresholds(sloThresholds)
			slor.SetThresholds(sloThresholds)
		}
	}

	plan := d.GetRawConfig()
	if tf, ok := d.GetOk("timeframe"); ok {
		if plan.GetAttr("timeframe").IsNull() {
			d.Set("timeframe", nil)
		} else {
			slo.SetTimeframe(datadogV1.SLOTimeframe(tf.(string)))
			slor.SetTimeframe(datadogV1.SLOTimeframe(tf.(string)))
		}
	}

	if targetValue, ok := d.GetOk("target_threshold"); ok {
		if plan.GetAttr("target_threshold").IsNull() {
			d.Set("target_threshold", nil)
		} else if f, ok := floatOk(targetValue); ok {
			slo.SetTargetThreshold(f)
			slor.SetTargetThreshold(f)
		}
	}

	if warningValue, ok := d.GetOk("warning_threshold"); ok {
		if plan.GetAttr("warning_threshold").IsNull() {
			d.Set("warning_threshold", nil)
		} else if f, ok := floatOk(warningValue); ok {
			slo.SetWarningThreshold(f)
			slor.SetWarningThreshold(f)
		}
	}

	return slo, slor, nil
}

func floatOk(val interface{}) (float64, bool) {
	switch val := val.(type) {
	case float64:
		return val, true
	case *float64:
		return *val, true
	case string:
		f, err := strconv.ParseFloat(val, 64)
		if err == nil {
			return f, true
		}
	case *string:
		f, err := strconv.ParseFloat(*val, 64)
		if err == nil {
			return f, true
		}
	default:
		return 0, false
	}
	return 0, false
}

func resourceDatadogServiceLevelObjectiveCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	_, slor, err := buildServiceLevelObjectiveStructs(d)
	if err != nil {
		return diag.FromErr(err)
	}
	sloResp, httpResponse, err := apiInstances.GetServiceLevelObjectivesApiV1().CreateSLO(auth, *slor)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating service level objective")
	}
	if err := utils.CheckForUnparsed(sloResp); err != nil {
		return diag.FromErr(err)
	}

	slo := &sloResp.GetData()[0]
	d.SetId(slo.GetId())

	return updateSLOState(d, slo)
}

func resourceDatadogServiceLevelObjectiveRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	sloResp, httpresp, err := apiInstances.GetServiceLevelObjectivesApiV1().GetSLO(auth, d.Id())
	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpresp, "error getting service level objective")
	}
	if err := utils.CheckForUnparsed(sloResp); err != nil {
		return diag.FromErr(err)
	}

	return updateSLOStateFromRead(d, sloResp.Data)
}

// Builds the corresponding terraform representation of the SLO's SLI specification
func buildTerraformSliSpecification(sliSpec *datadogV1.SLOSliSpec) []map[string]interface{} {
	rawSliSpec := make([]map[string]interface{}, 0)
	if sliSpec.SLOTimeSliceSpec != nil {
		rawTimeSliceSpec := make([]map[string]interface{}, 0)
		comparator := sliSpec.SLOTimeSliceSpec.TimeSlice.GetComparator()
		threshold := sliSpec.SLOTimeSliceSpec.TimeSlice.GetThreshold()
		query := sliSpec.SLOTimeSliceSpec.TimeSlice.GetQuery()
		rawFormulas := make([]map[string]interface{}, 0)
		for _, formula := range query.GetFormulas() {
			rawFormula := map[string]interface{}{"formula_expression": formula.GetFormula()}
			rawFormulas = append(rawFormulas, rawFormula)
		}
		rawQueries := make([]map[string]interface{}, 0)
		for _, q := range query.GetQueries() {
			rawMetricQueries := make([]map[string]interface{}, 0)
			rawQuery := map[string]interface{}{
				"name":        q.FormulaAndFunctionMetricQueryDefinition.GetName(),
				"data_source": q.FormulaAndFunctionMetricQueryDefinition.GetDataSource(),
				"query":       q.FormulaAndFunctionMetricQueryDefinition.GetQuery(),
			}
			rawMetricQueries = append(rawMetricQueries, rawQuery)
			rawQueries = append(rawQueries, map[string]interface{}{"metric_query": rawMetricQueries})
		}
		rawQuery := make([]map[string]interface{}, 0)
		rawQuery = append(rawQuery, map[string]interface{}{
			"formula": rawFormulas,
			"query":   rawQueries,
		})
		rawTimeSliceCond := map[string]interface{}{
			"comparator": comparator,
			"threshold":  threshold,
			"query":      rawQuery,
		}
		if queryInterval, ok := sliSpec.SLOTimeSliceSpec.TimeSlice.GetQueryIntervalSecondsOk(); ok {
			rawTimeSliceCond["query_interval_seconds"] = *queryInterval
		}
		rawTimeSliceSpec = append(rawTimeSliceSpec, rawTimeSliceCond)
		rawSliSpec = append(rawSliSpec, map[string]interface{}{"time_slice": rawTimeSliceSpec})
	}
	return rawSliSpec
}

func updateSLOState(d *schema.ResourceData, slo *datadogV1.ServiceLevelObjective) diag.Diagnostics {
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
	tags = append(tags, slo.GetTags()...)

	if err := d.Set("name", slo.GetName()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", slo.GetDescription()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("type", slo.GetType()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tags", tags); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("thresholds", thresholds); err != nil {
		return diag.FromErr(err)
	}
	if timeframe, ok := slo.GetTimeframeOk(); ok {
		if err := d.Set("timeframe", timeframe); err != nil {
			return diag.FromErr(err)
		}
	}
	if target, ok := slo.GetTargetThresholdOk(); ok {
		if err := d.Set("target_threshold", target); err != nil {
			return diag.FromErr(err)
		}
	}
	if warning, ok := slo.GetWarningThresholdOk(); ok {
		if err := d.Set("warning_threshold", warning); err != nil {
			return diag.FromErr(err)
		}
	}
	switch slo.GetType() {
	case datadogV1.SLOTYPE_MONITOR:
		// monitor type
		if len(slo.GetMonitorIds()) > 0 {
			if err := d.Set("monitor_ids", slo.GetMonitorIds()); err != nil {
				return diag.FromErr(err)
			}
		}
		if err := d.Set("groups", slo.GetGroups()); err != nil {
			return diag.FromErr(err)
		}
	case datadogV1.SLOTYPE_TIME_SLICE:
		// time slice type
		sliSpec := slo.GetSliSpecification()
		tfSliSpec := buildTerraformSliSpecification(&sliSpec)
		if err := d.Set("sli_specification", tfSliSpec); err != nil {
			return diag.FromErr(err)
		}
	default:
		// metric type
		query := make(map[string]interface{})
		q := slo.GetQuery()
		query["numerator"] = q.GetNumerator()
		query["denominator"] = q.GetDenominator()
		if err := d.Set("query", []map[string]interface{}{query}); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

// This duplicates updateSLOState for the SLOResponseData structure, which has mostly the same interface
func updateSLOStateFromRead(d *schema.ResourceData, slo *datadogV1.SLOResponseData) diag.Diagnostics {
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
	tags = append(tags, slo.GetTags()...)

	if err := d.Set("name", slo.GetName()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", slo.GetDescription()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("type", slo.GetType()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tags", tags); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("thresholds", thresholds); err != nil {
		return diag.FromErr(err)
	}
	if timeframe, ok := slo.GetTimeframeOk(); ok {
		if err := d.Set("timeframe", timeframe); err != nil {
			return diag.FromErr(err)
		}
	}
	if target, ok := slo.GetTargetThresholdOk(); ok {
		if err := d.Set("target_threshold", target); err != nil {
			return diag.FromErr(err)
		}
	}
	if warning, ok := slo.GetWarningThresholdOk(); ok {
		if err := d.Set("warning_threshold", warning); err != nil {
			return diag.FromErr(err)
		}
	}
	switch slo.GetType() {
	case datadogV1.SLOTYPE_MONITOR:
		// monitor type
		if len(slo.GetMonitorIds()) > 0 {
			if err := d.Set("monitor_ids", slo.GetMonitorIds()); err != nil {
				return diag.FromErr(err)
			}
		}
		if err := d.Set("groups", slo.GetGroups()); err != nil {
			return diag.FromErr(err)
		}
	case datadogV1.SLOTYPE_TIME_SLICE:
		// time slice type
		sliSpec := slo.GetSliSpecification()
		tfSliSpec := buildTerraformSliSpecification(&sliSpec)
		if err := d.Set("sli_specification", tfSliSpec); err != nil {
			return diag.FromErr(err)
		}
	default:
		// metric type
		query := make(map[string]interface{})
		q := slo.GetQuery()
		query["numerator"] = q.GetNumerator()
		query["denominator"] = q.GetDenominator()
		if err := d.Set("query", []map[string]interface{}{query}); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func resourceDatadogServiceLevelObjectiveUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	slo, _, err := buildServiceLevelObjectiveStructs(d)

	if err != nil {
		return diag.FromErr(err)
	}
	updatedSLO, httpResponse, err := apiInstances.GetServiceLevelObjectivesApiV1().UpdateSLO(auth, d.Id(), *slo)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error updating service level objective")
	}
	if err := utils.CheckForUnparsed(updatedSLO); err != nil {
		return diag.FromErr(err)
	}

	return updateSLOState(d, &updatedSLO.GetData()[0])
}

func resourceDatadogServiceLevelObjectiveDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	var err error

	var httpResponse *http.Response
	if d.Get("force_delete").(bool) {
		_, httpResponse, err = apiInstances.GetServiceLevelObjectivesApiV1().DeleteSLO(auth, d.Id(),
			*datadogV1.NewDeleteSLOOptionalParameters().WithForce("true"),
		)
	} else {
		_, httpResponse, err = apiInstances.GetServiceLevelObjectivesApiV1().DeleteSLO(auth, d.Id())
	}
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error deleting service level objective")
	}
	return nil

}

func trimStateValue(val interface{}) string {
	return strings.TrimSpace(val.(string))
}

// Ignore any diff for trimmed state values.
func diffTrimmedValues(_, old, new string, _ *schema.ResourceData) bool {
	return strings.TrimSpace(old) == strings.TrimSpace(new)
}
