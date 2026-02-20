package dashboardmapping

// field_groups_complex.go — Batch C field group additions
//
// New reusable FieldSpec groups needed by Batch C (complex/structural widgets).
// All groups are named after their OpenAPI components/schemas/ counterparts.

// ============================================================
// Query Table Widget Field Groups
// ============================================================

// apmStatsQueryColumnFields corresponds to column entries inside
// ApmStatsQueryDefinition.columns.
var apmStatsQueryColumnFields = []FieldSpec{
	{HCLKey: "name", Type: TypeString, OmitEmpty: false},
	{HCLKey: "alias", Type: TypeString, OmitEmpty: true},
	{HCLKey: "order", Type: TypeString, OmitEmpty: true},
	{HCLKey: "cell_display_mode", Type: TypeString, OmitEmpty: true},
}

// apmStatsQueryFields corresponds to OpenAPI
// components/schemas/ApmStatsQueryDefinition.
// Used by query_table apm_stats_query requests.
var apmStatsQueryFields = []FieldSpec{
	{HCLKey: "service", Type: TypeString, OmitEmpty: false},
	{HCLKey: "name", Type: TypeString, OmitEmpty: false},
	{HCLKey: "env", Type: TypeString, OmitEmpty: false},
	{HCLKey: "primary_tag", Type: TypeString, OmitEmpty: false},
	{HCLKey: "row_type", Type: TypeString, OmitEmpty: false},
	{HCLKey: "resource", Type: TypeString, OmitEmpty: true},
	{HCLKey: "columns", Type: TypeBlockList, OmitEmpty: true, Children: apmStatsQueryColumnFields},
}

// queryTableWidgetConditionalFormatFields is the same as widgetConditionalFormatFields
// but reused here for clarity in context.

// queryTableOldRequestFields corresponds to OpenAPI
// components/schemas/TableWidgetRequest for the old-style (non-formula) requests.
// Includes: q, apm_query, log_query, rum_query, security_query, apm_stats_query,
// process_query, conditional_formats, aggregator, alias, limit, order, cell_display_mode.
// Formula requests are handled via post-processing (buildQueryTableFormulaRequestJSON).
var queryTableOldRequestFields = []FieldSpec{
	{HCLKey: "q", Type: TypeString, OmitEmpty: true},
	{HCLKey: "apm_query", Type: TypeBlock, OmitEmpty: true, Children: logQueryDefinitionFields},
	{HCLKey: "log_query", Type: TypeBlock, OmitEmpty: true, Children: logQueryDefinitionFields},
	{HCLKey: "process_query", Type: TypeBlock, OmitEmpty: true, Children: processQueryDefinitionFields},
	{HCLKey: "rum_query", Type: TypeBlock, OmitEmpty: true, Children: logQueryDefinitionFields},
	{HCLKey: "security_query", Type: TypeBlock, OmitEmpty: true, Children: logQueryDefinitionFields},
	{HCLKey: "apm_stats_query", Type: TypeBlock, OmitEmpty: true, Children: apmStatsQueryFields},
	// conditional_formats (old-style requests have these at the request level)
	{HCLKey: "conditional_formats", Type: TypeBlockList, OmitEmpty: true,
		Children: widgetConditionalFormatFields},
	{HCLKey: "aggregator", Type: TypeString, OmitEmpty: true},
	{HCLKey: "alias", Type: TypeString, OmitEmpty: true},
	{HCLKey: "limit", Type: TypeInt, OmitEmpty: true},
	{HCLKey: "order", Type: TypeString, OmitEmpty: true},
	// cell_display_mode is a []string in old-style requests
	{HCLKey: "cell_display_mode", Type: TypeStringList, OmitEmpty: true},
	// text_formats is a 2D array - needs custom handling, not FieldSpec
	// text_formats is handled in post-processing
}

// ============================================================
// List Stream Widget Field Groups
// ============================================================

// listStreamColumnFields corresponds to OpenAPI
// components/schemas/ListStreamColumn.
var listStreamColumnFields = []FieldSpec{
	{HCLKey: "field", Type: TypeString, OmitEmpty: false},
	{HCLKey: "width", Type: TypeString, OmitEmpty: false},
}

// listStreamGroupByFields corresponds to the group_by block inside
// ListStreamQuery.
var listStreamGroupByFields = []FieldSpec{
	{HCLKey: "facet", Type: TypeString, OmitEmpty: false},
}

// listStreamSortFields corresponds to the sort block inside ListStreamQuery.
var listStreamSortFields = []FieldSpec{
	{HCLKey: "column", Type: TypeString, OmitEmpty: false},
	{HCLKey: "order", Type: TypeString, OmitEmpty: false},
}

// listStreamQueryFields corresponds to OpenAPI
// components/schemas/ListStreamQuery.
var listStreamQueryFields = []FieldSpec{
	{HCLKey: "data_source", Type: TypeString, OmitEmpty: false},
	{HCLKey: "query_string", Type: TypeString, OmitEmpty: false},
	{HCLKey: "event_size", Type: TypeString, OmitEmpty: true},
	{HCLKey: "clustering_pattern_field_path", Type: TypeString, OmitEmpty: true},
	{HCLKey: "storage", Type: TypeString, OmitEmpty: true},
	// indexes: OmitEmpty — only present when set in HCL
	{HCLKey: "indexes", Type: TypeStringList, OmitEmpty: true},
	// group_by: TypeBlockList
	{HCLKey: "group_by", Type: TypeBlockList, OmitEmpty: true, Children: listStreamGroupByFields},
	// sort: TypeBlock (MaxItems:1)
	{HCLKey: "sort", Type: TypeBlock, OmitEmpty: true, Children: listStreamSortFields},
}

// listStreamRequestFields corresponds to OpenAPI
// components/schemas/ListStreamWidgetRequest.
var listStreamRequestFields = []FieldSpec{
	// columns: HCL plural → JSON plural (same key)
	{HCLKey: "columns", Type: TypeBlockList, OmitEmpty: false, Children: listStreamColumnFields},
	// response_format is required
	{HCLKey: "response_format", Type: TypeString, OmitEmpty: false},
	// query: TypeBlock (MaxItems:1)
	{HCLKey: "query", Type: TypeBlock, OmitEmpty: false, Children: listStreamQueryFields},
}

// ============================================================
// SLO Widget Field Groups
// ============================================================
// slo widget has no request blocks, just simple fields.
// All fields are at the definition level.

// ============================================================
// SLO List Widget Field Groups
// ============================================================

// sloListSortFields corresponds to the sort block inside SLOListWidgetQuery.
var sloListSortFields = []FieldSpec{
	{HCLKey: "column", Type: TypeString, OmitEmpty: false},
	{HCLKey: "order", Type: TypeString, OmitEmpty: false},
}

// sloListQueryFields corresponds to OpenAPI
// components/schemas/SLOListWidgetQuery.
var sloListQueryFields = []FieldSpec{
	{HCLKey: "query_string", Type: TypeString, OmitEmpty: false},
	{HCLKey: "limit", Type: TypeInt, OmitEmpty: true},
	// sort: TypeBlockList (can be multiple)
	{HCLKey: "sort", Type: TypeBlockList, OmitEmpty: true, Children: sloListSortFields},
}

// sloListRequestFields corresponds to OpenAPI
// components/schemas/SLOListWidgetRequest.
var sloListRequestFields = []FieldSpec{
	{HCLKey: "request_type", Type: TypeString, OmitEmpty: false},
	// query: TypeBlock (MaxItems:1)
	{HCLKey: "query", Type: TypeBlock, OmitEmpty: false, Children: sloListQueryFields},
}

// ============================================================
// Split Graph Widget Field Groups
// ============================================================

// splitDimensionFields corresponds to OpenAPI
// components/schemas/SplitDimension.
var splitDimensionFields = []FieldSpec{
	{HCLKey: "one_graph_per", Type: TypeString, OmitEmpty: false},
}

// splitSortComputeFields corresponds to OpenAPI
// components/schemas/SplitConfigSortCompute.
var splitSortComputeFields = []FieldSpec{
	{HCLKey: "aggregation", Type: TypeString, OmitEmpty: true},
	{HCLKey: "metric", Type: TypeString, OmitEmpty: false},
}

// splitSortFields corresponds to OpenAPI
// components/schemas/SplitSort.
var splitSortFields = []FieldSpec{
	{HCLKey: "order", Type: TypeString, OmitEmpty: false},
	// compute: optional single-element block
	{HCLKey: "compute", Type: TypeBlock, OmitEmpty: true, Children: splitSortComputeFields},
}

// splitVectorEntryFields corresponds to the split_vector entry items.
var splitVectorEntryFields = []FieldSpec{
	{HCLKey: "tag_key", Type: TypeString, OmitEmpty: false},
	{HCLKey: "tag_values", Type: TypeStringList, OmitEmpty: false},
}

// staticSplitsEntryFields corresponds to the static_splits item
// (a single block with split_vector list).
var staticSplitsEntryFields = []FieldSpec{
	{HCLKey: "split_vector", Type: TypeBlockList, OmitEmpty: false, Children: splitVectorEntryFields},
}

// splitConfigFields corresponds to OpenAPI
// components/schemas/SplitConfig.
// Note: static_splits is NOT included here because it maps to a 2D JSON array
// that requires custom handling. See buildSplitConfigStaticSplitsJSON.
var splitConfigFields = []FieldSpec{
	// split_dimensions: HCL plural → JSON plural
	{HCLKey: "split_dimensions", Type: TypeBlockList, OmitEmpty: false,
		Children: splitDimensionFields},
	{HCLKey: "limit", Type: TypeInt, OmitEmpty: true},
	// sort: TypeBlock (MaxItems:1, Required in HCL)
	{HCLKey: "sort", Type: TypeBlock, OmitEmpty: false, Children: splitSortFields},
	// static_splits handled by custom code (buildSplitConfigStaticSplitsJSON)
}

// ============================================================
// Group Widget Field Groups
// ============================================================
// Group widget's "widget" list is handled recursively via custom code,
// not via FieldSpec, because each child widget needs type dispatch.

// ============================================================
// Powerpack Widget Field Groups
// ============================================================

// powerpackTVarContentFields corresponds to OpenAPI
// components/schemas/PowerpackTemplateVariableContents.
var powerpackTVarContentFields = []FieldSpec{
	{HCLKey: "name", Type: TypeString, OmitEmpty: false},
	{HCLKey: "prefix", Type: TypeString, OmitEmpty: true},
	{HCLKey: "values", Type: TypeStringList, OmitEmpty: false},
}

// powerpackTemplateVariableFields corresponds to the template_variables block
// inside PowerpackWidgetDefinition. Contains controlled_externally and
// controlled_by_powerpack sub-blocks, each a list of tvar content objects.
var powerpackTemplateVariableFields = []FieldSpec{
	{HCLKey: "controlled_externally", Type: TypeBlockList, OmitEmpty: true,
		Children: powerpackTVarContentFields},
	{HCLKey: "controlled_by_powerpack", Type: TypeBlockList, OmitEmpty: true,
		Children: powerpackTVarContentFields},
}
