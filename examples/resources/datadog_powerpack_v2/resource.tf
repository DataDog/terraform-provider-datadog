# Manage Datadog Powerpacks
resource "datadog_powerpack_v2" "foo" {
  description = "Created using the Datadog provider in terraform"
  live_span   = "4h"

  layout {
    height = 10
    width  = 3
    x      = 1
    y      = 0
  }

  template_variables {
    defaults = ["defaults"]
    name     = "datacenter"
  }

  widget {
    timeseries_definition {
      request {
        formula {
          formula_expression = "my_query_1"
        }
        query {
          metric_query {
            data_source = "metrics"
            query       = "avg:system.cpu.user{*} by {env}"
            name        = "my_query_1"
            aggregator  = "avg"
          }
        }
      }
      title = "CPU Usage"
    }
  }
}
