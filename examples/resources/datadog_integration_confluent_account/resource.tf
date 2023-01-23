
# Create new integration_confluent_account resource


resource "datadog_integration_confluent_account" "foo" {
  api_key    = "TESTAPIKEY123"
  api_secret = "test-api-secret-123"
  resources {
    id            = "resource-id-123"
    resource_type = "kafka"
    tags          = ["myTag", "myTag2:myValue"]
  }
  tags = ["myTag", "myTag2:myValue"]

}