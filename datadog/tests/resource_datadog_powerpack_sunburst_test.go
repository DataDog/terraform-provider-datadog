package test

import (
	"testing"
)

const datadogPowerpackSunburstTest = `
resource "datadog_powerpack" "sunburst_powerpack" {
	name         = "{{uniq}}"
    tags = ["tag:foo1"]
	description   = "Created using the Datadog provider in Terraform"
	template_variables {
		defaults = ["defaults"]
		name     = "datacenter"
	}
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

}
`

var datadogPowerpackSunburstTestAsserts = []string{
	// Powerpack metadata
	"name = {{uniq}}",
	"description = Created using the Datadog provider in Terraform",
	"widget.# = 2",
	"tags.# = 1",
	"tags.0 = tag:foo1",
	// Sunburst widget
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
	// Template Variables
	"template_variables.# = 1",
	"template_variables.0.name = datacenter",
	"template_variables.0.defaults.# = 1",
	"template_variables.0.defaults.0 = defaults",
}

func TestAccDatadogPowerpackSunburst(t *testing.T) {
	testAccDatadogPowerpackWidgetUtil(t, datadogPowerpackSunburstTest, "datadog_powerpack.sunburst_powerpack", datadogPowerpackSunburstTestAsserts)
}
