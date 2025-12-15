# Example: AWS Cloud Inventory Sync Config
resource "datadog_cloud_inventory_sync_config" "aws_example" {
  cloud_provider = "aws"

  aws_account_id            = "123456789012"
  destination_bucket_name   = "my-inventory-bucket"
  destination_bucket_region = "us-east-1"
  destination_prefix        = "inventory/"
}

# Example: Azure Cloud Inventory Sync Config
resource "datadog_cloud_inventory_sync_config" "azure_example" {
  cloud_provider = "azure"

  azure_client_id       = "00000000-0000-0000-0000-000000000000"
  azure_tenant_id       = "00000000-0000-0000-0000-000000000000"
  azure_subscription_id = "00000000-0000-0000-0000-000000000000"
  azure_resource_group  = "my-resource-group"
  azure_storage_account = "mystorageaccount"
  azure_container       = "inventory-container"
}

# Example: GCP Cloud Inventory Sync Config
resource "datadog_cloud_inventory_sync_config" "gcp_example" {
  cloud_provider = "gcp"

  gcp_project_id              = "my-gcp-project"
  gcp_destination_bucket_name = "my-inventory-bucket"
  gcp_source_bucket_name      = "my-source-bucket"
  gcp_service_account_email   = "serviceaccount@example.com"
}

