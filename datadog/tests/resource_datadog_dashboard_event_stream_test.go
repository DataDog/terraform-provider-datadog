package test

import (
	"testing"
)

const datadogDashboardEventStreamConfig = `
resource "datadog_dashboard" "event_stream_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "free"
	is_read_only  = "true"
	
	widget {
		event_stream_definition {
			title = "Widget Title"
			title_align = "right"
			title_size = "16"
			tags_execution = "and"
			query = "*"
			event_size = "l"
			live_span = "4h"
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

const datadogDashboardEventStreamConfigImport = `
resource "datadog_dashboard" "event_stream_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "free"
	is_read_only  = "true"
	
	widget {
		event_stream_definition {
			title = "Widget Title"
			title_align = "right"
			title_size = "16"
			tags_execution = "and"
			query = "*"
			event_size = "l"
			live_span = "4h"
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

var datadogDashboardEventStreamAsserts = []string{
	"widget.0.widget_layout.0.x = 5",
	"widget.0.event_stream_definition.0.title_size = 16",
	"widget.0.event_stream_definition.0.tags_execution = and",
	"title = {{uniq}}",
	"widget.0.widget_layout.0.y = 5",
	"widget.0.event_stream_definition.0.title_align = right",
	"widget.0.event_stream_definition.0.live_span = 4h",
	"widget.0.widget_layout.0.width = 32",
	"widget.0.event_stream_definition.0.event_size = l",
	"layout_type = free",
	"description = Created using the Datadog provider in Terraform",
	"is_read_only = true",
	"widget.0.event_stream_definition.0.query = *",
	"widget.0.event_stream_definition.0.title = Widget Title",
	"widget.0.widget_layout.0.height = 43",
}

func TestAccDatadogDashboardEventStream(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardEventStreamConfig, "datadog_dashboard.event_stream_dashboard", datadogDashboardEventStreamAsserts)
}

func TestAccDatadogDashboardEventStream_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardEventStreamConfigImport, "datadog_dashboard.event_stream_dashboard")
}
