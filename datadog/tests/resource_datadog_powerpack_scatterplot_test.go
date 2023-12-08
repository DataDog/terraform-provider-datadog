package test

import (
	"testing"
)

const datadogPowerpackScatterplotTest = `
resource "datadog_powerpack" "scatterplot_powerpack" {
	name         = "{{uniq}}"
    tags = ["tag:foo1"]
	description   = "Created using the Datadog provider in Terraform"
	template_variables {
		defaults = ["defaults"]
		name     = "datacenter"
	}
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
			custom_link {
				link = "https://app.datadoghq.com/dashboard/lists"
				is_hidden = true
				override_label = "logs"
			}
		}
	}
}
`

var datadogPowerpackScatterplotTestAsserts = []string{
	// Powerpack metadata
	"name = {{uniq}}",
	"description = Created using the Datadog provider in Terraform",
	"widget.# = 1",
	"tags.# = 1",
	"tags.0 = tag:foo1",
	// scatterplot widget
	"widget.0.scatterplot_definition.0.xaxis.0.min = 0",
	"widget.0.scatterplot_definition.0.color_by_groups.0 = app",
	"widget.0.scatterplot_definition.0.title = system.mem.used and system.cpu.user by service,team,app colored by app",
	"widget.0.scatterplot_definition.0.xaxis.0.max = 100",
	"widget.0.scatterplot_definition.0.yaxis.0.scale = log",
	"widget.0.scatterplot_definition.0.title_size = 16",
	"widget.0.scatterplot_definition.0.yaxis.0.min = 1",
	"widget.0.scatterplot_definition.0.yaxis.0.label = mem (Gib)",
	"widget.0.scatterplot_definition.0.xaxis.0.include_zero = false",
	"widget.0.scatterplot_definition.0.request.0.x.0.q = avg:system.cpu.user{account:prod} by {service, team, app}",
	"widget.0.scatterplot_definition.0.title_align = right",
	"widget.0.scatterplot_definition.0.request.0.x.0.aggregator = avg",
	"widget.0.scatterplot_definition.0.yaxis.0.include_zero = false",
	"widget.0.scatterplot_definition.0.yaxis.0.max =",
	"widget.0.scatterplot_definition.0.request.0.y.0.q = avg:system.mem.used{env:prod} by {service, team, app}",
	"widget.0.scatterplot_definition.0.xaxis.0.label = cpu (%)",
	"widget.0.scatterplot_definition.0.request.0.y.0.aggregator = avg",
	"widget.0.scatterplot_definition.0.xaxis.0.scale = log",
	"widget.0.scatterplot_definition.0.custom_link.# = 2",
	"widget.0.scatterplot_definition.0.custom_link.0.label = Test Custom Link label",
	"widget.0.scatterplot_definition.0.custom_link.0.link = https://app.datadoghq.com/dashboard/lists",
	"widget.0.scatterplot_definition.0.custom_link.1.override_label = logs",
	"widget.0.scatterplot_definition.0.custom_link.1.link = https://app.datadoghq.com/dashboard/lists",
	"widget.0.scatterplot_definition.0.custom_link.1.is_hidden = true",
	// Template Variables
	"template_variables.# = 1",
	"template_variables.0.name = datacenter",
	"template_variables.0.defaults.# = 1",
	"template_variables.0.defaults.0 = defaults",
}

func TestAccDatadogPowerpackScatterplot(t *testing.T) {
	testAccDatadogPowerpackWidgetUtil(t, datadogPowerpackScatterplotTest, "datadog_powerpack.scatterplot_powerpack", datadogPowerpackScatterplotTestAsserts)
}
