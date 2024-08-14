package test

import (
	"testing"
)

const datadogPowerpackIFrameTest = `
resource "datadog_powerpack" "iframe_powerpack" {
	name         = "{{uniq}}"
    tags = ["tag:foo1"]
	description   = "Created using the Datadog provider in Terraform"
	template_variables {
		defaults = ["defaults"]
		name     = "datacenter"
	}
    widget {
      iframe_definition {
        url = "https://google.com"
      }
    }
}
`

var datadogPowerpackIFrameTestAsserts = []string{
	// Powerpack metadata
	"description = Created using the Datadog provider in Terraform",
	"widget.# = 1",
	"tags.# = 1",
	"tags.0 = tag:foo1",
	// IFrame widget
	"widget.0.iframe_definition.0.url = https://google.com",
	// Template Variables
	"template_variables.# = 1",
	"template_variables.0.name = datacenter",
	"template_variables.0.defaults.# = 1",
	"template_variables.0.defaults.0 = defaults",
}

func TestAccDatadogPowerpackIFrame(t *testing.T) {
	testAccDatadogPowerpackWidgetUtil(t, datadogPowerpackIFrameTest, "datadog_powerpack.iframe_powerpack", datadogPowerpackIFrameTestAsserts)
}
