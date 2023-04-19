resource "datadog_integration_confluent_account" "foo" {
  api_key    = "TESTAPIKEY123"
  api_secret = "test-api-secret-123"
  tags       = ["mytag", "mytag2:myvalue"]
}

# Create new integration_confluent_resource resource
resource "datadog_integration_confluent_resource" "foo" {
  account_id    = datadog_integration_confluent_account.foo.id
  resource_id   = "123456"
  resource_type = "kafka"
  tags          = ["mytag", "mytag2:myvalue"]
}