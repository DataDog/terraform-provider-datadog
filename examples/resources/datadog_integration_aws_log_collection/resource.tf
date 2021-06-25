# Create a new Datadog - Amazon Web Services integration log collection
resource "datadog_integration_aws_log_collection" "main" {
  account_id = "1234567890"
  services   = ["lambda"]
}
