package test

import (
	"testing"
)

const datadogPowerpackToplistTest = `
resource "datadog_powerpack" "toplist_powerpack" {
	name         = "{{uniq}}"
    tags = ["tag:foo1"]
	description   = "Created using the Datadog provider in Terraform"
	template_variables {
		defaults = ["defaults"]
		name     = "datacenter"
	}
  widget {
    toplist_definition {
      request {
        q = "avg:system.cpu.user{app:general} by {datacenter}"
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
      title = "Widget Title"
    }
  }
}
`

var datadogPowerpackToplistTestAsserts = []string{
	// Powerpack metadata
	"description = Created using the Datadog provider in Terraform",
	"widget.# = 1",
	"tags.# = 1",
	"tags.0 = tag:foo1",
	// Toplist widget
	"widget.0.toplist_definition.0.request.0.q = avg:system.cpu.user{app:general} by {datacenter}",
	"widget.0.toplist_definition.0.request.0.conditional_formats.0.comparator = <",
	"widget.0.toplist_definition.0.request.0.conditional_formats.0.value = 2",
	"widget.0.toplist_definition.0.request.0.conditional_formats.0.palette = white_on_green",
	"widget.0.toplist_definition.0.request.0.conditional_formats.1.comparator = >",
	"widget.0.toplist_definition.0.request.0.conditional_formats.1.value = 2.2",
	"widget.0.toplist_definition.0.request.0.conditional_formats.1.palette = white_on_red",
	"widget.0.toplist_definition.0.title = Widget Title",
	// Template Variables
	"template_variables.# = 1",
	"template_variables.0.name = datacenter",
	"template_variables.0.defaults.# = 1",
	"template_variables.0.defaults.0 = defaults",
}

func TestAccDatadogPowerpackToplist(t *testing.T) {
	testAccDatadogPowerpackWidgetUtil(t, datadogPowerpackToplistTest, "datadog_powerpack.toplist_powerpack", datadogPowerpackToplistTestAsserts)
}
