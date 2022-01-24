package test

import (
	"testing"
)

const datadogDashboardTreemapConfig = `
resource "datadog_dashboard" "treemap_dashboard" {
  title        = "{{uniq}}"
  description  = "Created using the Datadog provider in Terraform"
  layout_type  = "ordered"
  is_read_only = true
  widget {
   	treemap_definition {
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
      }
    }
  }
}
`

var datadogDashboardTreemapAsserts = []string{
	"description = Created using the Datadog provider in Terraform",
	"layout_type = ordered",
	"title = {{uniq}}",
	"is_read_only = true",
	"widget.0.treemap_definition.0.request.0.formula.0.formula_expression = my_query_1 + my_query_2",
	"widget.0.treemap_definition.0.request.0.formula.0.alias = my ff query",
	"widget.0.treemap_definition.0.request.0.query.0.metric_query.0.data_source = metrics",
	"widget.0.treemap_definition.0.request.0.query.0.metric_query.0.query = avg:system.cpu.user{foo:bar} by {env}",
	"widget.0.treemap_definition.0.request.0.query.0.metric_query.0.name = my_query_1",
	"widget.0.treemap_definition.0.request.0.query.0.metric_query.0.aggregator = sum",
}

func TestAccDatadogDashboardTreemap(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardTreemapConfig, "datadog_dashboard.treemap_dashboard", datadogDashboardTreemapAsserts)
}

func TestAccDatadogDashboardTreemap_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardTreemapConfig, "datadog_dashboard.treemap_dashboard")
}
