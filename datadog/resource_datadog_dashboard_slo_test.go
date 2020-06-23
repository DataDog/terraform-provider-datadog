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
//    "title": "TF - SLO Example",
//    "created_at": "2020-06-09T13:41:36.039693+00:00",
//    "modified_at": "2020-06-09T13:41:58.724155+00:00",
//    "author_handle": "--redacted--m",
//    "widgets": [
//        {
//            "definition": {
//                "time_windows": [
//                    "90d",
//                    "previous_week",
//                    "month_to_date"
//                ],
//                "title_size": "16",
//                "show_error_budget": true,
//                "title": "",
//                "title_align": "center",
//                "slo_id": "b4c7739b2af25f9d947f828730357832",
//                "view_mode": "both",
//                "view_type": "detail",
//                "type": "slo"
//            },
//            "layout": {
//                "y": 3,
//                "x": 5,
//                "height": 21,
//                "width": 60
//            },
//            "id": 0
//        }
//    ],
//    "layout_type": "free"
//}

const datadogDashboardSLOConfig = `
resource "datadog_dashboard" "slo_dashboard" {
	title         = "Acceptance Test SLO Widget Dashboard"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"

	widget {
		service_level_objective_definition {
			time_windows = ["90d","previous_week","month_to_date"]
			title_size = "16"
			show_error_budget = true
			title = ""
			title_align = "center"
			slo_id = "b4c7739b2af25f9d947f828730357832"
			view_mode = "both"
			view_type = "detail"
		}
	}
}
`

var datadogDashboardSLOAsserts = []string{
	"widget.0.service_level_objective_definition.0.title_size = 16",
	"is_read_only = true",
	"title = Acceptance Test SLO Widget Dashboard",
	"widget.0.service_level_objective_definition.0.slo_id = b4c7739b2af25f9d947f828730357832",
	"widget.0.service_level_objective_definition.0.view_mode = both",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.service_level_objective_definition.0.time_windows.1 = previous_week",
	"widget.0.service_level_objective_definition.0.title_align = center",
	"widget.0.service_level_objective_definition.0.view_type = detail",
	"widget.0.service_level_objective_definition.0.show_error_budget = true",
	"widget.0.service_level_objective_definition.0.time_windows.0 = 90d",
	"widget.0.service_level_objective_definition.0.title =",
	"widget.0.service_level_objective_definition.0.time_windows.2 = month_to_date",
	"layout_type = ordered",
}

func TestAccDatadogDashboardSLO(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardSLOConfig, "datadog_dashboard.slo_dashboard", datadogDashboardSLOAsserts)
}

func TestAccDatadogDashboardSLO_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtil_import(t, datadogDashboardSLOConfig, "datadog_dashboard.slo_dashboard")
}
