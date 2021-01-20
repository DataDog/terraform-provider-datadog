data "datadog_monitor" "test" {
  name_filter = "My awesome monitor"
  monitor_tags_filter = ["foo:bar"]
}
