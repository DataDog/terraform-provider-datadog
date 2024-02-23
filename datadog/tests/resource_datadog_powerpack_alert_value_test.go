package test

import (
	"testing"
)

const datadogPowerpackAlertValueTest = `
resource "datadog_powerpack" "alert_value_powerpack" {
	name         = "{{uniq}}"
    tags = ["tag:foo1"]
	description   = "Created using the Datadog provider in Terraform"
	template_variables {
		defaults = ["defaults"]
		name     = "datacenter"
	}
	widget {
		alert_value_definition {
			alert_id = "895605"
		}
	}
	widget {
		alert_value_definition {
			alert_id = "895606"
			precision = 1
			unit = "b"
			title_size = "16"
			title_align = "center"
			title = "Widget Title"
			text_align = "center"
		}
	}
}
`

var datadogPowerpackAlertValueTestAsserts = []string{
	// Powerpack metadata
	"description = Created using the Datadog provider in Terraform",
	"widget.# = 2",
	"tags.# = 1",
	"tags.0 = tag:foo1",

	// Alert Value widgets
	"widget.0.alert_value_definition.0.title_align =",
	"widget.1.alert_value_definition.0.title_align = center",
	"widget.1.alert_value_definition.0.text_align = center",
	"widget.1.layout.% = 0",
	"widget.0.alert_value_definition.0.precision = 0",
	"widget.1.alert_value_definition.0.title_size = 16",
	"widget.1.alert_value_definition.0.precision = 1",
	"widget.0.alert_value_definition.0.title_size =",
	"widget.1.alert_value_definition.0.alert_id = 895606",
	"widget.0.alert_value_definition.0.text_align =",
	"widget.0.alert_value_definition.0.title =",
	"widget.0.alert_value_definition.0.unit =",
	"widget.1.alert_value_definition.0.title = Widget Title",
	"widget.0.alert_value_definition.0.alert_id = 895605",
	"widget.1.alert_value_definition.0.unit = b",

	// Template Variables
	"template_variables.# = 1",
	"template_variables.0.name = datacenter",
	"template_variables.0.defaults.# = 1",
	"template_variables.0.defaults.0 = defaults",
}

func TestAccDatadogPowerpackAlertValue(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogPowerpackAlertValueTest, "datadog_powerpack.alert_value_powerpack", datadogPowerpackAlertValueTestAsserts)
}
