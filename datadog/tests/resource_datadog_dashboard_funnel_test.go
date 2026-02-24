package test

import (
	"testing"
)

const datadogDashboardFunnelConfig = `
resource "datadog_dashboard" "funnel_dashboard" {
	title       = "{{uniq}}"
	layout_type = "ordered"

	widget {
		funnel_definition {
			request {
				request_type = "funnel"
				query {
					data_source  = "rum"
					query_string = "@browser.name:Chrome"

					step {
						facet = "@view.name"
						value = "/logs"
					}

					step {
						facet = "@view.name"
						value = "/apm/home"
					}
				}
			}
			title       = "Funnel Widget"
			title_size  = "16"
			title_align = "left"
		}
	}
}
`

var datadogDashboardFunnelAsserts = []string{
	"title = {{uniq}}",
	"layout_type = ordered",
	"widget.0.funnel_definition.0.request.0.request_type = funnel",
	"widget.0.funnel_definition.0.request.0.query.0.data_source = rum",
	"widget.0.funnel_definition.0.request.0.query.0.query_string = @browser.name:Chrome",
	"widget.0.funnel_definition.0.request.0.query.0.step.0.facet = @view.name",
	"widget.0.funnel_definition.0.request.0.query.0.step.0.value = /logs",
	"widget.0.funnel_definition.0.request.0.query.0.step.1.facet = @view.name",
	"widget.0.funnel_definition.0.request.0.query.0.step.1.value = /apm/home",
	"widget.0.funnel_definition.0.title = Funnel Widget",
	"widget.0.funnel_definition.0.title_size = 16",
	"widget.0.funnel_definition.0.title_align = left",
}

func TestAccDatadogDashboardFunnel(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardFunnelConfig, "datadog_dashboard.funnel_dashboard", datadogDashboardFunnelAsserts)
}

func TestAccDatadogDashboardFunnel_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardFunnelConfig, "datadog_dashboard.funnel_dashboard")
}
