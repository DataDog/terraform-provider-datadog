package test

import (
	"testing"
)

const datadogPowerpackChangeTest = `
resource "datadog_powerpack" "change_powerpack" {
	name         = "{{uniq}}"
    tags = ["tag:foo1"]
	description   = "Created using the Datadog provider in Terraform"
	template_variables {
		defaults = ["defaults"]
		name     = "datacenter"
	}
  	widget {
		change_definition {
		  request {
			q             = "avg:system.load.1{env:staging} by {account}"
			change_type   = "absolute"
			compare_to    = "week_before"
			increase_good = true
			order_by      = "name"
			order_dir     = "desc"
			show_present  = false
		  }
		  title     = "Widget Title"
		}
  	}
}
`

var datadogPowerpackChangeTestAsserts = []string{
	// Powerpack metadata
	"description = Created using the Datadog provider in Terraform",
	"widget.# = 1",
	"tags.# = 1",
	"tags.0 = tag:foo1",
	// Change widget
	"widget.0.change_definition.0.request.0.q = avg:system.load.1{env:staging} by {account}",
	"widget.0.change_definition.0.request.0.change_type = absolute",
	"widget.0.change_definition.0.request.0.compare_to = week_before",
	"widget.0.change_definition.0.request.0.increase_good = true",
	"widget.0.change_definition.0.request.0.order_by = name",
	"widget.0.change_definition.0.request.0.order_dir = desc",
	"widget.0.change_definition.0.request.0.show_present = false",
	"widget.0.change_definition.0.title = Widget Title",
	// Template Variables
	"template_variables.# = 1",
	"template_variables.0.name = datacenter",
	"template_variables.0.defaults.# = 1",
	"template_variables.0.defaults.0 = defaults",
}

func TestAccDatadogPowerpackChange(t *testing.T) {
	testAccDatadogPowerpackWidgetUtil(t, datadogPowerpackChangeTest, "datadog_powerpack.change_powerpack", datadogPowerpackChangeTestAsserts)
}
