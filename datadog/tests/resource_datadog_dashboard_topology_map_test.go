package test

import "testing"

const datadogDashboardTopologyMapConfig = `
resource "datadog_dashboard" "topology_map_dashboard" {
  title         = "{{uniq}}"
  description   = "Created using the Datadog provider in Terraform"
  layout_type   = "free"
  is_read_only  = "true"

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
			height = 43
			width = 32
			x = 5
			y = 5
		}
  }
}
`

var datadogDashboardTopologyMapAsserts = []string{
	"title = {{uniq}}",
	"widget.0.widget_layout.0.width = 32",
	"widget.0.topology_map_definition.0.title_align = left",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.widget_layout.0.x = 5",
	"widget.0.topology_map_definition.0.title_size = 16",
	"layout_type = free",
	"widget.0.topology_map_definition.0.request.0.request_type = topology",
	"widget.0.topology_map_definition.0.request.0.query.data_source = service_map",
	"widget.0.topology_map_definition.0.request.0.query.data_source = service_map",
	"widget.0.topology_map_definition.0.request.0.query.data_source = service_map",
	"widget.0.topology_map_definition.0.request.0.query.filters.0 = env:prod",
	"widget.0.topology_map_definition.0.request.0.query.service = master-db",
	"is_read_only = true",
	"widget.0.widget_layout.0.y = 5",
	"widget.0.topology_map_definition.0.title = env: prod, datacenter:dc1, service: master-db",
	"widget.0.widget_layout.0.height = 43",
	"widget.0.topology_map_definition.0.filters.1 = datacenter:dc1",
	"widget.0.topology_map_definition.0.custom_link.# = 2",
	"widget.0.topology_map_definition.0.custom_link.0.label = Test Custom Link label",
	"widget.0.topology_map_definition.0.custom_link.0.link = https://app.datadoghq.com/dashboard/lists",
	"widget.0.topology_map_definition.0.custom_link.1.override_label = logs",
	"widget.0.topology_map_definition.0.custom_link.1.link = https://app.datadoghq.com/dashboard/lists",
	"widget.0.topology_map_definition.0.custom_link.1.is_hidden = true",
}

func TestAccDatadogDashboardServiceMap(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardTopologyMapConfig, "datadog_dashboard.topology_map_dashboard", datadogDashboardTopologyMapAsserts)
}

func TestAccDatadogDashboardServiceMap_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardTopologyMapConfig, "datadog_dashboard.topology_map_dashboard")
}
