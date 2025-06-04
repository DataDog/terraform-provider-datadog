# Manage Datadog compliance resource evaluation filters
resource "datadog_compliance_resource_evaluation_filter" "basic_filter" {
  tags           = ["tag1:val1"]
  cloud_provider = "aws"
  resource_id    = "000000000000"
}
