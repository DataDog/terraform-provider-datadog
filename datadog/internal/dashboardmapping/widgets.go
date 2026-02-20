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
	{HCLKey: "palette", Type: TypeString, OmitEmpty: true},
	{HCLKey: "line_type", Type: TypeString, OmitEmpty: true},
	{HCLKey: "line_width", Type: TypeString, OmitEmpty: true},
	{HCLKey: "order_by", Type: TypeString, OmitEmpty: true},
}

// timeseriesWidgetMetadataFields corresponds to the inline metadata object
// on TimeseriesWidgetRequest (no standalone OpenAPI $ref; defined inline).
var timeseriesWidgetMetadataFields = []FieldSpec{
	{HCLKey: "expression", Type: TypeString, OmitEmpty: false},
	{HCLKey: "alias_name", Type: TypeString, OmitEmpty: true},
}

// timeseriesWidgetRequestFields corresponds to OpenAPI
// components/schemas/TimeseriesWidgetRequest.
// HCL key: "request" (singular), JSON key: "requests" (plural).
var timeseriesWidgetRequestFields = []FieldSpec{
	{HCLKey: "q", Type: TypeString, OmitEmpty: true},
	{HCLKey: "display_type", Type: TypeString, OmitEmpty: true},
	// on_right_yaxis is always emitted even when false — cassette confirms both true and false appear
	{HCLKey: "on_right_yaxis", Type: TypeBool, OmitEmpty: false},
	{HCLKey: "style", Type: TypeBlock, OmitEmpty: true, Children: timeseriesWidgetRequestStyleFields},
	{HCLKey: "metadata", Type: TypeBlockList, OmitEmpty: true, Children: timeseriesWidgetMetadataFields},
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
}

// timeseriesWidgetSpec corresponds to OpenAPI
// components/schemas/TimeseriesWidgetDefinition.
var timeseriesWidgetSpec = WidgetSpec{
	HCLKey:   "timeseries_definition",
	JSONType: "timeseries",
	Fields: []FieldSpec{
		// show_legend is always emitted even when false — cassette confirms false appears when not set
		{HCLKey: "show_legend", Type: TypeBool, OmitEmpty: false},
		{HCLKey: "legend_size", Type: TypeString, OmitEmpty: true},
		{HCLKey: "legend_layout", Type: TypeString, OmitEmpty: true},
		{HCLKey: "legend_columns", Type: TypeStringList, OmitEmpty: true},
		// WidgetAxis — used twice for the two y-axes
		{HCLKey: "yaxis", Type: TypeBlock, OmitEmpty: true, Children: widgetAxisFields},
		{HCLKey: "right_yaxis", Type: TypeBlock, OmitEmpty: true, Children: widgetAxisFields},
		// WidgetMarker: HCL singular "marker" → JSON plural "markers"
		{HCLKey: "marker", JSONKey: "markers", Type: TypeBlockList, OmitEmpty: true, Children: widgetMarkerFields},
		// WidgetEvent: HCL singular "event" → JSON plural "events"
		{HCLKey: "event", JSONKey: "events", Type: TypeBlockList, OmitEmpty: true, Children: widgetEventFields},
		// TimeseriesWidgetRequest: HCL singular "request" → JSON plural "requests"
		// OmitEmpty: false — always emit even if empty (matches SDK behavior in cassettes)
		{HCLKey: "request", JSONKey: "requests", Type: TypeBlockList, OmitEmpty: false,
			Children: timeseriesWidgetRequestFields},
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
	timeseriesWidgetSpec,
}

// concatWidgetSpecs merges multiple WidgetSpec slices into one.
func concatWidgetSpecs(slices ...[]WidgetSpec) []WidgetSpec {
	var result []WidgetSpec
	for _, s := range slices {
		result = append(result, s...)
	}
	return result
}
