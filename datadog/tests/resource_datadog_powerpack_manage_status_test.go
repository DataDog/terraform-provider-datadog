package test

import (
	"testing"
)

const datadogPowerpackManageStatusTest = `
resource "datadog_powerpack" "manage_status_powerpack" {
	name         = "{{uniq}}"
    tags = ["tag:foo1"]
	description   = "Created using the Datadog provider in Terraform"
	template_variables {
		defaults = ["defaults"]
		name     = "datacenter"
	}
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
			height = 6
			width = 4
			x = 5
			y = 5
		}
	}
}
`

var datadogPowerpackManageStatusTestAsserts = []string{
	// Powerpack metadata
	"description = Created using the Datadog provider in Terraform",
	"widget.# = 1",
	"tags.# = 1",
	"tags.0 = tag:foo1",

	// Manage Status widgets
	"widget.0.manage_status_definition.0.color_preference = background",
	"widget.0.widget_layout.0.x = 5",
	"widget.0.widget_layout.0.height = 6",
	"widget.0.manage_status_definition.0.display_format = countsAndList",
	"widget.0.widget_layout.0.width = 4",
	"widget.0.manage_status_definition.0.hide_zero_counts = true",
	"widget.0.widget_layout.0.y = 5",
	"widget.0.manage_status_definition.0.summary_type = combined",
	"widget.0.manage_status_definition.0.show_last_triggered = true",
	"widget.0.manage_status_definition.0.title =",
	"widget.0.manage_status_definition.0.title_size = 20",
	"widget.0.manage_status_definition.0.sort = triggered,desc",
	"widget.0.manage_status_definition.0.title_align = center",
	"widget.0.manage_status_definition.0.show_priority = false",
	"widget.0.manage_status_definition.0.query = env:prod group_status:alert",
	// Template Variables
	"template_variables.# = 1",
	"template_variables.0.name = datacenter",
	"template_variables.0.defaults.# = 1",
	"template_variables.0.defaults.0 = defaults",
}

func TestAccDatadogPowerpackManageStatus(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogPowerpackManageStatusTest, "datadog_powerpack.manage_status_powerpack", datadogPowerpackManageStatusTestAsserts)
}
