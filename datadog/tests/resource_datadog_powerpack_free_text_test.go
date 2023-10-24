package test

import (
	"testing"
)

const datadogPowerpackFreeTextConfig = `
resource "datadog_powerpack" "free_text_powerpack" {
	name         = "{{uniq}}"
    tags = ["tag:foo1"]
	description   = "Created using the Datadog provider in Terraform"
	
	widget {
		free_text_definition {
			color = "#eb364b"
			text = "Free Text"
			font_size = "56"
			text_align = "left"
		}
		widget_layout {
			height = 4
			width = 5
			x = 5
			y = 5
		}
	}
}
`

var datadogPowerpackFreeTextAsserts = []string{
	"tags.# = 1",
	"tags.0 = tag:foo1",
	"widget.0.widget_layout.0.y = 5",
	"widget.0.free_text_definition.0.text = Free Text",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.free_text_definition.0.font_size = 56",
	"widget.0.free_text_definition.0.color = #eb364b",
	"widget.0.widget_layout.0.width = 5",
	"widget.0.widget_layout.0.height = 4",
	"widget.0.free_text_definition.0.text_align = left",
	"widget.0.widget_layout.0.x = 5",
	"widget.0.layout.% = 0",
}

func TestAccDatadogPowerpackFreeText(t *testing.T) {
	testAccDatadogPowerpackWidgetUtil(t, datadogPowerpackFreeTextConfig, "datadog_powerpack.free_text_powerpack", datadogPowerpackFreeTextAsserts)
}
