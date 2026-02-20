package dashboardmapping

// field_groups_requests.go — Batch B field group additions

// widgetRequestStyleFields corresponds to OpenAPI WidgetStyle.
// Used by: distribution, heatmap, sunburst, toplist (request-level style).
var widgetRequestStyleFields = []FieldSpec{
	{HCLKey: "palette", Type: TypeString, OmitEmpty: true},
}

// widgetConditionalFormatFields corresponds to OpenAPI WidgetConditionalFormat.
// Used by: query_value, toplist requests.
var widgetConditionalFormatFields = []FieldSpec{
	{HCLKey: "comparator", Type: TypeString, OmitEmpty: false},
	{HCLKey: "value", Type: TypeFloat, OmitEmpty: false},
	{HCLKey: "palette", Type: TypeString, OmitEmpty: false},
	{HCLKey: "custom_bg_color", Type: TypeString, OmitEmpty: true},
	{HCLKey: "custom_fg_color", Type: TypeString, OmitEmpty: true},
	{HCLKey: "image_url", Type: TypeString, OmitEmpty: true},
	// Emitted even when false (cassette-verified)
	{HCLKey: "hide_value", Type: TypeBool, OmitEmpty: false},
	{HCLKey: "timeframe", Type: TypeString, OmitEmpty: true},
	{HCLKey: "metric", Type: TypeString, OmitEmpty: true},
}

// hostmapRequestFillSizeFields corresponds to OpenAPI HostMapRequest.
// Used by: hostmap fill and size sub-blocks.
var hostmapRequestFillSizeFields = append([]FieldSpec{
	{HCLKey: "q", Type: TypeString, OmitEmpty: true},
}, standardQueryFields...)

// hostmapStyleFields corresponds to the inline style block on HostMapWidgetDefinition.
var hostmapStyleFields = []FieldSpec{
	{HCLKey: "palette", Type: TypeString, OmitEmpty: true},
	{HCLKey: "palette_flip", Type: TypeBool, OmitEmpty: true},
	{HCLKey: "fill_min", Type: TypeString, OmitEmpty: true},
	{HCLKey: "fill_max", Type: TypeString, OmitEmpty: true},
}

// geomapStyleFields corresponds to the style block on GeomapWidgetDefinition.
var geomapStyleFields = []FieldSpec{
	// Both required — emitted even when false
	{HCLKey: "palette", Type: TypeString, OmitEmpty: false},
	{HCLKey: "palette_flip", Type: TypeBool, OmitEmpty: false},
}

// sunburstLegendInlineFields corresponds to OpenAPI SunburstWidgetLegendInlineAutomatic.
// Kept named (3 fields); sunburstLegendTableFields (1 field) inlined into sunburstWidgetSpec.
// geomapViewFields (1 field) inlined into geomapWidgetSpec.
var sunburstLegendInlineFields = []FieldSpec{
	{HCLKey: "type", Type: TypeString, OmitEmpty: false},
	// Emitted even when false (cassette-verified)
	{HCLKey: "hide_value", Type: TypeBool, OmitEmpty: false},
	{HCLKey: "hide_percent", Type: TypeBool, OmitEmpty: false},
}

// timeseriesBackgroundFields corresponds to OpenAPI TimeseriesBackground.
// Used by: query_value timeseries_background block.
var timeseriesBackgroundFields = []FieldSpec{
	{HCLKey: "type", Type: TypeString, OmitEmpty: false},
	{HCLKey: "yaxis", Type: TypeBlock, OmitEmpty: true, Children: widgetAxisFields},
}

// scatterplotXYRequestFields corresponds to OpenAPI ScatterPlotRequest.
// Used by: scatterplot x and y sub-blocks.
var scatterplotXYRequestFields = append([]FieldSpec{
	{HCLKey: "q", Type: TypeString, OmitEmpty: true},
	{HCLKey: "aggregator", Type: TypeString, OmitEmpty: true},
}, standardQueryFields...)

// toplistWidgetStyleDisplayFields corresponds to the display sub-block inside toplist style.
var toplistWidgetStyleDisplayFields = []FieldSpec{
	{HCLKey: "type", Type: TypeString, OmitEmpty: false},
}

// toplistWidgetStyleFields corresponds to OpenAPI ToplistWidgetStyle.
// Note: "display" is a single JSON object (MaxItems:1 in HCL); we use TypeBlock
// to emit a single object rather than an array.
var toplistWidgetStyleFields = []FieldSpec{
	{HCLKey: "display", Type: TypeBlock, OmitEmpty: true, Children: toplistWidgetStyleDisplayFields},
	{HCLKey: "palette", Type: TypeString, OmitEmpty: true},
	{HCLKey: "scaling", Type: TypeString, OmitEmpty: true},
}

// topologyQueryFields corresponds to the inline query block on TopologyRequest.
var topologyQueryFields = []FieldSpec{
	{HCLKey: "data_source", Type: TypeString, OmitEmpty: false},
	{HCLKey: "service", Type: TypeString, OmitEmpty: false},
	{HCLKey: "filters", Type: TypeStringList, OmitEmpty: false},
}
