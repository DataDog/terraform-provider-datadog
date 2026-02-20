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

// queryTableWidgetConditionalFormatFields is the same as widgetConditionalFormatFields
// but reused here for clarity in context.

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
	{HCLKey: "process_query", Type: TypeBlock, OmitEmpty: true, Description: "The process query to use in the widget.", Children: processQueryDefinitionFields},
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
		ValidValues: []string{"avg", "last", "max", "min", "sum", "percentile"},
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
		Description: "A list of display modes for each table cell.",
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

// ============================================================
// List Stream Widget Field Groups
// ============================================================

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

// ============================================================
// Split Graph Widget Field Groups
// ============================================================

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
