package test

import (
	"testing"
)

const datadogPowerpackImageTest = `
resource "datadog_powerpack" "image_powerpack" {
	name         = "{{uniq}}"
    tags = ["tag:foo1"]
	description   = "Created using the Datadog provider in Terraform"
	template_variables {
		defaults = ["defaults"]
		name     = "datacenter"
	}
  widget {
	image_definition {
	  url    = "https://google.com"
	  sizing = "fit"
	  margin = "small"
	}
  }
}
`

var datadogPowerpackImageTestAsserts = []string{
	// Powerpack metadata
	"description = Created using the Datadog provider in Terraform",
	"widget.# = 1",
	"tags.# = 1",
	"tags.0 = tag:foo1",
	// Image widget
	"widget.0.image_definition.0.url = https://google.com",
	"widget.0.image_definition.0.sizing = fit",
	"widget.0.image_definition.0.margin = small",

	// Template Variables
	"template_variables.# = 1",
	"template_variables.0.name = datacenter",
	"template_variables.0.defaults.# = 1",
	"template_variables.0.defaults.0 = defaults",
}

func TestAccDatadogPowerpackImage(t *testing.T) {
	testAccDatadogPowerpackWidgetUtil(t, datadogPowerpackImageTest, "datadog_powerpack.image_powerpack", datadogPowerpackImageTestAsserts)
}
