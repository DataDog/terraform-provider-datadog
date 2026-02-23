package dashboardmapping

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// widgets.go
//
// All WidgetSpec declarations and the allWidgetSpecs registry.
//
// Each WidgetSpec maps an HCL definition block key to a JSON widget type string
// and declares the widget-specific FieldSpecs (field groups live in field_groups.go).
//
// Sections:
//   - Common Widget Fields (var CommonWidgetFields is in field_groups.go)
//   - Simple Widgets (no request blocks)
//   - Request Widgets (standard log/apm/formula query requests)
//   - Complex Widgets (custom post-processing required)
//   - Widget Registry

// ============================================================
// Simple Widgets (no request blocks)
// ============================================================

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
	Description: "The definition for a Alert Value widget.",
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
	Description: "The definition for an Iframe widget.",
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
		// Default: true in schema — always emitted (old schema had Default:true)
		{
			HCLKey:      "has_background",
			Type:        TypeBool,
			OmitEmpty:   false,
			Default:     true,
			Description: "Whether to display a background or not.",
		},
		{
			HCLKey:      "has_border",
			Type:        TypeBool,
			OmitEmpty:   false,
			Default:     true,
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
			HCLKey:      "content",
			Type:        TypeString,
			OmitEmpty:   false,
			Required:    true,
			Description: "The content of the note.",
			// Use Validate (not ValidateDiag) so the error message includes the
			// full attribute path, matching the existing test assertion.
			Validate: validation.StringIsNotEmpty,
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
	Description: "The definition for a Event Stream widget.",
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
	Description: "The definition for a Event Timeline widget.",
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
	Description: "The definition for an Log Stream widget.",
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
		{HCLKey: "sort", Type: TypeBlock, OmitEmpty: true,
			Description: "The facet and order to sort the data, for example: `{\"column\": \"time\", \"order\": \"desc\"}`.",
			Children: []FieldSpec{
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
	Description: "The definition for an Manage Status widget.",
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
			Description: "Array of workflow inputs to map to dashboard template variables.",
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
		widgetCustomLinkField,
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
		widgetCustomLinkField,
	},
}

// TraceServiceWidgetSpec corresponds to OpenAPI ServiceSummaryWidgetDefinition.
var TraceServiceWidgetSpec = WidgetSpec{
	HCLKey:      "trace_service_definition",
	JSONType:    "trace_service",
	Description: "The definition for a Trace Service widget.",
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

// ============================================================
// Request Widgets (standard log/apm/formula query requests)
// ============================================================

// TimeseriesWidgetSpec corresponds to OpenAPI
// components/schemas/TimeseriesWidgetDefinition.
var TimeseriesWidgetSpec = WidgetSpec{
	HCLKey:      "timeseries_definition",
	JSONType:    "timeseries",
	Description: "The definition for a Timeseries widget.",
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
		widgetCustomLinkField,
	},
}

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
	HCLKey:      "change_definition",
	JSONType:    "change",
	Description: "The definition for a Change widget.",
	Fields: []FieldSpec{
		{HCLKey: "request", JSONKey: "requests", Type: TypeBlockList, OmitEmpty: false,
			Description: "A nested block describing the request to use when displaying the widget. Multiple request blocks are allowed using the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query` or `process_query` is required within the request block).",
			Children:    changeWidgetRequestFields},
		widgetCustomLinkField,
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
	HCLKey:      "distribution_definition",
	JSONType:    "distribution",
	Description: "The definition for a Distribution widget.",
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
	HCLKey:      "heatmap_definition",
	JSONType:    "heatmap",
	Description: "The definition for a Heatmap widget.",
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
		widgetCustomLinkField,
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
	HCLKey:      "hostmap_definition",
	JSONType:    "hostmap",
	Description: "The definition for a Hostmap widget.",
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
		widgetCustomLinkField,
	},
}

// QueryValueWidgetSpec corresponds to OpenAPI QueryValueWidgetDefinition.
var queryValueRequestFields = append([]FieldSpec{
	{HCLKey: "q", Type: TypeString, OmitEmpty: true,
		Description: "The metric query to use for this widget."},
	{HCLKey: "audit_query", Type: TypeBlock, OmitEmpty: true,
		Description: "The query to use for this widget.",
		Children:    logQueryDefinitionFields},
	{HCLKey: "aggregator", Type: TypeString, OmitEmpty: true,
		Description: "The aggregator to use for time aggregation.",
		ValidValues: []string{"avg", "min", "max", "sum", "last", "area", "l2norm", "percentile"}},
	{HCLKey: "conditional_formats", Type: TypeBlockList, OmitEmpty: true,
		Description: "Conditional formats allow you to set the color of your widget content or background depending on the rule applied to your data. Multiple `conditional_formats` blocks are allowed using the structure below.",
		Children:    widgetConditionalFormatFields},
}, standardQueryFields...)

var QueryValueWidgetSpec = WidgetSpec{
	HCLKey:      "query_value_definition",
	JSONType:    "query_value",
	Description: "The definition for a Query Value widget.",
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
		widgetCustomLinkField,
	},
}

// ToplistWidgetSpec corresponds to OpenAPI ToplistWidgetDefinition.
var toplistWidgetRequestFields = append([]FieldSpec{
	{HCLKey: "q", Type: TypeString, OmitEmpty: true,
		Description: "The metric query to use for this widget."},
	{HCLKey: "audit_query", Type: TypeBlock, OmitEmpty: true,
		Description: "The query to use for this widget.",
		Children:    logQueryDefinitionFields},
	{HCLKey: "conditional_formats", Type: TypeBlockList, OmitEmpty: true,
		Description: "Conditional formats allow you to set the color of your widget content or background, depending on a rule applied to your data. Multiple `conditional_formats` blocks are allowed using the structure below.",
		Children:    widgetConditionalFormatFields},
	{HCLKey: "style", Type: TypeBlock, OmitEmpty: true,
		Description: "Define request for the widget's style.",
		Children:    widgetRequestStyleFields},
}, standardQueryFields...)

var ToplistWidgetSpec = WidgetSpec{
	HCLKey:      "toplist_definition",
	JSONType:    "toplist",
	Description: "The definition for a Toplist widget.",
	Fields: []FieldSpec{
		{HCLKey: "style", Type: TypeBlock, OmitEmpty: true,
			Description: "The style of the widget",
			Children:    toplistWidgetStyleFields},
		{HCLKey: "request", JSONKey: "requests", Type: TypeBlockList, OmitEmpty: false,
			Description: "A nested block describing the request to use when displaying the widget. Multiple `request` blocks are allowed using the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query` or `process_query` is required within the `request` block).",
			Children:    toplistWidgetRequestFields},
		widgetCustomLinkField,
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
	// SchemaOnly: JSON building is handled by buildScatterplotTableJSON (injected as requests.table),
	// not by the FieldSpec engine (which would incorrectly emit it as requests.scatterplot_table).
	{HCLKey: "scatterplot_table", Type: TypeBlockList, OmitEmpty: true, SchemaOnly: true,
		Description: "Scatterplot request containing formulas and functions.",
		Children:    scatterplotTableRequestFields},
}

var ScatterplotWidgetSpec = WidgetSpec{
	HCLKey:      "scatterplot_definition",
	JSONType:    "scatterplot",
	Description: "The definition for a Scatterplot widget.",
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
		widgetCustomLinkField,
	},
}

// SunburstWidgetSpec corresponds to OpenAPI SunburstWidgetDefinition.
// The JSON "legend" field is polymorphic; HCL uses separate legend_inline and legend_table blocks.
var sunburstWidgetRequestFields = append([]FieldSpec{
	{HCLKey: "q", Type: TypeString, OmitEmpty: true,
		Description: "The metric query to use for this widget."},
	{HCLKey: "network_query", Type: TypeBlock, OmitEmpty: true,
		Description: "The query to use for this widget.",
		Children:    logQueryDefinitionFields},
	{HCLKey: "audit_query", Type: TypeBlock, OmitEmpty: true,
		Description: "The query to use for this widget.",
		Children:    logQueryDefinitionFields},
	{HCLKey: "style", Type: TypeBlock, OmitEmpty: true,
		Description: "Define style for the widget's request.",
		Children:    widgetRequestStyleFields},
}, standardQueryFields...)

var SunburstWidgetSpec = WidgetSpec{
	HCLKey:      "sunburst_definition",
	JSONType:    "sunburst",
	Description: "The definition for a Sunburst widget.",
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
		widgetCustomLinkField,
	},
}

// GeomapWidgetSpec corresponds to OpenAPI GeomapWidgetDefinition.
var geomapWidgetRequestFields = []FieldSpec{
	{HCLKey: "q", Type: TypeString, OmitEmpty: true,
		Description: "The metric query to use for this widget."},
	{HCLKey: "log_query", Type: TypeBlock, OmitEmpty: true,
		Description: "The query to use for this widget.",
		Children:    logQueryDefinitionFields},
	{HCLKey: "rum_query", Type: TypeBlock, OmitEmpty: true,
		Description: "The query to use for this widget.",
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
	HCLKey:      "geomap_definition",
	JSONType:    "geomap",
	Description: "The definition for a Geomap widget.",
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
		widgetCustomLinkField,
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
//
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
	HCLKey:      "treemap_definition",
	JSONType:    "treemap",
	Description: "The definition for a Treemap widget.",
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
	HCLKey:      "topology_map_definition",
	JSONType:    "topology_map",
	Description: "The definition for a Topology Map widget.",
	Fields: []FieldSpec{
		{HCLKey: "request", JSONKey: "requests", Type: TypeBlockList, OmitEmpty: false,
			Description: "A nested block describing the request to use when displaying the widget. Multiple request blocks are allowed using the structure below (`query` and `request_type` are required within the request).",
			Children:    topologyRequestFields},
		widgetCustomLinkField,
	},
}

// ============================================================
// Complex Widgets (custom post-processing required)
// ============================================================

// QueryTableWidgetSpec corresponds to OpenAPI
// components/schemas/TableWidgetDefinition.
// Formula requests are handled by post-processing (buildQueryTableFormulaRequestsJSON).
var QueryTableWidgetSpec = WidgetSpec{
	HCLKey:      "query_table_definition",
	JSONType:    "query_table",
	Description: "The definition for a Query Table widget.",
	Fields: []FieldSpec{
		// has_search_bar: OmitEmpty — only present when explicitly set
		{
			HCLKey:      "has_search_bar",
			Type:        TypeString,
			OmitEmpty:   true,
			Description: "Controls the display of the search bar.",
			ValidValues: []string{"always", "never", "auto"},
		},
		// request: HCL singular "request" → JSON plural "requests"
		// OmitEmpty: false — always emit even if empty
		{
			HCLKey:      "request",
			JSONKey:     "requests",
			Type:        TypeBlockList,
			OmitEmpty:   false,
			Description: "A nested block describing the request to use when displaying the widget. Multiple `request` blocks are allowed using the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query`, `apm_stats_query` or `process_query` is required within the `request` block).",
			Children:    queryTableOldRequestFields,
		},
		widgetCustomLinkField,
	},
}

// ListStreamWidgetSpec corresponds to OpenAPI
// components/schemas/ListStreamWidgetDefinition.
var ListStreamWidgetSpec = WidgetSpec{
	HCLKey:      "list_stream_definition",
	JSONType:    "list_stream",
	Description: "The definition for a List Stream widget.",
	Fields: []FieldSpec{
		// request: HCL singular "request" → JSON plural "requests"
		{
			HCLKey:      "request",
			JSONKey:     "requests",
			Type:        TypeBlockList,
			OmitEmpty:   false,
			Required:    true,
			Description: "Nested block describing the requests to use when displaying the widget. Multiple `request` blocks are allowed with the structure below.",
			Children:    listStreamRequestFields,
		},
	},
}

// SLOWidgetSpec corresponds to OpenAPI
// components/schemas/SLOWidgetDefinition.
var SLOWidgetSpec = WidgetSpec{
	HCLKey:      "service_level_objective_definition",
	JSONType:    "slo",
	Description: "The definition for a Service Level Objective widget.",
	Fields: []FieldSpec{
		// Required fields
		{
			HCLKey:      "slo_id",
			Type:        TypeString,
			OmitEmpty:   false,
			Required:    true,
			Description: "The ID of the service level objective used by the widget.",
		},
		{
			HCLKey:      "view_type",
			Type:        TypeString,
			OmitEmpty:   false,
			Required:    true,
			Description: "The type of view to use when displaying the widget. Only `detail` is supported.",
		},
		{
			HCLKey:      "view_mode",
			Type:        TypeString,
			OmitEmpty:   false,
			Required:    true,
			Description: "The view mode for the widget.",
			ValidValues: []string{"overall", "component", "both"},
		},
		{
			HCLKey:      "time_windows",
			Type:        TypeStringList,
			OmitEmpty:   false,
			Required:    true,
			Description: "A list of time windows to display in the widget.",
			ValidValues: []string{"7d", "30d", "90d", "week_to_date", "previous_week", "month_to_date", "previous_month", "global_time"},
		},
		// Optional fields
		{HCLKey: "show_error_budget", Type: TypeBool, OmitEmpty: true, Description: "Whether to show the error budget or not."},
		{HCLKey: "global_time_target", Type: TypeString, OmitEmpty: true, Description: "The global time target of the widget."},
		{HCLKey: "additional_query_filters", Type: TypeString, OmitEmpty: true, Description: "Additional filters applied to the SLO query."},
	},
}

// SLOListWidgetSpec corresponds to OpenAPI
// components/schemas/SLOListWidgetDefinition.
var SLOListWidgetSpec = WidgetSpec{
	HCLKey:      "slo_list_definition",
	JSONType:    "slo_list",
	Description: "The definition for an SLO (Service Level Objective) List widget.",
	Fields: []FieldSpec{
		// request: HCL singular "request" → JSON plural "requests"
		{
			HCLKey:      "request",
			JSONKey:     "requests",
			Type:        TypeBlockList,
			OmitEmpty:   false,
			Required:    true,
			Description: "A nested block describing the request to use when displaying the widget. Exactly one `request` block is allowed.",
			Children:    sloListRequestFields,
		},
	},
}

// SplitGraphWidgetSpec corresponds to OpenAPI
// components/schemas/SplitGraphWidgetDefinition.
// The JSON type is "split_group" (not "split_graph").
// source_widget_definition is handled by post-processing (buildSplitGraphSourceWidgetJSON).
var SplitGraphWidgetSpec = WidgetSpec{
	HCLKey:      "split_graph_definition",
	JSONType:    "split_group",
	Description: "The definition for a Split Graph widget.",
	Fields: []FieldSpec{
		{
			HCLKey:      "size",
			Type:        TypeString,
			OmitEmpty:   false,
			Required:    true,
			Description: "Size of the individual graphs in the split.",
		},
		// has_uniform_y_axes: always emitted (OmitEmpty: false) — cassette shows false when not set
		{HCLKey: "has_uniform_y_axes", Type: TypeBool, OmitEmpty: false, Description: "Normalize y axes across graphs."},
		// split_config: TypeBlock (MaxItems:1, Required)
		{
			HCLKey:      "split_config",
			Type:        TypeBlock,
			OmitEmpty:   false,
			Required:    true,
			Description: "Encapsulates all user choices about how to split a graph.",
			Children:    splitConfigFields,
		},
		// source_widget_definition: handled by post-processing
		// title is in CommonWidgetFields
	},
}

// GroupWidgetSpec corresponds to OpenAPI
// components/schemas/GroupWidgetDefinition.
// The nested "widget" list is handled by post-processing (buildGroupWidgetsJSON).
var GroupWidgetSpec = WidgetSpec{
	HCLKey:      "group_definition",
	JSONType:    "group",
	Description: "The definition for a Group widget.",
	Fields: []FieldSpec{
		{
			HCLKey:      "layout_type",
			Type:        TypeString,
			OmitEmpty:   false,
			Required:    true,
			Description: "The layout type of the group.",
			ValidValues: []string{"ordered"},
		},
		{HCLKey: "background_color", Type: TypeString, OmitEmpty: true, Description: "The background color of the group title, options: `vivid_blue`, `vivid_purple`, `vivid_pink`, `vivid_orange`, `vivid_yellow`, `vivid_green`, `blue`, `purple`, `pink`, `orange`, `yellow`, `green`, `gray` or `white`"},
		{HCLKey: "banner_img", Type: TypeString, OmitEmpty: true, Description: "The image URL to display as a banner for the group."},
		// Default: true — preserved from original schema (show_title defaults to visible)
		{HCLKey: "show_title", Type: TypeBool, OmitEmpty: false, Default: true, Description: "Whether to show the title or not."},
		// "widget" is handled by post-processing
		// title is in CommonWidgetFields
	},
}

// PowerpackWidgetSpec corresponds to OpenAPI
// components/schemas/PowerpackWidgetDefinition.
var PowerpackWidgetSpec = WidgetSpec{
	HCLKey:      "powerpack_definition",
	JSONType:    "powerpack",
	Description: "The definition for a Powerpack widget.",
	Fields: []FieldSpec{
		{HCLKey: "powerpack_id", Type: TypeString, OmitEmpty: false, Required: true, Description: "UUID of the associated powerpack."},
		{HCLKey: "background_color", Type: TypeString, OmitEmpty: true, Description: "The background color of the powerpack title."},
		{HCLKey: "banner_img", Type: TypeString, OmitEmpty: true, Description: "URL of image to display as a banner for the powerpack."},
		{HCLKey: "show_title", Type: TypeBool, OmitEmpty: true, Description: "Whether to show the title of the powerpack."},
		// template_variables: TypeBlock (MaxItems:1) containing two TypeBlockLists
		{
			HCLKey:      "template_variables",
			Type:        TypeBlock,
			OmitEmpty:   true,
			Description: "The list of template variables for this powerpack.",
			Children:    powerpackTemplateVariableFields,
		},
		// title is in CommonWidgetFields
	},
}

// ============================================================
// Widget Registry
// ============================================================

// allWidgetSpecs is the complete ordered registry of all implemented widget types.
// The engine iterates this slice to dispatch build/flatten for each widget.
var allWidgetSpecs = []WidgetSpec{
	// Simple widgets (no request blocks)
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
	// Request widgets (standard log/apm/formula query requests)
	TimeseriesWidgetSpec,
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
	// Complex widgets (custom post-processing required)
	QueryTableWidgetSpec,
	ListStreamWidgetSpec,
	SLOWidgetSpec,
	SLOListWidgetSpec,
	SplitGraphWidgetSpec,
	GroupWidgetSpec,
	PowerpackWidgetSpec,
}

// concatWidgetSpecs merges multiple WidgetSpec slices into one.
// Retained for backward compatibility with any external callers.
func concatWidgetSpecs(slices ...[]WidgetSpec) []WidgetSpec {
	var result []WidgetSpec
	for _, s := range slices {
		result = append(result, s...)
	}
	return result
}

// widgetLayoutFieldSpecs are the fields inside the widget_layout block.
var widgetLayoutFieldSpecs = []FieldSpec{
	{HCLKey: "x", Type: TypeInt, Required: true, Description: "The position of the widget on the x (horizontal) axis. Must be greater than or equal to 0."},
	{HCLKey: "y", Type: TypeInt, Required: true, Description: "The position of the widget on the y (vertical) axis. Must be greater than or equal to 0."},
	{HCLKey: "width", Type: TypeInt, Required: true, Description: "The width of the widget."},
	{HCLKey: "height", Type: TypeInt, Required: true, Description: "The height of the widget."},
	{HCLKey: "is_column_break", Type: TypeBool, OmitEmpty: true, Description: "Whether the widget should be the first one on the second column in high density or not. Only one widget in the dashboard should have this property set to `true`."},
}

// splitGraphSourceWidgetSchema builds the *schema.Schema for the source_widget_definition
// block inside split_graph_definition. It includes all widget types except group, powerpack,
// and split_group (which cannot be source widgets).
func splitGraphSourceWidgetSchema() *schema.Schema {
	inner := make(map[string]*schema.Schema)
	for _, spec := range allWidgetSpecs {
		if spec.JSONType == "group" || spec.JSONType == "powerpack" || spec.JSONType == "split_group" {
			continue
		}
		inner[spec.HCLKey] = WidgetSpecToSchemaBlock(spec)
	}
	return &schema.Schema{
		Type:        schema.TypeList,
		Required:    true,
		MaxItems:    1,
		Description: "The original widget we are splitting on.",
		Elem:        &schema.Resource{Schema: inner},
	}
}

// AllWidgetSchemasMap returns the schema map for all widget definition types,
// including widget_layout and id wrapper fields. If excludePowerpackOnly is true,
// powerpack and split_graph definitions are excluded (for use by the powerpack resource).
func AllWidgetSchemasMap(excludePowerpackOnly bool) map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The ID of the widget.",
		},
		"widget_layout": {
			Type:        schema.TypeList,
			MaxItems:    1,
			Optional:    true,
			Description: "The layout of the widget on a 'free' dashboard.",
			Elem:        &schema.Resource{Schema: FieldSpecsToSchema(widgetLayoutFieldSpecs)},
		},
	}
	for _, spec := range allWidgetSpecs {
		if excludePowerpackOnly && (spec.JSONType == "powerpack" || spec.JSONType == "split_group") {
			continue
		}
		s[spec.HCLKey] = WidgetSpecToSchemaBlock(spec)
	}
	// Inject source_widget_definition into split_graph_definition schema.
	// This block is dynamically generated from allWidgetSpecs (excluding group/powerpack/split_group)
	// and cannot be expressed as a static FieldSpec.
	if splitGraphSchema, ok := s["split_graph_definition"]; ok {
		splitGraphSchema.Elem.(*schema.Resource).Schema["source_widget_definition"] = splitGraphSourceWidgetSchema()
	}
	return s
}

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
	{HCLKey: "q", Type: TypeString, OmitEmpty: true, Description: "The metric query to use for this widget."},
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
	{HCLKey: "log_query", Type: TypeBlock, OmitEmpty: true, Description: "The query to use for this widget.", Children: logQueryDefinitionFields},
	{HCLKey: "apm_query", Type: TypeBlock, OmitEmpty: true, Description: "The query to use for this widget.", Children: logQueryDefinitionFields},
	{HCLKey: "rum_query", Type: TypeBlock, OmitEmpty: true, Description: "The query to use for this widget.", Children: logQueryDefinitionFields},
	{HCLKey: "network_query", Type: TypeBlock, OmitEmpty: true, Description: "The query to use for this widget.", Children: logQueryDefinitionFields},
	{HCLKey: "security_query", Type: TypeBlock, OmitEmpty: true, Description: "The query to use for this widget.", Children: logQueryDefinitionFields},
	{HCLKey: "audit_query", Type: TypeBlock, OmitEmpty: true, Description: "The query to use for this widget.", Children: logQueryDefinitionFields},
	{HCLKey: "profile_metrics_query", Type: TypeBlock, OmitEmpty: true, Description: "The query to use for this widget.", Children: logQueryDefinitionFields},
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
