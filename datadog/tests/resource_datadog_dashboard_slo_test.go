package test

import (
	"testing"
)

const datadogDashboardSLOConfig = `
resource "datadog_dashboard" "slo_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"

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
		}
	}
}
`

var datadogDashboardSLOAsserts = []string{
	"widget.0.service_level_objective_definition.0.title_size = 16",
	"is_read_only = true",
	"title = {{uniq}}",
	"widget.0.service_level_objective_definition.0.slo_id = b4c7739b2af25f9d947f828730357832",
	"widget.0.service_level_objective_definition.0.view_mode = both",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.service_level_objective_definition.0.time_windows.1 = previous_week",
	"widget.0.service_level_objective_definition.0.title_align = center",
	"widget.0.service_level_objective_definition.0.view_type = detail",
	"widget.0.service_level_objective_definition.0.show_error_budget = true",
	"widget.0.service_level_objective_definition.0.time_windows.0 = 90d",
	"widget.0.service_level_objective_definition.0.title =",
	"widget.0.service_level_objective_definition.0.time_windows.2 = global_time",
	"widget.0.service_level_objective_definition.0.global_time_target = 99.0",
	"layout_type = ordered",
}

func TestAccDatadogDashboardSLO(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardSLOConfig, "datadog_dashboard.slo_dashboard", datadogDashboardSLOAsserts)
}

func TestAccDatadogDashboardSLO_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardSLOConfig, "datadog_dashboard.slo_dashboard")
}
