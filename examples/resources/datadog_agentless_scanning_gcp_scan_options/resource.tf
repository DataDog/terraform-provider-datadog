# Configure agentless scanning for a GCP project
resource "datadog_agentless_scanning_gcp_scan_options" "example" {
  gcp_project_id     = "company-project-prod"
  cloud_function     = true
  vuln_containers_os = true
  vuln_host_os       = true
  # compliance_host  = true  # Optional. Defaults to false. Enables host compliance benchmark scanning.
}
