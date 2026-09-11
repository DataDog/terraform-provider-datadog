package test

import "testing"

const datadogDashboardV2GeomapRequestParityConfig = `
resource "datadog_dashboard_v2" "geomap_request_parity_dashboard" {
  title       = "{{uniq}}"
  layout_type = "ordered"

  widget {
    geomap_definition {
      title = "Regional traffic"

      style {
        palette      = "hostmap_blues"
        palette_flip = false
      }

      view {
        focus = "WORLD"
      }

      request {
        response_format = "scalar"

        formula {
          formula_expression = "query1"
        }

        query {
          event_query {
            data_source = "rum"
            name        = "query1"
            indexes     = ["*"]

            search {
              query = "@type:session"
            }

            compute {
              aggregation = "count"
            }
          }
        }

        conditional_formats {
          comparator = ">"
          value      = 1000
          palette    = "white_on_green"
        }

        sort {
          count = 250

          order_by {
            group_sort {
              name  = "country"
              order = "desc"
            }
          }
        }
      }

      request {
        response_format = "event_list"

        columns {
          field = "@network.client.geoip.location.latitude"
          width = "auto"
        }

        columns {
          field = "@network.client.geoip.location.longitude"
          width = "auto"
        }

        list_stream_query {
          data_source  = "logs_stream"
          query_string = ""
          storage      = "hot"
        }

        style {
          color_by = "status"
        }

        text_format {
          match {
            type  = "is"
            value = "error"
          }

          palette = "white_on_red"
        }
      }
    }
  }
}
`

var datadogDashboardV2GeomapRequestParityAsserts = []string{
	"title = {{uniq}}",
	"widget.0.geomap_definition.0.title = Regional traffic",
	"widget.0.geomap_definition.0.request.0.response_format = scalar",
	"widget.0.geomap_definition.0.request.0.conditional_formats.0.comparator = >",
	"widget.0.geomap_definition.0.request.0.conditional_formats.0.value = 1000",
	"widget.0.geomap_definition.0.request.0.sort.0.count = 250",
	"widget.0.geomap_definition.0.request.0.sort.0.order_by.0.group_sort.0.name = country",
	"widget.0.geomap_definition.0.request.0.sort.0.order_by.0.group_sort.0.order = desc",
	"widget.0.geomap_definition.0.request.1.response_format = event_list",
	"widget.0.geomap_definition.0.request.1.columns.0.field = @network.client.geoip.location.latitude",
	"widget.0.geomap_definition.0.request.1.columns.0.width = auto",
	"widget.0.geomap_definition.0.request.1.columns.1.field = @network.client.geoip.location.longitude",
	"widget.0.geomap_definition.0.request.1.columns.1.width = auto",
	"widget.0.geomap_definition.0.request.1.list_stream_query.0.data_source = logs_stream",
	"widget.0.geomap_definition.0.request.1.list_stream_query.0.storage = hot",
	"widget.0.geomap_definition.0.request.1.style.0.color_by = status",
	"widget.0.geomap_definition.0.request.1.text_format.0.match.0.type = is",
	"widget.0.geomap_definition.0.request.1.text_format.0.match.0.value = error",
	"widget.0.geomap_definition.0.request.1.text_format.0.palette = white_on_red",
}

func TestAccDatadogDashboardV2GeomapRequestParity(t *testing.T) {
	config, name := datadogDashboardV2GeomapRequestParityConfig, "datadog_dashboard_v2.geomap_request_parity_dashboard"
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardV2GeomapRequestParity", config, name, datadogDashboardV2GeomapRequestParityAsserts)
}

func TestAccDatadogDashboardV2GeomapRequestParity_import(t *testing.T) {
	config, name := datadogDashboardV2GeomapRequestParityConfig, "datadog_dashboard_v2.geomap_request_parity_dashboard"
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardV2GeomapRequestParity_import", config, name)
}
