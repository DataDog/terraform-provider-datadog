package datadog

import (
	"testing"
)

// JSON export used as test scenario
//{
//    "notify_list": [],
//    "description": "",
//    "author_name": "--redacted--",
//    "template_variables": [],
//    "is_read_only": false,
//    "id": "--redacted--",
//    "title": "TF - Top List Example",
//    "url": "--redacted--",
//    "created_at": "2020-06-09T12:07:23.772156+00:00",
//    "modified_at": "2020-06-09T12:10:02.808703+00:00",
//    "author_handle": "--redacted--",
//    "widgets": [
//        {
//            "definition": {
//                "title_size": "16",
//                "title": "Avg of system.core.user over account:prod by service,app",
//                "title_align": "right",
//                "time": {
//                    "live_span": "1w"
//                },
//                "requests": [
//                    {
//                        "q": "top(avg:system.core.user{account:prod} by {service,app}, 10, 'sum', 'desc')",
//                        "conditional_formats": [
//                            {
//                                "palette": "white_on_red",
//                                "value": 15000,
//                                "comparator": ">"
//                            }
//                        ]
//                    }
//                ],
//                "type": "toplist"
//            },
//            "layout": {
//                "y": 1,
//                "x": 1,
//                "height": 15,
//                "width": 47
//            },
//            "id": 0
//        }
//    ],
//    "layout_type": "free"
//}

const datadogDashboardTopListConfig = `
resource "datadog_dashboard" "top_list_dashboard" {
	title         = "Acceptance Test Top List Widget Dashboard"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"

	widget {
		toplist_definition {
			title_size = "16"
			title = "Avg of system.core.user over account:prod by service,app"
			title_align = "right"
			time = {
				live_span = "1w"
			}
			request {
				q = "top(avg:system.core.user{account:prod} by {service,app}, 10, 'sum', 'desc')"
				conditional_formats {
					palette = "white_on_red"
					value = 15000
					comparator = ">"
				}
			}
		}
	}
}
`

var datadogDashboardTopListAsserts = []string{
	"description = Created using the Datadog provider in Terraform",
	"widget.0.toplist_definition.0.request.0.conditional_formats.0.timeframe =",
	"widget.0.toplist_definition.0.request.0.conditional_formats.0.image_url =",
	"layout_type = ordered",
	"widget.0.toplist_definition.0.request.0.conditional_formats.0.comparator = >",
	"title = Acceptance Test Top List Widget Dashboard",
	"widget.0.toplist_definition.0.request.0.conditional_formats.0.custom_bg_color =",
	"widget.0.toplist_definition.0.request.0.conditional_formats.0.palette = white_on_red",
	"is_read_only = true",
	"widget.0.toplist_definition.0.time.live_span = 1w",
	"widget.0.toplist_definition.0.time.% = 1",
	"widget.0.toplist_definition.0.request.0.conditional_formats.0.hide_value = false",
	"widget.0.toplist_definition.0.request.0.q = top(avg:system.core.user{account:prod} by {service,app}, 10, 'sum', 'desc')",
	"widget.0.toplist_definition.0.title_size = 16",
	"widget.0.toplist_definition.0.title_align = right",
	"widget.0.toplist_definition.0.request.0.conditional_formats.0.value = 15000",
	"widget.0.toplist_definition.0.title = Avg of system.core.user over account:prod by service,app",
	"widget.0.toplist_definition.0.request.0.conditional_formats.0.custom_fg_color =",
}

func TestAccDatadogDashboardTopList(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardTopListConfig, "datadog_dashboard.top_list_dashboard", datadogDashboardTopListAsserts)
}

func TestAccDatadogDashboardTopList_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtil_import(t, datadogDashboardTopListConfig, "datadog_dashboard.top_list_dashboard")
}
