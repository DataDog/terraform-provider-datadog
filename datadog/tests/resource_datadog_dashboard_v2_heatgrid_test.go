package test

import "testing"

const datadogDashboardHeatgridConfig = `
resource "datadog_dashboard_v2" "heatgrid_dashboard" {
  title       = "{{uniq}}"
  layout_type = "ordered"

  widget {
    heatgrid_definition {
      title = "CPU by host"

      request {
        query {
          metric_query {
            data_source = "metrics"
            name        = "query1"
            query       = "avg:system.cpu.user{*} by {host}"
          }
        }
        formula {
          formula_expression = "query1"
        }
      }

      sort {
        nesting_display = "flat"
        sort_by {
          label {
            order = "asc"
          }
        }
      }

      color {
        mode   = "discrete"
        source = "custom"
        bin {
          color = jsonencode("#00FF00")
          label = "Normal"
        }
        bin {
          color       = jsonencode("#FF0000")
          lower_bound = 80
          label       = "High"
        }
      }

      label_column {
        text_format {
          match {
            type  = "starts_with"
            value = "api-"
          }
          replace {
            substring {
              substring = "api-"
              with      = ""
            }
          }
        }
      }
    }
  }
}
`

var datadogDashboardHeatgridAsserts = []string{
	"title = {{uniq}}",
	"widget.0.heatgrid_definition.0.title = CPU by host",
	"widget.0.heatgrid_definition.0.request.0.query.0.metric_query.0.data_source = metrics",
	"widget.0.heatgrid_definition.0.request.0.formula.0.formula_expression = query1",
	"widget.0.heatgrid_definition.0.sort.0.sort_by.0.label.0.order = asc",
	"widget.0.heatgrid_definition.0.color.0.mode = discrete",
	"widget.0.heatgrid_definition.0.color.0.source = custom",
	"widget.0.heatgrid_definition.0.color.0.bin.0.label = Normal",
	"widget.0.heatgrid_definition.0.color.0.bin.1.lower_bound = 80",
	"widget.0.heatgrid_definition.0.color.0.bin.1.label = High",
	"widget.0.heatgrid_definition.0.label_column.0.text_format.0.match.0.type = starts_with",
	"widget.0.heatgrid_definition.0.label_column.0.text_format.0.replace.0.substring.0.substring = api-",
}

func TestAccDatadogDashboardV2Heatgrid(t *testing.T) {
	config, name := datadogDashboardHeatgridConfig, "datadog_dashboard_v2.heatgrid_dashboard"
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardV2Heatgrid", config, name, datadogDashboardHeatgridAsserts)
}

func TestAccDatadogDashboardV2Heatgrid_import(t *testing.T) {
	config, name := datadogDashboardHeatgridConfig, "datadog_dashboard_v2.heatgrid_dashboard"
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardV2Heatgrid_import", config, name)
}
