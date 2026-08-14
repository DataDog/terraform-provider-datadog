resource "datadog_status_page" "example" {
  name               = "Platform Status"
  domain_prefix      = "platform"
  type               = "internal"
  visualization_type = "bars_and_uptime_percentage"
}
