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
//   "created_at":"2020-10-07T20:43:57.231383+00:00",
//   "modified_at":"2020-10-07T20:47:26.214103+00:00",
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
//            "yaxis":{
//               "include_zero":false,
//               "max":"100"
//            },
//            "title_align":"center",
//            "events":[
//               {
//                  "q":"env:prod",
//                  "tags_execution":"and"
//               }
//            ],
//            "show_legend":true,
//            "time":{
//               "live_span":"1mo"
//            },
//            "title":"Avg of system.cpu.user over account:prod by app",
//            "legend_size":"2",
//            "type":"heatmap",
//            "requests":[
//               {
//                  "q":"avg:system.cpu.user{account:prod} by {app}",
//                  "style":{
//                     "palette":"blue"
//                  }
//               }
//            ]
//         },
//         "id": "--redacted--"
//      }
//   ],
//   "layout_type":"ordered"
//}

const datadogDashboardHeatMapConfig = `
resource "datadog_dashboard" "heatmap_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"

	widget {
		heatmap_definition {
			title = "Avg of system.cpu.user over account:prod by app"
			title_align = "center"
			title_size = "16"
			yaxis {
				max = "100"
			}
			request {
				q = "avg:system.cpu.user{account:prod} by {app}"
				style {
					palette = "blue"
				}
			}

			time = {
				live_span = "1mo"
			}
			event {
				q = "env:prod"
				tags_execution = "and"
			}
			show_legend = true
			legend_size = "2"
			custom_link {
				link = "https://app.datadoghq.com/dashboard/lists"
				label = "Test Custom Link label"
			}
		}
	}
}
`

var datadogDashboardHeatMapAsserts = []string{
	"title = {{uniq}}",
	"description = Created using the Datadog provider in Terraform",
	"layout_type = ordered",
	"is_read_only = true",
	"widget.0.heatmap_definition.0.title = Avg of system.cpu.user over account:prod by app",
	"widget.0.heatmap_definition.0.title_align = center",
	"widget.0.heatmap_definition.0.title_size = 16",
	"widget.0.heatmap_definition.0.request.0.q = avg:system.cpu.user{account:prod} by {app}",
	"widget.0.heatmap_definition.0.request.0.style.0.palette = blue",
	"widget.0.heatmap_definition.0.yaxis.0.include_zero = false",
	"widget.0.heatmap_definition.0.yaxis.0.label =",
	"widget.0.heatmap_definition.0.yaxis.0.max = 100",
	"widget.0.heatmap_definition.0.yaxis.0.scale =",
	"widget.0.heatmap_definition.0.yaxis.0.min =",
	"widget.0.heatmap_definition.0.time.live_span = 1mo",
	"widget.0.heatmap_definition.0.event.0.q = env:prod",
	"widget.0.heatmap_definition.0.event.0.tags_execution = and",
	"widget.0.heatmap_definition.0.show_legend = true",
	"widget.0.heatmap_definition.0.legend_size = 2",
	"widget.0.heatmap_definition.0.custom_link.# = 1",
	"widget.0.heatmap_definition.0.custom_link.0.label = Test Custom Link label",
	"widget.0.heatmap_definition.0.custom_link.0.link = https://app.datadoghq.com/dashboard/lists",
}

func TestAccDatadogDashboardHeatMap(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardHeatMapConfig, "datadog_dashboard.heatmap_dashboard", datadogDashboardHeatMapAsserts)
}

func TestAccDatadogDashboardHeatMap_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtil_import(t, datadogDashboardHeatMapConfig, "datadog_dashboard.heatmap_dashboard")
}
