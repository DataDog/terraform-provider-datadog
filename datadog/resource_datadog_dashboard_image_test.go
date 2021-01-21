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
//    "title": "TF - Image Example",
//    "created_at": "2020-06-09T13:35:49.700883+00:00",
//    "modified_at": "2020-06-09T13:36:10.777106+00:00",
//    "author_handle": "--redacted--",
//    "widgets": [
//        {
//            "definition": {
//                "url": "https://i.picsum.photos/id/826/200/300.jpg",
//                "sizing": "fit",
//                "margin": "small",
//                "type": "image"
//            },
//            "layout": {
//                "y": 2,
//                "x": 8,
//                "height": 12,
//                "width": 12
//            },
//            "id": 0
//        }
//    ],
//    "layout_type": "free"
//}

const datadogDashboardImageConfigDeprecated = `
resource "datadog_dashboard" "image_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "free"
	is_read_only  = "true"

	widget {
		image_definition {
			url = "https://i.picsum.photos/id/826/200/300.jpg"
			sizing = "fit"
			margin = "small"
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

var datadogDashboardImageAssertsDeprecated = []string{
	"widget.0.image_definition.0.sizing = fit",
	"title = {{uniq}}",
	"widget.0.layout.y = 5",
	"widget.0.layout.x = 5",
	"widget.0.image_definition.0.margin = small",
	"widget.0.layout.height = 43",
	"layout_type = free",
	"widget.0.layout.width = 32",
	"is_read_only = true",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.image_definition.0.url = https://i.picsum.photos/id/826/200/300.jpg",
}

func TestAccDatadogDashboardImageDeprecated(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardImageConfigDeprecated, "datadog_dashboard.image_dashboard", datadogDashboardImageAssertsDeprecated)
}

const datadogDashboardImageConfig = `
resource "datadog_dashboard" "image_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "free"
	is_read_only  = "true"

	widget {
		image_definition {
			url = "https://i.picsum.photos/id/826/200/300.jpg"
			sizing = "fit"
			margin = "small"
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

var datadogDashboardImageAsserts = []string{
	"widget.0.image_definition.0.sizing = fit",
	"title = {{uniq}}",
	"widget.0.widget_layout.0.y = 5",
	"widget.0.widget_layout.0.x = 5",
	"widget.0.image_definition.0.margin = small",
	"widget.0.widget_layout.0.height = 43",
	"layout_type = free",
	"widget.0.widget_layout.0.width = 32",
	"is_read_only = true",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.image_definition.0.url = https://i.picsum.photos/id/826/200/300.jpg",
}

func TestAccDatadogDashboardImage(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardImageConfig, "datadog_dashboard.image_dashboard", datadogDashboardImageAsserts)
}

func TestAccDatadogDashboardImage_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtil_import(t, datadogDashboardImageConfig, "datadog_dashboard.image_dashboard")
}
