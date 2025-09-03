# Configure agentless scanning for an AWS account
resource "datadog_agentless_scanning_aws_scan_options" "example" {
  aws_account_id     = "123456789012"
  lambda             = true
  sensitive_data     = false
  vuln_containers_os = true
  vuln_host_os       = true
}
