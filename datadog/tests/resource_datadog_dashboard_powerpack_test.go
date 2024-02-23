package test

import "testing"

const datadogDashboardPowerpackConfig = `
resource "datadog_dashboard" "powerpack_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"

  widget {
    powerpack_definition {
      powerpack_id = "3c3096da-7cec-11ee-9b6f-da7ad0900002"
      background_color = "blue"
      banner_img = "https://imgix.datadoghq.com/img/about/presskit/logo-v/dd_vertical_white.png"
      show_title = true
      title = "Powerpack Widget Test"
      template_variables {
        controlled_by_powerpack {
          name = "var"
          values = ["default", "values", "here"]
          prefix= "pre"
        }
        controlled_externally {
          name = "test"
          values = ["test"]
          prefix= "dc"
        }
      }
    }
  }
}
`

var datadogDashboardPowerpackAsserts = []string{
	"description = Created using the Datadog provider in Terraform",
	"title = {{uniq}}",
	"widget.0.powerpack_definition.0.background_color = blue",
	"widget.0.powerpack_definition.0.title = Powerpack Widget Test",
	"widget.0.powerpack_definition.0.show_title = true",
	"widget.0.powerpack_definition.0.banner_img = https://imgix.datadoghq.com/img/about/presskit/logo-v/dd_vertical_white.png",
	"widget.0.powerpack_definition.0.powerpack_id = 3c3096da-7cec-11ee-9b6f-da7ad0900002",
	"widget.0.powerpack_definition.0.template_variables.0.controlled_by_powerpack.0.name = var",
	"widget.0.powerpack_definition.0.template_variables.0.controlled_by_powerpack.0.prefix = pre",
	"widget.0.powerpack_definition.0.template_variables.0.controlled_by_powerpack.0.values.# = 3",
	"widget.0.powerpack_definition.0.template_variables.0.controlled_by_powerpack.0.values.0 = default",
	"widget.0.powerpack_definition.0.template_variables.0.controlled_by_powerpack.0.values.1 = values",
	"widget.0.powerpack_definition.0.template_variables.0.controlled_by_powerpack.0.values.2 = here",
	"widget.0.powerpack_definition.0.template_variables.0.controlled_externally.0.name = test",
	"widget.0.powerpack_definition.0.template_variables.0.controlled_externally.0.values.0 = test",
	"widget.0.powerpack_definition.0.template_variables.0.controlled_externally.0.prefix = dc",

	"layout_type = ordered",
	"is_read_only = true",
}

func TestAccDatadogDashboardPowerpack(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardPowerpackConfig, "datadog_dashboard.powerpack_dashboard", datadogDashboardPowerpackAsserts)
}
