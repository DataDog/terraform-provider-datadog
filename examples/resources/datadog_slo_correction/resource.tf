# Create a new Datadog SLO correction. slo_id can be derived from slo resource or specify an slo id of an existing SLO.

resource "datadog_service_level_objective" "example_slo" {
  name        = "example slo"
  type        = "metric"
  description = "some updated description about example_slo SLO"
  query {
    numerator   = "sum:my.metric{type:good}.as_count()"
    denominator = "sum:my.metric{type:good}.as_count() + sum:my.metric{type:bad}.as_count()"
  }

  thresholds {
    timeframe = "7d"
    target    = 99.5
    warning   = 99.8
  }
  tags = ["foo:bar"]
}
resource "datadog_slo_correction" "example_slo_correction" {
  category    = "Scheduled Maintenance"
  description = "correction example"
  start       = 1735707000
  end         = 1735718600
  slo_id      = datadog_service_level_objective.example_slo.id
  timezone    = "UTC"
}
resource "datadog_slo_correction" "example_slo_correction_with_recurrence" {
  category    = "Scheduled Maintenance"
  description = "correction example with recurrence"
  start       = 1735707000
  rrule       = "FREQ=DAILY;INTERVAL=3"
  duration    = 3600
  slo_id      = datadog_service_level_objective.example_slo.id
  timezone    = "UTC"
}
