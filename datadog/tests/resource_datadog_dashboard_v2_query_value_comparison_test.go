package test

import "testing"

const datadogDashboardV2QueryValueComparisonConfig = `
resource "datadog_dashboard_v2" "query_value_comparison_dashboard" {
  title       = "{{uniq}}"
  layout_type = "ordered"

  widget {
    query_value_definition {
      title = "CPU comparison"

      request {
        q = "avg:system.cpu.user{*}"

        comparison {
          type           = "both"
          directionality = "increase_better"

          duration {
            type = "custom_timeframe"

            custom_timeframe {
              from = 1779290190000
              to   = 1779894990000
            }
          }
        }
      }
    }
  }
}
`

var datadogDashboardV2QueryValueComparisonAsserts = []string{
	"title = {{uniq}}",
	"widget.0.query_value_definition.0.title = CPU comparison",
	"widget.0.query_value_definition.0.request.0.q = avg:system.cpu.user{*}",
	"widget.0.query_value_definition.0.request.0.comparison.0.type = both",
	"widget.0.query_value_definition.0.request.0.comparison.0.directionality = increase_better",
	"widget.0.query_value_definition.0.request.0.comparison.0.duration.0.type = custom_timeframe",
	"widget.0.query_value_definition.0.request.0.comparison.0.duration.0.custom_timeframe.0.from = 1779290190000",
	"widget.0.query_value_definition.0.request.0.comparison.0.duration.0.custom_timeframe.0.to = 1779894990000",
}

func TestAccDatadogDashboardV2QueryValueComparison(t *testing.T) {
	config, name := datadogDashboardV2QueryValueComparisonConfig, "datadog_dashboard_v2.query_value_comparison_dashboard"
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardV2QueryValueComparison", config, name, datadogDashboardV2QueryValueComparisonAsserts)
}

func TestAccDatadogDashboardV2QueryValueComparison_import(t *testing.T) {
	config, name := datadogDashboardV2QueryValueComparisonConfig, "datadog_dashboard_v2.query_value_comparison_dashboard"
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardV2QueryValueComparison_import", config, name)
}
