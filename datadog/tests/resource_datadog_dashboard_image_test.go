package test

import (
	"testing"
)

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
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardImageConfig, "datadog_dashboard.image_dashboard")
}
