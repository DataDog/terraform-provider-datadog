package test

import (
	"testing"
)

const datadogDashboardEventTimelineConfig = `
resource "datadog_dashboard" "event_timeline_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "free"
	is_read_only  = "true"
	
	widget {
		event_timeline_definition {
			title = "Widget Title"
			title_align = "right"
			title_size = "16"
			tags_execution = "and"
			query = "status:error"
			live_span = "1h"
		}
		widget_layout {
			height = 43
			width = 32
			x = 5
			y = 5
		}
	}
	widget {
		event_timeline_definition {
			title = "Widget Title"
			title_align = "right"
			title_size = "16"
			tags_execution = "and"
			query = "status:error"
			time = {
				live_span = "1h"
			}
		}
		layout = {
			height = 43
			width = 32
			x = 5
			y = 5
		}
	}
}
`
const datadogDashboardEventTimelineConfigImport = `
resource "datadog_dashboard" "event_timeline_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "free"
	is_read_only  = "true"
	
	widget {
		event_timeline_definition {
			title = "Widget Title"
			title_align = "right"
			title_size = "16"
			tags_execution = "and"
			query = "status:error"
			live_span = "1h"
		}
		widget_layout {
			height = 43
			width = 32
			x = 5
			y = 5
		}
	}
}
`

var datadogDashboardEventTimelineAsserts = []string{
	"widget.0.widget_layout.0.y = 5",
	"widget.0.event_timeline_definition.0.title_align = right",
	"widget.0.widget_layout.0.x = 5",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.event_timeline_definition.0.live_span = 1h",
	"title = {{uniq}}",
	"is_read_only = true",
	"widget.0.widget_layout.0.width = 32",
	"widget.0.event_timeline_definition.0.title_size = 16",
	"layout_type = free",
	"widget.0.event_timeline_definition.0.query = status:error",
	"widget.0.event_timeline_definition.0.title = Widget Title",
	"widget.0.event_timeline_definition.0.tags_execution = and",
	"widget.0.widget_layout.0.height = 43",
	// Deprecated widget
	"widget.1.layout.y = 5",
	"widget.1.event_timeline_definition.0.title_align = right",
	"widget.1.layout.x = 5",
	"widget.1.event_timeline_definition.0.time.live_span = 1h",
	"widget.1.layout.width = 32",
	"widget.1.event_timeline_definition.0.title_size = 16",
	"widget.1.event_timeline_definition.0.query = status:error",
	"widget.1.event_timeline_definition.0.title = Widget Title",
	"widget.1.event_timeline_definition.0.tags_execution = and",
	"widget.1.layout.height = 43",
}

func TestAccDatadogDashboardEventTimeline(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardEventTimelineConfig, "datadog_dashboard.event_timeline_dashboard", datadogDashboardEventTimelineAsserts)
}

func TestAccDatadogDashboardEventTimeline_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardEventTimelineConfigImport, "datadog_dashboard.event_timeline_dashboard")
}
