
# Create new integration_confluent_resource resource


resource "datadog_integration_confluent_resource" "foo" {
  account_id = "UPDATE ME"

  resource_type = "kafka"
  tags          = ["myTag", "myTag2:myValue"]

}