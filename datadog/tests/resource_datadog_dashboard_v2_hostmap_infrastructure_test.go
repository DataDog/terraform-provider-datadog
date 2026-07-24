package test

import "testing"

const datadogDashboardV2HostmapInfrastructureConfig = `
resource "datadog_dashboard_v2" "hostmap_infrastructure_dashboard" {
  title       = "{{uniq}}"
  layout_type = "ordered"

  widget {
    hostmap_definition {
      title = "Infrastructure host map"

      request {
        request_type = "infrastructure_hostmap"
        node_type    = "host"
        filter       = "env:prod"

        group_by {
          column = "tags"
          key    = "service"
        }

        enrichment {
          response_format = "scalar"

          query {
            metric_query {
              data_source = "metrics"
              name        = "query1"
              query       = "avg:system.cpu.user{*} by {host}"
            }
          }

          formula {
            formula_expression = "query1"
            alias              = "CPU usage"
            dimension          = "fill"
          }
        }

        style {
          palette      = "hostmap_blues"
          palette_flip = true
          fill_min     = 0
          fill_max     = 100
        }

        conditional_formats {
          comparator = ">"
          value      = 80
          palette    = "white_on_red"
          hide_value = false
        }

        no_group_hosts  = true
        no_metric_hosts = true

        child {
          request_type = "infrastructure_hostmap"
          node_type    = "container"
          filter       = "kube_namespace:store"

          group_by {
            column = "tags"
            key    = "kube_namespace"
          }

          enrichment {
            response_format = "scalar"

            query {
              metric_query {
                data_source = "metrics"
                name        = "query1"
                query       = "avg:container.cpu.usage{*} by {container_id}"
              }
            }

            formula {
              formula_expression = "query1"
              dimension          = "size"
            }
          }
        }
      }
    }
  }
}
`

var datadogDashboardV2HostmapInfrastructureAsserts = []string{
	"title = {{uniq}}",
	"widget.0.hostmap_definition.0.title = Infrastructure host map",
	"widget.0.hostmap_definition.0.request.0.request_type = infrastructure_hostmap",
	"widget.0.hostmap_definition.0.request.0.node_type = host",
	"widget.0.hostmap_definition.0.request.0.filter = env:prod",
	"widget.0.hostmap_definition.0.request.0.group_by.0.column = tags",
	"widget.0.hostmap_definition.0.request.0.group_by.0.key = service",
	"widget.0.hostmap_definition.0.request.0.enrichment.0.response_format = scalar",
	"widget.0.hostmap_definition.0.request.0.enrichment.0.query.0.metric_query.0.data_source = metrics",
	"widget.0.hostmap_definition.0.request.0.enrichment.0.query.0.metric_query.0.name = query1",
	"widget.0.hostmap_definition.0.request.0.enrichment.0.formula.0.formula_expression = query1",
	"widget.0.hostmap_definition.0.request.0.enrichment.0.formula.0.alias = CPU usage",
	"widget.0.hostmap_definition.0.request.0.enrichment.0.formula.0.dimension = fill",
	"widget.0.hostmap_definition.0.request.0.style.0.palette = hostmap_blues",
	"widget.0.hostmap_definition.0.request.0.style.0.palette_flip = true",
	"widget.0.hostmap_definition.0.request.0.style.0.fill_min = 0",
	"widget.0.hostmap_definition.0.request.0.style.0.fill_max = 100",
	"widget.0.hostmap_definition.0.request.0.conditional_formats.0.comparator = >",
	"widget.0.hostmap_definition.0.request.0.conditional_formats.0.value = 80",
	"widget.0.hostmap_definition.0.request.0.no_group_hosts = true",
	"widget.0.hostmap_definition.0.request.0.no_metric_hosts = true",
	"widget.0.hostmap_definition.0.request.0.child.0.request_type = infrastructure_hostmap",
	"widget.0.hostmap_definition.0.request.0.child.0.node_type = container",
	"widget.0.hostmap_definition.0.request.0.child.0.filter = kube_namespace:store",
	"widget.0.hostmap_definition.0.request.0.child.0.enrichment.0.formula.0.dimension = size",
}

func TestAccDatadogDashboardV2HostmapInfrastructure(t *testing.T) {
	config, name := datadogDashboardV2HostmapInfrastructureConfig, "datadog_dashboard_v2.hostmap_infrastructure_dashboard"
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardV2HostmapInfrastructure", config, name, datadogDashboardV2HostmapInfrastructureAsserts)
}

func TestAccDatadogDashboardV2HostmapInfrastructure_import(t *testing.T) {
	config, name := datadogDashboardV2HostmapInfrastructureConfig, "datadog_dashboard_v2.hostmap_infrastructure_dashboard"
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardV2HostmapInfrastructure_import", config, name)
}
