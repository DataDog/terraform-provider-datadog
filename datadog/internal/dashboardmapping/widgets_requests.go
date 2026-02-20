package dashboardmapping

// widgets_requests.go — Batch B
//
// Widget types with standard query-based requests: change, distribution, heatmap,
// hostmap, query_value, toplist, scatterplot, sunburst, geomap, treemap, topology_map.

// ChangeWidgetSpec corresponds to OpenAPI ChangeWidgetDefinition.
var changeWidgetRequestFields = append([]FieldSpec{
	{HCLKey: "q", Type: TypeString, OmitEmpty: true,
		Description: "The metric query to use for this widget."},
	{HCLKey: "change_type", Type: TypeString, OmitEmpty: true,
		Description: "Whether to show absolute or relative change.",
		ValidValues: []string{"absolute", "relative"}},
	{HCLKey: "compare_to", Type: TypeString, OmitEmpty: true,
		Description: "Choose from when to compare current data to.",
		ValidValues: []string{"hour_before", "day_before", "week_before", "month_before"}},
	// Emitted even when false (cassette-verified)
	{HCLKey: "increase_good", Type: TypeBool, OmitEmpty: false,
		Description: "A Boolean indicating whether an increase in the value is good (displayed in green) or not (displayed in red)."},
	{HCLKey: "order_by", Type: TypeString, OmitEmpty: true,
		Description: "What to order by.",
		ValidValues: []string{"change", "name", "present", "past"}},
	{HCLKey: "order_dir", Type: TypeString, OmitEmpty: true,
		Description: "Widget sorting method.",
		ValidValues: []string{"asc", "desc"}},
	{HCLKey: "show_present", Type: TypeBool, OmitEmpty: false,
		Description: "If set to `true`, displays the current value."},
}, standardQueryFields...)

var ChangeWidgetSpec = WidgetSpec{
	HCLKey:   "change_definition",
	JSONType: "change",
	Fields: []FieldSpec{
		{HCLKey: "request", JSONKey: "requests", Type: TypeBlockList, OmitEmpty: false,
			Description: "A nested block describing the request to use when displaying the widget. Multiple request blocks are allowed using the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query` or `process_query` is required within the request block).",
			Children:    changeWidgetRequestFields},
	},
}

// DistributionWidgetSpec corresponds to OpenAPI DistributionWidgetDefinition.
var distributionWidgetRequestFields = append([]FieldSpec{
	{HCLKey: "q", Type: TypeString, OmitEmpty: true,
		Description: "The metric query to use for this widget."},
	{HCLKey: "style", Type: TypeBlock, OmitEmpty: true,
		Description: "The style of the widget graph. One nested block is allowed using the structure below.",
		Children:    widgetRequestStyleFields},
	// apm_stats_query is supported by distribution requests (OpenAPI DistributionWidgetRequest)
	{HCLKey: "apm_stats_query", Type: TypeBlock, OmitEmpty: true,
		Description: "The APM stats query to use in the widget.",
		Children:    apmStatsQueryFields},
}, standardQueryFields...)

// distributionWidgetXAxisFields corresponds to OpenAPI DistributionWidgetXAxis.
// Differs from WidgetAxis: include_zero uses OmitEmpty: true (not false).
var distributionWidgetXAxisFields = []FieldSpec{
	{HCLKey: "scale", Type: TypeString, OmitEmpty: true,
		Description: "Specify the scale type, options: `linear`, `log`, `pow`, `sqrt`."},
	{HCLKey: "min", Type: TypeString, OmitEmpty: true,
		Description: "Specify the minimum value to show on the Y-axis."},
	{HCLKey: "max", Type: TypeString, OmitEmpty: true,
		Description: "Specify the maximum value to show on the Y-axis."},
	{HCLKey: "include_zero", Type: TypeBool, OmitEmpty: true,
		Description: "Always include zero or fit the axis to the data range."},
}

// distributionWidgetYAxisFields extends distributionWidgetXAxisFields with a label field.
var distributionWidgetYAxisFields = append(
	append([]FieldSpec{}, distributionWidgetXAxisFields...),
	FieldSpec{HCLKey: "label", Type: TypeString, OmitEmpty: true,
		Description: "The label of the axis to display on the graph."},
)

var DistributionWidgetSpec = WidgetSpec{
	HCLKey:   "distribution_definition",
	JSONType: "distribution",
	Fields: []FieldSpec{
		// show_legend: OmitEmpty: false — cassette confirms false is emitted when not set
		{HCLKey: "show_legend", Type: TypeBool, OmitEmpty: false,
			Description: "Whether or not to show the legend on this widget."},
		{HCLKey: "legend_size", Type: TypeString, OmitEmpty: true,
			Description: "The size of the legend displayed in the widget."},
		{HCLKey: "xaxis", Type: TypeBlock, OmitEmpty: true,
			Description: "A nested block describing the X-Axis Controls. Exactly one nested block is allowed using the structure below.",
			Children:    distributionWidgetXAxisFields},
		{HCLKey: "yaxis", Type: TypeBlock, OmitEmpty: true,
			Description: "A nested block describing the Y-Axis Controls. Exactly one nested block is allowed using the structure below.",
			Children:    distributionWidgetYAxisFields},
		{HCLKey: "request", JSONKey: "requests", Type: TypeBlockList, OmitEmpty: false,
			Description: "A nested block describing the request to use when displaying the widget. Multiple request blocks are allowed using the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query` or `process_query` is required within the request block).",
			Children:    distributionWidgetRequestFields},
	},
}

// HeatmapWidgetSpec corresponds to OpenAPI HeatMapWidgetDefinition.
var heatmapWidgetRequestFields = append([]FieldSpec{
	{HCLKey: "q", Type: TypeString, OmitEmpty: true,
		Description: "The metric query to use for this widget."},
	{HCLKey: "style", Type: TypeBlock, OmitEmpty: true,
		Description: "The style of the widget graph. One nested block is allowed using the structure below.",
		Children:    widgetRequestStyleFields},
}, standardQueryFields...)

var HeatmapWidgetSpec = WidgetSpec{
	HCLKey:   "heatmap_definition",
	JSONType: "heatmap",
	Fields: []FieldSpec{
		// show_legend: OmitEmpty: false — cassette confirms false is emitted when not set
		{HCLKey: "show_legend", Type: TypeBool, OmitEmpty: false,
			Description: "Whether or not to show the legend on this widget."},
		{HCLKey: "legend_size", Type: TypeString, OmitEmpty: true,
			Description: "The size of the legend displayed in the widget."},
		{HCLKey: "yaxis", Type: TypeBlock, OmitEmpty: true,
			Description: "A nested block describing the Y-Axis Controls. The structure of this block is described below.",
			Children:    widgetAxisFields},
		{HCLKey: "event", JSONKey: "events", Type: TypeBlockList, OmitEmpty: true,
			Description: "The definition of the event to overlay on the graph. Multiple `event` blocks are allowed using the structure below.",
			Children:    widgetEventFields},
		{HCLKey: "request", JSONKey: "requests", Type: TypeBlockList, OmitEmpty: false,
			Description: "A nested block describing the request to use when displaying the widget. Multiple `request` blocks are allowed using the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query` or `process_query` is required within the request block).",
			Children:    heatmapWidgetRequestFields},
	},
}

// HostmapWidgetSpec corresponds to OpenAPI HostMapWidgetDefinition.
// "requests" is a JSON *object* (fill + size keys), not an array.
var hostmapRequestInnerFields = []FieldSpec{
	{HCLKey: "fill", Type: TypeBlock, OmitEmpty: true,
		Description: "The query used to fill the map. Exactly one nested block is allowed using the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query` or `process_query` is required within the request block).",
		Children:    hostmapRequestFillSizeFields},
	{HCLKey: "size", Type: TypeBlock, OmitEmpty: true,
		Description: "The query used to size the map. Exactly one nested block is allowed using the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query` or `process_query` is required within the request block).",
		Children:    hostmapRequestFillSizeFields},
}

var HostmapWidgetSpec = WidgetSpec{
	HCLKey:   "hostmap_definition",
	JSONType: "hostmap",
	Fields: []FieldSpec{
		// TypeBlock (not TypeBlockList) — "requests" is a JSON object, not array
		{HCLKey: "request", JSONKey: "requests", Type: TypeBlock, OmitEmpty: true,
			Description: "A nested block describing the request to use when displaying the widget. Multiple `request` blocks are allowed using the structure below.",
			Children:    hostmapRequestInnerFields},
		{HCLKey: "node_type", Type: TypeString, OmitEmpty: true,
			Description: "The type of node used.",
			ValidValues: []string{"host", "container"}},
		{HCLKey: "no_metric_hosts", Type: TypeBool, OmitEmpty: true,
			Description: "A Boolean indicating whether to show nodes with no metrics."},
		{HCLKey: "no_group_hosts", Type: TypeBool, OmitEmpty: true,
			Description: "A Boolean indicating whether to show ungrouped nodes."},
		{HCLKey: "group", Type: TypeStringList, OmitEmpty: true,
			Description: "The list of tags to group nodes by."},
		{HCLKey: "scope", Type: TypeStringList, OmitEmpty: true,
			Description: "The list of tags to filter nodes by."},
		{HCLKey: "style", Type: TypeBlock, OmitEmpty: true,
			Description: "The style of the widget graph. One nested block is allowed using the structure below.",
			Children:    hostmapStyleFields},
	},
}

// QueryValueWidgetSpec corresponds to OpenAPI QueryValueWidgetDefinition.
var queryValueRequestFields = append([]FieldSpec{
	{HCLKey: "q", Type: TypeString, OmitEmpty: true,
		Description: "The metric query to use for this widget."},
	{HCLKey: "audit_query", Type: TypeBlock, OmitEmpty: true,
		Description: "The audit query to use in the widget.",
		Children:    logQueryDefinitionFields},
	{HCLKey: "aggregator", Type: TypeString, OmitEmpty: true,
		Description: "The aggregator to use for time aggregation.",
		ValidValues: []string{"avg", "last", "max", "min", "sum", "percentile"}},
	{HCLKey: "conditional_formats", Type: TypeBlockList, OmitEmpty: true,
		Description: "Conditional formats allow you to set the color of your widget content or background depending on the rule applied to your data. Multiple `conditional_formats` blocks are allowed using the structure below.",
		Children:    widgetConditionalFormatFields},
}, standardQueryFields...)

var QueryValueWidgetSpec = WidgetSpec{
	HCLKey:   "query_value_definition",
	JSONType: "query_value",
	Fields: []FieldSpec{
		// Emitted even when false/0 (cassette-verified)
		{HCLKey: "autoscale", Type: TypeBool, OmitEmpty: false,
			Description: "A Boolean indicating whether to automatically scale the tile."},
		{HCLKey: "custom_unit", Type: TypeString, OmitEmpty: true,
			Description: "The unit for the value displayed in the widget."},
		{HCLKey: "precision", Type: TypeInt, OmitEmpty: false,
			Description: "The precision to use when displaying the tile."},
		{HCLKey: "text_align", Type: TypeString, OmitEmpty: true,
			Description: "The alignment of the widget's text.",
			ValidValues: []string{"center", "left", "right"}},
		{HCLKey: "timeseries_background", Type: TypeBlock, OmitEmpty: true,
			Description: "Set a timeseries on the widget background.",
			Children:    timeseriesBackgroundFields},
		{HCLKey: "request", JSONKey: "requests", Type: TypeBlockList, OmitEmpty: false,
			Description: "A nested block describing the request to use when displaying the widget. Multiple `request` blocks are allowed using the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query` or `process_query` is required within the `request` block).",
			Children:    queryValueRequestFields},
	},
}

// ToplistWidgetSpec corresponds to OpenAPI ToplistWidgetDefinition.
var toplistWidgetRequestFields = append([]FieldSpec{
	{HCLKey: "q", Type: TypeString, OmitEmpty: true,
		Description: "The metric query to use for this widget."},
	{HCLKey: "audit_query", Type: TypeBlock, OmitEmpty: true,
		Description: "The audit query to use in the widget.",
		Children:    logQueryDefinitionFields},
	{HCLKey: "conditional_formats", Type: TypeBlockList, OmitEmpty: true,
		Description: "Conditional formats allow you to set the color of your widget content or background, depending on a rule applied to your data. Multiple `conditional_formats` blocks are allowed using the structure below.",
		Children:    widgetConditionalFormatFields},
	{HCLKey: "style", Type: TypeBlock, OmitEmpty: true,
		Description: "Define request for the widget's style.",
		Children:    widgetRequestStyleFields},
}, standardQueryFields...)

var ToplistWidgetSpec = WidgetSpec{
	HCLKey:   "toplist_definition",
	JSONType: "toplist",
	Fields: []FieldSpec{
		{HCLKey: "style", Type: TypeBlock, OmitEmpty: true,
			Description: "The style of the widget",
			Children:    toplistWidgetStyleFields},
		{HCLKey: "request", JSONKey: "requests", Type: TypeBlockList, OmitEmpty: false,
			Description: "A nested block describing the request to use when displaying the widget. Multiple `request` blocks are allowed using the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query` or `process_query` is required within the `request` block).",
			Children:    toplistWidgetRequestFields},
	},
}

// ScatterplotWidgetSpec corresponds to OpenAPI ScatterPlotWidgetDefinition.
// "requests" is a JSON *object* (x + y keys), not an array.
var scatterplotRequestOuterFields = []FieldSpec{
	{HCLKey: "x", Type: TypeBlock, OmitEmpty: true,
		Description: "The query used for the X-Axis. Exactly one nested block is allowed using the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query`, `apm_stats_query` or `process_query` is required within the block).",
		Children:    scatterplotXYRequestFields},
	{HCLKey: "y", Type: TypeBlock, OmitEmpty: true,
		Description: "The query used for the Y-Axis. Exactly one nested block is allowed using the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query`, `apm_stats_query` or `process_query` is required within the block).",
		Children:    scatterplotXYRequestFields},
	{HCLKey: "scatterplot_table", Type: TypeBlockList, OmitEmpty: true,
		Description: "Scatterplot request containing formulas and functions.",
		Children:    scatterplotTableRequestFields},
}

var ScatterplotWidgetSpec = WidgetSpec{
	HCLKey:   "scatterplot_definition",
	JSONType: "scatterplot",
	Fields: []FieldSpec{
		// TypeBlock (not TypeBlockList) — "requests" is a JSON object, not array
		{HCLKey: "request", JSONKey: "requests", Type: TypeBlock, OmitEmpty: false,
			Description: "A nested block describing the request to use when displaying the widget. Exactly one `request` block is allowed using the structure below.",
			Children:    scatterplotRequestOuterFields},
		{HCLKey: "xaxis", Type: TypeBlock, OmitEmpty: true,
			Description: "A nested block describing the X-Axis Controls. Exactly one nested block is allowed using the structure below.",
			Children:    widgetAxisFields},
		{HCLKey: "yaxis", Type: TypeBlock, OmitEmpty: true,
			Description: "A nested block describing the Y-Axis Controls. Exactly one nested block is allowed using the structure below.",
			Children:    widgetAxisFields},
		{HCLKey: "color_by_groups", Type: TypeStringList, OmitEmpty: true,
			Description: "List of groups used for colors."},
	},
}

// SunburstWidgetSpec corresponds to OpenAPI SunburstWidgetDefinition.
// The JSON "legend" field is polymorphic; HCL uses separate legend_inline and legend_table blocks.
var sunburstWidgetRequestFields = append([]FieldSpec{
	{HCLKey: "q", Type: TypeString, OmitEmpty: true,
		Description: "The metric query to use for this widget."},
	{HCLKey: "network_query", Type: TypeBlock, OmitEmpty: true,
		Description: "The network query to use in the widget.",
		Children:    logQueryDefinitionFields},
	{HCLKey: "audit_query", Type: TypeBlock, OmitEmpty: true,
		Description: "The audit query to use in the widget.",
		Children:    logQueryDefinitionFields},
	{HCLKey: "style", Type: TypeBlock, OmitEmpty: true,
		Description: "Define style for the widget's request.",
		Children:    widgetRequestStyleFields},
}, standardQueryFields...)

var SunburstWidgetSpec = WidgetSpec{
	HCLKey:   "sunburst_definition",
	JSONType: "sunburst",
	Fields: []FieldSpec{
		{HCLKey: "hide_total", Type: TypeBool, OmitEmpty: true,
			Description: "Whether or not to show the total value in the widget."},
		// Both map to JSON "legend"; engine post-processing disambiguates on flatten
		{HCLKey: "legend_inline", JSONKey: "legend", Type: TypeBlock, OmitEmpty: true,
			Description: "Used to configure the inline legend. Cannot be used in conjunction with legend_table.",
			Children:    sunburstLegendInlineFields},
		{HCLKey: "legend_table", JSONKey: "legend", Type: TypeBlock, OmitEmpty: true,
			Description: "Used to configure the table legend. Cannot be used in conjunction with legend_inline.",
			Children: []FieldSpec{
				{HCLKey: "type", Type: TypeString, OmitEmpty: false, Required: true,
					Description: "The type of legend (table or none)."},
			}},
		{HCLKey: "request", JSONKey: "requests", Type: TypeBlockList, OmitEmpty: false,
			Description: "Nested block describing the request to use when displaying the widget. Multiple `request` blocks are allowed with the structure below (exactly one of `q`, `log_query` or `rum_query` is required within the `request` block).",
			Children:    sunburstWidgetRequestFields},
	},
}

// GeomapWidgetSpec corresponds to OpenAPI GeomapWidgetDefinition.
var geomapWidgetRequestFields = []FieldSpec{
	{HCLKey: "q", Type: TypeString, OmitEmpty: true,
		Description: "The metric query to use for this widget."},
	{HCLKey: "log_query", Type: TypeBlock, OmitEmpty: true,
		Description: "The log query to use in the widget.",
		Children:    logQueryDefinitionFields},
	{HCLKey: "rum_query", Type: TypeBlock, OmitEmpty: true,
		Description: "The RUM query to use in the widget.",
		Children:    logQueryDefinitionFields},
	// FormulaAndFunction query/formula fields
	{HCLKey: "query", Type: TypeBlockList, OmitEmpty: true,
		Description: "A list of queries to use in the widget.",
		Children:    formulaAndFunctionQueryFields},
	{HCLKey: "formula", Type: TypeBlockList, OmitEmpty: true,
		Description: "A list of formulas to use in the widget.",
		Children:    widgetFormulaFields},
}

var GeomapWidgetSpec = WidgetSpec{
	HCLKey:   "geomap_definition",
	JSONType: "geomap",
	Fields: []FieldSpec{
		{HCLKey: "style", Type: TypeBlock, OmitEmpty: false,
			Description: "The style of the widget graph. One nested block is allowed using the structure below.",
			Children:    geomapStyleFields},
		{HCLKey: "view", Type: TypeBlock, OmitEmpty: false,
			Description: "The view of the world that the map should render.",
			Children: []FieldSpec{
				{HCLKey: "focus", Type: TypeString, OmitEmpty: false, Required: true,
					Description: "The two-letter ISO code of a country to focus the map on (or `WORLD`)."},
			}},
		{HCLKey: "request", JSONKey: "requests", Type: TypeBlockList, OmitEmpty: false,
			Description: "A nested block describing the request to use when displaying the widget. Multiple `request` blocks are allowed using the structure below (exactly one of `q`, `log_query` or `rum_query` is required within the `request` block).",
			Children:    geomapWidgetRequestFields},
	},
}

// TreemapWidgetSpec corresponds to OpenAPI TreeMapWidgetDefinition.
// Formula-only — formula/query dispatch is handled by the engine.
// Notes:
//   - "title" is omitted here: it is already included in commonWidgetFields.
//   - "color_by" is omitted here: engine.go unconditionally injects "color_by": "user"
//     via post-processor after BuildEngineJSON runs, making a FieldSpec entry unreachable.
//   - "request" FieldSpec is omitted: treemap is formula-only; the engine post-processor
//     handles "requests" in the build direction via isFormulaCapableWidget("treemap").
//   - "custom_links" uses the plural HCL key (matching the existing schema in
//     getTreemapDefinitionSchema), unlike all other widgets which use "custom_link" singular.
// treemapRequestFields contains the HCL schema for treemap request blocks.
// Treemap only supports formula/query style — no old-style requests.
var treemapRequestFields = []FieldSpec{
	{HCLKey: "query", Type: TypeBlockList, OmitEmpty: true,
		Description: "A list of queries to use in the widget.",
		Children:    formulaAndFunctionQueryFields},
	{HCLKey: "formula", Type: TypeBlockList, OmitEmpty: true,
		Description: "A list of formulas to use in the widget.",
		Children:    widgetFormulaFields},
}

var TreemapWidgetSpec = WidgetSpec{
	HCLKey:   "treemap_definition",
	JSONType: "treemap",
	Fields: []FieldSpec{
		// request block: treemap uses formula/query only
		{HCLKey: "request", JSONKey: "requests", Type: TypeBlockList, OmitEmpty: true,
			Description: "Nested block describing the request to use when displaying the widget.",
			Children:    treemapRequestFields},
		// custom_links: both HCL and JSON keys are plural for treemap (schema-verified).
		{HCLKey: "custom_links", JSONKey: "custom_links", Type: TypeBlockList, OmitEmpty: true,
			Description: "A nested block describing a custom link. Multiple `custom_links` blocks are allowed using the structure below.",
			Children:    widgetCustomLinkFields},
	},
}

// TopologyMapWidgetSpec corresponds to OpenAPI TopologyMapWidgetDefinition.
// Note: TopologyRequest.Query is a singular struct (*TopologyQuery) in the API,
// so "query" maps to a JSON object (TypeBlock), not a list (TypeBlockList).
// The HCL schema uses TypeList with an implicit single element.
var topologyRequestFields = []FieldSpec{
	{HCLKey: "request_type", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The request type for the Topology request ('topology')."},
	{HCLKey: "query", Type: TypeBlock, OmitEmpty: false,
		Description: "The query for a Topology request.",
		Children:    topologyQueryFields},
}

var TopologyMapWidgetSpec = WidgetSpec{
	HCLKey:   "topology_map_definition",
	JSONType: "topology_map",
	Fields: []FieldSpec{
		{HCLKey: "request", JSONKey: "requests", Type: TypeBlockList, OmitEmpty: false,
			Description: "A nested block describing the request to use when displaying the widget. Multiple request blocks are allowed using the structure below (`query` and `request_type` are required within the request).",
			Children:    topologyRequestFields},
	},
}

var requestWidgetSpecs = []WidgetSpec{
	ChangeWidgetSpec,
	DistributionWidgetSpec,
	HeatmapWidgetSpec,
	HostmapWidgetSpec,
	QueryValueWidgetSpec,
	ToplistWidgetSpec,
	ScatterplotWidgetSpec,
	SunburstWidgetSpec,
	GeomapWidgetSpec,
	TreemapWidgetSpec,
	TopologyMapWidgetSpec,
}
