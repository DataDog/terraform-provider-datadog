package test

import (
	"testing"
)

const datadogDashboardSankeyConfig = `
resource "datadog_dashboard_v2" "sankey_dashboard" {
    title       = "{{uniq}}"
    layout_type = "ordered"
    widget {
        sankey_definition {
            title = "RUM Sankey"
            request {
                rum_request {
                    query {
                        data_source  = "rum"
                        query_string = "@type:view"
                        mode         = "source"
                    }
                }
            }
        }
    }
}
`

var datadogDashboardSankeyAsserts = []string{
	"title = {{uniq}}",
	"widget.0.sankey_definition.0.title = RUM Sankey",
	"widget.0.sankey_definition.0.request.0.rum_request.0.query.0.data_source = rum",
	"widget.0.sankey_definition.0.request.0.rum_request.0.query.0.query_string = @type:view",
	"widget.0.sankey_definition.0.request.0.rum_request.0.query.0.mode = source",
}

func TestAccDatadogDashboardV2Sankey(t *testing.T) {
	config, name := datadogDashboardSankeyConfig, "datadog_dashboard_v2.sankey_dashboard"
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardV2Sankey", config, name, datadogDashboardSankeyAsserts)
}

func TestAccDatadogDashboardV2Sankey_import(t *testing.T) {
	config, name := datadogDashboardSankeyConfig, "datadog_dashboard_v2.sankey_dashboard"
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardV2Sankey_import", config, name)
}
