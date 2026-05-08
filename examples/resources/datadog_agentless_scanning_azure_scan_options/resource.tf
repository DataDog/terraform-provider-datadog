# Configure agentless scanning for an Azure subscription
resource "datadog_agentless_scanning_azure_scan_options" "example" {
  azure_subscription_id = "12345678-1234-1234-1234-123456789012"
  compliance_host       = true
  vuln_containers_os    = true
  vuln_host_os          = true
}
