package test

import (
	"testing"
)

const datadogPowerpackSloTest = `
resource "datadog_powerpack" "service_level_objective_powerpack" {
	name         = "{{uniq}}"
    tags = ["tag:foo1"]
	description   = "Created using the Datadog provider in Terraform"
	template_variables {
		defaults = ["defaults"]
		name     = "datacenter"
	}
	widget {
		service_level_objective_definition {
			time_windows = ["90d","previous_week","global_time"]
			title_size = "16"
			show_error_budget = true
			title = ""
			title_align = "center"
			slo_id = "b4c7739b2af25f9d947f828730357832"
			view_mode = "both"
			view_type = "detail"
			global_time_target = "99.0"
			additional_query_filters = "!host:excluded_host"
		}
	}
}
`

var datadogPowerpackSloTestAsserts = []string{
	// Powerpack metadata
	"name = {{uniq}}",
	"description = Created using the Datadog provider in Terraform",
	"widget.# = 1",
	"tags.# = 1",
	"tags.0 = tag:foo1",
	// service level objective widget
	"widget.0.service_level_objective_definition.0.title_size = 16",
	"widget.0.service_level_objective_definition.0.slo_id = b4c7739b2af25f9d947f828730357832",
	"widget.0.service_level_objective_definition.0.view_mode = both",
	"widget.0.service_level_objective_definition.0.time_windows.1 = previous_week",
	"widget.0.service_level_objective_definition.0.title_align = center",
	"widget.0.service_level_objective_definition.0.view_type = detail",
	"widget.0.service_level_objective_definition.0.show_error_budget = true",
	"widget.0.service_level_objective_definition.0.time_windows.0 = 90d",
	"widget.0.service_level_objective_definition.0.title =",
	"widget.0.service_level_objective_definition.0.time_windows.2 = global_time",
	"widget.0.service_level_objective_definition.0.global_time_target = 99.0",
	"widget.0.service_level_objective_definition.0.additional_query_filters = !host:excluded_host",
	// Template Variables
	"template_variables.# = 1",
	"template_variables.0.name = datacenter",
	"template_variables.0.defaults.# = 1",
	"template_variables.0.defaults.0 = defaults",
}

func TestAccDatadogPowerpackSlo(t *testing.T) {
	testAccDatadogPowerpackWidgetUtil(t, datadogPowerpackSloTest, "datadog_powerpack.service_level_objective_powerpack", datadogPowerpackSloTestAsserts)
}
