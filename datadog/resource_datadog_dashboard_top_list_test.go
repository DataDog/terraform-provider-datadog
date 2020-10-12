package datadog

import (
	"testing"
)

// JSON export used as test scenario
//{
//   "notify_list":[],
//   "description":"Created using the Datadog provider in Terraform",
//   "author_name":"--redacted--",
//   "template_variable_presets":[],
//   "template_variables":[],
//   "is_read_only":true,
//   "id":"--redacted--",
//   "title":"{{uniq}}",
//   "url":"--redacted--",
//   "created_at":"2020-10-07T21:36:23.968919+00:00",
//   "modified_at":"2020-10-07T21:36:23.968919+00:00",
//   "author_handle":"--redacted--",
//   "widgets":[
//      {
//         "definition":{
//            "custom_links":[
//               {
//                  "link":"https://app.datadoghq.com/dashboard/lists",
//                  "label":"Test Custom Link label"
//               }
//            ],
//            "title_size":"16",
//            "title":"Avg of system.core.user over account:prod by service,app",
//            "title_align":"right",
//            "time":{
//               "live_span":"1w"
//            },
//            "requests":[
//               {
//                  "q":"top(avg:system.core.user{account:prod} by {service,app}, 10, 'sum', 'desc')",
//                  "conditional_formats":[
//                     {
//                        "palette":"white_on_red",
//                        "hide_value":false,
//                        "value":15000,
//                        "comparator":">"
//                     }
//                  ]
//               }
//            ],
//            "type":"toplist"
//         },
//         "id": "--redacted--"
//      }
//   ],
//   "layout_type":"ordered"
//}

const datadogDashboardTopListConfig = `
resource "datadog_dashboard" "top_list_dashboard" {
	title         = "{{uniq}}"
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
      		custom_link {
				link = "https://app.datadoghq.com/dashboard/lists"
				label = "Test Custom Link label"
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
	"title = {{uniq}}",
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
	"widget.0.toplist_definition.0.custom_link.# = 1",
	"widget.0.toplist_definition.0.custom_link.0.label = Test Custom Link label",
	"widget.0.toplist_definition.0.custom_link.0.link = https://app.datadoghq.com/dashboard/lists",
}

func TestAccDatadogDashboardTopList(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardTopListConfig, "datadog_dashboard.top_list_dashboard", datadogDashboardTopListAsserts)
}

func TestAccDatadogDashboardTopList_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtil_import(t, datadogDashboardTopListConfig, "datadog_dashboard.top_list_dashboard")
}
