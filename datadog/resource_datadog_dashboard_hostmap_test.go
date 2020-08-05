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
//    "title": "TF - Hostmap Example",
//    "created_at": "2020-06-09T13:05:44.107887+00:00",
//    "modified_at": "2020-06-09T13:07:21.567789+00:00",
//    "author_handle": "--redacted--",
//    "widgets": [
//        {
//            "definition": {
//                "style": {
//                    "fill_min": "10",
//                    "fill_max": "30",
//                    "palette": "YlOrRd",
//                    "palette_flip": true
//                },
//                "title_size": "16",
//                "title": "system.cpu.idle, system.cpu.user",
//                "title_align": "center",
//                "node_type": "host",
//                "no_metric_hosts": true,
//                "group": [
//                    "region"
//                ],
//                "requests": {
//                    "size": {
//                        "q": "max:system.cpu.user{env:prod} by {host}"
//                    },
//                    "fill": {
//                        "q": "avg:system.cpu.idle{env:prod} by {host}"
//                    }
//                },
//                "no_group_hosts": true,
//                "type": "hostmap",
//                "scope": [
//                    "env:prod"
//                ]
//            },
//            "layout": {
//                "y": 2,
//                "x": 3,
//                "height": 22,
//                "width": 47
//            },
//            "id": 0
//        }
//    ],
//    "layout_type": "free"
//}

const datadogDashboardHostMapConfig = `
resource "datadog_dashboard" "hostmap_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"

	widget {
		hostmap_definition {
			style {
				fill_min = "10"
				fill_max = "30"
				palette = "YlOrRd"
				palette_flip = true
			}
			node_type = "host"
			no_metric_hosts = "true"
			group = ["region"]
			request {
				size {
					q = "max:system.cpu.user{env:prod} by {host}"
				}
				fill {
					q = "avg:system.cpu.idle{env:prod} by {host}"
				}
			}
			no_group_hosts = "true"
			scope = ["env:prod"]
			title = "system.cpu.idle, system.cpu.user"
			title_align = "right"
			title_size = "16"
		}
	}
}
`

var datadogDashboardHostMapAsserts = []string{
	"widget.0.hostmap_definition.0.style.0.palette_flip = true",
	"widget.0.hostmap_definition.0.request.0.fill.0.q = avg:system.cpu.idle{env:prod} by {host}",
	"widget.0.hostmap_definition.0.title = system.cpu.idle, system.cpu.user",
	"widget.0.hostmap_definition.0.node_type = host",
	"widget.0.hostmap_definition.0.title_align = right",
	"widget.0.hostmap_definition.0.no_metric_hosts = true",
	"description = Created using the Datadog provider in Terraform",
	"layout_type = ordered",
	"widget.0.hostmap_definition.0.style.0.palette = YlOrRd",
	"widget.0.hostmap_definition.0.scope.0 = env:prod",
	"widget.0.hostmap_definition.0.title_size = 16",
	"widget.0.hostmap_definition.0.style.0.fill_max = 30",
	"widget.0.hostmap_definition.0.style.0.fill_min = 10",
	"widget.0.hostmap_definition.0.no_group_hosts = true",
	"widget.0.hostmap_definition.0.request.0.size.0.q = max:system.cpu.user{env:prod} by {host}",
	"is_read_only = true",
	"title = {{uniq}}",
	"widget.0.hostmap_definition.0.group.0 = region",
}

func TestAccDatadogDashboardHostMap(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardHostMapConfig, "datadog_dashboard.hostmap_dashboard", datadogDashboardHostMapAsserts)
}

func TestAccDatadogDashboardHostMap_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtil_import(t, datadogDashboardHostMapConfig, "datadog_dashboard.hostmap_dashboard")
}
