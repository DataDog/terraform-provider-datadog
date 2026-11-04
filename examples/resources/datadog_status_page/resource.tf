# Create new datadog_status_page resource

resource "datadog_status_page" "acme_status_page" {
  name               = "Acme Status"
  type               = "public"
  domain_prefix      = "acme-status"
  visualization_type = "bars_and_uptime_percentage"
}

# Status page with a grid layout, an RSS feed, and Microsoft Teams and webhook subscriptions
resource "datadog_status_page" "acme_grid_status_page" {
  name                           = "Acme Services"
  type                           = "public"
  domain_prefix                  = "acme-services"
  visualization_type             = "component_name_only"
  visualization_layout           = "grid"
  subscriptions_enabled          = true
  rss_feed_enabled               = true
  ms_teams_subscriptions_enabled = true
  webhooks_subscriptions_enabled = true
}
