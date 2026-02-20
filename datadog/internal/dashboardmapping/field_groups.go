package dashboardmapping

// field_groups.go
//
// Reusable []FieldSpec variables that mirror OpenAPI components/schemas/ entries.
// Each variable is named after the OpenAPI schema it corresponds to (camelCase).
// A comment on each variable identifies the OpenAPI schema and which widget types use it.
//
// Also contains:
//   - commonWidgetFields: FieldSpecs shared by most widget definition types
//   - Dashboard top-level fields and template variable field groups

// ============================================================
// Reusable FieldSpec Groups (mirroring OpenAPI $ref schemas)
// ============================================================

// widgetCustomLinkFields corresponds to OpenAPI components/schemas/WidgetCustomLink.
// Used by: timeseries, toplist, query_value, change, distribution, heatmap, hostmap,
//
//	geomap, scatterplot, service_map, sunburst, table, topology_map, treemap,
//	run_workflow (15 widget types).
//
// HCL key: "custom_link" (singular Terraform convention)
// JSON key: "custom_links" (plural, matching OpenAPI)
var widgetCustomLinkFields = []FieldSpec{
	{HCLKey: "label", Type: TypeString, OmitEmpty: true,
		Description: "The label for the custom link URL."},
	{HCLKey: "link", Type: TypeString, OmitEmpty: false,
		Description: "The URL of the custom link."},
	{HCLKey: "is_hidden", Type: TypeBool, OmitEmpty: true,
		Description: "The flag for toggling context menu link visibility."},
	{HCLKey: "override_label", Type: TypeString, OmitEmpty: true,
		Description: "The label ID that refers to a context menu link item. When `override_label` is provided, the client request omits the label field."},
}

// widgetTimeField corresponds to OpenAPI components/schemas/WidgetLegacyLiveSpan
// (the live_span variant of WidgetTime, which is the form used by HCL).
// HCL flattens this to a single "live_span" string field on the widget definition,
// which maps to {"time": {"live_span": "..."}} in JSON via JSONPath.
// Used by: 21+ widget types.
var widgetTimeField = FieldSpec{
	HCLKey:      "live_span",
	JSONPath:    "time.live_span",
	Type:        TypeString,
	OmitEmpty:   true,
	Description: "The timeframe to use when displaying the widget.",
}

// widgetAxisFields corresponds to OpenAPI components/schemas/WidgetAxis.
// Used by: timeseries (yaxis + right_yaxis), distribution, heatmap, scatterplot.
var widgetAxisFields = []FieldSpec{
	{HCLKey: "label", Type: TypeString, OmitEmpty: true,
		Description: "The label of the axis to display on the graph."},
	{HCLKey: "min", Type: TypeString, OmitEmpty: true,
		Description: "Specify the minimum value to show on the Y-axis."},
	{HCLKey: "max", Type: TypeString, OmitEmpty: true,
		Description: "Specify the maximum value to show on the Y-axis."},
	{HCLKey: "scale", Type: TypeString, OmitEmpty: true,
		Description: "Specify the scale type, options: `linear`, `log`, `pow`, `sqrt`."},
	// include_zero is always emitted even when false (OmitEmpty: false)
	// confirmed by cassette: "include_zero": false appears in right_yaxis
	{HCLKey: "include_zero", Type: TypeBool, OmitEmpty: false,
		Description: "Always include zero or fit the axis to the data range."},
}

// widgetMarkerFields corresponds to OpenAPI components/schemas/WidgetMarker.
// Used by: timeseries, distribution, heatmap.
// HCL key: "marker" (singular), JSON key: "markers" (plural).
var widgetMarkerFields = []FieldSpec{
	{HCLKey: "value", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "A mathematical expression describing the marker, for example: `y > 1`, `-5 < y < 0`, `y = 19`."},
	{HCLKey: "display_type", Type: TypeString, OmitEmpty: true,
		Description: "How the marker lines are displayed, options are one of {`error`, `warning`, `info`, `ok`} combined with one of {`dashed`, `solid`, `bold`}. Example: `error dashed`."},
	{HCLKey: "label", Type: TypeString, OmitEmpty: true,
		Description: "A label for the line or range."},
}

// widgetEventFields corresponds to OpenAPI components/schemas/WidgetEvent.
// Used by: timeseries, heatmap.
// HCL key: "event" (singular), JSON key: "events" (plural).
var widgetEventFields = []FieldSpec{
	{HCLKey: "q", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The event query to use in the widget."},
	{HCLKey: "tags_execution", Type: TypeString, OmitEmpty: true,
		Description: "The execution method for multi-value filters."},
}

// logQueryDefinitionGroupBySortFields corresponds to OpenAPI
// components/schemas/LogQueryDefinitionGroupBySort.
var logQueryDefinitionGroupBySortFields = []FieldSpec{
	{HCLKey: "aggregation", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The aggregation method."},
	{HCLKey: "order", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "Widget sorting methods.",
		ValidValues: []string{"asc", "desc"}},
	{HCLKey: "facet", Type: TypeString, OmitEmpty: true,
		Description: "The facet name."},
}

// logQueryDefinitionGroupByFields corresponds to OpenAPI
// components/schemas/LogQueryDefinitionGroupBy.
var logQueryDefinitionGroupByFields = []FieldSpec{
	{HCLKey: "facet", Type: TypeString, OmitEmpty: false,
		Description: "The facet name."},
	{HCLKey: "limit", Type: TypeInt, OmitEmpty: true,
		Description: "The maximum number of items in the group."},
	// HCL key: "sort_query" (disambiguates from other sort fields in HCL)
	// JSON key: "sort" (OpenAPI property name)
	{HCLKey: "sort_query", JSONKey: "sort", Type: TypeBlock, OmitEmpty: true,
		Description: "A list of exactly one element describing the sort query to use.",
		Children:    logQueryDefinitionGroupBySortFields},
}

// logsQueryComputeFields corresponds to OpenAPI components/schemas/LogsQueryCompute.
var logsQueryComputeFields = []FieldSpec{
	{HCLKey: "aggregation", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The aggregation method."},
	{HCLKey: "facet", Type: TypeString, OmitEmpty: true,
		Description: "The facet name."},
	{HCLKey: "interval", Type: TypeInt, OmitEmpty: true,
		Description: "Define the time interval in seconds."},
}

// logQueryDefinitionFields corresponds to OpenAPI components/schemas/LogQueryDefinition.
// Used by request fields: log_query, apm_query, rum_query, network_query,
//
//	security_query, audit_query, event_query, profile_metrics_query.
//
// That is: the same FieldSpec is reused for all 8 query-type fields on a request.
// HCL flattens "search.query" to "search_query" via JSONPath.
// HCL uses "compute_query" instead of "compute" (disambiguates from other uses); JSONKey: "compute".
var logQueryDefinitionFields = []FieldSpec{
	{HCLKey: "index", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The name of the index to query."},
	// search_query (flat HCL) → {"search": {"query": "..."}} (nested JSON) via JSONPath
	{HCLKey: "search_query", JSONPath: "search.query", Type: TypeString, OmitEmpty: false,
		Description: "The search query to use."},
	// HCL "compute_query" → JSON "compute" (renamed to avoid ambiguity in HCL)
	{HCLKey: "compute_query", JSONKey: "compute", Type: TypeBlock, OmitEmpty: true,
		Description: "`compute_query` or `multi_compute` is required. The map keys are listed below.",
		Children:    logsQueryComputeFields},
	// multi_compute → JSON "multi_compute" (same key, list of compute objects)
	{HCLKey: "multi_compute", Type: TypeBlockList, OmitEmpty: true,
		Description: "`compute_query` or `multi_compute` is required. Multiple `multi_compute` blocks are allowed using the structure below.",
		Children:    logsQueryComputeFields},
	// HCL key: "group_by" (same in HCL and JSON — no pluralization applied to this field)
	{HCLKey: "group_by", Type: TypeBlockList, OmitEmpty: true,
		Description: "Multiple `group_by` blocks are allowed using the structure below.",
		Children:    logQueryDefinitionGroupByFields},
}

// processQueryDefinitionFields corresponds to OpenAPI
// components/schemas/ProcessQueryDefinition.
// Used by timeseries and other widgets that support process metrics.
var processQueryDefinitionFields = []FieldSpec{
	{HCLKey: "metric", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "Your chosen metric."},
	{HCLKey: "search_by", Type: TypeString, OmitEmpty: true,
		Description: "Your chosen search term."},
	{HCLKey: "filter_by", Type: TypeStringList, OmitEmpty: true,
		Description: "A list of processes."},
	{HCLKey: "limit", Type: TypeInt, OmitEmpty: true,
		Description: "The max number of items in the filter list."},
}

// ============================================================
// FormulaAndFunction Query Field Groups
// ============================================================

// formulaAndFunctionMetricQueryFields corresponds to OpenAPI
// FormulaAndFunctionMetricQueryDefinition.
var formulaAndFunctionMetricQueryFields = []FieldSpec{
	{HCLKey: "data_source", Type: TypeString, OmitEmpty: false, Default: "metrics",
		Description: "The data source for metrics queries."},
	{HCLKey: "query", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The metrics query definition."},
	{HCLKey: "aggregator", Type: TypeString, OmitEmpty: true,
		Description: "The aggregation methods available for metrics queries.",
		ValidValues: []string{"avg", "last", "max", "min", "sum", "percentile"}},
	{HCLKey: "name", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The name of the query for use in formulas."},
	{HCLKey: "cross_org_uuids", Type: TypeStringList, OmitEmpty: true, MaxItems: 1,
		Description: "The source organization UUID for cross organization queries. Feature in Private Beta."},
	{HCLKey: "semantic_mode", Type: TypeString, OmitEmpty: true,
		Description: "Semantic mode for metrics queries. This determines how metrics from different sources are combined or displayed."},
}

// formulaAndFunctionEventQueryComputeFields corresponds to the compute block inside
// FormulaAndFunctionEventQueryDefinition.
var formulaAndFunctionEventQueryComputeFields = []FieldSpec{
	{HCLKey: "aggregation", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The aggregation methods for event platform queries.",
		ValidValues: []string{"count", "cardinality", "median", "pc75", "pc90", "pc95", "pc98", "pc99", "sum", "min", "max", "avg"}},
	{HCLKey: "interval", Type: TypeInt, OmitEmpty: true,
		Description: "A time interval in milliseconds."},
	{HCLKey: "metric", Type: TypeString, OmitEmpty: true,
		Description: "The measurable attribute to compute."},
}

// formulaAndFunctionEventQueryGroupBySortFields corresponds to the sort block inside
// FormulaAndFunctionEventQueryGroupBy.
var formulaAndFunctionEventQueryGroupBySortFields = []FieldSpec{
	{HCLKey: "aggregation", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The aggregation methods for the event platform queries.",
		ValidValues: []string{"count", "cardinality", "median", "pc75", "pc90", "pc95", "pc98", "pc99", "sum", "min", "max", "avg"}},
	{HCLKey: "metric", Type: TypeString, OmitEmpty: true,
		Description: "The metric used for sorting group by results."},
	{HCLKey: "order", Type: TypeString, OmitEmpty: true,
		Description: "Direction of sort.",
		ValidValues: []string{"asc", "desc"}},
}

// formulaAndFunctionEventQueryGroupByFields corresponds to FormulaAndFunctionEventQueryGroupBy.
var formulaAndFunctionEventQueryGroupByFields = []FieldSpec{
	{HCLKey: "facet", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The event facet."},
	{HCLKey: "limit", Type: TypeInt, OmitEmpty: true,
		Description: "The number of groups to return."},
	{HCLKey: "sort", Type: TypeBlock, OmitEmpty: true,
		Description: "The options for sorting group by results.",
		Children:    formulaAndFunctionEventQueryGroupBySortFields},
}

// formulaAndFunctionEventQuerySearchFields corresponds to the search block inside
// FormulaAndFunctionEventQueryDefinition.
var formulaAndFunctionEventQuerySearchFields = []FieldSpec{
	{HCLKey: "query", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The events search string."},
}

// formulaAndFunctionEventQueryFields corresponds to OpenAPI
// FormulaAndFunctionEventQueryDefinition.
var formulaAndFunctionEventQueryFields = []FieldSpec{
	{HCLKey: "data_source", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The data source for event platform-based queries.",
		ValidValues: []string{"logs", "spans", "network", "rum", "security_signals", "profiles", "audit", "events", "ci_tests", "ci_pipelines", "incident_analytics", "database_queries"}},
	{HCLKey: "storage", Type: TypeString, OmitEmpty: true,
		Description: "Storage location (private beta)."},
	{HCLKey: "search", Type: TypeBlock, OmitEmpty: true,
		Description: "The search options.",
		Children:    formulaAndFunctionEventQuerySearchFields},
	{HCLKey: "indexes", Type: TypeStringList, OmitEmpty: true,
		Description: "An array of index names to query in the stream."},
	{HCLKey: "cross_org_uuids", Type: TypeStringList, OmitEmpty: true, MaxItems: 1,
		Description: "The source organization UUID for cross organization queries. Feature in Private Beta."},
	{HCLKey: "compute", Type: TypeBlockList, OmitEmpty: false, Required: true,
		Description: "The compute options.",
		Children:    formulaAndFunctionEventQueryComputeFields},
	{HCLKey: "group_by", Type: TypeBlockList, OmitEmpty: true,
		Description: "Group by options.",
		Children:    formulaAndFunctionEventQueryGroupByFields},
	{HCLKey: "name", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The name of query for use in formulas."},
}

// formulaAndFunctionProcessQueryFields corresponds to OpenAPI
// FormulaAndFunctionProcessQueryDefinition.
var formulaAndFunctionProcessQueryFields = []FieldSpec{
	{HCLKey: "cross_org_uuids", Type: TypeStringList, OmitEmpty: true, MaxItems: 1,
		Description: "The source organization UUID for cross organization queries. Feature in Private Beta."},
	{HCLKey: "data_source", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The data source for process queries.",
		ValidValues: []string{"process", "container"}},
	{HCLKey: "metric", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The process metric name."},
	{HCLKey: "text_filter", Type: TypeString, OmitEmpty: true,
		Description: "The text to use as a filter."},
	{HCLKey: "tag_filters", Type: TypeStringList, OmitEmpty: true,
		Description: "An array of tags to filter by."},
	{HCLKey: "limit", Type: TypeInt, OmitEmpty: true,
		Description: "The number of hits to return."},
	{HCLKey: "sort", Type: TypeString, OmitEmpty: false, Default: "desc",
		Description: "The direction of the sort.",
		ValidValues: []string{"asc", "desc"}},
	{HCLKey: "aggregator", Type: TypeString, OmitEmpty: true,
		Description: "The aggregation methods available for metrics queries.",
		ValidValues: []string{"avg", "last", "max", "min", "sum", "percentile"}},
	{HCLKey: "is_normalized_cpu", Type: TypeBool, OmitEmpty: true,
		Description: "Whether to normalize the CPU percentages."},
	{HCLKey: "name", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The name of query for use in formulas."},
}

// formulaAndFunctionApmDependencyStatsQueryFields corresponds to OpenAPI
// FormulaAndFunctionApmDependencyStatsQueryDefinition.
var formulaAndFunctionApmDependencyStatsQueryFields = []FieldSpec{
	{HCLKey: "data_source", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The data source for APM Dependency Stats queries.",
		ValidValues: []string{"apm_dependency_stats"}},
	{HCLKey: "cross_org_uuids", Type: TypeStringList, OmitEmpty: true, MaxItems: 1,
		Description: "The source organization UUID for cross organization queries. Feature in Private Beta."},
	{HCLKey: "env", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "APM environment."},
	{HCLKey: "stat", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "APM statistic.",
		ValidValues: []string{"avg_duration", "avg_root_duration", "avg_spans_per_trace", "error_rate", "pct_exec_time", "pct_of_traces", "total_traces_count"}},
	{HCLKey: "operation_name", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "Name of operation on service."},
	{HCLKey: "resource_name", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "APM resource."},
	{HCLKey: "service", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "APM service."},
	{HCLKey: "primary_tag_name", Type: TypeString, OmitEmpty: true,
		Description: "The name of the second primary tag used within APM; required when `primary_tag_value` is specified."},
	{HCLKey: "primary_tag_value", Type: TypeString, OmitEmpty: true,
		Description: "Filter APM data by the second primary tag. `primary_tag_name` must also be specified."},
	{HCLKey: "is_upstream", Type: TypeBool, OmitEmpty: true,
		Description: "Determines whether stats for upstream or downstream dependencies should be queried."},
	{HCLKey: "name", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The name of query for use in formulas."},
}

// formulaAndFunctionApmResourceStatsQueryFields corresponds to OpenAPI
// FormulaAndFunctionApmResourceStatsQueryDefinition.
var formulaAndFunctionApmResourceStatsQueryFields = []FieldSpec{
	{HCLKey: "data_source", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The data source for APM Resource Stats queries.",
		ValidValues: []string{"apm_resource_stats"}},
	{HCLKey: "cross_org_uuids", Type: TypeStringList, OmitEmpty: true, MaxItems: 1,
		Description: "The source organization UUID for cross organization queries. Feature in Private Beta."},
	{HCLKey: "env", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "APM environment."},
	{HCLKey: "name", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The name of query for use in formulas."},
	{HCLKey: "stat", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "APM statistic.",
		ValidValues: []string{"errors", "error_rate", "hits", "latency_avg", "latency_distribution", "latency_max", "latency_p50", "latency_p75", "latency_p90", "latency_p95", "latency_p99"}},
	{HCLKey: "operation_name", Type: TypeString, OmitEmpty: true,
		Description: "Name of operation on service."},
	{HCLKey: "resource_name", Type: TypeString, OmitEmpty: true,
		Description: "APM resource."},
	{HCLKey: "service", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "APM service."},
	{HCLKey: "primary_tag_name", Type: TypeString, OmitEmpty: true,
		Description: "The name of the second primary tag used within APM; required when `primary_tag_value` is specified."},
	{HCLKey: "primary_tag_value", Type: TypeString, OmitEmpty: true,
		Description: "Filter APM data by the second primary tag. `primary_tag_name` must also be specified."},
	{HCLKey: "group_by", Type: TypeStringList, OmitEmpty: true,
		Description: "Array of fields to group results by."},
}

// formulaAndFunctionSLOQueryFields corresponds to OpenAPI
// FormulaAndFunctionSLOQueryDefinition.
var formulaAndFunctionSLOQueryFields = []FieldSpec{
	{HCLKey: "data_source", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The data source for SLO queries.",
		ValidValues: []string{"slo"}},
	{HCLKey: "cross_org_uuids", Type: TypeStringList, OmitEmpty: true, MaxItems: 1,
		Description: "The source organization UUID for cross organization queries. Feature in Private Beta."},
	{HCLKey: "slo_id", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "ID of an SLO to query."},
	{HCLKey: "measure", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "SLO measures queries.",
		ValidValues: []string{"good_events", "bad_events", "slo_status", "error_budget_remaining", "burn_rate", "error_budget_burndown"}},
	{HCLKey: "name", Type: TypeString, OmitEmpty: true,
		Description: "The name of query for use in formulas."},
	{HCLKey: "group_mode", Type: TypeString, OmitEmpty: false, Default: "overall",
		Description: "Group mode to query measures.",
		ValidValues: []string{"overall", "components"}},
	{HCLKey: "slo_query_type", Type: TypeString, OmitEmpty: false, Default: "metric",
		Description: "type of the SLO to query.",
		ValidValues: []string{"metric", "time_slice"}},
	{HCLKey: "additional_query_filters", Type: TypeString, OmitEmpty: true,
		Description: "Additional filters applied to the SLO query."},
}

// formulaAndFunctionCloudCostQueryFields corresponds to OpenAPI
// FormulaAndFunctionCloudCostQueryDefinition.
var formulaAndFunctionCloudCostQueryFields = []FieldSpec{
	{HCLKey: "data_source", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The data source for cloud cost queries.",
		ValidValues: []string{"cloud_cost"}},
	{HCLKey: "cross_org_uuids", Type: TypeStringList, OmitEmpty: true, MaxItems: 1,
		Description: "The source organization UUID for cross organization queries. Feature in Private Beta."},
	{HCLKey: "query", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The cloud cost query definition."},
	{HCLKey: "aggregator", Type: TypeString, OmitEmpty: true,
		Description: "The aggregation methods available for cloud cost queries.",
		ValidValues: []string{"avg", "last", "max", "min", "sum", "percentile"}},
	{HCLKey: "name", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The name of the query for use in formulas."},
}

// formulaAndFunctionQueryFields corresponds to the elements within a "query" TypeBlockList.
// Each element is a oneOf FormulaAndFunction query type — exactly one sub-block should be set.
// HCL: query { metric_query { ... } } or query { event_query { ... } } etc.
var formulaAndFunctionQueryFields = []FieldSpec{
	{HCLKey: "metric_query", Type: TypeBlock, OmitEmpty: true,
		Description: "A timeseries formula and functions metrics query.",
		Children:    formulaAndFunctionMetricQueryFields},
	{HCLKey: "event_query", Type: TypeBlock, OmitEmpty: true,
		Description: "A timeseries formula and functions events query.",
		Children:    formulaAndFunctionEventQueryFields},
	{HCLKey: "process_query", Type: TypeBlock, OmitEmpty: true,
		Description: "The process query using formulas and functions.",
		Children:    formulaAndFunctionProcessQueryFields},
	{HCLKey: "apm_dependency_stats_query", Type: TypeBlock, OmitEmpty: true,
		Description: "The APM Dependency Stats query using formulas and functions.",
		Children:    formulaAndFunctionApmDependencyStatsQueryFields},
	{HCLKey: "apm_resource_stats_query", Type: TypeBlock, OmitEmpty: true,
		Description: "The APM Resource Stats query using formulas and functions.",
		Children:    formulaAndFunctionApmResourceStatsQueryFields},
	{HCLKey: "slo_query", Type: TypeBlock, OmitEmpty: true,
		Description: "The SLO query using formulas and functions.",
		Children:    formulaAndFunctionSLOQueryFields},
	{HCLKey: "cloud_cost_query", Type: TypeBlock, OmitEmpty: true,
		Description: "The Cloud Cost query using formulas and functions.",
		Children:    formulaAndFunctionCloudCostQueryFields},
}

// standardQueryFields are the legacy query-source fields present on most request types.
// Used by: change, distribution, heatmap, query_value, toplist, sunburst requests.
var standardQueryFields = []FieldSpec{
	{HCLKey: "log_query", Type: TypeBlock, OmitEmpty: true,
		Description: "The log query to use in the widget.",
		Children:    logQueryDefinitionFields},
	{HCLKey: "apm_query", Type: TypeBlock, OmitEmpty: true,
		Description: "The APM query to use in the widget.",
		Children:    logQueryDefinitionFields},
	{HCLKey: "rum_query", Type: TypeBlock, OmitEmpty: true,
		Description: "The RUM query to use in the widget.",
		Children:    logQueryDefinitionFields},
	{HCLKey: "security_query", Type: TypeBlock, OmitEmpty: true,
		Description: "The security query to use in the widget.",
		Children:    logQueryDefinitionFields},
	{HCLKey: "process_query", Type: TypeBlock, OmitEmpty: true,
		Description: "The process query to use in the widget. The structure of this block is described below.",
		Children:    processQueryDefinitionFields},
	// FormulaAndFunction query/formula fields
	{HCLKey: "query", Type: TypeBlockList, OmitEmpty: true,
		Description: "A list of queries to use in the widget.",
		Children:    formulaAndFunctionQueryFields},
	{HCLKey: "formula", Type: TypeBlockList, OmitEmpty: true,
		Description: "A list of formulas to use in the widget.",
		Children:    widgetFormulaFields},
}

// ============================================================
// WidgetFormula Field Groups
// ============================================================

// widgetFormulaLimitFields corresponds to OpenAPI components/schemas/WidgetFormulaLimit.
// Used inside widgetFormulaFields as the "limit" block.
var widgetFormulaLimitFields = []FieldSpec{
	{HCLKey: "count", Type: TypeInt, OmitEmpty: true,
		Description: "The number of results to return."},
	{HCLKey: "order", Type: TypeString, OmitEmpty: true,
		Description: "The direction of the sort.",
		ValidValues: []string{"asc", "desc"}},
}

// widgetFormulaStyleFields corresponds to OpenAPI components/schemas/WidgetFormulaStyle.
// Styling options for a single formula (per-formula palette and palette_index).
// Distinct from request-level style (which has palette, line_type, line_width for timeseries).
var widgetFormulaStyleFields = []FieldSpec{
	{HCLKey: "palette", Type: TypeString, OmitEmpty: true,
		Description: "The color palette used to display the formula. A guide to the available color palettes can be found at https://docs.datadoghq.com/dashboards/guide/widget_colors."},
	{HCLKey: "palette_index", Type: TypeInt, OmitEmpty: true,
		Description: "Index specifying which color to use within the palette."},
}

// widgetFormulaCellDisplayModeOptionsFields corresponds to OpenAPI
// components/schemas/WidgetFormulaCellDisplayModeOptions.
// Only meaningful when cell_display_mode == "trend".
var widgetFormulaCellDisplayModeOptionsFields = []FieldSpec{
	{HCLKey: "trend_type", Type: TypeString, OmitEmpty: true,
		Description: "The type of trend line to display. Valid values are `area`, `line`, and `bars`.",
		ValidValues: []string{"area", "line", "bars"}},
	{HCLKey: "y_scale", Type: TypeString, OmitEmpty: true,
		Description: "The scale of the y-axis. Valid values are `shared` and `independent`.",
		ValidValues: []string{"shared", "independent"}},
}

// numberFormatUnitCanonicalFields corresponds to the canonical sub-block of
// WidgetNumberFormatUnit.
var numberFormatUnitCanonicalFields = []FieldSpec{
	{HCLKey: "per_unit_name", Type: TypeString, OmitEmpty: true,
		Description: "per unit name. If you want to represent megabytes/s, you set 'unit_name' = 'megabyte' and 'per_unit_name = 'second'"},
	{HCLKey: "unit_name", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "Unit name. It should be in singular form ('megabyte' and not 'megabytes')"},
}

// numberFormatUnitCustomFields corresponds to the custom sub-block of
// WidgetNumberFormatUnit.
var numberFormatUnitCustomFields = []FieldSpec{
	{HCLKey: "label", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "Unit label"},
}

// numberFormatUnitFields corresponds to the unit block inside WidgetNumberFormat.
var numberFormatUnitFields = []FieldSpec{
	{HCLKey: "canonical", Type: TypeBlock, OmitEmpty: true,
		Description: "Canonical Units",
		Children:    numberFormatUnitCanonicalFields},
	{HCLKey: "custom", Type: TypeBlock, OmitEmpty: true,
		Description: "Use custom (non canonical metrics)",
		Children:    numberFormatUnitCustomFields},
}

// numberFormatUnitScaleFields corresponds to the unit_scale block inside WidgetNumberFormat.
var numberFormatUnitScaleFields = []FieldSpec{
	{HCLKey: "unit_name", Type: TypeString, OmitEmpty: false, Required: true,
		Description: ""},
}

// widgetNumberFormatFields corresponds to OpenAPI WidgetNumberFormat.
// Used inside widgetFormulaFields as the "number_format" block.
var widgetNumberFormatFields = []FieldSpec{
	{HCLKey: "unit", Type: TypeBlock, OmitEmpty: false, Required: true,
		Description: "Unit of the number format. ",
		Children:    numberFormatUnitFields},
	{HCLKey: "unit_scale", Type: TypeBlock, OmitEmpty: true,
		Description: "",
		Children:    numberFormatUnitScaleFields},
}

// widgetFormulaFields corresponds to OpenAPI components/schemas/WidgetFormula.
// Used by all formula-capable widgets for the per-formula FieldSpec mapping.
// Applied via BuildEngineJSON/FlattenEngineJSON in the per-formula loops of
// buildFormulaQueryRequestJSON and buildScalarFormulaQueryRequestJSON.
//
// Note: "style" here is the per-formula style (palette, palette_index), distinct from
// the request-level style block (palette, line_type, line_width on timeseries). The
// two styles live at different JSON levels and do not conflict.
var widgetFormulaFields = []FieldSpec{
	// formula_expression (HCL) → formula (JSON)
	{HCLKey: "formula_expression", JSONKey: "formula", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "A string expression built from queries, formulas, and functions."},
	{HCLKey: "alias", Type: TypeString, OmitEmpty: true,
		Description: "An expression alias."},
	{HCLKey: "limit", Type: TypeBlock, OmitEmpty: true,
		Description: "The options for limiting results returned.",
		Children:    widgetFormulaLimitFields},
	{HCLKey: "cell_display_mode", Type: TypeString, OmitEmpty: true,
		Description: "A list of display modes for each table cell. Valid values are `number`, `bar`, and `trend`.",
		ValidValues: []string{"number", "bar", "trend"}},
	{HCLKey: "cell_display_mode_options", Type: TypeBlock, OmitEmpty: true,
		Description: "Options for the cell display mode. Only used when `cell_display_mode` is set to `trend`.",
		Children:    widgetFormulaCellDisplayModeOptionsFields},
	{HCLKey: "conditional_formats", Type: TypeBlockList, OmitEmpty: true,
		Description: "Conditional formats allow you to set the color of your widget content or background depending on the rule applied to your data. Multiple `conditional_formats` blocks are allowed using the structure below.",
		Children:    widgetConditionalFormatFields},
	{HCLKey: "style", Type: TypeBlock, OmitEmpty: true,
		Description: "Styling options for widget formulas.",
		Children:    widgetFormulaStyleFields},
	{HCLKey: "number_format", Type: TypeBlock, OmitEmpty: true,
		Description: "Number formatting options for the formula.",
		Children:    widgetNumberFormatFields},
}

// ============================================================
// Common Widget Fields
// ============================================================

// CommonWidgetFields are the FieldSpecs shared by most widget definition types.
// They are merged automatically into every WidgetSpec by the engine.
var CommonWidgetFields = []FieldSpec{
	// Inline properties on widget definitions (no OpenAPI $ref, common by convention)
	{HCLKey: "title", Type: TypeString, OmitEmpty: true,
		Description: "The title of the widget."},
	{HCLKey: "title_size", Type: TypeString, OmitEmpty: true,
		Description: "The size of the widget's title (defaults to 16)."},
	{HCLKey: "title_align", Type: TypeString, OmitEmpty: true,
		Description: "The alignment of the widget's title.",
		ValidValues: []string{"center", "left", "right"}},
	// WidgetTime: live_span (HCL) → {"time": {"live_span": "..."}} (JSON)
	widgetTimeField,
	// WidgetCustomLink: HCL "custom_link" (singular) → JSON "custom_links" (plural)
	{HCLKey: "custom_link", JSONKey: "custom_links", Type: TypeBlockList, OmitEmpty: true,
		Description: "A nested block describing a custom link. Multiple `custom_link` blocks are allowed using the structure below.",
		Children:    widgetCustomLinkFields},
}

// ============================================================
// Dashboard Top-Level Fields
// ============================================================

// templateVariableFields corresponds to OpenAPI DashboardTemplateVariable.
// HCL key: "template_variable" (singular), JSON key: "template_variables" (plural).
var templateVariableFields = []FieldSpec{
	{HCLKey: "name", Type: TypeString, OmitEmpty: false},
	{HCLKey: "prefix", Type: TypeString, OmitEmpty: true},
	{HCLKey: "default", Type: TypeString, OmitEmpty: true},
	{HCLKey: "defaults", Type: TypeStringList, OmitEmpty: true},
	{HCLKey: "available_values", Type: TypeStringList, OmitEmpty: true},
}

// templateVariablePresetValueFields corresponds to OpenAPI DashboardTemplateVariablePresetValue.
// Used inside template_variable_preset blocks.
// HCL key: "template_variable" (singular), JSON key: "template_variables" (plural).
var templateVariablePresetValueFields = []FieldSpec{
	{HCLKey: "name", Type: TypeString, OmitEmpty: true},
	{HCLKey: "value", Type: TypeString, OmitEmpty: true},
	{HCLKey: "values", Type: TypeStringList, OmitEmpty: true},
}

// templateVariablePresetFields corresponds to OpenAPI DashboardTemplateVariablePreset.
// HCL key: "template_variable_preset" (singular), JSON key: "template_variable_presets" (plural).
var templateVariablePresetFields = []FieldSpec{
	{HCLKey: "name", Type: TypeString, OmitEmpty: true},
	// template_variable (singular HCL) → template_variables (plural JSON)
	// OmitEmpty: false — even empty presets get "template_variables": [] (cassette-verified)
	{HCLKey: "template_variable", JSONKey: "template_variables", Type: TypeBlockList, OmitEmpty: false,
		Children: templateVariablePresetValueFields},
}

// dashboardTopLevelFields are the top-level fields of the Dashboard object.
var dashboardTopLevelFields = []FieldSpec{
	{HCLKey: "title", Type: TypeString, OmitEmpty: false},
	{HCLKey: "description", Type: TypeString, OmitEmpty: false},
	{HCLKey: "layout_type", Type: TypeString, OmitEmpty: false},
	{HCLKey: "reflow_type", Type: TypeString, OmitEmpty: true},
	// notify_list: always send [], never omit (OmitEmpty: false)
	{HCLKey: "notify_list", Type: TypeStringList, OmitEmpty: false},
	// tags: always send [], never omit
	{HCLKey: "tags", Type: TypeStringList, OmitEmpty: false},
	// template_variable (HCL singular) → template_variables (JSON plural)
	{HCLKey: "template_variable", JSONKey: "template_variables", Type: TypeBlockList, OmitEmpty: false,
		Children: templateVariableFields},
	// template_variable_preset (HCL singular) → template_variable_presets (JSON plural)
	{HCLKey: "template_variable_preset", JSONKey: "template_variable_presets", Type: TypeBlockList, OmitEmpty: false,
		Children: templateVariablePresetFields},
	// restricted_roles: omit when empty
	{HCLKey: "restricted_roles", Type: TypeStringList, OmitEmpty: true},
	// is_read_only: kept in schema for backward compat; omit when false
	{HCLKey: "is_read_only", Type: TypeBool, OmitEmpty: true},
}
