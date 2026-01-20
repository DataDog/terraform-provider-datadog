# AWS Cloud Inventory Sync Config
resource "datadog_cloud_inventory_sync_config" "aws_example" {
  cloud_provider = "aws"

  aws {
    aws_account_id            = "123456789012"
    destination_bucket_name   = "my-inventory-bucket"
    destination_bucket_region = "us-east-1"
    destination_prefix        = "inventory/"
  }
}

# Azure Cloud Inventory Sync Config
resource "datadog_cloud_inventory_sync_config" "azure_example" {
  cloud_provider = "azure"

  azure {
    client_id       = "00000000-0000-0000-0000-000000000000"
    tenant_id       = "00000000-0000-0000-0000-000000000000"
    subscription_id = "00000000-0000-0000-0000-000000000000"
    resource_group  = "my-resource-group"
    storage_account = "mystorageaccount"
    container       = "inventory"
  }
}

# GCP Cloud Inventory Sync Config
resource "datadog_cloud_inventory_sync_config" "gcp_example" {
  cloud_provider = "gcp"

  gcp {
    project_id              = "my-gcp-project"
    destination_bucket_name = "my-inventory-bucket"
    source_bucket_name      = "my-source-bucket"
    service_account_email   = "sa@my-gcp-project.iam.gserviceaccount.com"
  }
}
