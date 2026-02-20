package dashboardmapping

// widgets_simple.go — Batch A
//
// Widget types with no request blocks: alert_graph, alert_value, free_text, iframe,
// image, note, event_stream, event_timeline, check_status, log_stream,
// manage_status, run_workflow, service_map, trace_service.

// AlertGraphWidgetSpec corresponds to OpenAPI AlertGraphWidgetDefinition.
var AlertGraphWidgetSpec = WidgetSpec{
	HCLKey:      "alert_graph_definition",
	JSONType:    "alert_graph",
	Description: "The definition for a Alert Graph widget.",
	Fields: []FieldSpec{
		{
			HCLKey:      "alert_id",
			Type:        TypeString,
			OmitEmpty:   false,
			Required:    true,
			Description: "The ID of the monitor used by the widget.",
		},
		{
			HCLKey:      "viz_type",
			Type:        TypeString,
			OmitEmpty:   false,
			Required:    true,
			Description: "Type of visualization to use when displaying the widget.",
			ValidValues: []string{"timeseries", "toplist"},
		},
	},
}

// AlertValueWidgetSpec corresponds to OpenAPI AlertValueWidgetDefinition.
var AlertValueWidgetSpec = WidgetSpec{
	HCLKey:      "alert_value_definition",
	JSONType:    "alert_value",
	Description: "The definition for an Alert Value widget.",
	Fields: []FieldSpec{
		{
			HCLKey:      "alert_id",
			Type:        TypeString,
			OmitEmpty:   false,
			Required:    true,
			Description: "The ID of the monitor used by the widget.",
		},
		{
			HCLKey:      "precision",
			Type:        TypeInt,
			OmitEmpty:   true,
			Description: "The precision to use when displaying the value. Use `*` for maximum precision.",
		},
		{
			HCLKey:      "unit",
			Type:        TypeString,
			OmitEmpty:   true,
			Description: "The unit for the value displayed in the widget.",
		},
		{
			HCLKey:      "text_align",
			Type:        TypeString,
			OmitEmpty:   true,
			Description: "The alignment of the text in the widget.",
			ValidValues: []string{"center", "left", "right"},
		},
	},
}

// FreeTextWidgetSpec corresponds to OpenAPI FreeTextWidgetDefinition.
var FreeTextWidgetSpec = WidgetSpec{
	HCLKey:      "free_text_definition",
	JSONType:    "free_text",
	Description: "The definition for a Free Text widget.",
	Fields: []FieldSpec{
		{
			HCLKey:      "text",
			Type:        TypeString,
			OmitEmpty:   false,
			Required:    true,
			Description: "The text to display in the widget.",
		},
		{
			HCLKey:      "color",
			Type:        TypeString,
			OmitEmpty:   true,
			Description: "The color of the text in the widget.",
		},
		{
			HCLKey:      "font_size",
			Type:        TypeString,
			OmitEmpty:   true,
			Description: "The size of the text in the widget.",
		},
		{
			HCLKey:      "text_align",
			Type:        TypeString,
			OmitEmpty:   true,
			Description: "The alignment of the text in the widget.",
			ValidValues: []string{"center", "left", "right"},
		},
	},
}

// IFrameWidgetSpec corresponds to OpenAPI IFrameWidgetDefinition.
var IFrameWidgetSpec = WidgetSpec{
	HCLKey:      "iframe_definition",
	JSONType:    "iframe",
	Description: "The definition for an IFrame widget.",
	Fields: []FieldSpec{
		{
			HCLKey:      "url",
			Type:        TypeString,
			OmitEmpty:   false,
			Required:    true,
			Description: "The URL to use as a data source for the widget.",
		},
	},
}

// ImageWidgetSpec corresponds to OpenAPI ImageWidgetDefinition.
var ImageWidgetSpec = WidgetSpec{
	HCLKey:      "image_definition",
	JSONType:    "image",
	Description: "The definition for an Image widget.",
	Fields: []FieldSpec{
		{
			HCLKey:      "url",
			Type:        TypeString,
			OmitEmpty:   false,
			Required:    true,
			Description: "The URL to use as a data source for the widget.",
		},
		{
			HCLKey:      "url_dark_theme",
			Type:        TypeString,
			OmitEmpty:   true,
			Description: "The URL in dark mode to use as a data source for the widget.",
		},
		{
			HCLKey:      "sizing",
			Type:        TypeString,
			OmitEmpty:   true,
			Description: "The preferred method to adapt the dimensions of the image. The values are based on the image `object-fit` CSS properties. Note: `zoom`, `fit` and `center` values are deprecated.",
			ValidValues: []string{"fill", "contain", "cover", "none", "scale-down", "zoom", "fit", "center"},
		},
		{
			HCLKey:      "margin",
			Type:        TypeString,
			OmitEmpty:   true,
			Description: "The margins to use around the image. Note: `small` and `large` values are deprecated.",
			ValidValues: []string{"sm", "md", "lg", "small", "large"},
		},
		// Default: true in schema — always emitted
		{
			HCLKey:      "has_background",
			Type:        TypeBool,
			OmitEmpty:   false,
			Description: "Whether to display a background or not.",
		},
		{
			HCLKey:      "has_border",
			Type:        TypeBool,
			OmitEmpty:   false,
			Description: "Whether to display a border or not.",
		},
		{
			HCLKey:      "horizontal_align",
			Type:        TypeString,
			OmitEmpty:   true,
			Description: "The horizontal alignment for the widget.",
			ValidValues: []string{"center", "left", "right"},
		},
		{
			HCLKey:      "vertical_align",
			Type:        TypeString,
			OmitEmpty:   true,
			Description: "The vertical alignment for the widget.",
			ValidValues: []string{"center", "top", "bottom"},
		},
	},
}

// NoteWidgetSpec corresponds to OpenAPI NoteWidgetDefinition.
var NoteWidgetSpec = WidgetSpec{
	HCLKey:      "note_definition",
	JSONType:    "note",
	Description: "The definition for a Note widget.",
	Fields: []FieldSpec{
		{
			HCLKey:       "content",
			Type:         TypeString,
			OmitEmpty:    false,
			Required:     true,
			Description:  "The content of the note.",
			ValidateDiag: ValidateStringIsNotEmpty,
		},
		{
			HCLKey:      "background_color",
			Type:        TypeString,
			OmitEmpty:   true,
			Description: "The background color of the note.",
		},
		{
			HCLKey:      "font_size",
			Type:        TypeString,
			OmitEmpty:   true,
			Description: "The size of the text.",
		},
		{
			HCLKey:      "text_align",
			Type:        TypeString,
			OmitEmpty:   true,
			Description: "The alignment of the widget's text.",
			ValidValues: []string{"center", "left", "right"},
		},
		{
			HCLKey:      "vertical_align",
			Type:        TypeString,
			OmitEmpty:   true,
			Description: "The vertical alignment for the widget.",
			ValidValues: []string{"center", "top", "bottom"},
		},
		// Default: true — always emitted; old schema had Default:true so we preserve that
		{
			HCLKey:      "has_padding",
			Type:        TypeBool,
			OmitEmpty:   false,
			Default:     true,
			Description: "Whether to add padding or not.",
		},
		// Emitted even when false (cassette-verified)
		{
			HCLKey:      "show_tick",
			Type:        TypeBool,
			OmitEmpty:   false,
			Description: "Whether to show a tick or not.",
		},
		{
			HCLKey:      "tick_pos",
			Type:        TypeString,
			OmitEmpty:   true,
			Description: "When `tick = true`, a string with a percent sign indicating the position of the tick, for example: `tick_pos = \"50%\"` is centered alignment.",
		},
		{
			HCLKey:      "tick_edge",
			Type:        TypeString,
			OmitEmpty:   true,
			Description: "When `tick = true`, a string indicating on which side of the widget the tick should be displayed.",
			ValidValues: []string{"bottom", "left", "right", "top"},
		},
	},
}

// EventStreamWidgetSpec corresponds to OpenAPI EventStreamWidgetDefinition.
var EventStreamWidgetSpec = WidgetSpec{
	HCLKey:      "event_stream_definition",
	JSONType:    "event_stream",
	Description: "The definition for an Event Stream widget.",
	Fields: []FieldSpec{
		{
			HCLKey:      "query",
			Type:        TypeString,
			OmitEmpty:   false,
			Required:    true,
			Description: "The query to use in the widget.",
		},
		{
			HCLKey:      "event_size",
			Type:        TypeString,
			OmitEmpty:   true,
			Description: "The size to use to display an event.",
			ValidValues: []string{"s", "l"},
		},
		{
			HCLKey:      "tags_execution",
			Type:        TypeString,
			OmitEmpty:   true,
			Description: "The execution method for multi-value filters, options: `and` or `or`.",
		},
	},
}

// EventTimelineWidgetSpec corresponds to OpenAPI EventTimelineWidgetDefinition.
var EventTimelineWidgetSpec = WidgetSpec{
	HCLKey:      "event_timeline_definition",
	JSONType:    "event_timeline",
	Description: "The definition for an Event Timeline widget.",
	Fields: []FieldSpec{
		{
			HCLKey:      "query",
			Type:        TypeString,
			OmitEmpty:   false,
			Required:    true,
			Description: "The query to use in the widget.",
		},
		{
			HCLKey:      "tags_execution",
			Type:        TypeString,
			OmitEmpty:   true,
			Description: "The execution method for multi-value filters, options: `and` or `or`.",
		},
	},
}

// CheckStatusWidgetSpec corresponds to OpenAPI CheckStatusWidgetDefinition.
var CheckStatusWidgetSpec = WidgetSpec{
	HCLKey:      "check_status_definition",
	JSONType:    "check_status",
	Description: "The definition for a Check Status widget.",
	Fields: []FieldSpec{
		{
			HCLKey:      "check",
			Type:        TypeString,
			OmitEmpty:   false,
			Required:    true,
			Description: "The check to use in the widget.",
		},
		{
			HCLKey:      "grouping",
			Type:        TypeString,
			OmitEmpty:   false,
			Required:    true,
			Description: "The kind of grouping to use.",
			ValidValues: []string{"check", "cluster"},
		},
		{
			HCLKey:      "group",
			Type:        TypeString,
			OmitEmpty:   true,
			Description: "The check group to use in the widget.",
		},
		{
			HCLKey:      "group_by",
			Type:        TypeStringList,
			OmitEmpty:   true,
			Description: "When `grouping = \"cluster\"`, indicates a list of tags to use for grouping.",
		},
		{
			HCLKey:      "tags",
			Type:        TypeStringList,
			OmitEmpty:   true,
			Description: "A list of tags to use in the widget.",
		},
	},
}

// LogStreamWidgetSpec corresponds to OpenAPI LogStreamWidgetDefinition.
var LogStreamWidgetSpec = WidgetSpec{
	HCLKey:      "log_stream_definition",
	JSONType:    "log_stream",
	Description: "The definition for a Log Stream widget.",
	Fields: []FieldSpec{
		// Always emitted as [] when empty (cassette-verified)
		{
			HCLKey:      "indexes",
			Type:        TypeStringList,
			OmitEmpty:   false,
			Description: "An array of index names to query in the stream.",
		},
		{
			HCLKey:      "query",
			Type:        TypeString,
			OmitEmpty:   true,
			Description: "The query to use in the widget.",
		},
		{
			HCLKey:      "columns",
			Type:        TypeStringList,
			OmitEmpty:   false,
			Description: "Stringified list of columns to use, for example: `[\"column1\",\"column2\",\"column3\"]`.",
		},
		{
			HCLKey:      "show_date_column",
			Type:        TypeBool,
			OmitEmpty:   false,
			Description: "If the date column should be displayed.",
		},
		{
			HCLKey:      "show_message_column",
			Type:        TypeBool,
			OmitEmpty:   false,
			Description: "If the message column should be displayed.",
		},
		{
			HCLKey:      "message_display",
			Type:        TypeString,
			OmitEmpty:   true,
			Description: "The number of log lines to display.",
			ValidValues: []string{"inline", "expanded-md", "expanded-lg"},
		},
		{HCLKey: "sort", Type: TypeBlock, OmitEmpty: true, Children: []FieldSpec{
			{
				HCLKey:      "column",
				Type:        TypeString,
				OmitEmpty:   false,
				Required:    true,
				Description: "The facet path for the column.",
			}, // required in OpenAPI
			{
				HCLKey:      "order",
				Type:        TypeString,
				OmitEmpty:   false,
				Required:    true,
				Description: "Widget sorting methods.",
				ValidValues: []string{"asc", "desc"},
			}, // required in OpenAPI
		}},
	},
}

// ManageStatusWidgetSpec corresponds to OpenAPI MonitorSummaryWidgetDefinition.
var ManageStatusWidgetSpec = WidgetSpec{
	HCLKey:      "manage_status_definition",
	JSONType:    "manage_status",
	Description: "The definition for a Manage Status widget.",
	Fields: []FieldSpec{
		{
			HCLKey:      "query",
			Type:        TypeString,
			OmitEmpty:   false,
			Required:    true,
			Description: "The query to use in the widget.",
		},
		{
			HCLKey:      "summary_type",
			Type:        TypeString,
			OmitEmpty:   true,
			Description: "The summary type to use.",
			ValidValues: []string{"monitors", "groups", "combined"},
		},
		{
			HCLKey:      "sort",
			Type:        TypeString,
			OmitEmpty:   true,
			Description: "The method to sort the monitors.",
			ValidValues: []string{"name", "group", "status", "tags", "triggered", "group,asc", "group,desc", "name,asc", "name,desc", "status,asc", "status,desc", "tags,asc", "tags,desc", "triggered,asc", "triggered,desc", "priority,asc", "priority,desc"},
		},
		{
			HCLKey:      "display_format",
			Type:        TypeString,
			OmitEmpty:   true,
			Description: "The display setting to use.",
			ValidValues: []string{"counts", "countsAndList", "list"},
		},
		{
			HCLKey:      "color_preference",
			Type:        TypeString,
			OmitEmpty:   true,
			Description: "Whether to colorize text or background.",
			ValidValues: []string{"background", "text"},
		},
		// Emitted even when false (cassette-verified)
		{
			HCLKey:      "hide_zero_counts",
			Type:        TypeBool,
			OmitEmpty:   false,
			Description: "A Boolean indicating whether to hide empty categories.",
		},
		{
			HCLKey:      "show_last_triggered",
			Type:        TypeBool,
			OmitEmpty:   false,
			Description: "A Boolean indicating whether to show when monitors/groups last triggered.",
		},
		{
			HCLKey:      "show_priority",
			Type:        TypeBool,
			OmitEmpty:   false,
			Description: "Whether to show the priorities column.",
		},
	},
}

// RunWorkflowWidgetSpec corresponds to OpenAPI RunWorkflowWidgetDefinition.
var RunWorkflowWidgetSpec = WidgetSpec{
	HCLKey:      "run_workflow_definition",
	JSONType:    "run_workflow",
	Description: "The definition for a Run Workflow widget.",
	Fields: []FieldSpec{
		{
			HCLKey:      "workflow_id",
			Type:        TypeString,
			OmitEmpty:   false,
			Required:    true,
			Description: "Workflow ID",
		},
		// HCL: "input" (singular) → JSON: "inputs" (plural)
		{HCLKey: "input", JSONKey: "inputs", Type: TypeBlockList, OmitEmpty: true,
			Children: []FieldSpec{
				{
					HCLKey:      "name",
					Type:        TypeString,
					OmitEmpty:   false,
					Required:    true,
					Description: "Name of the workflow input.",
				}, // required in OpenAPI
				{
					HCLKey:      "value",
					Type:        TypeString,
					OmitEmpty:   false,
					Required:    true,
					Description: "Dashboard template variable. Can be suffixed with `.value` or `.key`.",
				}, // required in OpenAPI
			}},
	},
}

// ServiceMapWidgetSpec corresponds to OpenAPI ServiceMapWidgetDefinition.
// Note: both HCL key and JSON type use "servicemap" (no underscore).
var ServiceMapWidgetSpec = WidgetSpec{
	HCLKey:      "servicemap_definition",
	JSONType:    "servicemap",
	Description: "The definition for a Service Map widget.",
	Fields: []FieldSpec{
		{
			HCLKey:      "service",
			Type:        TypeString,
			OmitEmpty:   false,
			Required:    true,
			Description: "The ID of the service to map.",
		},
		{
			HCLKey:      "filters",
			Type:        TypeStringList,
			OmitEmpty:   false,
			Required:    true,
			Description: "Your environment and primary tag (or `*` if enabled for your account).",
		},
	},
}

// TraceServiceWidgetSpec corresponds to OpenAPI ServiceSummaryWidgetDefinition.
var TraceServiceWidgetSpec = WidgetSpec{
	HCLKey:      "trace_service_definition",
	JSONType:    "trace_service",
	Description: "The definition for a Trace Service (Service Summary) widget.",
	Fields: []FieldSpec{
		{
			HCLKey:      "env",
			Type:        TypeString,
			OmitEmpty:   false,
			Required:    true,
			Description: "APM environment.",
		},
		{
			HCLKey:      "service",
			Type:        TypeString,
			OmitEmpty:   false,
			Required:    true,
			Description: "APM service.",
		},
		{
			HCLKey:      "span_name",
			Type:        TypeString,
			OmitEmpty:   false,
			Required:    true,
			Description: "APM span name",
		},
		// Emitted even when false (cassette-verified)
		{
			HCLKey:      "show_hits",
			Type:        TypeBool,
			OmitEmpty:   false,
			Description: "Whether to show the hits metrics or not",
		},
		{
			HCLKey:      "show_errors",
			Type:        TypeBool,
			OmitEmpty:   false,
			Description: "Whether to show the error metrics or not.",
		},
		{
			HCLKey:      "show_latency",
			Type:        TypeBool,
			OmitEmpty:   false,
			Description: "Whether to show the latency metrics or not.",
		},
		{
			HCLKey:      "show_breakdown",
			Type:        TypeBool,
			OmitEmpty:   false,
			Description: "Whether to show the latency breakdown or not.",
		},
		{
			HCLKey:      "show_distribution",
			Type:        TypeBool,
			OmitEmpty:   false,
			Description: "Whether to show the latency distribution or not.",
		},
		{
			HCLKey:      "show_resource_list",
			Type:        TypeBool,
			OmitEmpty:   false,
			Description: "Whether to show the resource list or not.",
		},
		{
			HCLKey:      "size_format",
			Type:        TypeString,
			OmitEmpty:   true,
			Description: "The size of the widget.",
			ValidValues: []string{"small", "medium", "large"},
		},
		{
			HCLKey:      "display_format",
			Type:        TypeString,
			OmitEmpty:   true,
			Description: "The number of columns to display.",
			ValidValues: []string{"one_column", "two_column", "three_column"},
		},
	},
}

var simpleWidgetSpecs = []WidgetSpec{
	AlertGraphWidgetSpec,
	AlertValueWidgetSpec,
	FreeTextWidgetSpec,
	IFrameWidgetSpec,
	ImageWidgetSpec,
	NoteWidgetSpec,
	EventStreamWidgetSpec,
	EventTimelineWidgetSpec,
	CheckStatusWidgetSpec,
	LogStreamWidgetSpec,
	ManageStatusWidgetSpec,
	RunWorkflowWidgetSpec,
	ServiceMapWidgetSpec,
	TraceServiceWidgetSpec,
}
