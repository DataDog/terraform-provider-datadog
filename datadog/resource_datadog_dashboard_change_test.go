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
//    "title": "TF - Change Example",
//    "created_at": "2020-06-09T13:01:22.235028+00:00",
//    "modified_at": "2020-06-09T13:03:15.925532+00:00",
//    "author_handle": "--redacted--",
//    "widgets": [
//        {
//            "definition": {
//                "title_size": "16",
//                "title": "Sum of system.cpu.user over * by service,account",
//                "title_align": "left",
//                "time": {
//                    "live_span": "1h"
//                },
//                "requests": [
//                    {
//                        "change_type": "relative",
//                        "order_dir": "desc",
//                        "compare_to": "day_before",
//                        "q": "sum:system.cpu.user{*} by {service,account}",
//                        "show_present": true,
//                        "increase_good": false,
//                        "order_by": "change"
//                    }
//                ],
//                "type": "change"
//            },
//            "layout": {
//                "y": 6,
//                "x": 9,
//                "height": 15,
//                "width": 47
//            },
//            "id": 1
//        }
//    ],
//    "layout_type": "free"
//}

const datadogDashboardChangeConfig = `
resource "datadog_dashboard" "change_dashboard" {
   	title         = "{{uniq}}"
   	description   = "Created using the Datadog provider in Terraform"
   	layout_type   = "ordered"
   	is_read_only  = true
	widget {
		change_definition {
			request {
				q = "sum:system.cpu.user{*} by {service,account}"
			}
		}
	}
	
	widget {
		change_definition {
			request {
				q = "sum:system.cpu.user{*} by {service,account}"
				compare_to = "day_before"
				increase_good = "false"
				order_by = "change"
				change_type = "absolute"
				order_dir = "desc"
				show_present = "true"
			}
			title = "Sum of system.cpu.user over * by service,account"
			title_size = "16"
			title_align = "left"
			time = {
				live_span = "1h"
			}
		}
	}
}
`

var datadogDashboardChangeAsserts = []string{
	"widget.0.change_definition.0.request.0.q = sum:system.cpu.user{*} by {service,account}",
	"widget.1.change_definition.0.title_align = left",
	"widget.1.change_definition.0.request.0.change_type = absolute",
	"widget.0.change_definition.0.request.0.order_dir =",
	"widget.0.change_definition.0.title_size =",
	"title = {{uniq}}",
	"widget.0.change_definition.0.request.0.change_type =",
	"widget.1.change_definition.0.title = Sum of system.cpu.user over * by service,account",
	"widget.1.change_definition.0.title_size = 16",
	"widget.1.change_definition.0.request.0.compare_to = day_before",
	"is_read_only = true",
	"widget.0.change_definition.0.title_align =",
	"widget.0.change_definition.0.title =",
	"widget.1.change_definition.0.request.0.q = sum:system.cpu.user{*} by {service,account}",
	"widget.1.change_definition.0.request.0.show_present = true",
	"widget.1.change_definition.0.request.0.order_by = change",
	"layout_type = ordered",
	"widget.1.change_definition.0.request.0.order_dir = desc",
	"widget.0.change_definition.0.request.0.increase_good = false",
	"widget.1.change_definition.0.request.0.increase_good = false",
	"widget.0.change_definition.0.request.0.show_present = false",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.change_definition.0.request.0.order_by =",
	"widget.1.change_definition.0.time.live_span = 1h",
	"widget.0.change_definition.0.request.0.compare_to =",
}

func TestAccDatadogDashboardChange(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardChangeConfig, "datadog_dashboard.change_dashboard", datadogDashboardChangeAsserts)
}

func TestAccDatadogDashboardChange_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtil_import(t, datadogDashboardChangeConfig, "datadog_dashboard.change_dashboard")
}
