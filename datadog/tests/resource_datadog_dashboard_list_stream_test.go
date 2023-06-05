package test

import (
	"testing"
)

const datadogDashboardListStreamConfig = `
resource "datadog_dashboard" "list_stream_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "free"
	is_read_only  = "true"

	widget {
		list_stream_definition {
			title = "List Stream 1"
			title_align = "right"
			title_size = "16"
			request {
				response_format = "event_list"
				query {
					data_source = "rum_issue_stream"
				}
				columns {
					field = "details"
					width = "auto"
				}
			}
			request {
				response_format = "event_list"
				query {
					data_source = "apm_issue_stream"
					query_string = "env: prod"
					indexes = ["timestamp", "message"]
				}
				columns {
					field = "details"
					width = "auto"
				}
			}
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

var datadogDashboardListStreamAsserts = []string{
	"description = Created using the Datadog provider in Terraform",
	"layout_type = free",
	"is_read_only = true",
	"title = {{uniq}}",
	"widget.0.list_stream_definition.0.request.0.response_format = event_list",
	"widget.0.list_stream_definition.0.request.0.query.0.data_source = rum_issue_stream",
	"widget.0.list_stream_definition.0.request.0.columns.0.field = details",
	"widget.0.list_stream_definition.0.request.0.columns.0.width = auto",
	"widget.0.list_stream_definition.0.request.1.response_format = event_list",
	"widget.0.list_stream_definition.0.request.1.query.0.data_source = apm_issue_stream",
	"widget.0.list_stream_definition.0.request.1.query.0.query_string = env: prod",
	"widget.0.list_stream_definition.0.request.1.query.0.indexes.0 = timestamp",
	"widget.0.list_stream_definition.0.request.1.query.0.indexes.1 = message",
	"widget.0.list_stream_definition.0.request.1.columns.0.field = details",
	"widget.0.list_stream_definition.0.request.1.columns.0.width = auto",
	"widget.0.widget_layout.0.height = 43",
	"widget.0.widget_layout.0.width = 32",
	"widget.0.widget_layout.0.x = 5",
	"widget.0.widget_layout.0.y = 5",
}

func TestAccDatadogDashboardListStream(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardListStreamConfig, "datadog_dashboard.list_stream_dashboard", datadogDashboardListStreamAsserts)
}

const datadogDashboardListStreamEventsConfig = `
resource "datadog_dashboard" "list_stream_event_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"

	widget {
		list_stream_definition {
			title = "List Stream 2"
			title_align = "right"
			title_size = "16"
			request {
				response_format = "event_list"
				query {
					data_source = "event_stream"
					query_string = "example.metric"
					event_size = "l"
					sort {
						column = "source"
						order = "desc"
					}
				}
				columns {
					field = "source"
					width = "auto"
				}
			}
		}
	}
}
`

var datadogDashboardListStreamEventsAsserts = []string{
	"description = Created using the Datadog provider in Terraform",
	"layout_type = ordered",
	"is_read_only = true",
	"title = {{uniq}}",
	"widget.0.list_stream_definition.0.request.0.response_format = event_list",
	"widget.0.list_stream_definition.0.request.0.query.0.data_source = event_stream",
	"widget.0.list_stream_definition.0.request.0.query.0.query_string = example.metric",
	"widget.0.list_stream_definition.0.request.0.query.0.event_size = l",
	"widget.0.list_stream_definition.0.request.0.query.0.sort.0.column = source",
	"widget.0.list_stream_definition.0.request.0.query.0.sort.0.order = desc",
	"widget.0.list_stream_definition.0.request.0.columns.0.field = source",
	"widget.0.list_stream_definition.0.request.0.columns.0.width = auto",
}

func TestAccDatadogDashboardListStreamEvents(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardListStreamEventsConfig, "datadog_dashboard.list_stream_event_dashboard", datadogDashboardListStreamEventsAsserts)
}
