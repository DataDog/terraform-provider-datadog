package test

import "testing"

const datadogDashboardV2ListStreamOpenAPIParityConfig = `
resource "datadog_dashboard_v2" "list_stream_openapi_parity_dashboard" {
  title       = "{{uniq}}"
  layout_type = "ordered"

  widget {
    list_stream_definition {
      title = "Transaction stream"

      request {
        response_format = "event_list"

        query {
          data_source  = "logs_transaction_stream"
          query_string = "service:web-store"
          version      = "sequential_query"

          compute {
            aggregation = "count"
            facet       = "resource_name"
          }

          group_by {
            facet = "service"
          }
        }

        columns {
          field = "timestamp"
          width = "auto"
        }
      }
    }
  }

  widget {
    list_stream_definition {
      title = "Issue stream"

      request {
        response_format = "event_list"

        query {
          data_source      = "issue_stream"
          query_string     = "service:web-store"
          states           = ["OPEN", "ACKNOWLEDGED"]
          assignee_uuids   = ["00000000-0000-0000-0000-000000000000"]
          suspected_causes = ["deployment"]
          team_handles     = ["platform"]
          persona          = "backend"
        }

        columns {
          field = "timestamp"
          width = "auto"
        }
      }
    }
  }
}
`

var datadogDashboardV2ListStreamOpenAPIParityAsserts = []string{
	"title = {{uniq}}",
	"widget.0.list_stream_definition.0.request.0.query.0.data_source = logs_transaction_stream",
	"widget.0.list_stream_definition.0.request.0.query.0.version = sequential_query",
	"widget.0.list_stream_definition.0.request.0.query.0.compute.0.aggregation = count",
	"widget.0.list_stream_definition.0.request.0.query.0.compute.0.facet = resource_name",
	"widget.0.list_stream_definition.0.request.0.query.0.group_by.0.facet = service",
	"widget.1.list_stream_definition.0.request.0.query.0.data_source = issue_stream",
	"widget.1.list_stream_definition.0.request.0.query.0.states.0 = OPEN",
	"widget.1.list_stream_definition.0.request.0.query.0.states.1 = ACKNOWLEDGED",
	"widget.1.list_stream_definition.0.request.0.query.0.assignee_uuids.0 = 00000000-0000-0000-0000-000000000000",
	"widget.1.list_stream_definition.0.request.0.query.0.suspected_causes.0 = deployment",
	"widget.1.list_stream_definition.0.request.0.query.0.team_handles.0 = platform",
	"widget.1.list_stream_definition.0.request.0.query.0.persona = backend",
}

func TestAccDatadogDashboardV2ListStreamOpenAPIParity(t *testing.T) {
	config, name := datadogDashboardV2ListStreamOpenAPIParityConfig, "datadog_dashboard_v2.list_stream_openapi_parity_dashboard"
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardV2ListStreamOpenAPIParity", config, name, datadogDashboardV2ListStreamOpenAPIParityAsserts)
}

func TestAccDatadogDashboardV2ListStreamOpenAPIParity_import(t *testing.T) {
	config, name := datadogDashboardV2ListStreamOpenAPIParityConfig, "datadog_dashboard_v2.list_stream_openapi_parity_dashboard"
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardV2ListStreamOpenAPIParity_import", config, name)
}
