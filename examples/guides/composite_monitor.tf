resource "datadog_monitor" "bar" {
  name = "Composite Monitor"
  type = "composite"
  message = "This is a message"
  query = "${datadog_monitor.foo.id} || ${datadog_synthetics_test.foo.monitor_id}"
}
