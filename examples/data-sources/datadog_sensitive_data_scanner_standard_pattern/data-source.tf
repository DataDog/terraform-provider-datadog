data "datadog_sensitive_data_scanner_standard_pattern" "aws_access_key_by_name" {
  filter = "AWS Access Key ID Scanner"
}

data "datadog_sensitive_data_scanner_standard_pattern" "aws_access_key_by_id" {
  standard_pattern_id = data.datadog_sensitive_data_scanner_standard_pattern.aws_access_key_by_name.id
}
