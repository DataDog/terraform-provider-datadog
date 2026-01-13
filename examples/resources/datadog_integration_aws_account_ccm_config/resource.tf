# Create new integration_aws_account_ccm_config resource

resource "datadog_integration_aws_account_ccm_config" "foo" {
  aws_account_config_id = "00000000-0000-0000-0000-000000000000"

  ccm_config {
    data_export_configs {
      report_name   = "cost-and-usage-report"
      report_prefix = "reports"
      report_type   = "CUR2.0"
      bucket_name   = "billing"
      bucket_region = "us-east-1"
    }
  }
}
