# Cloud Inventory Sync Configs can be imported using the ID format specific to each cloud provider.

# AWS format: aws:<aws_account_id>:<destination_bucket_name>
terraform import datadog_cloud_inventory_sync_config.aws_example aws:123456789012:my-inventory-bucket

# Azure format: azure:<azure_client_id>:<azure_storage_account>:<azure_container>
terraform import datadog_cloud_inventory_sync_config.azure_example azure:00000000-0000-0000-0000-000000000000:mystorageaccount:inventory-container

# GCP format: gcp:<gcp_project_id>:<gcp_destination_bucket_name>
terraform import datadog_cloud_inventory_sync_config.gcp_example gcp:my-gcp-project:my-inventory-bucket

