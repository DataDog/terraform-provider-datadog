resource "datadog_monitor" "process_alert_example" {
  name = "Process Alert Monitor"
  type = "process alert"
  message = "Multiple Java processes running on example-tag"
  query = "processes('java').over('example-tag').rollup('count').last('10m') > 1",
  thresholds {
    critical          = 1.0
    critical_recovery = 0.0
  }

  notify_no_data    = false
  renotify_interval = 60
}
