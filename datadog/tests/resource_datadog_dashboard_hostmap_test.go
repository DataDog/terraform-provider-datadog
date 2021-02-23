package test

import (
	"testing"
)

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
			custom_link {
				link = "https://app.datadoghq.com/dashboard/lists"
				label = "Test Custom Link label"
			}
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
	"widget.0.hostmap_definition.0.custom_link.# = 1",
	"widget.0.hostmap_definition.0.custom_link.0.label = Test Custom Link label",
	"widget.0.hostmap_definition.0.custom_link.0.link = https://app.datadoghq.com/dashboard/lists",
}

func TestAccDatadogDashboardHostMap(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardHostMapConfig, "datadog_dashboard.hostmap_dashboard", datadogDashboardHostMapAsserts)
}

func TestAccDatadogDashboardHostMap_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardHostMapConfig, "datadog_dashboard.hostmap_dashboard")
}
