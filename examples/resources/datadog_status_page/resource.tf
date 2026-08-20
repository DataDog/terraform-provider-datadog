# Create new datadog_status_page resource

resource "datadog_status_page" "example" {
  name               = "Example Status Page"
  type               = "public"
  domain_prefix      = "example"
  visualization_type = "bars_and_uptime_percentage"
}
