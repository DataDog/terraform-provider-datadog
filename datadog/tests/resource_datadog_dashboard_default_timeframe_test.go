package test

import (
	"testing"
)

const datadogDashboardDefaultTimeframeConfig = `
resource "datadog_dashboard" "default_timeframe_dashboard" {
  title       = "{{uniq}}"
  layout_type = "ordered"
  description = "Created using the Datadog provider in Terraform"

  default_timeframe {
    type  = "live"
    unit  = "week"
    value = 1
  }

  widget {
    note_definition {
      content = "Widget 1"
    }
  }
}
`

var datadogDashboardDefaultTimeframeAsserts = []string{
	"title = {{uniq}}",
	"layout_type = ordered",
	"description = Created using the Datadog provider in Terraform",
	"default_timeframe.# = 1",
	"default_timeframe.0.type = live",
	"default_timeframe.0.unit = week",
	"default_timeframe.0.value = 1",
	"widget.# = 1",
}

func TestAccDatadogDashboardDefaultTimeframe(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardDefaultTimeframeConfig, "datadog_dashboard.default_timeframe_dashboard", datadogDashboardDefaultTimeframeAsserts)
}

func TestAccDatadogDashboardDefaultTimeframe_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardDefaultTimeframeConfig, "datadog_dashboard.default_timeframe_dashboard")
}
