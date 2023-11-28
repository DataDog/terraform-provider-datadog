package test

import (
	"testing"
)

const datadogPowerpackTreeMapTest = `
resource "datadog_powerpack" "treemap_powerpack" {
	name         = "{{uniq}}"
    tags = ["tag:foo1"]
	description   = "Created using the Datadog provider in Terraform"
	template_variables {
		defaults = ["defaults"]
		name     = "datacenter"
	}
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

var datadogPowerpackTreeMapTestAsserts = []string{
	// Powerpack metadata
	"name = {{uniq}}",
	"description = Created using the Datadog provider in Terraform",
	"widget.# = 1",
	"tags.# = 1",
	"tags.0 = tag:foo1",
	// Treemap widget
	"widget.0.treemap_definition.0.request.0.formula.0.formula_expression = my_query_1 + my_query_2",
	"widget.0.treemap_definition.0.request.0.formula.0.alias = my ff query",
	"widget.0.treemap_definition.0.request.0.query.0.metric_query.0.data_source = metrics",
	"widget.0.treemap_definition.0.request.0.query.0.metric_query.0.query = avg:system.cpu.user{foo:bar} by {env}",
	"widget.0.treemap_definition.0.request.0.query.0.metric_query.0.name = my_query_1",
	"widget.0.treemap_definition.0.request.0.query.0.metric_query.0.aggregator = sum",
	// Template Variables
	"template_variables.# = 1",
	"template_variables.0.name = datacenter",
	"template_variables.0.defaults.# = 1",
	"template_variables.0.defaults.0 = defaults",
}

func TestAccDatadogPowerpackTreeMap(t *testing.T) {
	testAccDatadogPowerpackWidgetUtil(t, datadogPowerpackTreeMapTest, "datadog_powerpack.treemap_powerpack", datadogPowerpackTreeMapTestAsserts)
}
