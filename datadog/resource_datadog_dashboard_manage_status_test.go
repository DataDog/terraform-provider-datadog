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
//    "title": "TF - Manage Status Example",
//    "created_at": "2020-06-09T13:24:12.975441+00:00",
//    "modified_at": "2020-06-09T13:25:14.195517+00:00",
//    "author_handle": "--redacted--",
//    "widgets": [
//        {
//            "definition": {
//                "sort": "triggered,desc",
//                "count": 50,
//                "title_size": "20",
//                "title": "",
//                "title_align": "center",
//                "hide_zero_counts": true,
//                "start": 0,
//                "summary_type": "combined",
//                "color_preference": "background",
//                "query": "env:prod group_status:alert",
//                "show_last_triggered": true,
//                "type": "manage_status",
//                "display_format": "countsAndList"
//            },
//            "layout": {
//                "y": 3,
//                "x": 1,
//                "height": 25,
//                "width": 50
//            },
//            "id": 0
//        }
//    ],
//    "layout_type": "free"
//}

const datadogDashboardManageStatusConfig = `
resource "datadog_dashboard" "manage_status_dashboard" {
	title         = "Acceptance Test Manage Status Widget Dashboard"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "free"
	is_read_only  = "true"

	widget {
		manage_status_definition {
			sort = "triggered,desc"
			count = "50"
			title_size = "20"
			title = ""
			title_align = "center"
			hide_zero_counts = true
			start = "0"
			summary_type = "combined"
			color_preference = "background"
			query = "env:prod group_status:alert"
			show_last_triggered = true
			display_format = "countsAndList"
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

var datadogDashboardManageStatusAsserts = []string{
	"widget.0.manage_status_definition.0.color_preference = background",
	"widget.0.layout.x = 5",
	"widget.0.layout.height = 43",
	"layout_type = free",
	"is_read_only = true",
	"widget.0.manage_status_definition.0.display_format = countsAndList",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.layout.width = 32",
	"widget.0.manage_status_definition.0.hide_zero_counts = true",
	"widget.0.layout.y = 5",
	"widget.0.manage_status_definition.0.summary_type = combined",
	"widget.0.manage_status_definition.0.show_last_triggered = true",
	"widget.0.manage_status_definition.0.title =",
	"widget.0.manage_status_definition.0.title_size = 20",
	"widget.0.manage_status_definition.0.sort = triggered,desc",
	"widget.0.manage_status_definition.0.title_align = center",
	"widget.0.manage_status_definition.0.start = 0",
	"widget.0.manage_status_definition.0.count = 50",
	"title = Acceptance Test Manage Status Widget Dashboard",
	"widget.0.manage_status_definition.0.query = env:prod group_status:alert",
}

func TestAccDatadogDashboardManageStatus(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardManageStatusConfig, "datadog_dashboard.manage_status_dashboard", datadogDashboardManageStatusAsserts)
}

func TestAccDatadogDashboardManageStatus_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtil_import(t, datadogDashboardManageStatusConfig, "datadog_dashboard.manage_status_dashboard")
}
