
# Create new integration_fastly_account resource


resource "datadog_integration_fastly_account" "foo" {
  api_key = "ABCDEFG123"
  name    = "test-name"
  services {
    id   = "6abc7de6893AbcDe9fghIj"
    tags = ["myTag", "myTag2:myValue"]
  }

}