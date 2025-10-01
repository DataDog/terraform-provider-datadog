# Create new GCP Usage Cost configuration resource

resource "datadog_gcp_uc_config" "example" {
  billing_account_id  = "123456_ABCDEF_123456"
  bucket_name         = "my-gcp-cost-bucket"
  export_dataset_name = "billing_export"
  export_prefix       = "datadog_cloud_cost_detailed_usage_export"
  export_project_name = "my-gcp-project"
  service_account     = "datadog-cost-management@my-gcp-project.iam.gserviceaccount.com"
}
