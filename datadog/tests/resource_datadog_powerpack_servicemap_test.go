package test

import (
	"testing"
)

const datadogPowerpackServiceMapTest = `
resource "datadog_powerpack" "servicemap_powerpack" {
	name         = "{{uniq}}"
    tags = ["tag:foo1"]
	description   = "Created using the Datadog provider in Terraform"
	template_variables {
		defaults = ["defaults"]
		name     = "datacenter"
	}
    widget {
      servicemap_definition {
        service     = "master-db"
        filters     = ["env:prod"]
        title       = "env: prod, datacenter:dc1, service: master-db"
        title_size  = "16"
        title_align = "left"
      }
    }
}
`

var datadogPowerpackServiceMapTestAsserts = []string{
	// Powerpack metadata
	"description = Created using the Datadog provider in Terraform",
	"widget.# = 1",
	"tags.# = 1",
	"tags.0 = tag:foo1",
	// Servicemap widget
	"widget.0.servicemap_definition.0.service = master-db",
	"widget.0.servicemap_definition.0.filters.# = 1",
	"widget.0.servicemap_definition.0.filters.0 = env:prod",
	"widget.0.servicemap_definition.0.title = env: prod, datacenter:dc1, service: master-db",
	"widget.0.servicemap_definition.0.title_size = 16",
	"widget.0.servicemap_definition.0.title_align = left",

	// Template Variables
	"template_variables.# = 1",
	"template_variables.0.name = datacenter",
	"template_variables.0.defaults.# = 1",
	"template_variables.0.defaults.0 = defaults",
}

func TestAccDatadogPowerpackServicemap(t *testing.T) {
	testAccDatadogPowerpackWidgetUtil(t, datadogPowerpackServiceMapTest, "datadog_powerpack.servicemap_powerpack", datadogPowerpackServiceMapTestAsserts)
}
