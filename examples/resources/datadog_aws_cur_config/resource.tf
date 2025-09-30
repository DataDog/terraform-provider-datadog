# Create new aws_cur_config resource

resource "datadog_aws_cur_config" "foo" {
  account_id    = "123456789123"
  bucket_name   = "dd-cost-bucket"
  bucket_region = "us-east-1"
  report_name   = "dd-report-name"
  report_prefix = "dd-report-prefix"

  account_filters {
    include_new_accounts = true
    excluded_accounts    = ["123456789123", "123456789143"]
  }
}