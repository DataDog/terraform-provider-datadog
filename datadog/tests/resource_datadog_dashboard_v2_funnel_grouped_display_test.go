package test

import "testing"

const datadogDashboardV2FunnelGroupedDisplayConfig = `
resource "datadog_dashboard_v2" "funnel_grouped_display_dashboard" {
  title       = "{{uniq}}"
  layout_type = "ordered"

  widget {
    funnel_definition {
      title           = "Browser funnel"
      grouped_display = "side_by_side"

      request {
        query {
          data_source  = "rum"
          query_string = "@browser.name:Chrome"

          step {
            facet = "@view.name"
            value = "/home"
          }

          step {
            facet = "@view.name"
            value = "/checkout"
          }
        }
      }
    }
  }
}
`

var datadogDashboardV2FunnelGroupedDisplayAsserts = []string{
	"title = {{uniq}}",
	"widget.0.funnel_definition.0.title = Browser funnel",
	"widget.0.funnel_definition.0.grouped_display = side_by_side",
}

func TestAccDatadogDashboardV2FunnelGroupedDisplay(t *testing.T) {
	config, name := datadogDashboardV2FunnelGroupedDisplayConfig, "datadog_dashboard_v2.funnel_grouped_display_dashboard"
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardV2FunnelGroupedDisplay", config, name, datadogDashboardV2FunnelGroupedDisplayAsserts)
}

func TestAccDatadogDashboardV2FunnelGroupedDisplay_import(t *testing.T) {
	config, name := datadogDashboardV2FunnelGroupedDisplayConfig, "datadog_dashboard_v2.funnel_grouped_display_dashboard"
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardV2FunnelGroupedDisplay_import", config, name)
}
