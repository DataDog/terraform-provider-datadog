package test

import (
	"testing"
)

const datadogPowerpackTraceServiceTest = `
resource "datadog_powerpack" "trace_service_powerpack" {
	name = "{{uniq}}"
    tags = ["tag:foo1"]
	description   = "Created using the Datadog provider in Terraform"
	live_span = "4h"
	template_variables {
		defaults = ["defaults"]
		name     = "datacenter"
	}
  	widget {
		trace_service_definition {
		  display_format     = "three_column"
		  env                = "datadog.com"
		  service            = "alerting-cassandra"
		  show_breakdown     = true
		  show_distribution  = true
		  show_errors        = true
		  show_hits          = true
		  show_latency       = false
		  show_resource_list = false
		  size_format        = "large"
		  span_name          = "cassandra.query"
		  title              = "alerting-cassandra #env:datadog.com"
		  title_align        = "center"
		  title_size         = "13"
		}
  	}
}
`

var datadogPowerpackTraceServiceTestAsserts = []string{
	// Powerpack metadata
	"description = Created using the Datadog provider in Terraform",
	"widget.# = 1",
	"tags.# = 1",
	"tags.0 = tag:foo1",
	"live_span = 4h",
	// Trace Service widget
	"widget.0.trace_service_definition.0.display_format = three_column",
	"widget.0.trace_service_definition.0.env = datadog.com",
	"widget.0.trace_service_definition.0.service = alerting-cassandra",
	"widget.0.trace_service_definition.0.show_breakdown = true",
	"widget.0.trace_service_definition.0.show_distribution = true",
	"widget.0.trace_service_definition.0.show_errors = true",
	"widget.0.trace_service_definition.0.show_hits = true",
	"widget.0.trace_service_definition.0.show_latency = false",
	"widget.0.trace_service_definition.0.show_resource_list = false",
	"widget.0.trace_service_definition.0.size_format = large",
	"widget.0.trace_service_definition.0.span_name = cassandra.query",
	"widget.0.trace_service_definition.0.title = alerting-cassandra #env:datadog.com",
	"widget.0.trace_service_definition.0.title_align = center",
	"widget.0.trace_service_definition.0.title_size = 13",
	// Template Variables
	"template_variables.# = 1",
	"template_variables.0.name = datacenter",
	"template_variables.0.defaults.# = 1",
	"template_variables.0.defaults.0 = defaults",
}

func TestAccDatadogPowerpackTraceService(t *testing.T) {
	testAccDatadogPowerpackWidgetUtil(t, datadogPowerpackTraceServiceTest, "datadog_powerpack.trace_service_powerpack", datadogPowerpackTraceServiceTestAsserts)
}
