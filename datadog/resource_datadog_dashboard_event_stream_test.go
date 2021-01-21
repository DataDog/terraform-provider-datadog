package datadog

import (
	"testing"
)

// JSON export used as test scenario
//{
//    "notify_list": [],
//    "description": "",
//    "author_name": "--redacted--",
//    "id": "--redacted--",
//    "url": "--redacted--",
//    "template_variables": [],
//    "is_read_only": false,
//    "title": "TF - Event Strem Example",
//    "created_at": "2020-06-09T13:13:12.633530+00:00",
//    "modified_at": "2020-06-09T13:13:39.449243+00:00",
//    "author_handle": "--redacted--",
//    "widgets": [
//        {
//            "definition": {
//                "title_size": "16",
//                "title": "",
//                "title_align": "center",
//                "tags_execution": "and",
//                "time": {
//                    "live_span": "4h"
//                },
//                "query": "*",
//                "type": "event_stream",
//                "event_size": "l"
//            },
//            "layout": {
//                "y": 2,
//                "x": 0,
//                "height": 38,
//                "width": 47
//            },
//            "id": 0
//        }
//    ],
//    "layout_type": "free"
//}

const datadogDashboardEventStreamConfigDeprecated = `
resource "datadog_dashboard" "event_stream_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "free"
	is_read_only  = "true"
	
	widget {
		event_stream_definition {
			title = "Widget Title"
			title_align = "right"
			title_size = "16"
			tags_execution = "and"
			query = "*"
			event_size = "l"
			time = {
				live_span = "4h"
			}
		}
		layout = {
			height = 43
			width = 32
			x = 5
			y = 5
		}
	}
}
`

var datadogDashboardEventStreamAssertsDeprecated = []string{
	"widget.0.layout.x = 5",
	"widget.0.event_stream_definition.0.title_size = 16",
	"widget.0.event_stream_definition.0.tags_execution = and",
	"title = {{uniq}}",
	"widget.0.layout.y = 5",
	"widget.0.event_stream_definition.0.title_align = right",
	"widget.0.event_stream_definition.0.time.live_span = 4h",
	"widget.0.layout.width = 32",
	"widget.0.event_stream_definition.0.event_size = l",
	"layout_type = free",
	"description = Created using the Datadog provider in Terraform",
	"is_read_only = true",
	"widget.0.event_stream_definition.0.query = *",
	"widget.0.event_stream_definition.0.title = Widget Title",
	"widget.0.layout.height = 43",
}

func TestAccDatadogDashboardEventStreamDeprecated(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardEventStreamConfigDeprecated, "datadog_dashboard.event_stream_dashboard", datadogDashboardEventStreamAssertsDeprecated)
}

const datadogDashboardEventStreamConfig = `
resource "datadog_dashboard" "event_stream_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "free"
	is_read_only  = "true"
	
	widget {
		event_stream_definition {
			title = "Widget Title"
			title_align = "right"
			title_size = "16"
			tags_execution = "and"
			query = "*"
			event_size = "l"
			live_span = "4h"
		}
		widget_layout {
			height = 43
			width = 32
			x = 5
			y = 5
		}
	}
}
`

var datadogDashboardEventStreamAsserts = []string{
	"widget.0.widget_layout.0.x = 5",
	"widget.0.event_stream_definition.0.title_size = 16",
	"widget.0.event_stream_definition.0.tags_execution = and",
	"title = {{uniq}}",
	"widget.0.widget_layout.0.y = 5",
	"widget.0.event_stream_definition.0.title_align = right",
	"widget.0.event_stream_definition.0.live_span = 4h",
	"widget.0.widget_layout.0.width = 32",
	"widget.0.event_stream_definition.0.event_size = l",
	"layout_type = free",
	"description = Created using the Datadog provider in Terraform",
	"is_read_only = true",
	"widget.0.event_stream_definition.0.query = *",
	"widget.0.event_stream_definition.0.title = Widget Title",
	"widget.0.widget_layout.0.height = 43",
}

func TestAccDatadogDashboardEventStream(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardEventStreamConfig, "datadog_dashboard.event_stream_dashboard", datadogDashboardEventStreamAsserts)
}

func TestAccDatadogDashboardEventStream_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtil_import(t, datadogDashboardEventStreamConfig, "datadog_dashboard.event_stream_dashboard")
}
