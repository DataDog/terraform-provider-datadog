package dashboardmapping

// widgets.go
//
// Per-widget FieldSpec groups, WidgetSpec declarations, and the allWidgetSpecs registry.
//
// Each per-widget FieldSpec group corresponds to an OpenAPI widget request/definition schema.
// WidgetSpec declarations map HCL definition block keys to JSON widget type strings.
//
// Phase 1 implements: timeseries widget.

// ============================================================
// Timeseries Widget
// ============================================================

// timeseriesWidgetRequestStyleFields corresponds to OpenAPI
// components/schemas/WidgetRequestStyle (inline on TimeseriesWidgetRequest).
var timeseriesWidgetRequestStyleFields = []FieldSpec{
	{
		HCLKey:      "palette",
		Type:        TypeString,
		OmitEmpty:   true,
		Description: "A color palette to apply to the widget. The available options are available at: https://docs.datadoghq.com/dashboards/widgets/timeseries/#appearance.",
	},
	{
		HCLKey:      "line_type",
		Type:        TypeString,
		OmitEmpty:   true,
		Description: "The type of lines displayed.",
		ValidValues: []string{"dashed", "dotted", "solid"},
	},
	{
		HCLKey:      "line_width",
		Type:        TypeString,
		OmitEmpty:   true,
		Description: "The width of line displayed.",
		ValidValues: []string{"normal", "thick", "thin"},
	},
	{
		HCLKey:      "order_by",
		Type:        TypeString,
		OmitEmpty:   true,
		Description: "How to order series in timeseries visualizations.",
		ValidValues: []string{"tags", "values"},
	},
}

// timeseriesWidgetMetadataFields corresponds to the inline metadata object
// on TimeseriesWidgetRequest (no standalone OpenAPI $ref; defined inline).
var timeseriesWidgetMetadataFields = []FieldSpec{
	{
		HCLKey:      "expression",
		Type:        TypeString,
		OmitEmpty:   false,
		Required:    true,
		Description: "The expression name.",
	},
	{
		HCLKey:      "alias_name",
		Type:        TypeString,
		OmitEmpty:   true,
		Description: "The expression alias.",
	},
}

// timeseriesWidgetRequestFields corresponds to OpenAPI
// components/schemas/TimeseriesWidgetRequest.
// HCL key: "request" (singular), JSON key: "requests" (plural).
var timeseriesWidgetRequestFields = []FieldSpec{
	{HCLKey: "q", Type: TypeString, OmitEmpty: true},
	{
		HCLKey:      "display_type",
		Type:        TypeString,
		OmitEmpty:   true,
		Description: "How to display the marker lines.",
		ValidValues: []string{"area", "bars", "line", "overlay"},
	},
	// on_right_yaxis is always emitted even when false — cassette confirms both true and false appear
	{
		HCLKey:      "on_right_yaxis",
		Type:        TypeBool,
		OmitEmpty:   false,
		Description: "A Boolean indicating whether the request uses the right or left Y-Axis.",
	},
	{
		HCLKey:      "style",
		Type:        TypeBlock,
		OmitEmpty:   true,
		Description: "The style of the widget graph. Exactly one `style` block is allowed using the structure below.",
		Children:    timeseriesWidgetRequestStyleFields,
	},
	{
		HCLKey:      "metadata",
		Type:        TypeBlockList,
		OmitEmpty:   true,
		Description: "Used to define expression aliases. Multiple `metadata` blocks are allowed using the structure below.",
		Children:    timeseriesWidgetMetadataFields,
	},
	// The following 7 fields all use logQueryDefinitionFields (same OpenAPI $ref,
	// different JSON key per query source type):
	{HCLKey: "log_query", Type: TypeBlock, OmitEmpty: true, Children: logQueryDefinitionFields},
	{HCLKey: "apm_query", Type: TypeBlock, OmitEmpty: true, Children: logQueryDefinitionFields},
	{HCLKey: "rum_query", Type: TypeBlock, OmitEmpty: true, Children: logQueryDefinitionFields},
	{HCLKey: "network_query", Type: TypeBlock, OmitEmpty: true, Children: logQueryDefinitionFields},
	{HCLKey: "security_query", Type: TypeBlock, OmitEmpty: true, Children: logQueryDefinitionFields},
	{HCLKey: "audit_query", Type: TypeBlock, OmitEmpty: true, Children: logQueryDefinitionFields},
	{HCLKey: "profile_metrics_query", Type: TypeBlock, OmitEmpty: true, Children: logQueryDefinitionFields},
	// ProcessQueryDefinition
	{HCLKey: "process_query", Type: TypeBlock, OmitEmpty: true, Children: processQueryDefinitionFields},
	// FormulaAndFunction query/formula fields
	{HCLKey: "query", Type: TypeBlockList, OmitEmpty: true,
		Description: "A list of queries to use in the widget.",
		Children:    formulaAndFunctionQueryFields},
	{HCLKey: "formula", Type: TypeBlockList, OmitEmpty: true,
		Description: "A list of formulas to use in the widget.",
		Children:    widgetFormulaFields},
}

// TimeseriesWidgetSpec corresponds to OpenAPI
// components/schemas/TimeseriesWidgetDefinition.
var TimeseriesWidgetSpec = WidgetSpec{
	HCLKey:   "timeseries_definition",
	JSONType: "timeseries",
	Fields: []FieldSpec{
		// show_legend is always emitted even when false — cassette confirms false appears when not set
		{
			HCLKey:      "show_legend",
			Type:        TypeBool,
			OmitEmpty:   false,
			Description: "Whether or not to show the legend on this widget.",
		},
		{
			HCLKey:      "legend_size",
			Type:        TypeString,
			OmitEmpty:   true,
			Description: "The size of the legend displayed in the widget.",
			ValidValues: []string{"0", "2", "4", "8", "16", "auto"},
		},
		{
			HCLKey:      "legend_layout",
			Type:        TypeString,
			OmitEmpty:   true,
			Description: "The layout of the legend displayed in the widget.",
			ValidValues: []string{"auto", "horizontal", "vertical"},
		},
		{
			HCLKey:      "legend_columns",
			Type:        TypeStringList,
			OmitEmpty:   true,
			Description: "A list of columns to display in the legend.",
			ValidValues: []string{"value", "avg", "sum", "min", "max"},
			UseSet:      true,
		},
		// WidgetAxis — used twice for the two y-axes
		{
			HCLKey:      "yaxis",
			Type:        TypeBlock,
			OmitEmpty:   true,
			Description: "A nested block describing the Y-Axis Controls. The structure of this block is described below.",
			Children:    widgetAxisFields,
		},
		{
			HCLKey:      "right_yaxis",
			Type:        TypeBlock,
			OmitEmpty:   true,
			Description: "A nested block describing the right Y-Axis Controls. See the `on_right_yaxis` property for which request will use this axis. The structure of this block is described below.",
			Children:    widgetAxisFields,
		},
		// WidgetMarker: HCL singular "marker" → JSON plural "markers"
		{
			HCLKey:      "marker",
			JSONKey:     "markers",
			Type:        TypeBlockList,
			OmitEmpty:   true,
			Description: "A nested block describing the marker to use when displaying the widget. The structure of this block is described below. Multiple `marker` blocks are allowed within a given `tile_def` block.",
			Children:    widgetMarkerFields,
		},
		// WidgetEvent: HCL singular "event" → JSON plural "events"
		{
			HCLKey:      "event",
			JSONKey:     "events",
			Type:        TypeBlockList,
			OmitEmpty:   true,
			Description: "The definition of the event to overlay on the graph. Multiple `event` blocks are allowed using the structure below.",
			Children:    widgetEventFields,
		},
		// TimeseriesWidgetRequest: HCL singular "request" → JSON plural "requests"
		// OmitEmpty: false — always emit even if empty (matches SDK behavior in cassettes)
		{
			HCLKey:      "request",
			JSONKey:     "requests",
			Type:        TypeBlockList,
			OmitEmpty:   false,
			Description: "A nested block describing the request to use when displaying the widget. Multiple `request` blocks are allowed using the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `network_query`, `security_query` or `process_query` is required within the `request` block).",
			Children:    timeseriesWidgetRequestFields,
		},
	},
}

// allWidgetSpecs is the complete registry of all implemented widget types.
// It is assembled from per-batch sub-slices so that parallel agents can each
// work in their own file without merge conflicts on this registry.
//
// Batch files:
//
//	widgets_simple.go   — Batch A: no-request widgets (alert_graph, note, image, etc.)
//	widgets_requests.go — Batch B: request-based widgets (change, heatmap, toplist, etc.)
//	widgets_complex.go  — Batch C: complex/structural widgets (table, list_stream, slo, group, etc.)
var allWidgetSpecs = concatWidgetSpecs(
	coreWidgetSpecs,
	simpleWidgetSpecs,
	requestWidgetSpecs,
	complexWidgetSpecs,
)

// coreWidgetSpecs contains the Phase 1 timeseries implementation.
var coreWidgetSpecs = []WidgetSpec{
	TimeseriesWidgetSpec,
}

// concatWidgetSpecs merges multiple WidgetSpec slices into one.
func concatWidgetSpecs(slices ...[]WidgetSpec) []WidgetSpec {
	var result []WidgetSpec
	for _, s := range slices {
		result = append(result, s...)
	}
	return result
}
