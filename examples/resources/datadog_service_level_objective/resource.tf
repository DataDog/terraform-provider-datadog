# Metric-Based SLO
# Create a new Datadog service level objective
resource "datadog_service_level_objective" "foo" {
  name        = "Example Metric SLO"
  type        = "metric"
  description = "My custom metric SLO"
  query {
    numerator   = "sum:my.custom.count.metric{type:good_events}.as_count()"
    denominator = "sum:my.custom.count.metric{*}.as_count()"
  }

  thresholds {
    timeframe = "7d"
    target    = 99.9
    warning   = 99.99
  }

  thresholds {
    timeframe = "30d"
    target    = 99.9
    warning   = 99.99
  }

  timeframe         = "30d"
  target_threshold  = 99.9
  warning_threshold = 99.99

  tags = ["foo:bar", "baz"]
}


# Monitor-Based SLO
# Create a new Datadog service level objective
resource "datadog_service_level_objective" "bar" {
  name        = "Example Monitor SLO"
  type        = "monitor"
  description = "My custom monitor SLO"
  monitor_ids = [1, 2, 3]

  thresholds {
    timeframe = "7d"
    target    = 99.9
    warning   = 99.99
  }

  thresholds {
    timeframe = "30d"
    target    = 99.9
    warning   = 99.99
  }

  timeframe         = "30d"
  target_threshold  = 99.9
  warning_threshold = 99.99

  tags = ["foo:bar", "baz"]
}

resource "datadog_service_level_objective" "time_slice_slo" {
  name        = "Example Time Slice SLO"
  type        = "time_slice"
  description = "My custom time slice SLO"
  sli_specification {
    time_slice {
      query {
        formula {
          formula_expression = "query1"
        }
        query {
          metric_query {
            name  = "query1"
            query = "avg:my.custom.count.metric{*}.as_count()"
          }
        }
      }
      comparator = ">"
      threshold  = 0.9
    }
  }

  thresholds {
    timeframe = "7d"
    target    = 99.9
    warning   = 99.99
  }

  timeframe         = "7d"
  target_threshold  = 99.9
  warning_threshold = 99.99

  tags = ["service:myservice", "team:myteam"]
}
