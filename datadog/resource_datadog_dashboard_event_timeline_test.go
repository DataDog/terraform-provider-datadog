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
//    "title": "TF - Event Stream Example",
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

const datadogDashboardEventTimelineConfig = `
resource "datadog_dashboard" "event_timeline_dashboard" {
	title         = "Acceptance Test Event Timeline Widget Dashboard"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "free"
	is_read_only  = "true"
	
	widget {
		event_timeline_definition {
			title = "Widget Title"
			title_align = "right"
			title_size = "16"
			tags_execution = "and"
			query = "status:error"
			time = {
				live_span = "1h"
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

var datadogDashboardEventTimelineAsserts = []string{
	"widget.0.layout.y = 5",
	"widget.0.event_timeline_definition.0.title_align = right",
	"widget.0.layout.x = 5",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.event_timeline_definition.0.time.live_span = 1h",
	"title = Acceptance Test Event Timeline Widget Dashboard",
	"is_read_only = true",
	"widget.0.layout.width = 32",
	"widget.0.event_timeline_definition.0.title_size = 16",
	"layout_type = free",
	"widget.0.event_timeline_definition.0.query = status:error",
	"widget.0.event_timeline_definition.0.title = Widget Title",
	"widget.0.event_timeline_definition.0.tags_execution = and",
	"widget.0.layout.height = 43",
}

func TestAccDatadogDashboardEventTimeline(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardEventTimelineConfig, "datadog_dashboard.event_timeline_dashboard", datadogDashboardEventTimelineAsserts)
}

func TestAccDatadogDashboardEventTimeline_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtil_import(t, datadogDashboardEventTimelineConfig, "datadog_dashboard.event_timeline_dashboard")
}
