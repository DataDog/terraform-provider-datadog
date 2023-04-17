# Create new integration_cloudflare_account resource

resource "datadog_integration_cloudflare_account" "foo" {
  api_key = "12345678910abc"
  email   = "test-email@example.com"
  name    = "test-name"
}