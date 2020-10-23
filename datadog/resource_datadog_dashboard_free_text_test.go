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
//    "title": "Free Text - Example",
//    "created_at": "2020-06-09T13:38:43.135928+00:00",
//    "modified_at": "2020-06-09T13:39:37.055724+00:00",
//    "author_handle": "--redacted--",
//    "widgets": [
//        {
//            "definition": {
//                "color": "#eb364b",
//                "text": "Free Text",
//                "type": "free_text",
//                "font_size": "56",
//                "text_align": "left"
//            },
//            "layout": {
//                "y": -2,
//                "x": 1,
//                "height": 6,
//                "width": 24
//            },
//            "id": 0
//        }
//    ],
//    "layout_type": "free"
//}

const datadogDashboardFreeTextConfig = `
resource "datadog_dashboard" "free_text_dashboard" {
	title         = "Acceptance Test Free Text Widget Dashboard"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "free"
	is_read_only  = "true"
	
	widget {
		free_text_definition {
			color = "#eb364b"
			text = "Free Text"
			font_size = "56"
			text_align = "left"
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

var datadogDashboardFreeTextAsserts = []string{
	"widget.0.layout.y = 5",
	"widget.0.free_text_definition.0.text = Free Text",
	"layout_type = free",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.free_text_definition.0.font_size = 56",
	"is_read_only = true",
	"widget.0.free_text_definition.0.color = #eb364b",
	"widget.0.layout.width = 32",
	"widget.0.layout.height = 43",
	"widget.0.free_text_definition.0.text_align = left",
	"title = Acceptance Test Free Text Widget Dashboard",
	"widget.0.layout.x = 5",
}

func TestAccDatadogDashboardFreeText(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardFreeTextConfig, "datadog_dashboard.free_text_dashboard", datadogDashboardFreeTextAsserts)
}

func TestAccDatadogDashboardFreeText_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtil_import(t, datadogDashboardFreeTextConfig, "datadog_dashboard.free_text_dashboard")
}
