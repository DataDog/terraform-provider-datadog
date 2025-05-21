# Manage Datadog resource evaluation filters
resource "datadog_resource_evaluation_filter" "basic_filter" {
  tags           = ["tag1:val1"]
  cloud_provider = "aws"
  id             = "000000000000"
}
