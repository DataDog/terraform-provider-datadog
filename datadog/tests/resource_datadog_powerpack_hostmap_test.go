package test

import (
	"testing"
)

const datadogPowerpackHostMapTest = `
resource "datadog_powerpack" "hostmap_powerpack" {
	name         = "{{uniq}}"
    tags = ["tag:foo1"]
	description   = "Created using the Datadog provider in Terraform"
	template_variables {
		defaults = ["defaults"]
		name     = "datacenter"
	}
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
			custom_link {
				link = "https://app.datadoghq.com/dashboard/lists"
				is_hidden = true
				override_label = "logs"
			}
		}
	}
}
`

var datadogPowerpackHostMapTestAsserts = []string{
	// Powerpack metadata
	"name = {{uniq}}",
	"description = Created using the Datadog provider in Terraform",
	"widget.# = 1",
	"tags.# = 1",
	"tags.0 = tag:foo1",
	// Hostmap widget
	"widget.0.hostmap_definition.0.style.0.palette_flip = true",
	"widget.0.hostmap_definition.0.request.0.fill.0.q = avg:system.cpu.idle{env:prod} by {host}",
	"widget.0.hostmap_definition.0.title = system.cpu.idle, system.cpu.user",
	"widget.0.hostmap_definition.0.node_type = host",
	"widget.0.hostmap_definition.0.title_align = right",
	"widget.0.hostmap_definition.0.no_metric_hosts = true",
	"widget.0.hostmap_definition.0.style.0.palette = YlOrRd",
	"widget.0.hostmap_definition.0.scope.0 = env:prod",
	"widget.0.hostmap_definition.0.title_size = 16",
	"widget.0.hostmap_definition.0.style.0.fill_max = 30",
	"widget.0.hostmap_definition.0.style.0.fill_min = 10",
	"widget.0.hostmap_definition.0.no_group_hosts = true",
	"widget.0.hostmap_definition.0.request.0.size.0.q = max:system.cpu.user{env:prod} by {host}",
	"widget.0.hostmap_definition.0.group.0 = region",
	"widget.0.hostmap_definition.0.custom_link.# = 2",
	"widget.0.hostmap_definition.0.custom_link.0.label = Test Custom Link label",
	"widget.0.hostmap_definition.0.custom_link.0.link = https://app.datadoghq.com/dashboard/lists",
	"widget.0.hostmap_definition.0.custom_link.1.override_label = logs",
	"widget.0.hostmap_definition.0.custom_link.1.link = https://app.datadoghq.com/dashboard/lists",
	"widget.0.hostmap_definition.0.custom_link.1.is_hidden = true",

	// Template Variables
	"template_variables.# = 1",
	"template_variables.0.name = datacenter",
	"template_variables.0.defaults.# = 1",
	"template_variables.0.defaults.0 = defaults",
}

func TestAccDatadogPowerpackHostMap(t *testing.T) {
	testAccDatadogPowerpackWidgetUtil(t, datadogPowerpackHostMapTest, "datadog_powerpack.hostmap_powerpack", datadogPowerpackHostMapTestAsserts)
}
