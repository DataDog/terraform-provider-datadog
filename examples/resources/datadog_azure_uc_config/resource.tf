# Create new Azure Usage Cost configuration resource

resource "datadog_azure_uc_config" "example" {
  account_id = "12345678-1234-abcd-1234-123456789012"
  client_id  = "87654321-4321-dcba-4321-210987654321"
  scope      = "/subscriptions/12345678-1234-abcd-1234-123456789012"

  actual_bill_config {
    export_name       = "my-actual-export"
    export_path       = "exports/actual"
    storage_account   = "mystorageaccount"
    storage_container = "cost-exports"
  }

  amortized_bill_config {
    export_name       = "my-amortized-export"
    export_path       = "exports/amortized"
    storage_account   = "mystorageaccount"
    storage_container = "cost-exports"
  }
}