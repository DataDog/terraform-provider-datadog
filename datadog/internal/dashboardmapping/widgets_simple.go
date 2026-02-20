package dashboardmapping

// widgets_simple.go — Batch A
//
// Widget types with no request blocks: alert_graph, alert_value, free_text, iframe,
// image, note, event_stream, event_timeline, check_status, log_stream,
// manage_status, run_workflow, service_map, trace_service.

// alertGraphWidgetSpec corresponds to OpenAPI AlertGraphWidgetDefinition.
var alertGraphWidgetSpec = WidgetSpec{
	HCLKey:   "alert_graph_definition",
	JSONType: "alert_graph",
	Fields: []FieldSpec{
		{HCLKey: "alert_id", Type: TypeString, OmitEmpty: false},
		{HCLKey: "viz_type", Type: TypeString, OmitEmpty: false},
	},
}

// alertValueWidgetSpec corresponds to OpenAPI AlertValueWidgetDefinition.
var alertValueWidgetSpec = WidgetSpec{
	HCLKey:   "alert_value_definition",
	JSONType: "alert_value",
	Fields: []FieldSpec{
		{HCLKey: "alert_id", Type: TypeString, OmitEmpty: false},
		{HCLKey: "precision", Type: TypeInt, OmitEmpty: true},
		{HCLKey: "unit", Type: TypeString, OmitEmpty: true},
		{HCLKey: "text_align", Type: TypeString, OmitEmpty: true},
	},
}

// freeTextWidgetSpec corresponds to OpenAPI FreeTextWidgetDefinition.
var freeTextWidgetSpec = WidgetSpec{
	HCLKey:   "free_text_definition",
	JSONType: "free_text",
	Fields: []FieldSpec{
		{HCLKey: "text", Type: TypeString, OmitEmpty: false},
		{HCLKey: "color", Type: TypeString, OmitEmpty: true},
		{HCLKey: "font_size", Type: TypeString, OmitEmpty: true},
		{HCLKey: "text_align", Type: TypeString, OmitEmpty: true},
	},
}

// iframeWidgetSpec corresponds to OpenAPI IFrameWidgetDefinition.
var iframeWidgetSpec = WidgetSpec{
	HCLKey:   "iframe_definition",
	JSONType: "iframe",
	Fields: []FieldSpec{
		{HCLKey: "url", Type: TypeString, OmitEmpty: false},
	},
}

// imageWidgetSpec corresponds to OpenAPI ImageWidgetDefinition.
var imageWidgetSpec = WidgetSpec{
	HCLKey:   "image_definition",
	JSONType: "image",
	Fields: []FieldSpec{
		{HCLKey: "url", Type: TypeString, OmitEmpty: false},
		{HCLKey: "url_dark_theme", Type: TypeString, OmitEmpty: true},
		{HCLKey: "sizing", Type: TypeString, OmitEmpty: true},
		{HCLKey: "margin", Type: TypeString, OmitEmpty: true},
		// Default: true in schema — always emitted
		{HCLKey: "has_background", Type: TypeBool, OmitEmpty: false},
		{HCLKey: "has_border", Type: TypeBool, OmitEmpty: false},
		{HCLKey: "horizontal_align", Type: TypeString, OmitEmpty: true},
		{HCLKey: "vertical_align", Type: TypeString, OmitEmpty: true},
	},
}

// noteWidgetSpec corresponds to OpenAPI NoteWidgetDefinition.
var noteWidgetSpec = WidgetSpec{
	HCLKey:   "note_definition",
	JSONType: "note",
	Fields: []FieldSpec{
		{HCLKey: "content", Type: TypeString, OmitEmpty: false},
		{HCLKey: "background_color", Type: TypeString, OmitEmpty: true},
		{HCLKey: "font_size", Type: TypeString, OmitEmpty: true},
		{HCLKey: "text_align", Type: TypeString, OmitEmpty: true},
		{HCLKey: "vertical_align", Type: TypeString, OmitEmpty: true},
		// Default: true — always emitted
		{HCLKey: "has_padding", Type: TypeBool, OmitEmpty: false},
		// Emitted even when false (cassette-verified)
		{HCLKey: "show_tick", Type: TypeBool, OmitEmpty: false},
		{HCLKey: "tick_pos", Type: TypeString, OmitEmpty: true},
		{HCLKey: "tick_edge", Type: TypeString, OmitEmpty: true},
	},
}

// eventStreamWidgetSpec corresponds to OpenAPI EventStreamWidgetDefinition.
var eventStreamWidgetSpec = WidgetSpec{
	HCLKey:   "event_stream_definition",
	JSONType: "event_stream",
	Fields: []FieldSpec{
		{HCLKey: "query", Type: TypeString, OmitEmpty: false},
		{HCLKey: "event_size", Type: TypeString, OmitEmpty: true},
		{HCLKey: "tags_execution", Type: TypeString, OmitEmpty: true},
	},
}

// eventTimelineWidgetSpec corresponds to OpenAPI EventTimelineWidgetDefinition.
var eventTimelineWidgetSpec = WidgetSpec{
	HCLKey:   "event_timeline_definition",
	JSONType: "event_timeline",
	Fields: []FieldSpec{
		{HCLKey: "query", Type: TypeString, OmitEmpty: false},
		{HCLKey: "tags_execution", Type: TypeString, OmitEmpty: true},
	},
}

// checkStatusWidgetSpec corresponds to OpenAPI CheckStatusWidgetDefinition.
var checkStatusWidgetSpec = WidgetSpec{
	HCLKey:   "check_status_definition",
	JSONType: "check_status",
	Fields: []FieldSpec{
		{HCLKey: "check", Type: TypeString, OmitEmpty: false},
		{HCLKey: "grouping", Type: TypeString, OmitEmpty: false},
		{HCLKey: "group", Type: TypeString, OmitEmpty: true},
		{HCLKey: "group_by", Type: TypeStringList, OmitEmpty: true},
		{HCLKey: "tags", Type: TypeStringList, OmitEmpty: true},
	},
}

// logStreamWidgetSpec corresponds to OpenAPI LogStreamWidgetDefinition.
var logStreamWidgetSpec = WidgetSpec{
	HCLKey:   "log_stream_definition",
	JSONType: "log_stream",
	Fields: []FieldSpec{
		// Always emitted as [] when empty (cassette-verified)
		{HCLKey: "indexes", Type: TypeStringList, OmitEmpty: false},
		{HCLKey: "query", Type: TypeString, OmitEmpty: true},
		{HCLKey: "columns", Type: TypeStringList, OmitEmpty: false},
		{HCLKey: "show_date_column", Type: TypeBool, OmitEmpty: false},
		{HCLKey: "show_message_column", Type: TypeBool, OmitEmpty: false},
		{HCLKey: "message_display", Type: TypeString, OmitEmpty: true},
		{HCLKey: "sort", Type: TypeBlock, OmitEmpty: true, Children: []FieldSpec{
			{HCLKey: "column", Type: TypeString, OmitEmpty: false}, // required in OpenAPI
			{HCLKey: "order", Type: TypeString, OmitEmpty: false},  // required in OpenAPI
		}},
	},
}

// manageStatusWidgetSpec corresponds to OpenAPI MonitorSummaryWidgetDefinition.
var manageStatusWidgetSpec = WidgetSpec{
	HCLKey:   "manage_status_definition",
	JSONType: "manage_status",
	Fields: []FieldSpec{
		{HCLKey: "query", Type: TypeString, OmitEmpty: false},
		{HCLKey: "summary_type", Type: TypeString, OmitEmpty: true},
		{HCLKey: "sort", Type: TypeString, OmitEmpty: true},
		{HCLKey: "display_format", Type: TypeString, OmitEmpty: true},
		{HCLKey: "color_preference", Type: TypeString, OmitEmpty: true},
		// Emitted even when false (cassette-verified)
		{HCLKey: "hide_zero_counts", Type: TypeBool, OmitEmpty: false},
		{HCLKey: "show_last_triggered", Type: TypeBool, OmitEmpty: false},
		{HCLKey: "show_priority", Type: TypeBool, OmitEmpty: false},
	},
}

// runWorkflowWidgetSpec corresponds to OpenAPI RunWorkflowWidgetDefinition.
var runWorkflowWidgetSpec = WidgetSpec{
	HCLKey:   "run_workflow_definition",
	JSONType: "run_workflow",
	Fields: []FieldSpec{
		{HCLKey: "workflow_id", Type: TypeString, OmitEmpty: false},
		// HCL: "input" (singular) → JSON: "inputs" (plural)
		{HCLKey: "input", JSONKey: "inputs", Type: TypeBlockList, OmitEmpty: true,
			Children: []FieldSpec{
				{HCLKey: "name", Type: TypeString, OmitEmpty: false},  // required in OpenAPI
				{HCLKey: "value", Type: TypeString, OmitEmpty: false}, // required in OpenAPI
			}},
	},
}

// serviceMapWidgetSpec corresponds to OpenAPI ServiceMapWidgetDefinition.
// Note: both HCL key and JSON type use "servicemap" (no underscore).
var serviceMapWidgetSpec = WidgetSpec{
	HCLKey:   "servicemap_definition",
	JSONType: "servicemap",
	Fields: []FieldSpec{
		{HCLKey: "service", Type: TypeString, OmitEmpty: false},
		{HCLKey: "filters", Type: TypeStringList, OmitEmpty: false},
	},
}

// traceServiceWidgetSpec corresponds to OpenAPI ServiceSummaryWidgetDefinition.
var traceServiceWidgetSpec = WidgetSpec{
	HCLKey:   "trace_service_definition",
	JSONType: "trace_service",
	Fields: []FieldSpec{
		{HCLKey: "env", Type: TypeString, OmitEmpty: false},
		{HCLKey: "service", Type: TypeString, OmitEmpty: false},
		{HCLKey: "span_name", Type: TypeString, OmitEmpty: false},
		// Emitted even when false (cassette-verified)
		{HCLKey: "show_hits", Type: TypeBool, OmitEmpty: false},
		{HCLKey: "show_errors", Type: TypeBool, OmitEmpty: false},
		{HCLKey: "show_latency", Type: TypeBool, OmitEmpty: false},
		{HCLKey: "show_breakdown", Type: TypeBool, OmitEmpty: false},
		{HCLKey: "show_distribution", Type: TypeBool, OmitEmpty: false},
		{HCLKey: "show_resource_list", Type: TypeBool, OmitEmpty: false},
		{HCLKey: "size_format", Type: TypeString, OmitEmpty: true},
		{HCLKey: "display_format", Type: TypeString, OmitEmpty: true},
	},
}

var simpleWidgetSpecs = []WidgetSpec{
	alertGraphWidgetSpec,
	alertValueWidgetSpec,
	freeTextWidgetSpec,
	iframeWidgetSpec,
	imageWidgetSpec,
	noteWidgetSpec,
	eventStreamWidgetSpec,
	eventTimelineWidgetSpec,
	checkStatusWidgetSpec,
	logStreamWidgetSpec,
	manageStatusWidgetSpec,
	runWorkflowWidgetSpec,
	serviceMapWidgetSpec,
	traceServiceWidgetSpec,
}
