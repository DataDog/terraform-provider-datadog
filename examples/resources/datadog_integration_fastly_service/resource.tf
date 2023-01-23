
# Create new integration_fastly_service resource


resource "datadog_integration_fastly_service" "foo" {
  account_id = "UPDATE ME"

  tags = ["myTag", "myTag2:myValue"]

}