package test

import (
	"testing"
)

const datadogDashboardBarChartConfig = `
resource "datadog_dashboard_v2" "bar_chart_dashboard" {
    title       = "{{uniq}}"
    layout_type = "ordered"
    widget {
        bar_chart_definition {
            title = "Avg of system.cpu.user"
            request {
                formula {
                    formula_expression = "query1"
                }
                query {
                    metric_query {
                        data_source = "metrics"
                        name        = "query1"
                        query       = "avg:system.cpu.user{*}"
                    }
                }
            }
        }
    }
}
`

var datadogDashboardBarChartAsserts = []string{
	"title = {{uniq}}",
	"widget.0.bar_chart_definition.0.title = Avg of system.cpu.user",
	"widget.0.bar_chart_definition.0.request.0.formula.0.formula_expression = query1",
	"widget.0.bar_chart_definition.0.request.0.query.0.metric_query.0.data_source = metrics",
	"widget.0.bar_chart_definition.0.request.0.query.0.metric_query.0.name = query1",
	"widget.0.bar_chart_definition.0.request.0.query.0.metric_query.0.query = avg:system.cpu.user{*}",
}

func TestAccDatadogDashboardV2BarChart(t *testing.T) {
	config, name := datadogDashboardBarChartConfig, "datadog_dashboard_v2.bar_chart_dashboard"
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardV2BarChart", config, name, datadogDashboardBarChartAsserts)
}

func TestAccDatadogDashboardV2BarChart_import(t *testing.T) {
	config, name := datadogDashboardBarChartConfig, "datadog_dashboard_v2.bar_chart_dashboard"
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardV2BarChart_import", config, name)
}
