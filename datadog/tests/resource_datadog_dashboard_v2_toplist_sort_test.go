package test

import (
	"testing"
)

const datadogDashboardV2ToplistSortConfig = `
resource "datadog_dashboard_v2" "toplist_sort_dashboard" {
    title       = "{{uniq}}"
    layout_type = "ordered"
    widget {
        toplist_definition {
            title = "Top Hosts by CPU"
            request {
                formula {
                    formula_expression = "query1"
                }
                query {
                    metric_query {
                        name        = "query1"
                        query       = "avg:system.cpu.user{*} by {host}"
                        data_source = "metrics"
                        aggregator  = "avg"
                    }
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

var datadogDashboardV2ToplistSortAsserts = []string{
	"title = {{uniq}}",
	"widget.0.toplist_definition.0.title = Top Hosts by CPU",
	"widget.0.toplist_definition.0.request.0.formula.0.formula_expression = query1",
	"widget.0.toplist_definition.0.request.0.query.0.metric_query.0.name = query1",
	"widget.0.toplist_definition.0.request.0.query.0.metric_query.0.query = avg:system.cpu.user{*} by {host}",
	"widget.0.toplist_definition.0.request.0.sort.0.count = 10",
	"widget.0.toplist_definition.0.request.0.sort.0.order_by.0.formula_sort.0.index = 0",
	"widget.0.toplist_definition.0.request.0.sort.0.order_by.0.formula_sort.0.order = desc",
}

func TestAccDatadogDashboardV2ToplistSort(t *testing.T) {
	config, name := datadogDashboardV2ToplistSortConfig, "datadog_dashboard_v2.toplist_sort_dashboard"
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardV2ToplistSort", config, name, datadogDashboardV2ToplistSortAsserts)
}

func TestAccDatadogDashboardV2ToplistSort_import(t *testing.T) {
	config, name := datadogDashboardV2ToplistSortConfig, "datadog_dashboard_v2.toplist_sort_dashboard"
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardV2ToplistSort_import", config, name)
}
