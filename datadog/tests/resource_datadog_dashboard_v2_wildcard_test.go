package test

import (
	"testing"
)

const datadogDashboardWildcardConfig = `
resource "datadog_dashboard_v2" "wildcard_dashboard" {
    title       = "{{uniq}}"
    layout_type = "ordered"
    widget {
        wildcard_definition {
            title = "Custom Vega-Lite Chart"
            specification {
                type     = "vega-lite"
                contents = jsonencode({
                    "$schema" = "https://vega.github.io/schema/vega-lite/v5.json"
                    mark      = "bar"
                    data      = { name = "table1" }
                    encoding  = {
                        x = { field = "env", type = "nominal", sort = "-y" }
                        y = { field = "query1", type = "quantitative" }
                    }
                })
            }
            request {
                formula {
                    formula_expression = "query1"
                }
                query {
                    metric_query {
                        data_source = "metrics"
                        name        = "query1"
                        query       = "avg:system.cpu.user{*} by {env}"
                    }
                }
                response_format = "scalar"
            }
        }
    }
}
`

var datadogDashboardWildcardAsserts = []string{
	"title = {{uniq}}",
	"widget.0.wildcard_definition.0.title = Custom Vega-Lite Chart",
	"widget.0.wildcard_definition.0.specification.0.type = vega-lite",
	"widget.0.wildcard_definition.0.request.0.formula.0.formula_expression = query1",
	"widget.0.wildcard_definition.0.request.0.query.0.metric_query.0.data_source = metrics",
	"widget.0.wildcard_definition.0.request.0.query.0.metric_query.0.name = query1",
	"widget.0.wildcard_definition.0.request.0.query.0.metric_query.0.query = avg:system.cpu.user{*} by {env}",
}

func TestAccDatadogDashboardV2Wildcard(t *testing.T) {
	config, name := datadogDashboardWildcardConfig, "datadog_dashboard_v2.wildcard_dashboard"
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardV2Wildcard", config, name, datadogDashboardWildcardAsserts)
}

func TestAccDatadogDashboardV2Wildcard_import(t *testing.T) {
	config, name := datadogDashboardWildcardConfig, "datadog_dashboard_v2.wildcard_dashboard"
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardV2Wildcard_import", config, name)
}
