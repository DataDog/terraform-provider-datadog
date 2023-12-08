package test

import (
	"testing"
)

const datadogPowerpackEventStreamTest = `
resource "datadog_powerpack" "event_stream_powerpack" {
	name = "{{uniq}}"
    tags = ["tag:foo1"]
	description   = "Created using the Datadog provider in Terraform"
	live_span = "4h"
	template_variables {
		defaults = ["defaults"]
		name     = "datacenter"
	}
    widget {
		event_stream_definition {
		  query       = "*"
		  event_size  = "l"
		  title       = "Widget Title"
		  title_size  = 16
		  title_align = "right"
		}
	}
}
`

var datadogPowerpackEventStreamTestAsserts = []string{
	// Powerpack metadata
	"description = Created using the Datadog provider in Terraform",
	"widget.# = 1",
	"tags.# = 1",
	"tags.0 = tag:foo1",
	"live_span = 4h",
	// Event Stream widget
	"widget.0.event_stream_definition.0.query = *",
	"widget.0.event_stream_definition.0.event_size = l",
	"widget.0.event_stream_definition.0.title = Widget Title",
	"widget.0.event_stream_definition.0.title_size = 16",
	"widget.0.event_stream_definition.0.title_align = right",
	// Template Variables
	"template_variables.# = 1",
	"template_variables.0.name = datacenter",
	"template_variables.0.defaults.# = 1",
	"template_variables.0.defaults.0 = defaults",
}

func TestAccDatadogPowerpackEventStream(t *testing.T) {
	testAccDatadogPowerpackWidgetUtil(t, datadogPowerpackEventStreamTest, "datadog_powerpack.event_stream_powerpack", datadogPowerpackEventStreamTestAsserts)
}
