package datadog

import (
	"testing"
)

const datadogDashboardFreeTextConfig = `
resource "datadog_dashboard" "free_text_dashboard" {
	title         = "{{uniq}}"
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
		widget_layout {
			height = 43
			width = 32
			x = 5
			y = 5
		}
	}
}
`

var datadogDashboardFreeTextAsserts = []string{
	"widget.0.widget_layout.0.y = 5",
	"widget.0.free_text_definition.0.text = Free Text",
	"layout_type = free",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.free_text_definition.0.font_size = 56",
	"is_read_only = true",
	"widget.0.free_text_definition.0.color = #eb364b",
	"widget.0.widget_layout.0.width = 32",
	"widget.0.widget_layout.0.height = 43",
	"widget.0.free_text_definition.0.text_align = left",
	"title = {{uniq}}",
	"widget.0.widget_layout.0.x = 5",
	"widget.0.layout.% = 0",
}

func TestAccDatadogDashboardFreeText(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardFreeTextConfig, "datadog_dashboard.free_text_dashboard", datadogDashboardFreeTextAsserts)
}

func TestAccDatadogDashboardFreeText_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtil_import(t, datadogDashboardFreeTextConfig, "datadog_dashboard.free_text_dashboard")
}
