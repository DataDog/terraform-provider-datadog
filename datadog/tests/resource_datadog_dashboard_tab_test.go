package test

import (
	"testing"
)

const datadogDashboardTabConfig = `
resource "datadog_dashboard" "tab_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"

	widget {
		note_definition {
			content = "Widget 1"
		}
	}

	widget {
		note_definition {
			content = "Widget 2"
		}
	}

	widget {
		note_definition {
			content = "Widget 3"
		}
	}

	tab {
		name       = "Overview"
		widget_ids = ["@1", "@2"]
	}

	tab {
		name       = "Details"
		widget_ids = ["@3"]
	}
}
`

var datadogDashboardTabAsserts = []string{
	"title = {{uniq}}",
	"description = Created using the Datadog provider in Terraform",
	"layout_type = ordered",
	"widget.# = 3",
	"tab.# = 2",
	"tab.0.name = Overview",
	"tab.0.widget_ids.# = 2",
	"tab.0.widget_ids.0 = @1",
	"tab.0.widget_ids.1 = @2",
	"tab.1.name = Details",
	"tab.1.widget_ids.# = 1",
	"tab.1.widget_ids.0 = @3",
}

func TestAccDatadogDashboardTab(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardTabConfig, "datadog_dashboard.tab_dashboard", datadogDashboardTabAsserts)
}

func TestAccDatadogDashboardTab_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardTabConfig, "datadog_dashboard.tab_dashboard")
}
