resource "datadog_integration_fastly_account" "foo" {
  api_key = "ABCDEFG123"
  name    = "test-name"
}

# Create new integration_fastly_service resource
resource "datadog_integration_fastly_service" "foo" {
  account_id = datadog_integration_fastly_account.foo.id
  tags       = ["mytag", "mytag2:myvalue"]
  service_id = "my-service-id"
}