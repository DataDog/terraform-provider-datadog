package dashboardmapping

// widgets_requests.go — Batch B
//
// Widget types with standard query-based requests: change, distribution, heatmap,
// hostmap, query_value, toplist, scatterplot, sunburst, geomap, treemap, topology_map.

// changeWidgetSpec corresponds to OpenAPI ChangeWidgetDefinition.
var changeWidgetRequestFields = append([]FieldSpec{
	{HCLKey: "q", Type: TypeString, OmitEmpty: true},
	{HCLKey: "change_type", Type: TypeString, OmitEmpty: true},
	{HCLKey: "compare_to", Type: TypeString, OmitEmpty: true},
	// Emitted even when false (cassette-verified)
	{HCLKey: "increase_good", Type: TypeBool, OmitEmpty: false},
	{HCLKey: "order_by", Type: TypeString, OmitEmpty: true},
	{HCLKey: "order_dir", Type: TypeString, OmitEmpty: true},
	{HCLKey: "show_present", Type: TypeBool, OmitEmpty: false},
}, standardQueryFields...)

var changeWidgetSpec = WidgetSpec{
	HCLKey:   "change_definition",
	JSONType: "change",
	Fields: []FieldSpec{
		{HCLKey: "request", JSONKey: "requests", Type: TypeBlockList, OmitEmpty: false,
			Children: changeWidgetRequestFields},
	},
}

// distributionWidgetSpec corresponds to OpenAPI DistributionWidgetDefinition.
var distributionWidgetRequestFields = append([]FieldSpec{
	{HCLKey: "q", Type: TypeString, OmitEmpty: true},
	{HCLKey: "style", Type: TypeBlock, OmitEmpty: true, Children: widgetRequestStyleFields},
	// apm_stats_query is supported by distribution requests (OpenAPI DistributionWidgetRequest)
	{HCLKey: "apm_stats_query", Type: TypeBlock, OmitEmpty: true, Children: apmStatsQueryFields},
}, standardQueryFields...)

// distributionWidgetXAxisFields corresponds to OpenAPI DistributionWidgetXAxis.
// Differs from WidgetAxis: include_zero uses OmitEmpty: true (not false).
var distributionWidgetXAxisFields = []FieldSpec{
	{HCLKey: "scale", Type: TypeString, OmitEmpty: true},
	{HCLKey: "min", Type: TypeString, OmitEmpty: true},
	{HCLKey: "max", Type: TypeString, OmitEmpty: true},
	{HCLKey: "include_zero", Type: TypeBool, OmitEmpty: true},
}

// distributionWidgetYAxisFields extends distributionWidgetXAxisFields with a label field.
var distributionWidgetYAxisFields = append(
	append([]FieldSpec{}, distributionWidgetXAxisFields...),
	FieldSpec{HCLKey: "label", Type: TypeString, OmitEmpty: true},
)

var distributionWidgetSpec = WidgetSpec{
	HCLKey:   "distribution_definition",
	JSONType: "distribution",
	Fields: []FieldSpec{
		// show_legend: OmitEmpty: false — cassette confirms false is emitted when not set
		{HCLKey: "show_legend", Type: TypeBool, OmitEmpty: false},
		{HCLKey: "legend_size", Type: TypeString, OmitEmpty: true},
		{HCLKey: "xaxis", Type: TypeBlock, OmitEmpty: true, Children: distributionWidgetXAxisFields},
		{HCLKey: "yaxis", Type: TypeBlock, OmitEmpty: true, Children: distributionWidgetYAxisFields},
		{HCLKey: "request", JSONKey: "requests", Type: TypeBlockList, OmitEmpty: false,
			Children: distributionWidgetRequestFields},
	},
}

// heatmapWidgetSpec corresponds to OpenAPI HeatMapWidgetDefinition.
var heatmapWidgetRequestFields = append([]FieldSpec{
	{HCLKey: "q", Type: TypeString, OmitEmpty: true},
	{HCLKey: "style", Type: TypeBlock, OmitEmpty: true, Children: widgetRequestStyleFields},
}, standardQueryFields...)

var heatmapWidgetSpec = WidgetSpec{
	HCLKey:   "heatmap_definition",
	JSONType: "heatmap",
	Fields: []FieldSpec{
		// show_legend: OmitEmpty: false — cassette confirms false is emitted when not set
		{HCLKey: "show_legend", Type: TypeBool, OmitEmpty: false},
		{HCLKey: "legend_size", Type: TypeString, OmitEmpty: true},
		{HCLKey: "yaxis", Type: TypeBlock, OmitEmpty: true, Children: widgetAxisFields},
		{HCLKey: "event", JSONKey: "events", Type: TypeBlockList, OmitEmpty: true,
			Children: widgetEventFields},
		{HCLKey: "request", JSONKey: "requests", Type: TypeBlockList, OmitEmpty: false,
			Children: heatmapWidgetRequestFields},
	},
}

// hostmapWidgetSpec corresponds to OpenAPI HostMapWidgetDefinition.
// "requests" is a JSON *object* (fill + size keys), not an array.
var hostmapRequestInnerFields = []FieldSpec{
	{HCLKey: "fill", Type: TypeBlock, OmitEmpty: true, Children: hostmapRequestFillSizeFields},
	{HCLKey: "size", Type: TypeBlock, OmitEmpty: true, Children: hostmapRequestFillSizeFields},
}

var hostmapWidgetSpec = WidgetSpec{
	HCLKey:   "hostmap_definition",
	JSONType: "hostmap",
	Fields: []FieldSpec{
		// TypeBlock (not TypeBlockList) — "requests" is a JSON object, not array
		{HCLKey: "request", JSONKey: "requests", Type: TypeBlock, OmitEmpty: true,
			Children: hostmapRequestInnerFields},
		{HCLKey: "node_type", Type: TypeString, OmitEmpty: true},
		{HCLKey: "no_metric_hosts", Type: TypeBool, OmitEmpty: true},
		{HCLKey: "no_group_hosts", Type: TypeBool, OmitEmpty: true},
		{HCLKey: "group", Type: TypeStringList, OmitEmpty: true},
		{HCLKey: "scope", Type: TypeStringList, OmitEmpty: true},
		{HCLKey: "style", Type: TypeBlock, OmitEmpty: true, Children: hostmapStyleFields},
	},
}

// queryValueWidgetSpec corresponds to OpenAPI QueryValueWidgetDefinition.
var queryValueRequestFields = append([]FieldSpec{
	{HCLKey: "q", Type: TypeString, OmitEmpty: true},
	{HCLKey: "audit_query", Type: TypeBlock, OmitEmpty: true, Children: logQueryDefinitionFields},
	{HCLKey: "aggregator", Type: TypeString, OmitEmpty: true},
	{HCLKey: "conditional_formats", Type: TypeBlockList, OmitEmpty: true,
		Children: widgetConditionalFormatFields},
}, standardQueryFields...)

var queryValueWidgetSpec = WidgetSpec{
	HCLKey:   "query_value_definition",
	JSONType: "query_value",
	Fields: []FieldSpec{
		// Emitted even when false/0 (cassette-verified)
		{HCLKey: "autoscale", Type: TypeBool, OmitEmpty: false},
		{HCLKey: "custom_unit", Type: TypeString, OmitEmpty: true},
		{HCLKey: "precision", Type: TypeInt, OmitEmpty: false},
		{HCLKey: "text_align", Type: TypeString, OmitEmpty: true},
		{HCLKey: "timeseries_background", Type: TypeBlock, OmitEmpty: true,
			Children: timeseriesBackgroundFields},
		{HCLKey: "request", JSONKey: "requests", Type: TypeBlockList, OmitEmpty: false,
			Children: queryValueRequestFields},
	},
}

// toplistWidgetSpec corresponds to OpenAPI ToplistWidgetDefinition.
var toplistWidgetRequestFields = append([]FieldSpec{
	{HCLKey: "q", Type: TypeString, OmitEmpty: true},
	{HCLKey: "audit_query", Type: TypeBlock, OmitEmpty: true, Children: logQueryDefinitionFields},
	{HCLKey: "conditional_formats", Type: TypeBlockList, OmitEmpty: true,
		Children: widgetConditionalFormatFields},
	{HCLKey: "style", Type: TypeBlock, OmitEmpty: true, Children: widgetRequestStyleFields},
}, standardQueryFields...)

var toplistWidgetSpec = WidgetSpec{
	HCLKey:   "toplist_definition",
	JSONType: "toplist",
	Fields: []FieldSpec{
		{HCLKey: "style", Type: TypeBlock, OmitEmpty: true, Children: toplistWidgetStyleFields},
		{HCLKey: "request", JSONKey: "requests", Type: TypeBlockList, OmitEmpty: false,
			Children: toplistWidgetRequestFields},
	},
}

// scatterplotWidgetSpec corresponds to OpenAPI ScatterPlotWidgetDefinition.
// "requests" is a JSON *object* (x + y keys), not an array.
var scatterplotRequestOuterFields = []FieldSpec{
	{HCLKey: "x", Type: TypeBlock, OmitEmpty: true, Children: scatterplotXYRequestFields},
	{HCLKey: "y", Type: TypeBlock, OmitEmpty: true, Children: scatterplotXYRequestFields},
}

var scatterplotWidgetSpec = WidgetSpec{
	HCLKey:   "scatterplot_definition",
	JSONType: "scatterplot",
	Fields: []FieldSpec{
		// TypeBlock (not TypeBlockList) — "requests" is a JSON object, not array
		{HCLKey: "request", JSONKey: "requests", Type: TypeBlock, OmitEmpty: false,
			Children: scatterplotRequestOuterFields},
		{HCLKey: "xaxis", Type: TypeBlock, OmitEmpty: true, Children: widgetAxisFields},
		{HCLKey: "yaxis", Type: TypeBlock, OmitEmpty: true, Children: widgetAxisFields},
		{HCLKey: "color_by_groups", Type: TypeStringList, OmitEmpty: true},
	},
}

// sunburstWidgetSpec corresponds to OpenAPI SunburstWidgetDefinition.
// The JSON "legend" field is polymorphic; HCL uses separate legend_inline and legend_table blocks.
var sunburstWidgetRequestFields = append([]FieldSpec{
	{HCLKey: "q", Type: TypeString, OmitEmpty: true},
	{HCLKey: "network_query", Type: TypeBlock, OmitEmpty: true, Children: logQueryDefinitionFields},
	{HCLKey: "audit_query", Type: TypeBlock, OmitEmpty: true, Children: logQueryDefinitionFields},
	{HCLKey: "style", Type: TypeBlock, OmitEmpty: true, Children: widgetRequestStyleFields},
}, standardQueryFields...)

var sunburstWidgetSpec = WidgetSpec{
	HCLKey:   "sunburst_definition",
	JSONType: "sunburst",
	Fields: []FieldSpec{
		{HCLKey: "hide_total", Type: TypeBool, OmitEmpty: true},
		// Both map to JSON "legend"; engine post-processing disambiguates on flatten
		{HCLKey: "legend_inline", JSONKey: "legend", Type: TypeBlock, OmitEmpty: true,
			Children: sunburstLegendInlineFields},
		{HCLKey: "legend_table", JSONKey: "legend", Type: TypeBlock, OmitEmpty: true,
			Children: []FieldSpec{
				{HCLKey: "type", Type: TypeString, OmitEmpty: false},
			}},
		{HCLKey: "request", JSONKey: "requests", Type: TypeBlockList, OmitEmpty: false,
			Children: sunburstWidgetRequestFields},
	},
}

// geomapWidgetSpec corresponds to OpenAPI GeomapWidgetDefinition.
var geomapWidgetRequestFields = []FieldSpec{
	{HCLKey: "q", Type: TypeString, OmitEmpty: true},
	{HCLKey: "log_query", Type: TypeBlock, OmitEmpty: true, Children: logQueryDefinitionFields},
	{HCLKey: "rum_query", Type: TypeBlock, OmitEmpty: true, Children: logQueryDefinitionFields},
}

var geomapWidgetSpec = WidgetSpec{
	HCLKey:   "geomap_definition",
	JSONType: "geomap",
	Fields: []FieldSpec{
		{HCLKey: "style", Type: TypeBlock, OmitEmpty: false, Children: geomapStyleFields},
		{HCLKey: "view", Type: TypeBlock, OmitEmpty: false, Children: []FieldSpec{
			{HCLKey: "focus", Type: TypeString, OmitEmpty: false},
		}},
		{HCLKey: "request", JSONKey: "requests", Type: TypeBlockList, OmitEmpty: false,
			Children: geomapWidgetRequestFields},
	},
}

// treemapWidgetSpec corresponds to OpenAPI TreeMapWidgetDefinition.
// Formula-only — formula/query dispatch is handled by the engine.
// Notes:
//   - "title" is omitted here: it is already included in commonWidgetFields.
//   - "color_by" is omitted here: engine.go unconditionally injects "color_by": "user"
//     via post-processor after BuildEngineJSON runs, making a FieldSpec entry unreachable.
//   - "request" FieldSpec is omitted: treemap is formula-only; the engine post-processor
//     handles "requests" in the build direction via isFormulaCapableWidget("treemap").
//   - "custom_links" uses the plural HCL key (matching the existing schema in
//     getTreemapDefinitionSchema), unlike all other widgets which use "custom_link" singular.
var treemapWidgetSpec = WidgetSpec{
	HCLKey:   "treemap_definition",
	JSONType: "treemap",
	Fields: []FieldSpec{
		// custom_links: both HCL and JSON keys are plural for treemap (schema-verified).
		{HCLKey: "custom_links", JSONKey: "custom_links", Type: TypeBlockList, OmitEmpty: true,
			Children: widgetCustomLinkFields},
	},
}

// topologyMapWidgetSpec corresponds to OpenAPI TopologyMapWidgetDefinition.
// Note: TopologyRequest.Query is a singular struct (*TopologyQuery) in the API,
// so "query" maps to a JSON object (TypeBlock), not a list (TypeBlockList).
// The HCL schema uses TypeList with an implicit single element.
var topologyRequestFields = []FieldSpec{
	{HCLKey: "request_type", Type: TypeString, OmitEmpty: false},
	{HCLKey: "query", Type: TypeBlock, OmitEmpty: false, Children: topologyQueryFields},
}

var topologyMapWidgetSpec = WidgetSpec{
	HCLKey:   "topology_map_definition",
	JSONType: "topology_map",
	Fields: []FieldSpec{
		{HCLKey: "request", JSONKey: "requests", Type: TypeBlockList, OmitEmpty: false,
			Children: topologyRequestFields},
	},
}

var requestWidgetSpecs = []WidgetSpec{
	changeWidgetSpec,
	distributionWidgetSpec,
	heatmapWidgetSpec,
	hostmapWidgetSpec,
	queryValueWidgetSpec,
	toplistWidgetSpec,
	scatterplotWidgetSpec,
	sunburstWidgetSpec,
	geomapWidgetSpec,
	treemapWidgetSpec,
	topologyMapWidgetSpec,
}
