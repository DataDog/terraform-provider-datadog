package test

import (
	"testing"
)

const datadogDashboardListStreamStorageConfig = `
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
					data_source = "logs_stream"
					query_string = "env: prod"
					indexes = ["timestamp", "message"]
					storage = "online_archives"
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

var datadogDashboardListStreamStorageAsserts = []string{
	"description = Created using the Datadog provider in Terraform",
	"layout_type = free",
	"is_read_only = true",
	"title = {{uniq}}",
	"widget.0.list_stream_definition.0.request.0.response_format = event_list",
	"widget.0.list_stream_definition.0.request.0.query.0.data_source = logs_stream",
	"widget.0.list_stream_definition.0.request.0.query.0.query_string = env: prod",
	"widget.0.list_stream_definition.0.request.0.query.0.indexes.0 = timestamp",
	"widget.0.list_stream_definition.0.request.0.query.0.indexes.1 = message",
	"widget.0.list_stream_definition.0.request.0.query.0.storage = online_archives",
	"widget.0.list_stream_definition.0.request.0.columns.0.field = details",
	"widget.0.list_stream_definition.0.request.0.columns.0.width = auto",
	"widget.0.widget_layout.0.height = 43",
	"widget.0.widget_layout.0.width = 32",
	"widget.0.widget_layout.0.x = 5",
	"widget.0.widget_layout.0.y = 5",
}

func TestAccDatadogDashboardListStreamStorage(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardListStreamStorageConfig, "datadog_dashboard.list_stream_dashboard", datadogDashboardListStreamStorageAsserts)
}
