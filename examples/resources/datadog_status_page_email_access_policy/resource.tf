# Create new datadog_status_page_email_access_policy resource

resource "datadog_status_page" "acme_private_status_page" {
  name               = "Acme Private Status"
  type               = "private"
  domain_prefix      = "acme-private-status"
  visualization_type = "bars_and_uptime_percentage"
}

# Allow every address in the example.com domain.
resource "datadog_status_page_email_access_policy" "allow_domain" {
  page_id = datadog_status_page.acme_private_status_page.id
  pattern = "@example.com"
  effect  = "allow"
}

# Deny a single address. Deny policies take precedence over allow policies.
resource "datadog_status_page_email_access_policy" "deny_address" {
  page_id = datadog_status_page.acme_private_status_page.id
  pattern = "contractor@example.com"
  effect  = "deny"
}
