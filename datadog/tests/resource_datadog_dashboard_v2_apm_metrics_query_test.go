package test

import "testing"

const datadogDashboardV2ApmMetricsQueryConfig = `
resource "datadog_dashboard_v2" "apm_metrics_query_dashboard" {
  title       = "{{uniq}}"
  layout_type = "ordered"

  widget {
    query_value_definition {
      title = "APM hits"

      request {
        formula {
          formula_expression = "query1"
        }

        query {
          apm_metrics_query {
            data_source    = "apm_metrics"
            name           = "query1"
            stat           = "hits"
            service        = "web-store"
            peer_tags      = ["peer.service:payments"]
            resource_hash  = "abc123"
            resource_name  = "GET /api/v1/users"
            operation_name = "web.request"
            operation_mode = "primary"
            query_filter   = "env:prod"
            group_by       = ["resource_name"]
            span_kind      = "server"
          }
        }
      }
    }
  }

  widget {
    distribution_definition {
      title = "APM latency distribution"

      request {
        request_type = "histogram"

        histogram_query {
          apm_metrics_query {
            data_source = "apm_metrics"
            name        = "query1"
            stat        = "latency_distribution"
            service     = "web-store"
          }
        }
      }
    }
  }
}
`

var datadogDashboardV2ApmMetricsQueryAsserts = []string{
	"title = {{uniq}}",
	"widget.0.query_value_definition.0.request.0.query.0.apm_metrics_query.0.data_source = apm_metrics",
	"widget.0.query_value_definition.0.request.0.query.0.apm_metrics_query.0.name = query1",
	"widget.0.query_value_definition.0.request.0.query.0.apm_metrics_query.0.stat = hits",
	"widget.0.query_value_definition.0.request.0.query.0.apm_metrics_query.0.service = web-store",
	"widget.0.query_value_definition.0.request.0.query.0.apm_metrics_query.0.peer_tags.0 = peer.service:payments",
	"widget.0.query_value_definition.0.request.0.query.0.apm_metrics_query.0.resource_hash = abc123",
	"widget.0.query_value_definition.0.request.0.query.0.apm_metrics_query.0.resource_name = GET /api/v1/users",
	"widget.0.query_value_definition.0.request.0.query.0.apm_metrics_query.0.operation_name = web.request",
	"widget.0.query_value_definition.0.request.0.query.0.apm_metrics_query.0.operation_mode = primary",
	"widget.0.query_value_definition.0.request.0.query.0.apm_metrics_query.0.query_filter = env:prod",
	"widget.0.query_value_definition.0.request.0.query.0.apm_metrics_query.0.group_by.0 = resource_name",
	"widget.0.query_value_definition.0.request.0.query.0.apm_metrics_query.0.span_kind = server",
	"widget.1.distribution_definition.0.request.0.request_type = histogram",
	"widget.1.distribution_definition.0.request.0.histogram_query.0.apm_metrics_query.0.data_source = apm_metrics",
	"widget.1.distribution_definition.0.request.0.histogram_query.0.apm_metrics_query.0.stat = latency_distribution",
}

func TestAccDatadogDashboardV2ApmMetricsQuery(t *testing.T) {
	config, name := datadogDashboardV2ApmMetricsQueryConfig, "datadog_dashboard_v2.apm_metrics_query_dashboard"
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardV2ApmMetricsQuery", config, name, datadogDashboardV2ApmMetricsQueryAsserts)
}

func TestAccDatadogDashboardV2ApmMetricsQuery_import(t *testing.T) {
	config, name := datadogDashboardV2ApmMetricsQueryConfig, "datadog_dashboard_v2.apm_metrics_query_dashboard"
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardV2ApmMetricsQuery_import", config, name)
}
