package datadog

import (
	"testing"
)

const datadogDashboardLogStreamConfig = `
resource "datadog_dashboard" "log_stream_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "free"
	is_read_only  = "true"

	widget {
		log_stream_definition {
			title = "Log Stream"
			title_align = "right"
			title_size = "16"
			show_message_column = "true"
			message_display = "expanded-md"
			query = "status:error env:prod"
			show_date_column = "true"
			indexes = ["main"]
			columns = ["core_host", "core_service"]
			live_span = "1d"
			sort {
				column = "time"
				order = "desc"
			}
		}
		widget_layout {
			height = 43
			width = 32
			x = 5
			y = 5
		}
	}
	widget {
		log_stream_definition {
			title = "Log Stream"
			title_align = "right"
			title_size = "16"
			show_message_column = "true"
			message_display = "expanded-md"
			query = "status:error env:prod"
			show_date_column = "true"
			indexes = ["main"]
			columns = ["core_host", "core_service"]
			time = {
				live_span = "1d"
			}
			sort {
				column = "time"
				order = "desc"
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

var datadogDashboardLogStreamAsserts = []string{
	"description = Created using the Datadog provider in Terraform",
	"widget.0.log_stream_definition.0.query = status:error env:prod",
	"widget.0.log_stream_definition.0.title_align = right",
	"widget.0.log_stream_definition.0.show_date_column = true",
	"widget.0.log_stream_definition.0.columns.0 = core_host",
	"layout_type = free",
	"widget.0.log_stream_definition.0.show_message_column = true",
	"widget.0.log_stream_definition.0.live_span = 1d",
	"widget.0.widget_layout.0.width = 32",
	"widget.0.widget_layout.0.x = 5",
	"is_read_only = true",
	"widget.0.log_stream_definition.0.message_display = expanded-md",
	"widget.0.widget_layout.0.height = 43",
	"title = {{uniq}}",
	"widget.0.log_stream_definition.0.columns.1 = core_service",
	"widget.0.log_stream_definition.0.title_size = 16",
	"widget.0.log_stream_definition.0.logset =",
	"widget.0.widget_layout.0.y = 5",
	"widget.0.log_stream_definition.0.sort.0.column = time",
	"widget.0.log_stream_definition.0.title = Log Stream",
	"widget.0.log_stream_definition.0.sort.0.order = desc",
	"widget.0.log_stream_definition.0.indexes.0 = main",
	// Deprecated widget
	"widget.1.log_stream_definition.0.query = status:error env:prod",
	"widget.1.log_stream_definition.0.title_align = right",
	"widget.1.log_stream_definition.0.show_date_column = true",
	"widget.1.log_stream_definition.0.columns.0 = core_host",
	"widget.1.log_stream_definition.0.show_message_column = true",
	"widget.1.log_stream_definition.0.time.live_span = 1d",
	"widget.1.layout.width = 32",
	"widget.1.layout.x = 5",
	"widget.1.log_stream_definition.0.message_display = expanded-md",
	"widget.1.layout.height = 43",
	"widget.1.log_stream_definition.0.columns.1 = core_service",
	"widget.1.log_stream_definition.0.title_size = 16",
	"widget.1.log_stream_definition.0.logset =",
	"widget.1.layout.y = 5",
	"widget.1.log_stream_definition.0.sort.0.column = time",
	"widget.1.log_stream_definition.0.title = Log Stream",
	"widget.1.log_stream_definition.0.sort.0.order = desc",
	"widget.1.log_stream_definition.0.indexes.0 = main",
}

func TestAccDatadogDashboardLogStream(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardLogStreamConfig, "datadog_dashboard.log_stream_dashboard", datadogDashboardLogStreamAsserts)
}

func TestAccDatadogDashboardLogStream_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtil_import(t, datadogDashboardLogStreamConfig, "datadog_dashboard.log_stream_dashboard")
}

const datadogDashboardLogStreamLogSetConfig = `
resource "datadog_dashboard" "log_stream_dashboard_logset" {
	title         = "Acceptance Test Log Stream Widget Dashboard"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "free"
	is_read_only  = "true"

	widget {
		log_stream_definition {
			title = "Log Stream"
			title_align = "right"
			title_size = "16"
			show_message_column = "true"
			message_display = "expanded-md"
			query = "status:error env:prod"
			show_date_column = "true"
			logset = 19
			columns = ["core_host", "core_service"]
			live_span = "1d"
			sort {
				column = "time"
				order = "desc"
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

var datadogDashboardLogStreamLogSetAsserts = []string{
	"description = Created using the Datadog provider in Terraform",
	"widget.0.log_stream_definition.0.query = status:error env:prod",
	"widget.0.log_stream_definition.0.title_align = right",
	"widget.0.log_stream_definition.0.show_date_column = true",
	"widget.0.log_stream_definition.0.columns.0 = core_host",
	"layout_type = free",
	"widget.0.log_stream_definition.0.show_message_column = true",
	"widget.0.log_stream_definition.0.live_span = 1d",
	"widget.0.widget_layout.0.width = 32",
	"widget.0.widget_layout.0.x = 5",
	"is_read_only = true",
	"widget.0.log_stream_definition.0.message_display = expanded-md",
	"widget.0.widget_layout.0.height = 43",
	"title = Acceptance Test Log Stream Widget Dashboard",
	"widget.0.log_stream_definition.0.columns.1 = core_service",
	"widget.0.log_stream_definition.0.title_size = 16",
	"widget.0.log_stream_definition.0.logset = 19",
	"widget.0.widget_layout.0.y = 5",
	"widget.0.log_stream_definition.0.sort.0.column = time",
	"widget.0.log_stream_definition.0.title = Log Stream",
	"widget.0.log_stream_definition.0.sort.0.order = desc",
}

func TestAccDatadogDashboardLogStreamLogSet(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardLogStreamLogSetConfig, "datadog_dashboard.log_stream_dashboard_logset", datadogDashboardLogStreamLogSetAsserts)
}

func TestAccDatadogDashboardLogStreamLogSet_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtil_import(t, datadogDashboardLogStreamLogSetConfig, "datadog_dashboard.log_stream_dashboard_logset")
}
