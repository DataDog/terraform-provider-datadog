package test

import (
	"testing"
)

const datadogPowerpackHeatMapTest = `
resource "datadog_powerpack" "heatmap_powerpack" {
	name         = "{{uniq}}"
    tags = ["tag:foo1"]
	description   = "Created using the Datadog provider in Terraform"
	template_variables {
		defaults = ["defaults"]
		name     = "datacenter"
	}
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
			custom_link {
				link = "https://app.datadoghq.com/dashboard/lists"
				is_hidden = true
				override_label = "logs"
			}
		}
	}
}
`

var datadogPowerpackHeatMapTestAsserts = []string{
	// Powerpack metadata
	"name = {{uniq}}",
	"description = Created using the Datadog provider in Terraform",
	"widget.# = 1",
	"tags.# = 1",
	"tags.0 = tag:foo1",
	// heatmap widget
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
	//"widget.0.heatmap_definition.0.event.0.q = env:prod",
	//"widget.0.heatmap_definition.0.event.0.tags_execution = and",
	"widget.0.heatmap_definition.0.show_legend = true",
	"widget.0.heatmap_definition.0.legend_size = 2",
	"widget.0.heatmap_definition.0.custom_link.# = 2",
	"widget.0.heatmap_definition.0.custom_link.0.label = Test Custom Link label",
	"widget.0.heatmap_definition.0.custom_link.0.link = https://app.datadoghq.com/dashboard/lists",
	"widget.0.heatmap_definition.0.custom_link.1.override_label = logs",
	"widget.0.heatmap_definition.0.custom_link.1.link = https://app.datadoghq.com/dashboard/lists",
	"widget.0.heatmap_definition.0.custom_link.1.is_hidden = true",
	// Template Variables
	"template_variables.# = 1",
	"template_variables.0.name = datacenter",
	"template_variables.0.defaults.# = 1",
	"template_variables.0.defaults.0 = defaults",
}

func TestAccDatadogPowerpackHeatMap(t *testing.T) {
	testAccDatadogPowerpackWidgetUtil(t, datadogPowerpackHeatMapTest, "datadog_powerpack.heatmap_powerpack", datadogPowerpackHeatMapTestAsserts)
}
