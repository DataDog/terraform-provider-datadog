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
//    "title": "TF - Alert Graph",
//    "created_at": "2020-06-09T13:26:19.396570+00:00",
//    "modified_at": "2020-06-09T13:27:20.634671+00:00",
//    "author_handle": "--redacted--",
//    "widgets": [
//        {
//            "definition": {
//                "title_size": "16",
//                "viz_type": "toplist",
//                "title": "",
//                "title_align": "center",
//                "alert_id": "10605849",
//                "type": "alert_graph",
//                "time": {
//                    "live_span": "10m"
//                }
//            },
//            "layout": {
//                "y": 1,
//                "x": 2,
//                "height": 15,
//                "width": 47
//            },
//            "id": 0
//        }
//    ],
//    "layout_type": "free"
//}

const datadogDashboardAlertGraphConfig = `
resource "datadog_dashboard" "alert_graph_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = true
	widget {
		alert_graph_definition {
			alert_id = "895605"
			viz_type = "timeseries"
		}
	}
	widget {
		alert_graph_definition {
			alert_id = "895606"
			viz_type = "toplist"
			title = "Widget Title"
			title_align = "right"
			title_size = "16"
			live_span = "1h"
		}
	}
}
`

const datadogDashboardAlertGraphConfigDeprecated = `
resource "datadog_dashboard" "alert_graph_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = true
	widget {
		alert_graph_definition {
			alert_id = "895605"
			viz_type = "timeseries"
		}
	}
	widget {
		alert_graph_definition {
			alert_id = "895606"
			viz_type = "toplist"
			title = "Widget Title"
			title_align = "right"
			title_size = "16"
			time = {
				live_span = "1h"
			}
		}
	}
}
`

var datadogDashboardAlertGraphAsserts = []string{
	"title = {{uniq}}",
	"widget.0.alert_graph_definition.0.alert_id = 895605",
	"widget.1.alert_graph_definition.0.time.% = 0",
	"widget.1.alert_graph_definition.0.title = Widget Title",
	"is_read_only = true",
	"widget.1.alert_graph_definition.0.title_size = 16",
	"widget.1.alert_graph_definition.0.viz_type = toplist",
	"widget.1.alert_graph_definition.0.live_span = 1h",
	"widget.1.alert_graph_definition.0.alert_id = 895606",
	"widget.0.alert_graph_definition.0.title_size =",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.alert_graph_definition.0.title_align =",
	"widget.0.alert_graph_definition.0.title =",
	"widget.1.alert_graph_definition.0.title_align = right",
	"layout_type = ordered",
	"widget.0.alert_graph_definition.0.viz_type = timeseries",
}

var datadogDashboardAlertGraphAssertsDeprecated = []string{
	"title = {{uniq}}",
	"widget.0.alert_graph_definition.0.alert_id = 895605",
	"widget.1.alert_graph_definition.0.time.% = 1",
	"widget.1.alert_graph_definition.0.title = Widget Title",
	"is_read_only = true",
	"widget.1.alert_graph_definition.0.title_size = 16",
	"widget.1.alert_graph_definition.0.viz_type = toplist",
	"widget.1.alert_graph_definition.0.time.live_span = 1h",
	"widget.1.alert_graph_definition.0.alert_id = 895606",
	"widget.0.alert_graph_definition.0.title_size =",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.alert_graph_definition.0.title_align =",
	"widget.0.alert_graph_definition.0.title =",
	"widget.1.alert_graph_definition.0.title_align = right",
	"layout_type = ordered",
	"widget.0.alert_graph_definition.0.viz_type = timeseries",
}

func TestAccDatadogDashboardAlertGraph(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardAlertGraphConfig, "datadog_dashboard.alert_graph_dashboard", datadogDashboardAlertGraphAsserts)
}

func TestAccDatadogDashboardAlertGraphDeprecated(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardAlertGraphConfigDeprecated, "datadog_dashboard.alert_graph_dashboard", datadogDashboardAlertGraphAssertsDeprecated)
}

func TestAccDatadogDashboardAlertGraph_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtil_import(t, datadogDashboardAlertGraphConfig, "datadog_dashboard.alert_graph_dashboard")
}
