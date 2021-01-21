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
//    "title": "TF - IFrame Example",
//    "created_at": "2020-06-09T13:36:50.905652+00:00",
//    "modified_at": "2020-06-09T13:37:07.261954+00:00",
//    "author_handle": "--redacted--",
//    "widgets": [
//        {
//            "definition": {
//                "url": "https://en.wikipedia.org/wiki/Datadog",
//                "type": "iframe"
//            },
//            "layout": {
//                "y": 2,
//                "x": 18,
//                "height": 12,
//                "width": 12
//            },
//            "id": 0
//        }
//    ],
//    "layout_type": "free"
//}

const datadogDashboardIFrameConfigDeprecated = `
resource "datadog_dashboard" "iframe_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "free"
	is_read_only  = "true"

	widget {
		iframe_definition {
			url = "https://en.wikipedia.org/wiki/Datadog"
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

var datadogDashboardIFrameAssertsDeprecated = []string{
	"description = Created using the Datadog provider in Terraform",
	"is_read_only = true",
	"widget.0.iframe_definition.0.url = https://en.wikipedia.org/wiki/Datadog",
	"widget.0.layout.height = 43",
	"title = {{uniq}}",
	"widget.0.layout.x = 5",
	"widget.0.layout.y = 5",
	"layout_type = free",
	"widget.0.layout.width = 32",
}

func TestAccDatadogDashboardIFrameDeprecated(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardIFrameConfigDeprecated, "datadog_dashboard.iframe_dashboard", datadogDashboardIFrameAssertsDeprecated)
}

const datadogDashboardIFrameConfig = `
resource "datadog_dashboard" "iframe_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "free"
	is_read_only  = "true"

	widget {
		iframe_definition {
			url = "https://en.wikipedia.org/wiki/Datadog"
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

var datadogDashboardIFrameAsserts = []string{
	"description = Created using the Datadog provider in Terraform",
	"is_read_only = true",
	"widget.0.iframe_definition.0.url = https://en.wikipedia.org/wiki/Datadog",
	"widget.0.widget_layout.0.height = 43",
	"title = {{uniq}}",
	"widget.0.widget_layout.0.x = 5",
	"widget.0.widget_layout.0.y = 5",
	"layout_type = free",
	"widget.0.widget_layout.0.width = 32",
}

func TestAccDatadogDashboardIFrame(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardIFrameConfig, "datadog_dashboard.iframe_dashboard", datadogDashboardIFrameAsserts)
}

func TestAccDatadogDashboardIFrame_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtil_import(t, datadogDashboardIFrameConfig, "datadog_dashboard.iframe_dashboard")
}
