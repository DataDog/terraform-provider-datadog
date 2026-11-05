# Create new datadog_status_page resource

resource "datadog_status_page" "acme_status_page" {
  name               = "Acme Status"
  type               = "public"
  domain_prefix      = "acme-status"
  visualization_type = "bars_and_uptime_percentage"
}
