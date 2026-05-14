package test

import (
	"testing"
)

// Sunburst requests in datadog_dashboard_v2 declare a `sort` block. The field
// is emitted by the formula request flattener (sunburst falls through to
// scalarFormulaRequestConfig); this test exercises the round-trip against a
// recorded cassette to confirm the schema addition holds end-to-end.

const datadogDashboardV2SunburstSortConfig = `
resource "datadog_dashboard_v2" "sunburst_sort_dashboard" {
    title       = "{{uniq}}"
    description = "Created using the Datadog provider in Terraform"
    layout_type = "ordered"
    widget {
        sunburst_definition {
            title = "Hits by host"
            request {
                formula {
                    formula_expression = "query1"
                }
                query {
                    metric_query {
                        data_source = "metrics"
                        name        = "query1"
                        query       = "sum:system.cpu.user{*} by {host}"
                        aggregator  = "sum"
                    }
                }
                sort {
                    count = 25
                    order_by {
                        formula_sort {
                            index = 0
                            order = "desc"
                        }
                    }
                }
            }
        }
    }
}
`

var datadogDashboardV2SunburstSortAsserts = []string{
	"title = {{uniq}}",
	"layout_type = ordered",
	"widget.0.sunburst_definition.0.title = Hits by host",
	"widget.0.sunburst_definition.0.request.0.formula.0.formula_expression = query1",
	"widget.0.sunburst_definition.0.request.0.query.0.metric_query.0.data_source = metrics",
	"widget.0.sunburst_definition.0.request.0.query.0.metric_query.0.name = query1",
	"widget.0.sunburst_definition.0.request.0.query.0.metric_query.0.query = sum:system.cpu.user{*} by {host}",
	"widget.0.sunburst_definition.0.request.0.query.0.metric_query.0.aggregator = sum",
	"widget.0.sunburst_definition.0.request.0.sort.0.count = 25",
	"widget.0.sunburst_definition.0.request.0.sort.0.order_by.0.formula_sort.0.index = 0",
	"widget.0.sunburst_definition.0.request.0.sort.0.order_by.0.formula_sort.0.order = desc",
}

func TestAccDatadogDashboardV2SunburstSort(t *testing.T) {
	config, name := datadogDashboardV2SunburstSortConfig, "datadog_dashboard_v2.sunburst_sort_dashboard"
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardV2SunburstSort", config, name, datadogDashboardV2SunburstSortAsserts)
}

func TestAccDatadogDashboardV2SunburstSort_import(t *testing.T) {
	config, name := datadogDashboardV2SunburstSortConfig, "datadog_dashboard_v2.sunburst_sort_dashboard"
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardV2SunburstSort_import", config, name)
}
