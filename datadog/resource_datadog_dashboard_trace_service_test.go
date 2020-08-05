package datadog

import (
	"testing"
)

// JSON export used as test scenario
//{
//    "notify_list": [],
//    "description": "",
//    "author_name": "--redacted--",
//    "template_variables": [],
//    "is_read_only": false,
//    "id": "--redacted--",
//    "title": "TF - Service Summary Example",
//    "url": "--redacted--",
//    "created_at": "2020-06-09T13:33:54.661635+00:00",
//    "modified_at": "2020-06-09T13:34:41.222757+00:00",
//    "author_handle": "--redacted--",
//    "widgets": [
//        {
//            "definition": {
//                "span_name": "postgres.connection.rollback",
//                "title_size": "16",
//                "service": "postgres",
//                "title": "postgres #env:datadoghq.com",
//                "size_format": "large",
//                "show_hits": true,
//                "show_latency": true,
//                "title_align": "center",
//                "show_errors": true,
//                "show_breakdown": true,
//                "env": "datadoghq.com",
//                "time": {
//                    "live_span": "30m"
//                },
//                "show_distribution": true,
//                "display_format": "three_column",
//                "type": "trace_service",
//                "show_resource_list": true
//            },
//            "layout": {
//                "y": 2,
//                "x": 1,
//                "height": 72,
//                "width": 72
//            },
//            "id": 0
//        }
//    ],
//    "layout_type": "free"
//}

const datadogDashboardTraceServiceConfig = `
resource "datadog_dashboard" "trace_service_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "free"
	is_read_only  = "true"

	widget {
		trace_service_definition {
			span_name = "postgres.connection.rollback"
			title_size = "16"
			service = "postgres"
			title = "postgres #env:datadoghq.com"
			size_format = "large"
			show_hits = true
			show_latency = true
			title_align = "center"
			show_errors = true
			show_breakdown = true
			env = "datadoghq.com"
			time = {
				live_span = "30m"
			}
			show_distribution = true
			display_format = "three_column"
			show_resource_list = true
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

var datadogDashboardTraceServiceAsserts = []string{
	"widget.0.trace_service_definition.0.show_distribution = true",
	"widget.0.trace_service_definition.0.title = postgres #env:datadoghq.com",
	"is_read_only = true",
	"widget.0.trace_service_definition.0.show_hits = true",
	"widget.0.trace_service_definition.0.span_name = postgres.connection.rollback",
	"widget.0.layout.height = 43",
	"widget.0.trace_service_definition.0.size_format = large",
	"widget.0.trace_service_definition.0.env = datadoghq.com",
	"widget.0.layout.width = 32",
	"layout_type = free",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.trace_service_definition.0.display_format = three_column",
	"widget.0.trace_service_definition.0.time.live_span = 30m",
	"widget.0.trace_service_definition.0.show_resource_list = true",
	"widget.0.trace_service_definition.0.show_errors = true",
	"widget.0.trace_service_definition.0.title_align = center",
	"widget.0.trace_service_definition.0.title_size = 16",
	"widget.0.trace_service_definition.0.show_breakdown = true",
	"widget.0.layout.x = 5",
	"widget.0.layout.y = 5",
	"widget.0.trace_service_definition.0.show_latency = true",
	"widget.0.trace_service_definition.0.service = postgres",
	"title = {{uniq}}",
}

func TestAccDatadogDashboardTraceService(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardTraceServiceConfig, "datadog_dashboard.trace_service_dashboard", datadogDashboardTraceServiceAsserts)
}

func TestAccDatadogDashboardTraceService_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtil_import(t, datadogDashboardTraceServiceConfig, "datadog_dashboard.trace_service_dashboard")
}
