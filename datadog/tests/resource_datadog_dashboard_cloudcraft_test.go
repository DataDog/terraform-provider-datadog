package test

import "testing"

const datadogDashboardCloudcraftConfig = `
resource "datadog_dashboard" "cloudcraft_dashboard" {
  title         = "{{uniq}}"
  description   = "Created using the Datadog provider in Terraform"
  layout_type   = "free"

  widget {
		cloudcraft_definition {
			query_string   = "service:web-store"
			provider       = "aws"
			overlay        = "Observability"
			overlay_filter = "env:prod"
			group_by       = ["region", "service"]
			projection     = "isometric"
			title          = "env: prod, service: web-store"
			title_size     = "16"
			title_align    = "left"
			live_span      = "1h"
      		custom_link {
				link  = "https://app.datadoghq.com/dashboard/lists"
				label = "Test Custom Link label"
			}
			custom_link {
				link           = "https://app.datadoghq.com/dashboard/lists"
				is_hidden      = true
				override_label = "logs"
			}
		}
		widget_layout {
			height = 43
			width  = 32
			x      = 5
			y      = 5
		}
  }
}
`

var datadogDashboardCloudcraftAsserts = []string{
	"title = {{uniq}}",
	"widget.0.widget_layout.0.width = 32",
	"widget.0.cloudcraft_definition.0.title_align = left",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.widget_layout.0.x = 5",
	"widget.0.cloudcraft_definition.0.title_size = 16",
	"layout_type = free",
	"widget.0.cloudcraft_definition.0.query_string = service:web-store",
	"widget.0.cloudcraft_definition.0.provider = aws",
	"widget.0.cloudcraft_definition.0.overlay = Observability",
	"widget.0.cloudcraft_definition.0.overlay_filter = env:prod",
	"widget.0.cloudcraft_definition.0.projection = isometric",
	"widget.0.cloudcraft_definition.0.live_span = 1h",
	"widget.0.cloudcraft_definition.0.group_by.0 = region",
	"widget.0.cloudcraft_definition.0.group_by.1 = service",
	"widget.0.widget_layout.0.y = 5",
	"widget.0.cloudcraft_definition.0.title = env: prod, service: web-store",
	"widget.0.widget_layout.0.height = 43",
	"widget.0.cloudcraft_definition.0.custom_link.# = 2",
	"widget.0.cloudcraft_definition.0.custom_link.0.label = Test Custom Link label",
	"widget.0.cloudcraft_definition.0.custom_link.0.link = https://app.datadoghq.com/dashboard/lists",
	"widget.0.cloudcraft_definition.0.custom_link.1.override_label = logs",
	"widget.0.cloudcraft_definition.0.custom_link.1.link = https://app.datadoghq.com/dashboard/lists",
	"widget.0.cloudcraft_definition.0.custom_link.1.is_hidden = true",
}

func TestAccDatadogDashboardCloudcraft(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardCloudcraftConfig, "datadog_dashboard.cloudcraft_dashboard", datadogDashboardCloudcraftAsserts)
}

func TestAccDatadogDashboardCloudcraft_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardCloudcraftConfig, "datadog_dashboard.cloudcraft_dashboard")
}
