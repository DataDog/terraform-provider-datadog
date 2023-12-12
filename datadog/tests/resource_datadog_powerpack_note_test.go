package test

import (
	"testing"
)

const datadogPowerpackNoteTest = `
resource "datadog_powerpack" "note_powerpack" {
	name         = "{{uniq}}"
    tags = ["tag:foo1"]
	description   = "Created using the Datadog provider in Terraform"
	template_variables {
		defaults = ["defaults"]
		name     = "datacenter"
	}
  widget {
    note_definition {
      content          = "note widget"
      background_color = "blue"
    }
  }
}
`

var datadogPowerpackNoteTestAsserts = []string{
	// Powerpack metadata
	"description = Created using the Datadog provider in Terraform",
	"widget.# = 1",
	"tags.# = 1",
	"tags.0 = tag:foo1",
	// Note widget
	"widget.0.note_definition.0.content = note widget",
	"widget.0.note_definition.0.background_color = blue",
	// Template Variables
	"template_variables.# = 1",
	"template_variables.0.name = datacenter",
	"template_variables.0.defaults.# = 1",
	"template_variables.0.defaults.0 = defaults",
}

func TestAccDatadogPowerpackNote(t *testing.T) {
	testAccDatadogPowerpackWidgetUtil(t, datadogPowerpackNoteTest, "datadog_powerpack.note_powerpack", datadogPowerpackNoteTestAsserts)
}
