package test

import (
	"testing"
)

const datadogDashboardManageStatusConfig = `
resource "datadog_dashboard" "manage_status_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "free"
	is_read_only  = "true"

	widget {
		manage_status_definition {
			sort = "triggered,desc"
			title_size = "20"
			title = ""
			title_align = "center"
			hide_zero_counts = true
			summary_type = "combined"
			color_preference = "background"
			query = "env:prod group_status:alert"
			show_last_triggered = true
			display_format = "countsAndList"
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

var datadogDashboardManageStatusAsserts = []string{
	"widget.0.manage_status_definition.0.color_preference = background",
	"widget.0.widget_layout.0.x = 5",
	"widget.0.widget_layout.0.height = 43",
	"layout_type = free",
	"is_read_only = true",
	"widget.0.manage_status_definition.0.display_format = countsAndList",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.widget_layout.0.width = 32",
	"widget.0.manage_status_definition.0.hide_zero_counts = true",
	"widget.0.widget_layout.0.y = 5",
	"widget.0.manage_status_definition.0.summary_type = combined",
	"widget.0.manage_status_definition.0.show_last_triggered = true",
	"widget.0.manage_status_definition.0.title =",
	"widget.0.manage_status_definition.0.title_size = 20",
	"widget.0.manage_status_definition.0.sort = triggered,desc",
	"widget.0.manage_status_definition.0.title_align = center",
	"widget.0.manage_status_definition.0.show_priority = false",
	"title = {{uniq}}",
	"widget.0.manage_status_definition.0.query = env:prod group_status:alert",
}

func TestAccDatadogDashboardManageStatus(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardManageStatusConfig, "datadog_dashboard.manage_status_dashboard", datadogDashboardManageStatusAsserts)
}

func TestAccDatadogDashboardManageStatus_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardManageStatusConfig, "datadog_dashboard.manage_status_dashboard")
}
