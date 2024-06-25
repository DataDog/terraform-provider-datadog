package test

import (
	"testing"
)

const datadogDashboardCrossOrgConfig = `
resource "datadog_dashboard" "cross_org_dashboard" {
  title        = "{{uniq}}"
  description  = "Created using the Datadog provider in Terraform"
  layout_type  = "ordered"
  is_read_only = true
  widget {
    timeseries_definition {
      request {
        formula {
          formula_expression = "my_query_1 + my_query_2"
          limit {
            count = 5
            order = "asc"
          }
          alias = "sum query"
        }
        formula {
          formula_expression = "my_query_1 * my_query_2"
          limit {
            count = 7
            order = "desc"
          }
          alias = "multiplicative query"
        }
        query {
          metric_query {
            data_source     = "metrics"
            query           = "avg:system.cpu.user{app:general} by {env}"
            name            = "my_query_1"
            aggregator      = "sum"
            cross_org_uuids = ["6434abde-xxxx-yyyy-zzzz-da7ad0900001"]
          }
        }
        query {
          metric_query {
            data_source     = "metrics"
            query           = "avg:system.cpu.user{app:general} by {env}"
            name            = "my_query_2"
            aggregator      = "sum"
            cross_org_uuids = ["6434abde-xxxx-yyyy-zzzz-da7ad0900001"]
          }
        }
      }
    }
  }
  widget {
    timeseries_definition {
      request {
        query {
          event_query {
            data_source = "logs"
            indexes     = ["days-3"]
            name        = "my_event_query"
            compute {
              aggregation = "count"
            }
            search {
              query = "abc"
            }
            cross_org_uuids = ["6434abde-xxxx-yyyy-zzzz-da7ad0900001"]
            group_by {
              facet = "host"
              sort {
                metric      = "@lambda.max_memory_used"
                aggregation = "avg"
                order       = "desc"
              }
              limit = 10
            }
          }
        }
      }
    }
  }
  widget {
    timeseries_definition {
      request {
        query {
          process_query {
            data_source       = "process"
            text_filter       = "abc"
            metric            = "process.stat.cpu.total_pct"
            limit             = 10
            tag_filters       = ["some_filter"]
            name              = "my_process_query"
            sort              = "asc"
            is_normalized_cpu = true
            cross_org_uuids   = ["6434abde-xxxx-yyyy-zzzz-da7ad0900001"]
          }
        }
      }
    }
  }
  widget {
    timeseries_definition {
      request {
        formula {
          formula_expression = "query1"
          alias              = "my cloud cost query"
        }
        query {
          cloud_cost_query {
            data_source     = "cloud_cost"
            query           = "sum:aws.cost.amortized{*}"
            name            = "query1"
            aggregator      = "sum"
            cross_org_uuids = ["6434abde-xxxx-yyyy-zzzz-da7ad0900001"]
          }
        }
      }
    }
  }
  widget {
    timeseries_definition {
      request {
        formula {
          formula_expression = "query1"
          alias              = "my slo query"
        }
        query {
          slo_query {
            data_source     = "slo"
            measure         = "good_events"
            slo_id          = "slo1"
            cross_org_uuids = ["6434abde-xxxx-yyyy-zzzz-da7ad0900001"]
          }
        }
      }
    }
  }
}
`

var datadogDashboardCrossOrgAsserts = []string{
	"is_read_only = true",
	"title = {{uniq}}",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.timeseries_definition.0.request.0.query.0.metric_query.0.cross_org_uuids.0 = 6434abde-xxxx-yyyy-zzzz-da7ad0900001",
	"widget.1.timeseries_definition.0.request.0.query.0.event_query.0.cross_org_uuids.0 = 6434abde-xxxx-yyyy-zzzz-da7ad0900001",
	"widget.2.timeseries_definition.0.request.0.query.0.process_query.0.cross_org_uuids.0 = 6434abde-xxxx-yyyy-zzzz-da7ad0900001",
	"widget.3.timeseries_definition.0.request.0.query.0.cloud_cost_query.0.cross_org_uuids.0 = 6434abde-xxxx-yyyy-zzzz-da7ad0900001",
	"widget.4.timeseries_definition.0.request.0.query.0.slo_query.0.cross_org_uuids.0 = 6434abde-xxxx-yyyy-zzzz-da7ad0900001",
	"layout_type = ordered",
}

func TestAccDatadogDashboardCrossOrg(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardCrossOrgConfig, "datadog_dashboard.cross_org_dashboard", datadogDashboardCrossOrgAsserts)
}

func TestAccDatadogDashboardCrossOrg_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardCrossOrgConfig, "datadog_dashboard.cross_org_dashboard")
}
