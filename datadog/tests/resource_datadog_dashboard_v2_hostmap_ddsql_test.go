package test

import "testing"

const datadogDashboardV2HostmapDDSQLConfig = `
resource "datadog_dashboard_v2" "hostmap_ddsql_dashboard" {
  title       = "{{uniq}}"
  layout_type = "ordered"

  widget {
    hostmap_definition {
      title = "DDSQL host map"

      request {
        request_type = "data_projection"
        limit        = 250

        query {
          data_source     = "dataset"
          dataset_provider = "ddsql_query"
          dataset_id       = "example-dataset"
          filter           = "service:web-store"
          limit            = 100

          sort {
            field {
              name  = "cpu_usage"
              order = "desc"
            }
          }
        }

        projection {
          type = "hostmap"

          dimension {
            column    = "entity_id"
            dimension = "node"
          }

          dimension {
            column    = "service"
            dimension = "group"
          }

          dimension {
            column    = "cpu_usage"
            dimension = "fill"
            alias     = "CPU"

            number_format {
              unit {
                custom {
                  label = "%"
                }
              }
            }
          }
        }
      }
    }
  }
}
`

var datadogDashboardV2HostmapDDSQLAsserts = []string{
	"title = {{uniq}}",
	"widget.0.hostmap_definition.0.request.0.request_type = data_projection",
	"widget.0.hostmap_definition.0.request.0.limit = 250",
	"widget.0.hostmap_definition.0.request.0.query.0.data_source = dataset",
	"widget.0.hostmap_definition.0.request.0.query.0.dataset_provider = ddsql_query",
	"widget.0.hostmap_definition.0.request.0.query.0.dataset_id = example-dataset",
	"widget.0.hostmap_definition.0.request.0.query.0.sort.0.field.0.name = cpu_usage",
	"widget.0.hostmap_definition.0.request.0.query.0.sort.0.field.0.order = desc",
	"widget.0.hostmap_definition.0.request.0.projection.0.type = hostmap",
	"widget.0.hostmap_definition.0.request.0.projection.0.dimension.0.dimension = node",
	"widget.0.hostmap_definition.0.request.0.projection.0.dimension.1.dimension = group",
	"widget.0.hostmap_definition.0.request.0.projection.0.dimension.2.dimension = fill",
	"widget.0.hostmap_definition.0.request.0.projection.0.dimension.2.alias = CPU",
	"widget.0.hostmap_definition.0.request.0.projection.0.dimension.2.number_format.0.unit.0.custom.0.label = %",
}

func TestAccDatadogDashboardV2HostmapDDSQL(t *testing.T) {
	config, name := datadogDashboardV2HostmapDDSQLConfig, "datadog_dashboard_v2.hostmap_ddsql_dashboard"
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardV2HostmapDDSQL", config, name, datadogDashboardV2HostmapDDSQLAsserts)
}

func TestAccDatadogDashboardV2HostmapDDSQL_import(t *testing.T) {
	config, name := datadogDashboardV2HostmapDDSQLConfig, "datadog_dashboard_v2.hostmap_ddsql_dashboard"
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardV2HostmapDDSQL_import", config, name)
}
