# Configure agentless scanning for an AWS account
resource "datadog_agentless_scanning_gcp_scan_options" "example" {
  gcp_project_id     = "company-project-prod"
  vuln_containers_os = true
  vuln_host_os       = true
}
