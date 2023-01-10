package test

import "testing"

const datadogDashboardStyleConfig = `
resource "datadog_dashboard" "style_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"

	widget {
		timeseries_definition {
			request {
				query {
					metric_query {
						query       = "avg:system.cpu.user{app:general} by {env}"
						name        = "query"
					}
				}		  
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
	"title = {{uniq}}",
	"description = Created using the Datadog provider in Terraform",
	"layout_type = ordered",
	"widget.0.timeseries_definition.0.request.0.formula.0.formula_expression = query",
	"widget.0.timeseries_definition.0.request.0.formula.0.style.0.palette = warm",
	"widget.0.timeseries_definition.0.request.0.formula.0.style.0.palette_index = 2",
}

func TestAccDatadogDashboardStyle(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardStyleConfig, "datadog_dashboard.style_dashboard", datadogDashboardStyleAsserts)
}

func TestAccDatadogDashboardStyle_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardStyleConfig, "datadog_dashboard.style_dashboard")
}
