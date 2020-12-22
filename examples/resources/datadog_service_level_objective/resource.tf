# Metric-Based SLO
# Create a new Datadog service level objective
resource "datadog_service_level_objective" "foo" {
  name               = "Example Metric SLO"
  type               = "metric"
  description        = "My custom metric SLO"
  query {
    numerator = "sum:my.custom.count.metric{type:good_events}.as_count()"
    denominator = "sum:my.custom.count.metric{*}.as_count()"
  }

  thresholds {
    timeframe = "7d"
    target = 99.9
    warning = 99.99
    target_display = "99.900"
    warning_display = "99.990"
  }

  thresholds {
    timeframe = "30d"
    target = 99.9
    warning = 99.99
    target_display = "99.900"
    warning_display = "99.990"
  }

  tags = ["foo:bar", "baz"]
}


# Monitor-Based SLO
# Create a new Datadog service level objective
resource "datadog_service_level_objective" "bar" {
  name               = "Example Monitor SLO"
  type               = "monitor"
  description        = "My custom monitor SLO"
  monitor_ids = [1, 2, 3]

  thresholds {
    timeframe = "7d"
    target = 99.9
    warning = 99.99
  }

  thresholds {
    timeframe = "30d"
    target = 99.9
    warning = 99.99
  }

  tags = ["foo:bar", "baz"]
}
