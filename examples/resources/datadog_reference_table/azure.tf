# Create a reference table from Azure Blob Storage
resource "datadog_reference_table" "azure_table" {
  table_name  = "warehouse_inventory"
  description = "Inventory data from Azure Blob Storage"
  source      = "AZURE"

  file_metadata {
    sync_enabled = true

    access_details {
      azure_detail {
        azure_tenant_id            = "cccccccc-4444-5555-6666-dddddddddddd"
        azure_client_id            = "aaaaaaaa-1111-2222-3333-bbbbbbbbbbbb"
        azure_storage_account_name = "datadogstorage"
        azure_container_name       = "reference-tables"
        file_path                  = "inventory/warehouse_stock.csv"
      }
    }
  }

  schema {
    primary_keys = ["sku"]

    fields {
      name = "sku"
      type = "STRING"
    }

    fields {
      name = "warehouse_id"
      type = "STRING"
    }

    fields {
      name = "stock_quantity"
      type = "INT32"
    }

    fields {
      name = "status"
      type = "STRING"
    }
  }

  tags = ["source:azure", "team:warehouse"]
}

