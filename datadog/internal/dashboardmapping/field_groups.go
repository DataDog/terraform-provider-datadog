package dashboardmapping

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

// field_groups.go
//
// All reusable []FieldSpec variables that mirror OpenAPI components/schemas/ entries.
// Each variable is named after the OpenAPI schema it corresponds to (camelCase).
// A comment on each variable identifies the OpenAPI schema and which widget types use it.
//
// Sections:
//   - Shared Widget Field Groups (OpenAPI: WidgetCustomLink, WidgetTime, etc.)
//   - Query Field Groups (OpenAPI: LogQueryDefinition, etc.)
//   - Formula Field Groups (OpenAPI: WidgetFormula, etc.)
//   - Widget-level Field Groups (per widget type: hostmap, geomap, scatterplot, etc.)
//   - Dashboard Top-Level Field Groups

// ============================================================
// Shared Widget Field Groups (OpenAPI: WidgetCustomLink, etc.)
// ============================================================
// (base groups used by commonWidgetFields)

// widgetCustomLinkFields corresponds to OpenAPI components/schemas/WidgetCustomLink.
// Used by: timeseries, toplist, query_value, change, heatmap, hostmap,
//
//	geomap, scatterplot, service_map, sunburst, table, topology_map,
//	run_workflow (13 widget types).
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

// widgetCustomLinkField is the standard custom_link FieldSpec entry used by
// widgets that support custom_links per the OpenAPI spec.
// HCL: "custom_link" (singular) → JSON: "custom_links" (plural)
var widgetCustomLinkField = FieldSpec{
	HCLKey:      "custom_link",
	JSONKey:     "custom_links",
	Type:        TypeBlockList,
	OmitEmpty:   true,
	Children:    widgetCustomLinkFields,
	Description: "A nested block describing a custom link. Multiple `custom_link` blocks are allowed using the structure below.",
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
	Description: "The timeframe to use when displaying the widget. Valid values are `1m`, `5m`, `10m`, `15m`, `30m`, `1h`, `4h`, `1d`, `2d`, `1w`, `1mo`, `3mo`, `6mo`, `week_to_date`, `month_to_date`, `1y`, `alert`.",
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

// ============================================================
// Query Field Groups (OpenAPI: LogQueryDefinition, etc.)
// ============================================================
// (log/apm/rum/process query groups + standardQueryFields)

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

// formulaAndFunctionMetricQueryFields corresponds to OpenAPI
// FormulaAndFunctionMetricQueryDefinition.
var formulaAndFunctionMetricQueryFields = []FieldSpec{
	{HCLKey: "data_source", Type: TypeString, OmitEmpty: false, Default: "metrics",
		Description: "The data source for metrics queries."},
	{HCLKey: "query", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The metrics query definition."},
	{HCLKey: "aggregator", Type: TypeString, OmitEmpty: true,
		Description: "The aggregation methods available for metrics queries.",
		ValidValues: []string{"avg", "min", "max", "sum", "last", "area", "l2norm", "percentile"}},
	{HCLKey: "name", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The name of the query for use in formulas."},
	{HCLKey: "cross_org_uuids", Type: TypeStringList, OmitEmpty: true, MaxItems: 1,
		Description: "The source organization UUID for cross organization queries. Feature in Private Beta."},
	{HCLKey: "semantic_mode", Type: TypeString, OmitEmpty: true,
		Description: "Semantic mode for metrics queries. This determines how metrics from different sources are combined or displayed. Valid values are `combined`, `native`."},
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
		ValidValues: []string{"logs", "spans", "network", "rum", "security_signals", "profiles", "audit", "events", "ci_tests", "ci_pipelines", "incident_analytics", "product_analytics", "on_call_events"}},
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
		ValidValues: []string{"avg", "min", "max", "sum", "last", "area", "l2norm", "percentile"}},
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
		Description: "The name of the second primary tag used within APM; required when `primary_tag_value` is specified. See https://docs.datadoghq.com/tracing/guide/setting_primary_tags_to_scope/#add-a-second-primary-tag-in-datadog."},
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
		Description: "The name of the second primary tag used within APM; required when `primary_tag_value` is specified. See https://docs.datadoghq.com/tracing/guide/setting_primary_tags_to_scope/#add-a-second-primary-tag-in-datadog."},
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
		ValidValues: []string{"good_events", "bad_events", "good_minutes", "bad_minutes", "slo_status", "error_budget_remaining", "burn_rate", "error_budget_burndown"}},
	{HCLKey: "name", Type: TypeString, OmitEmpty: true,
		Description: "The name of query for use in formulas."},
	{HCLKey: "group_mode", Type: TypeString, OmitEmpty: false, Default: "overall",
		Description: "Group mode to query measures.",
		ValidValues: []string{"overall", "components"}},
	{HCLKey: "slo_query_type", Type: TypeString, OmitEmpty: false, Default: "metric",
		Description: "type of the SLO to query.",
		ValidValues: []string{"metric", "monitor", "time_slice"}},
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
		ValidValues: []string{"avg", "min", "max", "sum", "last", "area", "l2norm", "percentile"}},
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
		Description: "The query to use for this widget.",
		Children:    logQueryDefinitionFields},
	{HCLKey: "apm_query", Type: TypeBlock, OmitEmpty: true,
		Description: "The query to use for this widget.",
		Children:    logQueryDefinitionFields},
	{HCLKey: "rum_query", Type: TypeBlock, OmitEmpty: true,
		Description: "The query to use for this widget.",
		Children:    logQueryDefinitionFields},
	{HCLKey: "security_query", Type: TypeBlock, OmitEmpty: true,
		Description: "The query to use for this widget.",
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
// Formula Field Groups (OpenAPI: WidgetFormula, etc.)
// ============================================================
// (widgetFormulaFields and its sub-groups, widgetConditionalFormatFields)

// widgetFormulaLimitFields corresponds to OpenAPI components/schemas/WidgetFormulaLimit.
// Used inside widgetFormulaFields as the "limit" block.
var widgetFormulaLimitFields = []FieldSpec{
	{HCLKey: "count", Type: TypeInt, OmitEmpty: true,
		Description: "The number of results to return."},
	{HCLKey: "order", Type: TypeString, OmitEmpty: true, Default: "desc",
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

// widgetConditionalFormatFields corresponds to OpenAPI WidgetConditionalFormat.
// Used by: query_value, toplist requests and formula fields.
var widgetConditionalFormatFields = []FieldSpec{
	{HCLKey: "comparator", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The comparator to use.",
		ValidValues: []string{"=", ">", ">=", "<", "<="}},
	{HCLKey: "value", Type: TypeFloat, OmitEmpty: false, Required: true,
		Description: "A value for the comparator."},
	{HCLKey: "palette", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The color palette to apply.",
		ValidValues: []string{"blue", "custom_bg", "custom_image", "custom_text", "gray_on_white", "grey", "green", "orange", "red", "red_on_white", "white_on_gray", "white_on_green", "green_on_white", "white_on_red", "white_on_yellow", "yellow_on_white", "black_on_light_yellow", "black_on_light_green", "black_on_light_red"}},
	{HCLKey: "custom_bg_color", Type: TypeString, OmitEmpty: true,
		Description: "The color palette to apply to the background, same values available as palette."},
	{HCLKey: "custom_fg_color", Type: TypeString, OmitEmpty: true,
		Description: "The color palette to apply to the foreground, same values available as palette."},
	{HCLKey: "image_url", Type: TypeString, OmitEmpty: true,
		Description: "Displays an image as the background."},
	// Emitted even when false (cassette-verified)
	{HCLKey: "hide_value", Type: TypeBool, OmitEmpty: false,
		Description: "Setting this to True hides values."},
	{HCLKey: "timeframe", Type: TypeString, OmitEmpty: true,
		Description: "Defines the displayed timeframe."},
	{HCLKey: "metric", Type: TypeString, OmitEmpty: true,
		Description: "The metric from the request to correlate with this conditional format."},
}

// ============================================================
// Widget-level Field Groups (per widget type)
// ============================================================
// (hostmap, geomap, scatterplot, sunburst, toplist, distribution axes,
//
//	timeseries background, query table, list stream, slo, split graph, powerpack groups)

// widgetRequestStyleFields corresponds to OpenAPI WidgetStyle.
// Used by: distribution, heatmap, sunburst, toplist (request-level style).
var widgetRequestStyleFields = []FieldSpec{
	{HCLKey: "palette", Type: TypeString, OmitEmpty: true,
		Description: "A color palette to apply to the widget. The available options are available at: https://docs.datadoghq.com/dashboards/widgets/timeseries/#appearance."},
}

// hostmapRequestFillSizeFields corresponds to OpenAPI HostMapRequest.
// Used by: hostmap fill and size sub-blocks.
var hostmapRequestFillSizeFields = append([]FieldSpec{
	{HCLKey: "q", Type: TypeString, OmitEmpty: true,
		Description: "The metric query to use for this widget."},
}, standardQueryFields...)

// hostmapStyleFields corresponds to the inline style block on HostMapWidgetDefinition.
var hostmapStyleFields = []FieldSpec{
	{HCLKey: "palette", Type: TypeString, OmitEmpty: true,
		Description: "A color palette to apply to the widget. The available options are available at: https://docs.datadoghq.com/dashboards/widgets/timeseries/#appearance."},
	{HCLKey: "palette_flip", Type: TypeBool, OmitEmpty: true,
		Description: "A Boolean indicating whether to flip the palette tones."},
	{HCLKey: "fill_min", Type: TypeString, OmitEmpty: true,
		Description: "The min value to use to color the map."},
	{HCLKey: "fill_max", Type: TypeString, OmitEmpty: true,
		Description: "The max value to use to color the map."},
}

// geomapStyleFields corresponds to the style block on GeomapWidgetDefinition.
var geomapStyleFields = []FieldSpec{
	// Both required — emitted even when false
	{HCLKey: "palette", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The color palette to apply to the widget."},
	{HCLKey: "palette_flip", Type: TypeBool, OmitEmpty: false, Required: true,
		Description: "A Boolean indicating whether to flip the palette tones."},
}

// sunburstLegendInlineFields corresponds to OpenAPI SunburstWidgetLegendInlineAutomatic.
// Kept named (3 fields); sunburstLegendTableFields (1 field) inlined into sunburstWidgetSpec.
// geomapViewFields (1 field) inlined into geomapWidgetSpec.
var sunburstLegendInlineFields = []FieldSpec{
	{HCLKey: "type", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The type of legend (inline or automatic)."},
	// Emitted even when false (cassette-verified)
	{HCLKey: "hide_value", Type: TypeBool, OmitEmpty: false,
		Description: "Whether to hide the values of the groups."},
	{HCLKey: "hide_percent", Type: TypeBool, OmitEmpty: false,
		Description: "Whether to hide the percentages of the groups."},
}

// timeseriesBackgroundFields corresponds to OpenAPI TimeseriesBackground.
// Used by: query_value timeseries_background block.
var timeseriesBackgroundFields = []FieldSpec{
	{HCLKey: "type", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "Whether the Timeseries is made using an area or bars.",
		ValidValues: []string{"area", "bars"}},
	{HCLKey: "yaxis", Type: TypeBlock, OmitEmpty: true,
		Description: "A nested block describing the Y-Axis Controls. Exactly one nested block is allowed using the structure below.",
		Children:    widgetAxisFields},
}

// scatterplotXYRequestFields corresponds to OpenAPI ScatterPlotRequest.
// Used by: scatterplot x and y sub-blocks.
var scatterplotXYRequestFields = append([]FieldSpec{
	{HCLKey: "q", Type: TypeString, OmitEmpty: true,
		Description: "The metric query to use for this widget."},
	{HCLKey: "aggregator", Type: TypeString, OmitEmpty: true,
		Description: "Aggregator used for the request.",
		ValidValues: []string{"avg", "min", "max", "sum", "last", "area", "l2norm", "percentile"}},
}, standardQueryFields...)

// scatterplotFormulaFields corresponds to OpenAPI ScatterplotWidgetFormula.
// Used in the scatterplot_table sub-block (different from widgetFormulaFields).
var scatterplotFormulaFields = []FieldSpec{
	{HCLKey: "formula_expression", JSONKey: "formula", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "A string expression built from queries, formulas, and functions."},
	{HCLKey: "dimension", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "Dimension of the Scatterplot.",
		ValidValues: []string{"x", "y", "radius", "color"}},
	{HCLKey: "alias", Type: TypeString, OmitEmpty: true,
		Description: "An expression alias."},
}

// scatterplotTableRequestFields corresponds to OpenAPI ScatterplotTableRequest.
// Used by: scatterplot scatterplot_table sub-block.
var scatterplotTableRequestFields = []FieldSpec{
	{HCLKey: "query", Type: TypeBlockList, OmitEmpty: true,
		Description: "A list of queries to use in the widget.",
		Children:    formulaAndFunctionQueryFields},
	{HCLKey: "formula", Type: TypeBlockList, OmitEmpty: true,
		Description: "A list of formulas to use in the widget.",
		Children:    scatterplotFormulaFields},
}

// toplistWidgetStyleDisplayFields corresponds to the display sub-block inside toplist style.
var toplistWidgetStyleDisplayFields = []FieldSpec{
	{HCLKey: "type", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The display type for the widget."},
}

// toplistWidgetStyleFields corresponds to OpenAPI ToplistWidgetStyle.
// Note: "display" is a single JSON object (MaxItems:1 in HCL); we use TypeBlock
// to emit a single object rather than an array.
var toplistWidgetStyleFields = []FieldSpec{
	{HCLKey: "display", Type: TypeBlock, OmitEmpty: true,
		Description: "The display mode for the widget.",
		Children:    toplistWidgetStyleDisplayFields},
	{HCLKey: "palette", Type: TypeString, OmitEmpty: true,
		Description: "The color palette for the widget."},
	{HCLKey: "scaling", Type: TypeString, OmitEmpty: true,
		Description: "The scaling mode for the widget."},
}

// topologyQueryFields corresponds to the inline query block on TopologyRequest.
var topologyQueryFields = []FieldSpec{
	{HCLKey: "data_source", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The data source for the Topology request ('service_map' or 'data_streams')."},
	{HCLKey: "service", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The ID of the service to map."},
	{HCLKey: "filters", Type: TypeStringList, OmitEmpty: false, Required: true,
		Description: "Your environment and primary tag (or `*` if enabled for your account)."},
}

// apmStatsQueryColumnFields corresponds to column entries inside
// ApmStatsQueryDefinition.columns.
var apmStatsQueryColumnFields = []FieldSpec{
	{
		HCLKey:      "name",
		Type:        TypeString,
		OmitEmpty:   false,
		Required:    true,
		Description: "The column name.",
	},
	{
		HCLKey:      "alias",
		Type:        TypeString,
		OmitEmpty:   true,
		Description: "A user-assigned alias for the column.",
	},
	{
		HCLKey:      "order",
		Type:        TypeString,
		OmitEmpty:   true,
		Description: "Widget sorting methods.",
		ValidValues: []string{"asc", "desc"},
	},
	{
		HCLKey:      "cell_display_mode",
		Type:        TypeString,
		OmitEmpty:   true,
		Description: "A list of display modes for each table cell.",
		ValidValues: []string{"number", "bar", "trend"},
	},
}

// apmStatsQueryFields corresponds to OpenAPI
// components/schemas/ApmStatsQueryDefinition.
// Used by query_table apm_stats_query requests.
var apmStatsQueryFields = []FieldSpec{
	{
		HCLKey:      "service",
		Type:        TypeString,
		OmitEmpty:   false,
		Required:    true,
		Description: "The service name.",
	},
	{
		HCLKey:      "name",
		Type:        TypeString,
		OmitEmpty:   false,
		Required:    true,
		Description: "The operation name associated with the service.",
	},
	{
		HCLKey:      "env",
		Type:        TypeString,
		OmitEmpty:   false,
		Required:    true,
		Description: "The environment name.",
	},
	{
		HCLKey:      "primary_tag",
		Type:        TypeString,
		OmitEmpty:   false,
		Required:    true,
		Description: "The organization's host group name and value.",
	},
	{
		HCLKey:      "row_type",
		Type:        TypeString,
		OmitEmpty:   false,
		Required:    true,
		Description: "The level of detail for the request.",
		ValidValues: []string{"service", "resource", "span"},
	},
	{
		HCLKey:      "resource",
		Type:        TypeString,
		OmitEmpty:   true,
		Description: "The resource name.",
	},
	{
		HCLKey:      "columns",
		Type:        TypeBlockList,
		OmitEmpty:   true,
		Description: "Column properties used by the front end for display.",
		Children:    apmStatsQueryColumnFields,
	},
}

// tableWidgetTextFormatMatchFields corresponds to TableWidgetTextFormatMatch.
var tableWidgetTextFormatMatchFields = []FieldSpec{
	{HCLKey: "type", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "Match or compare option.",
		ValidValues: []string{"is", "is_not", "contains", "does_not_contain", "starts_with", "ends_with"}},
	{HCLKey: "value", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "Table Widget Match String."},
}

// tableWidgetTextFormatReplaceFields corresponds to TableWidgetTextFormatReplace.
var tableWidgetTextFormatReplaceFields = []FieldSpec{
	{HCLKey: "type", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "Table widget text format replace all type.",
		ValidValues: []string{"all", "substring"}},
	{HCLKey: "with", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "Table Widget Match String."},
	{HCLKey: "substring", Type: TypeString, OmitEmpty: true,
		Description: "Text that will be replaced. Must be used with type `substring`."},
}

// tableWidgetTextFormatRuleFields corresponds to a single text_format rule block
// inside the text_formats list.
var tableWidgetTextFormatRuleFields = []FieldSpec{
	{HCLKey: "match", Type: TypeBlock, OmitEmpty: false, Required: true,
		Description: "Match rule for the table widget text format.",
		Children:    tableWidgetTextFormatMatchFields},
	{HCLKey: "palette", Type: TypeString, OmitEmpty: true,
		Description: "The color palette to apply.",
		ValidValues: []string{"white_on_red", "white_on_yellow", "white_on_green", "black_on_light_red", "black_on_light_yellow", "black_on_light_green", "red_on_white", "yellow_on_white", "green_on_white", "custom_bg", "custom_text"}},
	{HCLKey: "replace", Type: TypeBlock, OmitEmpty: true,
		Description: "Match rule for the table widget text format.",
		Children:    tableWidgetTextFormatReplaceFields},
	{HCLKey: "custom_bg_color", Type: TypeString, OmitEmpty: true,
		Description: "The custom color palette to apply to the background."},
	{HCLKey: "custom_fg_color", Type: TypeString, OmitEmpty: true,
		Description: "The custom color palette to apply to the foreground text."},
}

// tableWidgetTextFormatsFields is the outer text_formats block containing text_format rules.
// text_formats is a list, each element of which is a list of text_format rules.
var tableWidgetTextFormatsFields = []FieldSpec{
	{HCLKey: "text_format", Type: TypeBlockList, OmitEmpty: true,
		Description: "The text format to apply to the items in a table widget column.",
		Children:    tableWidgetTextFormatRuleFields},
}

// queryTableOldRequestFields corresponds to OpenAPI
// components/schemas/TableWidgetRequest for the old-style (non-formula) requests.
// Includes: q, apm_query, log_query, rum_query, security_query, apm_stats_query,
// process_query, conditional_formats, aggregator, alias, limit, order, cell_display_mode.
// Formula requests are handled via post-processing (buildQueryTableFormulaRequestJSON).
var queryTableOldRequestFields = []FieldSpec{
	{HCLKey: "q", Type: TypeString, OmitEmpty: true, Description: "The metric query to use for this widget."},
	{HCLKey: "apm_query", Type: TypeBlock, OmitEmpty: true, Description: "The query to use for this widget.", Children: logQueryDefinitionFields},
	{HCLKey: "log_query", Type: TypeBlock, OmitEmpty: true, Description: "The query to use for this widget.", Children: logQueryDefinitionFields},
	{HCLKey: "process_query", Type: TypeBlock, OmitEmpty: true, Description: "The process query to use in the widget. The structure of this block is described below.", Children: processQueryDefinitionFields},
	{HCLKey: "rum_query", Type: TypeBlock, OmitEmpty: true, Description: "The query to use for this widget.", Children: logQueryDefinitionFields},
	{HCLKey: "security_query", Type: TypeBlock, OmitEmpty: true, Description: "The query to use for this widget.", Children: logQueryDefinitionFields},
	{HCLKey: "apm_stats_query", Type: TypeBlock, OmitEmpty: true, Children: apmStatsQueryFields},
	// conditional_formats (old-style requests have these at the request level)
	{
		HCLKey:      "conditional_formats",
		Type:        TypeBlockList,
		OmitEmpty:   true,
		Description: "Conditional formats allow you to set the color of your widget content or background, depending on the rule applied to your data. Multiple `conditional_formats` blocks are allowed using the structure below.",
		Children:    widgetConditionalFormatFields,
	},
	{
		HCLKey:      "aggregator",
		Type:        TypeString,
		OmitEmpty:   true,
		Description: "The aggregator to use for time aggregation.",
		ValidValues: []string{"avg", "min", "max", "sum", "last", "area", "l2norm", "percentile"},
	},
	{HCLKey: "alias", Type: TypeString, OmitEmpty: true, Description: "The alias for the column name (defaults to metric name)."},
	{HCLKey: "limit", Type: TypeInt, OmitEmpty: true, Description: "The number of lines to show in the table."},
	{
		HCLKey:      "order",
		Type:        TypeString,
		OmitEmpty:   true,
		Description: "The sort order for the rows.",
		ValidValues: []string{"asc", "desc"},
	},
	// cell_display_mode is a []string in old-style requests
	{
		HCLKey:      "cell_display_mode",
		Type:        TypeStringList,
		OmitEmpty:   true,
		Description: "A list of display modes for each table cell. Valid values are `number`, `bar`.",
	},
	// text_formats: each element is a list of text_format blocks
	{HCLKey: "text_formats", Type: TypeBlockList, OmitEmpty: true,
		Description: "Text formats define how to format text in table widget content. Multiple `text_formats` blocks are allowed using the structure below. This resource is in beta and is subject to change.",
		Children:    tableWidgetTextFormatsFields},
	// FormulaAndFunction query/formula fields
	{HCLKey: "query", Type: TypeBlockList, OmitEmpty: true,
		Description: "A list of queries to use in the widget.",
		Children:    formulaAndFunctionQueryFields},
	{HCLKey: "formula", Type: TypeBlockList, OmitEmpty: true,
		Description: "A list of formulas to use in the widget.",
		Children:    widgetFormulaFields},
}

// listStreamColumnFields corresponds to OpenAPI
// components/schemas/ListStreamColumn.
var listStreamColumnFields = []FieldSpec{
	{HCLKey: "field", Type: TypeString, OmitEmpty: false, Description: "Widget column field."},
	{
		HCLKey:      "width",
		Type:        TypeString,
		OmitEmpty:   false,
		Description: "Widget column width.",
		ValidValues: []string{"auto", "compact", "full"},
	},
}

// listStreamGroupByFields corresponds to the group_by block inside
// ListStreamQuery.
var listStreamGroupByFields = []FieldSpec{
	{HCLKey: "facet", Type: TypeString, OmitEmpty: false, Required: true, Description: "Facet name"},
}

// listStreamSortFields corresponds to the sort block inside ListStreamQuery.
var listStreamSortFields = []FieldSpec{
	{HCLKey: "column", Type: TypeString, OmitEmpty: false, Required: true, Description: "The facet path for the column."},
	{
		HCLKey:      "order",
		Type:        TypeString,
		OmitEmpty:   false,
		Required:    true,
		Description: "Widget sorting methods.",
		ValidValues: []string{"asc", "desc"},
	},
}

// listStreamQueryFields corresponds to OpenAPI
// components/schemas/ListStreamQuery.
var listStreamQueryFields = []FieldSpec{
	{
		HCLKey:      "data_source",
		Type:        TypeString,
		OmitEmpty:   false,
		Required:    true,
		Description: "Source from which to query items to display in the stream.",
		ValidValues: []string{
			"logs_stream", "audit_stream", "ci_pipeline_stream", "ci_test_stream",
			"rum_issue_stream", "apm_issue_stream", "trace_stream", "logs_issue_stream",
			"logs_pattern_stream", "logs_transaction_stream", "event_stream", "rum_stream",
			"llm_observability_stream",
		},
	},
	{HCLKey: "query_string", Type: TypeString, OmitEmpty: false, Description: "Widget query."},
	{
		HCLKey:      "event_size",
		Type:        TypeString,
		OmitEmpty:   true,
		Description: "Size of events displayed in widget. Required if `data_source` is `event_stream`.",
		ValidValues: []string{"s", "l"},
	},
	{HCLKey: "clustering_pattern_field_path", Type: TypeString, OmitEmpty: true, Description: "Specifies the field for logs pattern clustering. Can only be used with `logs_pattern_stream`."},
	{HCLKey: "storage", Type: TypeString, OmitEmpty: true, Description: "Storage location (private beta)."},
	// indexes: OmitEmpty — only present when set in HCL
	{HCLKey: "indexes", Type: TypeStringList, OmitEmpty: true, Description: "List of indexes."},
	// group_by: TypeBlockList
	{
		HCLKey:      "group_by",
		Type:        TypeBlockList,
		OmitEmpty:   true,
		Description: "Group by configuration for the List Stream widget. Group by can only be used with `logs_pattern_stream` (up to 4 items) or `logs_transaction_stream` (one group by item is required) list stream source.",
		Children:    listStreamGroupByFields,
	},
	// sort: TypeBlock (MaxItems:1)
	{
		HCLKey:      "sort",
		Type:        TypeBlock,
		OmitEmpty:   true,
		Description: "The facet and order to sort the data, for example: `{\"column\": \"time\", \"order\": \"desc\"}`.",
		Children:    listStreamSortFields,
	},
}

// listStreamRequestFields corresponds to OpenAPI
// components/schemas/ListStreamWidgetRequest.
var listStreamRequestFields = []FieldSpec{
	// columns: HCL plural → JSON plural (same key)
	{
		HCLKey:      "columns",
		Type:        TypeBlockList,
		OmitEmpty:   false,
		Required:    true,
		Description: "Widget columns.",
		Children:    listStreamColumnFields,
	},
	// response_format is required
	{
		HCLKey:      "response_format",
		Type:        TypeString,
		OmitEmpty:   false,
		Required:    true,
		Description: "Widget response format.",
		ValidValues: []string{"event_list"},
	},
	// query: TypeBlock (MaxItems:1)
	{
		HCLKey:      "query",
		Type:        TypeBlock,
		OmitEmpty:   false,
		Required:    true,
		Description: "Updated list stream widget.",
		Children:    listStreamQueryFields,
	},
}

// funnelStepFields corresponds to OpenAPI components/schemas/FunnelStep.
// Used by funnelQueryFields.
var funnelStepFields = []FieldSpec{
	{HCLKey: "facet", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The facet of the step."},
	{HCLKey: "value", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The value of the step."},
}

// funnelQueryFields corresponds to OpenAPI components/schemas/FunnelQuery.
// Used by funnelRequestFields.
var funnelQueryFields = []FieldSpec{
	{HCLKey: "data_source", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The source from which to query items to display in the funnel.",
		ValidValues: []string{"rum"}},
	{HCLKey: "query_string", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The widget query."},
	// HCL: "step" (singular) → JSON: "steps" (plural)
	{HCLKey: "step", JSONKey: "steps", Type: TypeBlockList, OmitEmpty: false, Required: true,
		Description: "List of funnel steps.",
		Children:    funnelStepFields},
}

// funnelRequestFields corresponds to OpenAPI components/schemas/FunnelWidgetRequest.
// Used by FunnelWidgetSpec.
var funnelRequestFields = []FieldSpec{
	{HCLKey: "request_type", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The request type for the Funnel widget request.",
		ValidValues: []string{"funnel"}},
	{HCLKey: "query", Type: TypeBlock, OmitEmpty: false, Required: true,
		Description: "Updated funnel widget.",
		Children:    funnelQueryFields},
}

// sloListSortFields corresponds to the sort block inside SLOListWidgetQuery.
var sloListSortFields = []FieldSpec{
	{HCLKey: "column", Type: TypeString, OmitEmpty: false, Required: true, Description: "The facet path for the column."},
	{
		HCLKey:      "order",
		Type:        TypeString,
		OmitEmpty:   false,
		Required:    true,
		Description: "Widget sorting methods.",
		ValidValues: []string{"asc", "desc"},
	},
}

// sloListQueryFields corresponds to OpenAPI
// components/schemas/SLOListWidgetQuery.
var sloListQueryFields = []FieldSpec{
	{HCLKey: "query_string", Type: TypeString, OmitEmpty: false, Required: true, Description: "Widget query."},
	{HCLKey: "limit", Type: TypeInt, OmitEmpty: true, Description: "Maximum number of results to display in the table."},
	// sort: TypeBlockList (can be multiple)
	{
		HCLKey:      "sort",
		Type:        TypeBlockList,
		OmitEmpty:   true,
		Description: "The facet and order to sort the data, for example: `{\"column\": \"status.sli\", \"order\": \"desc\"}`.",
		Children:    sloListSortFields,
	},
}

// sloListRequestFields corresponds to OpenAPI
// components/schemas/SLOListWidgetRequest.
var sloListRequestFields = []FieldSpec{
	{
		HCLKey:      "request_type",
		Type:        TypeString,
		OmitEmpty:   false,
		Required:    true,
		Description: "The request type for the SLO List request.",
		ValidValues: []string{"slo_list"},
	},
	// query: TypeBlock (MaxItems:1)
	{
		HCLKey:      "query",
		Type:        TypeBlock,
		OmitEmpty:   false,
		Required:    true,
		Description: "Updated SLO List widget.",
		Children:    sloListQueryFields,
	},
}

// splitDimensionFields corresponds to OpenAPI
// components/schemas/SplitDimension.
var splitDimensionFields = []FieldSpec{
	{
		HCLKey:      "one_graph_per",
		Type:        TypeString,
		OmitEmpty:   false,
		Required:    true,
		Description: "The system interprets this attribute differently depending on the data source of the query being split. For metrics, it's a tag. For the events platform, it's an attribute or tag.",
	},
}

// splitSortComputeFields corresponds to OpenAPI
// components/schemas/SplitConfigSortCompute.
var splitSortComputeFields = []FieldSpec{
	{HCLKey: "aggregation", Type: TypeString, OmitEmpty: true, Description: "How to aggregate the sort metric for the purposes of ordering."},
	{HCLKey: "metric", Type: TypeString, OmitEmpty: false, Required: true, Description: "The metric to use for sorting graphs."},
}

// splitSortFields corresponds to OpenAPI
// components/schemas/SplitSort.
var splitSortFields = []FieldSpec{
	{
		HCLKey:      "order",
		Type:        TypeString,
		OmitEmpty:   false,
		Required:    true,
		Description: "Widget sorting methods.",
		ValidValues: []string{"asc", "desc"},
	},
	// compute: optional single-element block
	{
		HCLKey:      "compute",
		Type:        TypeBlock,
		OmitEmpty:   true,
		Description: "Defines the metric and aggregation used as the sort value",
		Children:    splitSortComputeFields,
	},
}

// splitVectorEntryFields corresponds to the split_vector entry items.
var splitVectorEntryFields = []FieldSpec{
	{HCLKey: "tag_key", Type: TypeString, OmitEmpty: false, Required: true},
	{HCLKey: "tag_values", Type: TypeStringList, OmitEmpty: false, Required: true},
}

// staticSplitsEntryFields corresponds to the static_splits item
// (a single block with split_vector list).
var staticSplitsEntryFields = []FieldSpec{
	{
		HCLKey:      "split_vector",
		Type:        TypeBlockList,
		OmitEmpty:   false,
		Required:    true,
		Description: "The split graph list contains a graph for each value of the split dimension.",
		Children:    splitVectorEntryFields,
	},
}

// splitConfigFields corresponds to OpenAPI
// components/schemas/SplitConfig.
// Note: static_splits is NOT included here because it maps to a 2D JSON array
// that requires custom handling. See buildSplitConfigStaticSplitsJSON.
var splitConfigFields = []FieldSpec{
	// split_dimensions: HCL plural → JSON plural
	{
		HCLKey:      "split_dimensions",
		Type:        TypeBlockList,
		OmitEmpty:   false,
		Required:    true,
		Description: "The property by which the graph splits",
		Children:    splitDimensionFields,
	},
	{HCLKey: "limit", Type: TypeInt, OmitEmpty: true, Description: "Maximum number of graphs to display in the widget."},
	// sort: TypeBlock (MaxItems:1, Required in HCL)
	{
		HCLKey:      "sort",
		Type:        TypeBlock,
		OmitEmpty:   false,
		Required:    true,
		Description: "Controls the order in which graphs appear in the split.",
		Children:    splitSortFields,
	},
	// static_splits: JSON serialization handled by custom code (buildSplitConfigStaticSplitsJSON)
	// but the FieldSpec entry is needed for schema generation.
	{
		HCLKey:      "static_splits",
		Type:        TypeBlockList,
		OmitEmpty:   true,
		SchemaOnly:  true,
		Description: "The property by which the graph splits",
		Children:    staticSplitsEntryFields,
	},
}

// powerpackTVarContentFields corresponds to OpenAPI
// components/schemas/PowerpackTemplateVariableContents.
var powerpackTVarContentFields = []FieldSpec{
	{HCLKey: "name", Type: TypeString, OmitEmpty: false, Required: true, Description: "The name of the variable."},
	{HCLKey: "prefix", Type: TypeString, OmitEmpty: true, Description: "The tag prefix associated with the variable. Only tags with this prefix appear in the variable dropdown."},
	{HCLKey: "values", Type: TypeStringList, OmitEmpty: false, Required: true, Description: "One or many template variable values within the saved view, which will be unioned together using `OR` if more than one is specified."},
}

// powerpackTemplateVariableFields corresponds to the template_variables block
// inside PowerpackWidgetDefinition. Contains controlled_externally and
// controlled_by_powerpack sub-blocks, each a list of tvar content objects.
var powerpackTemplateVariableFields = []FieldSpec{
	{
		HCLKey:      "controlled_externally",
		Type:        TypeBlockList,
		OmitEmpty:   true,
		Description: "Template variables controlled by the external resource, such as the dashboard this powerpack is on.",
		Children:    powerpackTVarContentFields,
	},
	{
		HCLKey:      "controlled_by_powerpack",
		Type:        TypeBlockList,
		OmitEmpty:   true,
		Description: "Template variables controlled at the powerpack level.",
		Children:    powerpackTVarContentFields,
	},
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
	// WidgetTime: hide_incomplete_cost_data (HCL) → {"time": {"hide_incomplete_cost_data": true}} (JSON)
	{HCLKey: "hide_incomplete_cost_data", JSONPath: "time.hide_incomplete_cost_data", Type: TypeBool, OmitEmpty: true,
		Computed:    true,
		Description: "Hide any portion of the widget's timeframe that is incomplete due to cost data not being available."},
}

// ============================================================
// Dashboard Top-Level Field Groups
// ============================================================
// (moved to field_groups_dashboard.go: dashboardTemplateVariableFields,
//  dashboardTemplateVariablePresetValueFields, dashboardTemplateVariablePresetFields)

// DashboardTopLevelFields are the top-level fields of the Dashboard object.
// Exported for use in resource_datadog_dashboard.go SchemaFunc.
// Descriptions and valid values are sourced from OpenAPI components/schemas/Dashboard.
var DashboardTopLevelFields = []FieldSpec{
	{HCLKey: "title", Type: TypeString, Required: true, OmitEmpty: false,
		Description: "The title of the dashboard."},

	// DashboardLayoutType enum: ordered | free (OpenAPI DashboardLayoutType)
	{HCLKey: "layout_type", Type: TypeString, Required: true, OmitEmpty: false,
		ForceNew:    true,
		ValidValues: []string{"ordered", "free"},
		Description: "The layout type of the dashboard."},

	// DashboardReflowType enum: auto | fixed (OpenAPI DashboardReflowType)
	{HCLKey: "reflow_type", Type: TypeString, OmitEmpty: true,
		ValidValues: []string{"auto", "fixed"},
		Description: "The reflow type of a new dashboard layout. Set this only when layout type is `ordered`. If set to `fixed`, the dashboard expects all widgets to have a layout, and if it's set to `auto`, widgets should not have layouts."},

	// OmitEmpty: true — description is only emitted when explicitly set.
	// Cassettes recorded without description have no "description" key.
	{HCLKey: "description", Type: TypeString, OmitEmpty: true,
		Description: "The description of the dashboard."},

	// url: Computed+Optional. Always suppress diff — value is assigned by API and cannot be updated.
	// SchemaOnly: managed by UpdateDashboardEngineState, not serialized to JSON.
	{HCLKey: "url", Type: TypeString, Computed: true, OmitEmpty: true, SchemaOnly: true,
		DiffSuppress: func(_, _, _ string, _ *schema.ResourceData) bool { return true },
		Description:  "The URL of the dashboard."},

	{HCLKey: "restricted_roles", Type: TypeStringList, UseSet: true, OmitEmpty: true,
		ConflictsWith: []string{"is_read_only"},
		Description:   "UUIDs of roles whose associated users are authorized to edit the dashboard."},

	// template_variable (HCL singular) → template_variables (JSON plural)
	{HCLKey: "template_variable", JSONKey: "template_variables",
		Type: TypeBlockList, OmitEmpty: false,
		Description: "The list of template variables for this dashboard.",
		Children:    dashboardTemplateVariableFields},

	// template_variable_preset (HCL singular) → template_variable_presets (JSON plural)
	{HCLKey: "template_variable_preset", JSONKey: "template_variable_presets",
		Type: TypeBlockList, OmitEmpty: false,
		Description: "The list of selectable template variable presets for this dashboard.",
		Children:    dashboardTemplateVariablePresetFields},

	// notify_list: always send [], never omit (OmitEmpty: false)
	{HCLKey: "notify_list", Type: TypeStringList, UseSet: true, OmitEmpty: false,
		Description: "The list of handles for the users to notify when changes are made to this dashboard."},

	// SchemaOnly: managed as side effects via updateDashboardLists, not serialized to JSON.
	{HCLKey: "dashboard_lists", Type: TypeIntList, UseSet: true, OmitEmpty: false, SchemaOnly: true,
		Description: "A list of dashboard lists this dashboard belongs to. This attribute should not be set if managing the corresponding dashboard lists using Terraform as it causes inconsistent behavior."},

	// dashboard_lists_removed: computed only via CustomizeDiff, never sent in JSON
	// SchemaOnly: managed as side effects via updateDashboardLists, not serialized to JSON.
	{HCLKey: "dashboard_lists_removed", Type: TypeIntList, UseSet: true, Computed: true, SchemaOnly: true,
		Description: "A list of dashboard lists this dashboard should be removed from. Internal only."},

	{HCLKey: "is_read_only", Type: TypeBool, Default: false, OmitEmpty: true,
		ConflictsWith: []string{"restricted_roles"},
		Deprecated:    "This field is deprecated and non-functional. Use `restricted_roles` instead to define which roles are required to edit the dashboard.",
		Description:   "Whether this dashboard is read-only."},

	// tags: always send [], never omit
	{HCLKey: "tags", Type: TypeStringList, MaxItems: 5, OmitEmpty: false,
		Description: "A list of tags assigned to the Dashboard. Only team names of the form `team:<name>` are supported."},
}
