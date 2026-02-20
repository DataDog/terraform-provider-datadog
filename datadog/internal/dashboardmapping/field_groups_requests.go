package dashboardmapping

// field_groups_requests.go — Batch B field group additions

// widgetRequestStyleFields corresponds to OpenAPI WidgetStyle.
// Used by: distribution, heatmap, sunburst, toplist (request-level style).
var widgetRequestStyleFields = []FieldSpec{
	{HCLKey: "palette", Type: TypeString, OmitEmpty: true,
		Description: "A color palette to apply to the widget. The available options are available at: https://docs.datadoghq.com/dashboards/widgets/timeseries/#appearance."},
}

// widgetConditionalFormatFields corresponds to OpenAPI WidgetConditionalFormat.
// Used by: query_value, toplist requests.
var widgetConditionalFormatFields = []FieldSpec{
	{HCLKey: "comparator", Type: TypeString, OmitEmpty: false, Required: true,
		Description: "The comparator to use.",
		ValidValues: []string{"<", "<=", ">", ">="}},
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
		ValidValues: []string{"avg", "last", "max", "min", "sum", "percentile"}},
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
