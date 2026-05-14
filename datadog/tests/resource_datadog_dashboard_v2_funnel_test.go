package test

import (
	"testing"
)

const datadogDashboardV2FunnelConfig = `
resource "datadog_dashboard_v2" "funnel_dashboard" {
    title       = "{{uniq}}"
    layout_type = "ordered"
    widget {
        funnel_definition {
            request {
                query {
                    query_string = "@browser.name:Chrome"
                    data_source  = "rum"
                    step {
                        facet = "@view.name"
                        value = "/home"
                    }
                    step {
                        facet = "@view.name"
                        value = "/checkout"
                    }
                }
            }
            title = "Browser Funnel"
        }
    }
}
`

var datadogDashboardV2FunnelAsserts = []string{
	"title = {{uniq}}",
	"widget.0.funnel_definition.0.title = Browser Funnel",
	"widget.0.funnel_definition.0.request.0.query.0.query_string = @browser.name:Chrome",
	"widget.0.funnel_definition.0.request.0.query.0.data_source = rum",
	"widget.0.funnel_definition.0.request.0.query.0.step.0.facet = @view.name",
	"widget.0.funnel_definition.0.request.0.query.0.step.0.value = /home",
	"widget.0.funnel_definition.0.request.0.query.0.step.1.facet = @view.name",
	"widget.0.funnel_definition.0.request.0.query.0.step.1.value = /checkout",
}

func TestAccDatadogDashboardV2Funnel(t *testing.T) {
	config, name := datadogDashboardV2FunnelConfig, "datadog_dashboard_v2.funnel_dashboard"
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardV2Funnel", config, name, datadogDashboardV2FunnelAsserts)
}

func TestAccDatadogDashboardV2Funnel_import(t *testing.T) {
	config, name := datadogDashboardV2FunnelConfig, "datadog_dashboard_v2.funnel_dashboard"
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardV2Funnel_import", config, name)
}
