resource "datadog_dashboard_v2" "display_fields" {
  title       = "Dashboard display options"
  layout_type = "ordered"

  widget {
    manage_status_definition {
      query                 = "tag:service:example"
      count                 = 50
      start                 = 0
      last_triggered_format = "relative"
      show_status           = true
      show_investigation    = false
    }
  }

  widget {
    slo_list_definition {
      request {
        request_type = "slo_list"
        query {
          query_string = "team:example"
          rollup {
            type = "month"
          }
        }
      }
    }
  }

  widget {
    timeseries_definition {
      request {
        formula {
          formula_expression = "events"
        }
        query {
          event_query {
            data_source = "events"
            name        = "events"
            compute {
              aggregation = "count"
            }
            group_by {
              facet                  = "host"
              should_exclude_missing = true
            }
          }
        }
        style {
          palette     = "dog_classic"
          color_order = "monotonic"
        }
      }
    }
  }
}
