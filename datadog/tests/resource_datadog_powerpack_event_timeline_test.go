package test

import (
	"testing"
)

const datadogPowerpackEventTimelineTest = `
resource "datadog_powerpack" "event_timeline_powerpack" {
	name         = "{{uniq}}"
    tags = ["tag:foo1"]
	description   = "Created using the Datadog provider in Terraform"
	template_variables {
		defaults = ["defaults"]
		name     = "datacenter"
	}
  widget {
    event_timeline_definition {
      title = "Widget Title"
      title_align = "right"
      title_size = "16"
      tags_execution = "and"
      query = "status:error"
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

var datadogPowerpackEventTimelineTestAsserts = []string{
	// Powerpack metadata
	"description = Created using the Datadog provider in Terraform",
	"widget.# = 1",
	"tags.# = 1",
	"tags.0 = tag:foo1",

	// Event timeline widgets
	"widget.0.widget_layout.0.y = 5",
	"widget.0.event_timeline_definition.0.title_align = right",
	"widget.0.widget_layout.0.x = 5",
	"widget.0.widget_layout.0.width = 3",
	"widget.0.event_timeline_definition.0.title_size = 16",
	"widget.0.event_timeline_definition.0.query = status:error",
	"widget.0.event_timeline_definition.0.title = Widget Title",
	"widget.0.event_timeline_definition.0.tags_execution = and",
	"widget.0.widget_layout.0.height = 4",

	// Template Variables
	"template_variables.# = 1",
	"template_variables.0.name = datacenter",
	"template_variables.0.defaults.# = 1",
	"template_variables.0.defaults.0 = defaults",
}

func TestAccDatadogPowerpackEventTimeline(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogPowerpackEventTimelineTest, "datadog_powerpack.event_timeline_powerpack", datadogPowerpackEventTimelineTestAsserts)
}
