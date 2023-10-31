package test

import (
	"testing"
)

const datadogPowerpackListStreamTest = `
resource "datadog_powerpack" "list_stream_powerpack" {
	name         = "{{uniq}}"
    tags = ["tag:foo1"]
	description   = "Created using the Datadog provider in Terraform"
	template_variables {
		defaults = ["defaults"]
		name     = "datacenter"
	}
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
			height = 5
			width = 5
			x = 5
			y = 5
		}
	}
}
`

var datadogPowerpackListStreamTestAsserts = []string{
	// Powerpack metadata
	"name = {{uniq}}",
	"description = Created using the Datadog provider in Terraform",
	"widget.# = 1",
	"tags.# = 1",
	"tags.0 = tag:foo1",
	// list_stream widget
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
	"widget.0.widget_layout.0.height = 5",
	"widget.0.widget_layout.0.width = 5",
	"widget.0.widget_layout.0.x = 5",
	"widget.0.widget_layout.0.y = 5",
	// Template Variables
	"template_variables.# = 1",
	"template_variables.0.name = datacenter",
	"template_variables.0.defaults.# = 1",
	"template_variables.0.defaults.0 = defaults",
}

func TestAccDatadogPowerpackListStream(t *testing.T) {
	testAccDatadogPowerpackWidgetUtil(t, datadogPowerpackListStreamTest, "datadog_powerpack.list_stream_powerpack", datadogPowerpackListStreamTestAsserts)
}
