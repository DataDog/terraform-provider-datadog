package test

import (
	"testing"
)

const DatadogPowerpackQueryValueTest = `
resource "datadog_powerpack" "query_value_powerpack" {
	name         = "{{uniq}}"
    tags = ["tag:foo1"]
	description   = "Created using the Datadog provider in Terraform"
	template_variables {
		defaults = ["defaults"]
		name     = "datacenter"
	}
      widget {
        query_value_definition {
          request {
            q          = "avg:system.load.1{env:staging} by {account}"
            aggregator = "sum"
            conditional_formats {
              comparator = "<"
              value      = "2"
              palette    = "white_on_green"
            }
            conditional_formats {
              comparator = ">"
              value      = "2.2"
              palette    = "white_on_red"
            }
          }
          autoscale   = true
          custom_unit = "xx"
          precision   = "4"
          text_align  = "right"
          title       = "Widget Title"
        }
      }
}
`

var DatadogPowerpackQueryValueTestAsserts = []string{
	// Powerpack metadata
	"description = Created using the Datadog provider in Terraform",
	"widget.# = 1",
	"tags.# = 1",
	"tags.0 = tag:foo1",
	// Change widget
	"widget.0.query_value_definition.0.request.0.q = avg:system.load.1{env:staging} by {account}",
	"widget.0.query_value_definition.0.request.0.aggregator = sum",
	"widget.0.query_value_definition.0.request.0.conditional_formats.0.comparator = <",
	"widget.0.query_value_definition.0.request.0.conditional_formats.0.value = 2",
	"widget.0.query_value_definition.0.request.0.conditional_formats.0.palette = white_on_green",
	"widget.0.query_value_definition.0.request.0.conditional_formats.1.comparator = >",
	"widget.0.query_value_definition.0.request.0.conditional_formats.1.value = 2.2",
	"widget.0.query_value_definition.0.request.0.conditional_formats.1.palette = white_on_red",
	"widget.0.query_value_definition.0.title = Widget Title",
	"widget.0.query_value_definition.0.title = Widget Title",
	"widget.0.query_value_definition.0.title = Widget Title",
	"widget.0.query_value_definition.0.title = Widget Title",
	"widget.0.query_value_definition.0.title = Widget Title",
	// Template Variables
	"template_variables.# = 1",
	"template_variables.0.name = datacenter",
	"template_variables.0.defaults.# = 1",
	"template_variables.0.defaults.0 = defaults",
}

func TestAccDatadogPowerpackQueryValue(t *testing.T) {
	testAccDatadogPowerpackWidgetUtil(t, DatadogPowerpackQueryValueTest, "datadog_powerpack.query_value_powerpack", DatadogPowerpackQueryValueTestAsserts)
}
