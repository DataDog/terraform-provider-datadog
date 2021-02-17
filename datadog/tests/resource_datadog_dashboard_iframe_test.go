package test

import (
	"testing"
)

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
