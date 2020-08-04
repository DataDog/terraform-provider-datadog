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
//    "title": "TF - Alert Value Example",
//    "created_at": "2020-06-09T13:28:26.474869+00:00",
//    "modified_at": "2020-06-09T13:29:06.581646+00:00",
//    "author_handle": "--redacted--",
//    "widgets": [
//        {
//            "definition": {
//                "title_size": "16",
//                "title": "",
//                "title_align": "center",
//                "text_align": "center",
//                "precision": 1,
//                "alert_id": "10605849",
//                "type": "alert_value",
//                "unit": "b"
//            },
//            "layout": {
//                "y": 2,
//                "x": 8,
//                "height": 8,
//                "width": 15
//            },
//            "id": 0
//        }
//    ],
//    "layout_type": "free"
//}

const datadogDashboardAlertValueConfig = `
resource "datadog_dashboard" "alert_value_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = true
	widget {
		alert_value_definition {
			alert_id = "895605"
		}
	}
	widget {
		alert_value_definition {
			alert_id = "895606"
			precision = 1
			unit = "b"
			title_size = "16"
			title_align = "center"
			title = "Widget Title"
			text_align = "center"
		}
	}
}
`

var datadogDashboardAlertValueAsserts = []string{
	"widget.0.alert_value_definition.0.title_align =",
	"widget.1.alert_value_definition.0.title_align = center",
	"widget.1.alert_value_definition.0.text_align = center",
	"widget.1.layout.% = 0",
	"title = {{uniq}}",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.alert_value_definition.0.precision = 0",
	"widget.1.alert_value_definition.0.title_size = 16",
	"widget.1.alert_value_definition.0.precision = 1",
	"widget.0.alert_value_definition.0.title_size =",
	"widget.1.alert_value_definition.0.alert_id = 895606",
	"widget.0.alert_value_definition.0.text_align =",
	"layout_type = ordered",
	"widget.0.alert_value_definition.0.title =",
	"widget.0.alert_value_definition.0.unit =",
	"widget.1.alert_value_definition.0.title = Widget Title",
	"widget.0.alert_value_definition.0.alert_id = 895605",
	"widget.1.alert_value_definition.0.unit = b",
	"is_read_only = true",
}

func TestAccDatadogDashboardAlertValue(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardAlertValueConfig, "datadog_dashboard.alert_value_dashboard", datadogDashboardAlertValueAsserts)
}

func TestAccDatadogDashboardAlertValue_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtil_import(t, datadogDashboardAlertValueConfig, "datadog_dashboard.alert_value_dashboard")
}
