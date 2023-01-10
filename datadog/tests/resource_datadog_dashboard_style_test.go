package test

import "testing"

const datadogDashboardStyleConfig = `
resource "datadog_dashboard" "style_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"

	widget {
		timeseries_definition {
			request {
				formula {
					formula_expression = "query"
					style {
						palette = "warm"
						palette_index = 2
					}
				}
			}
		}
	}
}
`

var datadogDashboardStyleAsserts = []string{
	"widget.0.service_level_objective_definition.0.title_size = 16",
	"is_read_only = true",
	"title = {{uniq}}",
	"description = Created using the Datadog provider in Terraform",
	"layout_type = ordered",
	"widget.0.timeseries_definition.request.formula.formula_expression = query",
	"widget.0.timeseries_definition.request.formula.style.palette = warm",
	"widget.0.timeseries_definition.request.formula.style.palette_index = 2",
}

func TestAccDatadogDashboardStyle(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardStyleConfig, "datadog_dashboard.style_dashboard", datadogDashboardStyleAsserts)
}

func TestAccDatadogDashboardStyle_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardStyleConfig, "datadog_dashboard.style_dashboard")
}
