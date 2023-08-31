package test

import (
	"testing"
)

const datadogDashboardSunburstConfig = `
resource "datadog_dashboard" "sunburst_dashboard" {
  title        = "{{uniq}}"
  description  = "Created using the Datadog provider in Terraform"
  layout_type  = "ordered"
  widget {
    sunburst_definition {
      request {
        formula {
          formula_expression = "my_query_1 + my_query_2"
          alias              = "my ff query"
        }
        query {
          metric_query {
            data_source = "metrics"
            query       = "avg:system.cpu.user{foo:bar} by {env}"
            name        = "my_query_1"
            aggregator  = "sum"
          }
        }
        style {
          palette = "dog_classic"
        }
      }
      hide_total = false
      legend_inline {
        type         = "automatic"
        hide_value   = true
        hide_percent = false
      }
      custom_link {
        link  = "https://app.datadoghq.com/dashboard/lists"
        label = "Test Custom Link label"
      }
    }
  }
  widget {
    sunburst_definition {
      request {
        query {
          event_query {
            data_source = "rum"
            indexes     = ["*"]
            name        = "query1"
            compute {
              aggregation = "count"
            }
            search {
              query = "abc"
            }
          }
        }
      }
      hide_total = true
      legend_table {
        type = "table"
      }
    }
  }
  widget {
    group_definition {
      title       = "{{uniq}}"
      layout_type = "ordered"
      widget {
        sunburst_definition {
          request {
            query {
              event_query {
                data_source = "rum"
                indexes     = ["*"]
                name        = "query1"
                compute {
                  aggregation = "count"
                }
                search {
                  query = "abc"
                }
              }
            }
          }
          hide_total = true
          legend_table {
            type = "table"
          }
        }
      }
    }
  }
}
`

var datadogDashboardSunburstAsserts = []string{
	"description = Created using the Datadog provider in Terraform",
	"layout_type = ordered",
	"title = {{uniq}}",
	"widget.0.sunburst_definition.0.request.0.formula.0.formula_expression = my_query_1 + my_query_2",
	"widget.0.sunburst_definition.0.request.0.formula.0.alias = my ff query",
	"widget.0.sunburst_definition.0.request.0.query.0.metric_query.0.data_source = metrics",
	"widget.0.sunburst_definition.0.request.0.query.0.metric_query.0.query = avg:system.cpu.user{foo:bar} by {env}",
	"widget.0.sunburst_definition.0.request.0.query.0.metric_query.0.name = my_query_1",
	"widget.0.sunburst_definition.0.request.0.query.0.metric_query.0.aggregator = sum",
	"widget.0.sunburst_definition.0.request.0.style.0.palette = dog_classic",
	"widget.0.sunburst_definition.0.hide_total = false",
	"widget.0.sunburst_definition.0.legend_inline.0.type = automatic",
	"widget.0.sunburst_definition.0.legend_inline.0.hide_value = true",
	"widget.0.sunburst_definition.0.legend_inline.0.hide_percent = false",
	"widget.0.sunburst_definition.0.legend_inline.0.hide_percent = false",
	"widget.0.sunburst_definition.0.custom_link.# = 1",
	"widget.0.sunburst_definition.0.custom_link.0.label = Test Custom Link label",
	"widget.0.sunburst_definition.0.custom_link.0.link = https://app.datadoghq.com/dashboard/lists",
	"widget.1.sunburst_definition.0.request.0.query.0.event_query.0.data_source = rum",
	"widget.1.sunburst_definition.0.request.0.query.0.event_query.0.indexes.# = 1",
	"widget.1.sunburst_definition.0.request.0.query.0.event_query.0.indexes.0 = *",
	"widget.1.sunburst_definition.0.request.0.query.0.event_query.0.name = query1",
	"widget.1.sunburst_definition.0.request.0.query.0.event_query.0.compute.0.aggregation = count",
	"widget.1.sunburst_definition.0.request.0.query.0.event_query.0.search.0.query = abc",
	"widget.1.sunburst_definition.0.hide_total = true",
	"widget.1.sunburst_definition.0.legend_table.0.type = table",
	"widget.2.group_definition.0.widget.0.sunburst_definition.0.request.0.query.0.event_query.0.data_source = rum",
	"widget.2.group_definition.0.widget.0.sunburst_definition.0.request.0.query.0.event_query.0.indexes.# = 1",
	"widget.2.group_definition.0.widget.0.sunburst_definition.0.request.0.query.0.event_query.0.indexes.0 = *",
	"widget.2.group_definition.0.widget.0.sunburst_definition.0.request.0.query.0.event_query.0.name = query1",
	"widget.2.group_definition.0.widget.0.sunburst_definition.0.request.0.query.0.event_query.0.compute.0.aggregation = count",
	"widget.2.group_definition.0.widget.0.sunburst_definition.0.request.0.query.0.event_query.0.search.0.query = abc",
	"widget.2.group_definition.0.widget.0.sunburst_definition.0.hide_total = true",
	"widget.2.group_definition.0.widget.0.sunburst_definition.0.legend_table.0.type = table",
}

func TestAccDatadogDashboardSunburst(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardSunburstConfig, "datadog_dashboard.sunburst_dashboard", datadogDashboardSunburstAsserts)
}

func TestAccDatadogDashboardSunburst_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardSunburstConfig, "datadog_dashboard.sunburst_dashboard")
}
