# Create a new Datadog webhook OAuth2 client credentials auth method.

resource "datadog_webhook_oauth2_client_credentials" "foo" {
  name             = "example-auth-method"
  access_token_url = "https://example.com/oauth2/token"
  client_id        = "example-client-id"
  client_secret    = "example-client-secret"
  audience         = "https://example.com/api"
  scope            = "read write"
}
