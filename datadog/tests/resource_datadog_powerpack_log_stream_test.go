package test

import (
	"testing"
)

const datadogPowerpackLogStreamTest = `
resource "datadog_powerpack" "logstream_powerpack" {
	name         = "{{uniq}}"
    tags = ["tag:foo1"]
	description   = "Created using the Datadog provider in Terraform"
	template_variables {
		defaults = ["defaults"]
		name     = "datacenter"
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
			sort {
				column = "time"
				order = "desc"
			}
		}
		widget_layout {
			height = 5
			width = 5
			x = 5
			y = 5
		}
	}
}
`

var datadogPowerpackLogStreamTestAsserts = []string{
	// Powerpack metadata
	"name = {{uniq}}",
	"description = Created using the Datadog provider in Terraform",
	"widget.# = 1",
	"tags.# = 1",
	"tags.0 = tag:foo1",
	// Log stream widget
	"widget.0.log_stream_definition.0.query = status:error env:prod",
	"widget.0.log_stream_definition.0.title_align = right",
	"widget.0.log_stream_definition.0.show_date_column = true",
	"widget.0.log_stream_definition.0.columns.0 = core_host",
	"widget.0.log_stream_definition.0.show_message_column = true",
	"widget.0.widget_layout.0.width = 5",
	"widget.0.widget_layout.0.x = 5",
	"widget.0.log_stream_definition.0.message_display = expanded-md",
	"widget.0.widget_layout.0.height = 5",
	"widget.0.log_stream_definition.0.columns.1 = core_service",
	"widget.0.log_stream_definition.0.title_size = 16",
	"widget.0.widget_layout.0.y = 5",
	"widget.0.log_stream_definition.0.sort.0.column = time",
	"widget.0.log_stream_definition.0.title = Log Stream",
	"widget.0.log_stream_definition.0.sort.0.order = desc",
	"widget.0.log_stream_definition.0.indexes.0 = main",
	// Template Variables
	"template_variables.# = 1",
	"template_variables.0.name = datacenter",
	"template_variables.0.defaults.# = 1",
	"template_variables.0.defaults.0 = defaults",
}

func TestAccDatadogPowerpackLogStream(t *testing.T) {
	testAccDatadogPowerpackWidgetUtil(t, datadogPowerpackLogStreamTest, "datadog_powerpack.logstream_powerpack", datadogPowerpackLogStreamTestAsserts)
}
