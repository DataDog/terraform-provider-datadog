package test

import (
	"testing"
)

const datadogPowerpackSloListTest = `
resource "datadog_powerpack" "slo_list_powerpack" {
	name         = "{{uniq}}"
    tags = ["tag:foo1"]
	description   = "Created using the Datadog provider in Terraform"
	template_variables {
		defaults = ["defaults"]
		name     = "datacenter"
	}
	widget {
		slo_list_definition {
			request {
				request_type = "slo_list"
				query {
					query_string = "env:prod AND service:my-app"
					limit = 30

					sort {
						column = "status.sli"
						order = "desc"
					}

				}
			}
			title = "my title"
			title_size = "16"
			title_align = "center"
		}
	}
}
`

var datadogPowerpackSloListTestAsserts = []string{
	// Powerpack metadata
	"name = {{uniq}}",
	"description = Created using the Datadog provider in Terraform",
	"widget.# = 1",
	"tags.# = 1",
	"tags.0 = tag:foo1",
	// Slo List widget
	"widget.0.slo_list_definition.0.request.0.query.0.query_string = env:prod AND service:my-app",
	"widget.0.slo_list_definition.0.request.0.query.0.limit = 30",
	"widget.0.slo_list_definition.0.title = my title",
	"widget.0.slo_list_definition.0.title_size = 16",
	"widget.0.slo_list_definition.0.title_align = center",
	"widget.0.slo_list_definition.0.request.0.query.0.sort.0.column = status.sli",
	"widget.0.slo_list_definition.0.request.0.query.0.sort.0.order = desc",
	// Template Variables
	"template_variables.# = 1",
	"template_variables.0.name = datacenter",
	"template_variables.0.defaults.# = 1",
	"template_variables.0.defaults.0 = defaults",
}

func TestAccDatadogPowerpackSloList(t *testing.T) {
	testAccDatadogPowerpackWidgetUtil(t, datadogPowerpackSloListTest, "datadog_powerpack.slo_list_powerpack", datadogPowerpackSloListTestAsserts)
}
