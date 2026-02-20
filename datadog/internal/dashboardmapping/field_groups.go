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
	{HCLKey: "label", Type: TypeString, OmitEmpty: true},
	{HCLKey: "link", Type: TypeString, OmitEmpty: false},
	{HCLKey: "is_hidden", Type: TypeBool, OmitEmpty: true},
	{HCLKey: "override_label", Type: TypeString, OmitEmpty: true},
}

// widgetTimeField corresponds to OpenAPI components/schemas/WidgetLegacyLiveSpan
// (the live_span variant of WidgetTime, which is the form used by HCL).
// HCL flattens this to a single "live_span" string field on the widget definition,
// which maps to {"time": {"live_span": "..."}} in JSON via JSONPath.
// Used by: 21+ widget types.
var widgetTimeField = FieldSpec{
	HCLKey:    "live_span",
	JSONPath:  "time.live_span",
	Type:      TypeString,
	OmitEmpty: true,
}

// widgetAxisFields corresponds to OpenAPI components/schemas/WidgetAxis.
// Used by: timeseries (yaxis + right_yaxis), distribution, heatmap, scatterplot.
var widgetAxisFields = []FieldSpec{
	{HCLKey: "label", Type: TypeString, OmitEmpty: true},
	{HCLKey: "min", Type: TypeString, OmitEmpty: true},
	{HCLKey: "max", Type: TypeString, OmitEmpty: true},
	{HCLKey: "scale", Type: TypeString, OmitEmpty: true},
	// include_zero is always emitted even when false (OmitEmpty: false)
	// confirmed by cassette: "include_zero": false appears in right_yaxis
	{HCLKey: "include_zero", Type: TypeBool, OmitEmpty: false},
}

// widgetMarkerFields corresponds to OpenAPI components/schemas/WidgetMarker.
// Used by: timeseries, distribution, heatmap.
// HCL key: "marker" (singular), JSON key: "markers" (plural).
var widgetMarkerFields = []FieldSpec{
	{HCLKey: "value", Type: TypeString, OmitEmpty: false}, // required in OpenAPI
	{HCLKey: "display_type", Type: TypeString, OmitEmpty: true},
	{HCLKey: "label", Type: TypeString, OmitEmpty: true},
}

// widgetEventFields corresponds to OpenAPI components/schemas/WidgetEvent.
// Used by: timeseries, heatmap.
// HCL key: "event" (singular), JSON key: "events" (plural).
var widgetEventFields = []FieldSpec{
	{HCLKey: "q", Type: TypeString, OmitEmpty: false},             // required in OpenAPI
	{HCLKey: "tags_execution", Type: TypeString, OmitEmpty: true}, // omit when empty
}

// logQueryDefinitionGroupBySortFields corresponds to OpenAPI
// components/schemas/LogQueryDefinitionGroupBySort.
var logQueryDefinitionGroupBySortFields = []FieldSpec{
	{HCLKey: "aggregation", Type: TypeString, OmitEmpty: false},
	{HCLKey: "order", Type: TypeString, OmitEmpty: false},
	{HCLKey: "facet", Type: TypeString, OmitEmpty: true},
}

// logQueryDefinitionGroupByFields corresponds to OpenAPI
// components/schemas/LogQueryDefinitionGroupBy.
var logQueryDefinitionGroupByFields = []FieldSpec{
	{HCLKey: "facet", Type: TypeString, OmitEmpty: false}, // required in OpenAPI
	{HCLKey: "limit", Type: TypeInt, OmitEmpty: true},
	// HCL key: "sort_query" (disambiguates from other sort fields in HCL)
	// JSON key: "sort" (OpenAPI property name)
	{HCLKey: "sort_query", JSONKey: "sort", Type: TypeBlock, OmitEmpty: true,
		Children: logQueryDefinitionGroupBySortFields},
}

// logsQueryComputeFields corresponds to OpenAPI components/schemas/LogsQueryCompute.
var logsQueryComputeFields = []FieldSpec{
	{HCLKey: "aggregation", Type: TypeString, OmitEmpty: false}, // required in OpenAPI
	{HCLKey: "facet", Type: TypeString, OmitEmpty: true},
	{HCLKey: "interval", Type: TypeInt, OmitEmpty: true},
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
	{HCLKey: "index", Type: TypeString, OmitEmpty: false},
	// search_query (flat HCL) → {"search": {"query": "..."}} (nested JSON) via JSONPath
	{HCLKey: "search_query", JSONPath: "search.query", Type: TypeString, OmitEmpty: false},
	// HCL "compute_query" → JSON "compute" (renamed to avoid ambiguity in HCL)
	{HCLKey: "compute_query", JSONKey: "compute", Type: TypeBlock, OmitEmpty: true,
		Children: logsQueryComputeFields},
	// multi_compute → JSON "multi_compute" (same key, list of compute objects)
	{HCLKey: "multi_compute", Type: TypeBlockList, OmitEmpty: true,
		Children: logsQueryComputeFields},
	// HCL key: "group_by" (same in HCL and JSON — no pluralization applied to this field)
	{HCLKey: "group_by", Type: TypeBlockList, OmitEmpty: true,
		Children: logQueryDefinitionGroupByFields},
}

// processQueryDefinitionFields corresponds to OpenAPI
// components/schemas/ProcessQueryDefinition.
// Used by timeseries and other widgets that support process metrics.
var processQueryDefinitionFields = []FieldSpec{
	{HCLKey: "metric", Type: TypeString, OmitEmpty: false},   // required in OpenAPI
	{HCLKey: "search_by", Type: TypeString, OmitEmpty: true}, // omit when empty
	{HCLKey: "filter_by", Type: TypeStringList, OmitEmpty: true},
	{HCLKey: "limit", Type: TypeInt, OmitEmpty: true},
}

// ============================================================
// Common Widget Fields
// ============================================================

// commonWidgetFields are the FieldSpecs shared by most widget definition types.
// They are merged automatically into every WidgetSpec by the engine.
var commonWidgetFields = []FieldSpec{
	// Inline properties on widget definitions (no OpenAPI $ref, common by convention)
	{HCLKey: "title", Type: TypeString, OmitEmpty: true},
	{HCLKey: "title_size", Type: TypeString, OmitEmpty: true},
	{HCLKey: "title_align", Type: TypeString, OmitEmpty: true},
	// WidgetTime: live_span (HCL) → {"time": {"live_span": "..."}} (JSON)
	widgetTimeField,
	// WidgetCustomLink: HCL "custom_link" (singular) → JSON "custom_links" (plural)
	{HCLKey: "custom_link", JSONKey: "custom_links", Type: TypeBlockList, OmitEmpty: true,
		Children: widgetCustomLinkFields},
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
	{HCLKey: "template_variable", JSONKey: "template_variables", Type: TypeBlockList, OmitEmpty: true,
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
