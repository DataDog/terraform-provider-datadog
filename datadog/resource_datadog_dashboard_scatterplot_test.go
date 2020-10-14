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
//   "created_at":"2020-10-07T21:18:21.603323+00:00",
//   "modified_at":"2020-10-07T21:18:21.603323+00:00",
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
//               "scale":"log",
//               "min":"1",
//               "label":"mem (Gib)"
//            },
//            "title_align":"right",
//            "color_by_groups":[
//               "app"
//            ],
//            "xaxis":{
//               "include_zero":false,
//               "max":"100",
//               "min":"0",
//               "scale":"log",
//               "label":"cpu (%)"
//            },
//            "time":{
//               "live_span":"15m"
//            },
//            "title":"system.mem.used and system.cpu.user by service,team,app colored by app",
//            "requests":{
//               "y":{
//                  "q":"avg:system.mem.used{env:prod} by {service, team, app}",
//                  "aggregator":"avg"
//               },
//               "x":{
//                  "q":"avg:system.cpu.user{account:prod} by {service, team, app}",
//                  "aggregator":"avg"
//               }
//            },
//            "type":"scatterplot"
//         },
//         "id": "--redacted--"
//      }
//   ],
//   "layout_type":"ordered"
//}

const datadogDashboardScatterplotConfig = `
resource "datadog_dashboard" "scatterplot_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"

	widget {
		scatterplot_definition {
			title_size = "16"
			yaxis {
				scale = "log"
				include_zero = false
				min = "1"
				label = "mem (Gib)"
			}
			title_align = "right"
			color_by_groups = ["app"]
			xaxis {
				scale = "log"
				max = "100"
				min = "0"
				label = "cpu (%)"
				include_zero = false
			}
			time = {
				live_span = "15m"
			}
			title = "system.mem.used and system.cpu.user by service,team,app colored by app"
			request {
				y {
					q = "avg:system.mem.used{env:prod} by {service, team, app}"
					aggregator = "avg"
				}
				x {
					q = "avg:system.cpu.user{account:prod} by {service, team, app}"
					aggregator = "avg"
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

var datadogDashboardScatterplotAsserts = []string{
	"widget.0.scatterplot_definition.0.xaxis.0.min = 0",
	"widget.0.scatterplot_definition.0.color_by_groups.0 = app",
	"widget.0.scatterplot_definition.0.title = system.mem.used and system.cpu.user by service,team,app colored by app",
	"widget.0.scatterplot_definition.0.xaxis.0.max = 100",
	"widget.0.scatterplot_definition.0.yaxis.0.scale = log",
	"is_read_only = true",
	"widget.0.scatterplot_definition.0.title_size = 16",
	"widget.0.scatterplot_definition.0.yaxis.0.min = 1",
	"widget.0.scatterplot_definition.0.yaxis.0.label = mem (Gib)",
	"widget.0.scatterplot_definition.0.xaxis.0.include_zero = false",
	"widget.0.scatterplot_definition.0.request.0.x.0.q = avg:system.cpu.user{account:prod} by {service, team, app}",
	"widget.0.scatterplot_definition.0.title_align = right",
	"layout_type = ordered",
	"title = {{uniq}}",
	"widget.0.scatterplot_definition.0.request.0.x.0.aggregator = avg",
	"widget.0.scatterplot_definition.0.yaxis.0.include_zero = false",
	"widget.0.scatterplot_definition.0.time.live_span = 15m",
	"widget.0.scatterplot_definition.0.yaxis.0.max =",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.scatterplot_definition.0.request.0.y.0.q = avg:system.mem.used{env:prod} by {service, team, app}",
	"widget.0.scatterplot_definition.0.xaxis.0.label = cpu (%)",
	"widget.0.scatterplot_definition.0.request.0.y.0.aggregator = avg",
	"widget.0.scatterplot_definition.0.xaxis.0.scale = log",
	"widget.0.scatterplot_definition.0.custom_link.# = 1",
	"widget.0.scatterplot_definition.0.custom_link.0.label = Test Custom Link label",
	"widget.0.scatterplot_definition.0.custom_link.0.link = https://app.datadoghq.com/dashboard/lists",
}

func TestAccDatadogDashboardScatterplot(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardScatterplotConfig, "datadog_dashboard.scatterplot_dashboard", datadogDashboardScatterplotAsserts)
}

func TestAccDatadogDashboardScatterplot_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtil_import(t, datadogDashboardScatterplotConfig, "datadog_dashboard.scatterplot_dashboard")
}
