package test

import "testing"

const datadogDashboardV2QueryTableSortConfig = `
resource "datadog_dashboard_v2" "query_table_sort_dashboard" {
  title       = "{{uniq}}"
  layout_type = "ordered"

  widget {
    query_table_definition {
      title = "Formula request sort"

      request {
        formula {
          formula_expression = "query1"
        }

        query {
          metric_query {
            data_source = "metrics"
            name        = "query1"
            query       = "avg:system.cpu.user{*} by {host}"
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

  widget {
    query_table_definition {
      title = "Legacy request sort"

      request {
        q = "avg:system.mem.used{*} by {host}"

        sort {
          count = 10

          order_by {
            group_sort {
              name  = "host"
              order = "asc"
            }
          }
        }
      }
    }
  }
}
`

var datadogDashboardV2QueryTableSortAsserts = []string{
	"title = {{uniq}}",
	"widget.0.query_table_definition.0.request.0.sort.0.count = 25",
	"widget.0.query_table_definition.0.request.0.sort.0.order_by.0.formula_sort.0.index = 0",
	"widget.0.query_table_definition.0.request.0.sort.0.order_by.0.formula_sort.0.order = desc",
	"widget.1.query_table_definition.0.request.0.sort.0.count = 10",
	"widget.1.query_table_definition.0.request.0.sort.0.order_by.0.group_sort.0.name = host",
	"widget.1.query_table_definition.0.request.0.sort.0.order_by.0.group_sort.0.order = asc",
}

func TestAccDatadogDashboardV2QueryTableSort(t *testing.T) {
	config, name := datadogDashboardV2QueryTableSortConfig, "datadog_dashboard_v2.query_table_sort_dashboard"
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardV2QueryTableSort", config, name, datadogDashboardV2QueryTableSortAsserts)
}

func TestAccDatadogDashboardV2QueryTableSort_import(t *testing.T) {
	config, name := datadogDashboardV2QueryTableSortConfig, "datadog_dashboard_v2.query_table_sort_dashboard"
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardV2QueryTableSort_import", config, name)
}
