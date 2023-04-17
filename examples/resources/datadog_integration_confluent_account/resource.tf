# Create new integration_confluent_account resource

resource "datadog_integration_confluent_account" "foo" {
  api_key    = "TESTAPIKEY123"
  api_secret = "test-api-secret-123"
  tags       = ["mytag", "mytag2:myvalue"]
}