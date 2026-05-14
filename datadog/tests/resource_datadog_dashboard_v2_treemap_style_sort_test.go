package test

import (
	"testing"
)

// Treemap requests in datadog_dashboard_v2 declare a `style` and `sort` block.
// Both fields are emitted by the formula request flattener (treemap falls
// through to scalarFormulaRequestConfig); this test exercises the round-trip
// against a recorded cassette to confirm the schema additions hold end-to-end.

const datadogDashboardV2TreemapStyleSortConfig = `
resource "datadog_dashboard_v2" "treemap_style_sort_dashboard" {
    title       = "{{uniq}}"
    description = "Created using the Datadog provider in Terraform"
    layout_type = "ordered"
    widget {
        treemap_definition {
            title = "Memory by service"
            request {
                formula {
                    formula_expression = "query1"
                }
                query {
                    metric_query {
                        data_source = "metrics"
                        name        = "query1"
                        query       = "sum:system.mem.total{*} by {service}"
                        aggregator  = "sum"
                    }
                }
                style {
                    palette = "datadog16"
                }
                sort {
                    count = 10
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

var datadogDashboardV2TreemapStyleSortAsserts = []string{
	"title = {{uniq}}",
	"layout_type = ordered",
	"widget.0.treemap_definition.0.title = Memory by service",
	"widget.0.treemap_definition.0.request.0.formula.0.formula_expression = query1",
	"widget.0.treemap_definition.0.request.0.query.0.metric_query.0.data_source = metrics",
	"widget.0.treemap_definition.0.request.0.query.0.metric_query.0.name = query1",
	"widget.0.treemap_definition.0.request.0.query.0.metric_query.0.query = sum:system.mem.total{*} by {service}",
	"widget.0.treemap_definition.0.request.0.query.0.metric_query.0.aggregator = sum",
	"widget.0.treemap_definition.0.request.0.style.0.palette = datadog16",
	"widget.0.treemap_definition.0.request.0.sort.0.count = 10",
	"widget.0.treemap_definition.0.request.0.sort.0.order_by.0.formula_sort.0.index = 0",
	"widget.0.treemap_definition.0.request.0.sort.0.order_by.0.formula_sort.0.order = desc",
}

func TestAccDatadogDashboardV2TreemapStyleSort(t *testing.T) {
	config, name := datadogDashboardV2TreemapStyleSortConfig, "datadog_dashboard_v2.treemap_style_sort_dashboard"
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardV2TreemapStyleSort", config, name, datadogDashboardV2TreemapStyleSortAsserts)
}

func TestAccDatadogDashboardV2TreemapStyleSort_import(t *testing.T) {
	config, name := datadogDashboardV2TreemapStyleSortConfig, "datadog_dashboard_v2.treemap_style_sort_dashboard"
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardV2TreemapStyleSort_import", config, name)
}
