package test

import (
	"testing"
)

const datadogPowerpackTopologyMapTest = `
resource "datadog_powerpack" "topology_map_powerpack" {
	name         = "{{uniq}}"
    tags = ["tag:foo1"]
	description   = "Created using the Datadog provider in Terraform"
	template_variables {
		defaults = ["defaults"]
		name     = "datacenter"
	}
  widget {
		topology_map_definition {
			request {
				request_type = "topology"
				query {
					data_source = "service_map"
					service = "master-db"
					filters = ["env:prod","datacenter:dc1"]
				}
			}
			title = "env: prod, datacenter:dc1, service: master-db"
			title_size = "16"
			title_align = "left"
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
		widget_layout {
			height = 4
			width = 3
			x = 5
			y = 5
		}
  }
}
`

var datadogPowerpackTopologyMapTestAsserts = []string{
	// Powerpack metadata
	"name = {{uniq}}",
	"description = Created using the Datadog provider in Terraform",
	"widget.# = 1",
	"tags.# = 1",
	"tags.0 = tag:foo1",
	// Topology map widget
	"widget.0.widget_layout.0.width = 3",
	"widget.0.topology_map_definition.0.title_align = left",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.widget_layout.0.x = 5",
	"widget.0.topology_map_definition.0.title_size = 16",
	"widget.0.topology_map_definition.0.request.0.request_type = topology",
	"widget.0.topology_map_definition.0.request.0.query.0.data_source = service_map",
	"widget.0.topology_map_definition.0.request.0.query.0.data_source = service_map",
	"widget.0.topology_map_definition.0.request.0.query.0.data_source = service_map",
	"widget.0.topology_map_definition.0.request.0.query.0.filters.0 = env:prod",
	"widget.0.topology_map_definition.0.request.0.query.0.service = master-db",
	"widget.0.topology_map_definition.0.request.0.query.0.filters.1 = datacenter:dc1",
	"widget.0.widget_layout.0.y = 5",
	"widget.0.topology_map_definition.0.title = env: prod, datacenter:dc1, service: master-db",
	"widget.0.widget_layout.0.height = 4",
	"widget.0.topology_map_definition.0.custom_link.# = 2",
	"widget.0.topology_map_definition.0.custom_link.0.label = Test Custom Link label",
	"widget.0.topology_map_definition.0.custom_link.0.link = https://app.datadoghq.com/dashboard/lists",
	"widget.0.topology_map_definition.0.custom_link.1.override_label = logs",
	"widget.0.topology_map_definition.0.custom_link.1.link = https://app.datadoghq.com/dashboard/lists",
	"widget.0.topology_map_definition.0.custom_link.1.is_hidden = true",
	// Template Variables
	"template_variables.# = 1",
	"template_variables.0.name = datacenter",
	"template_variables.0.defaults.# = 1",
	"template_variables.0.defaults.0 = defaults",
}

func TestAccDatadogPowerpackTopologyMap(t *testing.T) {
	testAccDatadogPowerpackWidgetUtil(t, datadogPowerpackTopologyMapTest, "datadog_powerpack.topology_map_powerpack", datadogPowerpackTopologyMapTestAsserts)
}
