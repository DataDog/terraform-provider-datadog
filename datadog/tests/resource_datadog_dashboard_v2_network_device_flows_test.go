package test

import "testing"

const datadogDashboardV2NetworkDeviceFlowsConfig = `
resource "datadog_dashboard_v2" "network_device_flows" {
  title       = "{{uniq}}"
  layout_type = "ordered"

  widget {
    timeseries_definition {
      request {
        display_type = "line"
        query {
          event_query {
            data_source = "network_device_flows"
            name        = "query1"
            search {
              query = "*"
            }
            compute {
              aggregation = "count"
            }
            group_by_fields {
              fields = ["@source.ip"]
              limit  = 10
            }
          }
        }
        formula {
          formula_expression = "query1"
        }
      }
    }
  }
}
`

func TestAccDatadogDashboardV2NetworkDeviceFlowsFormula(t *testing.T) {
	const name = "datadog_dashboard_v2.network_device_flows"
	testAccDatadogDashboardV2WidgetUtil(t, t.Name(), datadogDashboardV2NetworkDeviceFlowsConfig, name, []string{
		"widget.0.timeseries_definition.0.request.0.query.0.event_query.0.data_source = network_device_flows",
		"widget.0.timeseries_definition.0.request.0.query.0.event_query.0.group_by_fields.0.fields.0 = @source.ip",
		"widget.0.timeseries_definition.0.request.0.formula.0.formula_expression = query1",
	})
}
