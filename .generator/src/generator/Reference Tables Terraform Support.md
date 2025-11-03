# Terraform Support for Reference Tables \- Design Proposal

| Authors | [Guillaume Brizolier](mailto:guillaume.brizolier@datadoghq.com) |
| :---- | :---- |
| **Teams** | REDAPL Experiences |
| **Date** | Oct 29, 2025 |
| **Reviewers** | [Kevin Tung](mailto:kevin.tung@datadoghq.com)Not started  [Karan Bokil](mailto:karan.bokil@datadoghq.com)Approved  |
| **Status** | Review |

## Overview

This document proposes the design for adding Reference Tables support to the Datadog Terraform provider. The implementation will enable infrastructure-as-code management of reference table resources, with support for cloud file tables (S3, GCS, Azure) only.

**Terraform implementation**

- Resources: `datadog_reference_table`: Manage table lifecycle for cloud storage (S3/GCS/Azure) sources  
- Data sources: `datadog_reference_table`, `datadog_reference_table_rows`

**Key Design Decisions:**

- **Cloud Storage Only (v1)**: Initial release supports S3, GCS, and Azure Blob Storage sources. Terraform manages metadata; data syncs from external storage.  
- **LOCAL\_FILE Deferred**: Local file upload support deferred to future release (see "Why LOCAL\_FILE is Deferred" section)  
- **Schema evolution**: Only additive changes supported (adding fields). Destructive changes (remove fields, change primary\_keys/types) are not supported and will result in plan errors.  
- **Import**: Fully supported; all configuration is retrievable from API  
- **State**: Stores user-configured metadata (schema, tags, access\_details, sync\_enabled); computed values (row\_count, status) don't trigger drift

**Implementation:** Terraform Plugin Framework, schema change analysis in [ModifyPlan](https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification)

**Rationale for Cloud Storage Focus:**

Terraform users typically already manage infrastructure-as-code with cloud providers and have established patterns for storing data in S3, GCS, or Azure Blob Storage. This approach:
- Aligns with cloud-native best practices
- Leverages existing cloud storage infrastructure
- Provides automatic sync capabilities
- Avoids complex multipart upload orchestration in Terraform
- Simplifies state management and drift detection

## Resource Design

### `datadog_reference_table`

Primary resource for managing reference table lifecycle. Supports cloud storage sources: **S3**, **GCS**, and **Azure Blob Storage**.

**API Endpoint Mapping:**

| Operation | Endpoint | Notes |
| :---- | :---- | :---- |
| Create | POST /api/v2/reference-tables/tables | Create table with access\_details and sync\_enabled |
| Read | GET /api/v2/reference-tables/tables/{id} | Refresh computed attributes |
| Update | PATCH /api/v2/reference-tables/tables/{id} | Update metadata/schema/access\_details/sync\_enabled |
| Delete | DELETE /api/v2/reference-tables/tables/{id} | Remove table |
| Import | GET /api/v2/reference-tables/tables/{id} | Populate state from existing table |

**Schema structure:**

- Required: `table_name`, `source` (S3/GCS/AZURE), `schema` (with `primary_keys` and `fields` list), `file_metadata` block  
- Optional: `description`, `tags`  
- Computed: `id`, `row_count`, `status`, `created_by`, `last_updated_by`, `updated_at`

**Key behaviors:**

- Schema change analysis in ModifyPlan: additive changes (adding fields) update in-place via PATCH; destructive changes (removing fields, changing types/primary\_keys) return validation errors
- Drift detection compares user-configured attributes (schema, tags, sync\_enabled, access\_details) only
- Computed attributes (row\_count, status, timestamps) refresh but don't trigger drift

## Data Source Design

### `datadog_reference_table`

Query reference tables by `table_name` (through the list endpoint with exact name matching) or `id`. Returns full metadata including schema, row_count, status, and source configuration.

**Supported source types**: ALL source types are supported in the data source, including:
- **Manageable via Terraform resource**: S3, GCS, AZURE
- **Read-only (external integrations)**: SERVICENOW, SALESFORCE, DATABRICKS, SNOWFLAKE, LOCAL_FILE

The data source uses GET operations which return all source types. Read-only source types cannot be created/managed via the `datadog_reference_table` resource (they're managed externally), but they can be queried and their data can be read.

### `datadog_reference_table_rows`

Query specific rows by primary key values via GET /api/v2/reference-tables/tables/{id}/rows. Works with **all source types**.

## Why LOCAL\_FILE is Deferred

The initial release focuses exclusively on cloud storage sources (S3, GCS, Azure). LOCAL\_FILE support is deferred due to edge cases that pose significant UX risks:

### Core Edge Cases

**1. Drift Detection Brittleness**

Local file change detection creates fragile state management as the Reference Tables API currently does not have a good way to retrieve total row content:
- **Out-of-band changes**: If a user updates the table via UI/API, Terraform won't detect drift
- **No server-side validation**: Can't compare local file contents with server state (API doesn't store original file)

**2. Import Sync Risk**

The import workflow forces unexpected data synchronization:
```
1. terraform import datadog_reference_table.prod EXISTING_TABLE_ID
2. User adds file_path = "local_data.csv" to configuration
3. terraform plan shows: "Will update table with local file"
4. terraform apply → Overwrites production data!
```
This behavior is surprising and carries **intrinsic risk of accidental production data loss**. Users expect import to be read-only, not trigger data changes.

### Cloud Storage Advantages

Terraform users are **already managing cloud infrastructure** and typically have data in S3/GCS/Azure. Cloud storage sources:
- Eliminate file content drift detection issues (server manages sync state)
- Support automatic sync with `sync_enabled` flag
- Enable import without data overwrite risks (all config in API response)

## Implementation Details

### Schema Change Validation

ModifyPlan analyzes schema changes:

- **Additive (fields added)**: In-place update → PATCH with new schema  
- **Destructive (fields removed, primary\_keys changed, types changed)**: Plan error → User must manually delete and recreate

**Rationale for blocking destructive changes:**

Destructive schema changes require table recreation, which causes downtime of several minutes:
- Table must be deleted before new one is created (standard Terraform behavior)
- During this window, enrichment processors and log pipelines referencing the table will fail
- `table_name` uniqueness prevents `create_before_destroy` lifecycle mitigation
- Risk of accidental data loss if user doesn't understand the implications

Instead, users who need destructive changes must explicitly:
1. Remove the resource from Terraform (`terraform state rm`)
2. Recreate the table with new schema

This explicit workflow ensures users are aware of the breaking change and downtime.

| Change Type | Action | API Calls | Notes |
| :---- | :---- | :---- | :---- |
| Adding fields | In-place update | PATCH with new schema | API supports additive changes |
| Removing fields | Plan error | N/A | User must manually delete and recreate table |
| Changing primary\_keys | Plan error | N/A | User must manually delete and recreate table |
| Changing field types | Plan error | N/A | User must manually delete and recreate table |
| Cloud storage path/config | In-place update | PATCH with new access\_details | Update bucket/path without schema change |
| Sync settings | In-place update | PATCH with new sync\_enabled | Enable/disable automatic sync |
| Description or tags | In-place update | PATCH with new values | Metadata changes only |

### Cloud Storage Workflow

Terraform manages metadata/schema, not CSV data (external storage).

- **Create**: POST with `access_details` and `sync_enabled`; API reads CSV asynchronously from cloud storage
- **Update**: PATCH for metadata, sync settings, additive schema changes; destructive changes blocked
- **Delete**: Removes table; source CSV file in cloud storage remains

### State Management

- **Stored**: `table_name`, `source`, `schema`, `description`, `tags`, `file_metadata` (user-configured); `id`, `row_count`, `status`, timestamps (computed)
- **Drift detection**: Compares user-configured attributes only; computed attributes refresh but don't trigger drift

## Error Handling

- **Destructive schema changes** → Plan error with manual recreation guidance
- **API validation errors** (schema/format/quota) → AddError with details
- **Partial success** (status=DONE, error\_row\_count \> 0) → AddWarning, allow apply to succeed

## Import Support

Command: `terraform import datadog_reference_table.<local_name> <table_id>`

All configuration is retrievable from API. State is fully populated after import with drift detection working immediately. No forced synchronization or data overwrites.

## Possible Future Enhancements

1. **LOCAL\_FILE Resource Support**: Local file upload resource with multipart upload handling (deferred due to edge cases - see "Why LOCAL\_FILE is Deferred"). Note: LOCAL_FILE tables can already be queried via data sources.
2. **Destructive Schema Changes**: Support for removing fields or changing types/primary_keys with explicit acknowledgment flags (e.g., `allow_destructive_changes = true`) and clear downtime warnings
3. **Bulk Import**: Import multiple tables at once with `for_each` pattern support  
4. **Configurable Timeouts**: Resource-level create/update timeout configuration for large cloud storage files  
5. **Schema Validation**: Pre-flight schema validation before CREATE/UPDATE operations
6. **Lifecycle Hooks**: Support for `create_before_destroy` with temporary name suffixes (requires API support for name conflicts)
7. **List Data Source**: `datadog_reference_tables` (plural) data source for listing/filtering multiple tables at once

---

# Appendix: Terraform Configuration Examples

This section provides practical examples for implementing reference tables in Terraform.

## Resource: `datadog_reference_table`

### Example 1: S3 Cloud Storage Source

```hcl
resource "datadog_reference_table" "products_s3" {
  table_name  = "products_catalog"
  description = "Product catalog synced from S3"
  source      = "S3"
  
  file_metadata {
    sync_enabled = true
    
    access_details {
      aws_detail {
        aws_account_id  = "123456789000"
        aws_bucket_name = "my-data-bucket"
        file_path       = "reference-tables/products.csv"
      }
    }
  }
  
  schema {
    primary_keys = ["product_id"]
    
    field {
      name = "product_id"
      type = "STRING"
    }
    
    field {
      name = "product_name"
      type = "STRING"
    }
    
    field {
      name = "category"
      type = "STRING"
    }
    
    field {
      name = "price_cents"
      type = "INT32"
    }
  }
  
  tags = [
    "source:s3",
    "team:catalog",
    "env:production"
  ]
}
```

### Example 2: GCS Cloud Storage Source

```hcl
resource "datadog_reference_table" "locations_gcs" {
  table_name  = "store_locations"
  description = "Store location data from Google Cloud Storage"
  source      = "GCS"
  
  file_metadata {
    sync_enabled = true
    
    access_details {
      gcp_detail {
        gcp_project_id            = "my-gcp-project-12345"
        gcp_bucket_name           = "datadog-reference-tables"
        file_path                 = "data/store_locations.csv"
        gcp_service_account_email = "datadog-sa@my-gcp-project-12345.iam.gserviceaccount.com"
      }
    }
  }
  
  schema {
    primary_keys = ["location_id"]
    
    field {
      name = "location_id"
      type = "STRING"
    }
    
    field {
      name = "store_name"
      type = "STRING"
    }
    
    field {
      name = "region"
      type = "STRING"
    }
    
    field {
      name = "employee_count"
      type = "INT32"
    }
  }
  
  tags = ["source:gcs", "team:retail"]
}
```

### Example 3: Azure Blob Storage Source

```hcl
resource "datadog_reference_table" "inventory_azure" {
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
    
    field {
      name = "sku"
      type = "STRING"
    }
    
    field {
      name = "warehouse_id"
      type = "STRING"
    }
    
    field {
      name = "stock_quantity"
      type = "INT32"
    }
    
    field {
      name = "status"
      type = "STRING"
    }
  }
  
  tags = ["source:azure", "team:warehouse"]
}
```

### Example 4: Schema Evolution - Adding Fields (In-Place Update)

```hcl
# Before: original schema
resource "datadog_reference_table" "customers" {
  table_name  = "customer_data"
  description = "Customer reference data"
  source      = "S3"
  
  file_metadata {
    sync_enabled = true
    
    access_details {
      aws_detail {
        aws_account_id  = "123456789000"
        aws_bucket_name = "my-data-bucket"
        file_path       = "customers.csv"
      }
    }
  }
  
  schema {
    primary_keys = ["customer_id"]
    
    field {
      name = "customer_id"
      type = "STRING"
    }
    
    field {
      name = "name"
      type = "STRING"
    }
  }
  
  tags = ["team:sales"]
}

# After: add new fields (in-place update, no downtime)
# Step 1: Update the CSV in S3 with the new columns
# Step 2: Update Terraform config to add the new fields
resource "datadog_reference_table" "customers" {
  table_name  = "customer_data"
  description = "Customer reference data"
  source      = "S3"
  
  file_metadata {
    sync_enabled = true
    
    access_details {
      aws_detail {
        aws_account_id  = "123456789000"
        aws_bucket_name = "my-data-bucket"
        file_path       = "customers.csv"
      }
    }
  }
  
  schema {
    primary_keys = ["customer_id"]
    
    field {
      name = "customer_id"
      type = "STRING"
    }
    
    field {
      name = "name"
      type = "STRING"
    }
    
    # New fields added - in-place update via PATCH
    field {
      name = "email"
      type = "STRING"
    }
    
    field {
      name = "signup_year"
      type = "INT32"
    }
  }
  
  tags = ["team:sales"]
}

# Note: Removing fields or changing types/primary_keys will result in a plan error.
# For destructive changes, you must manually delete and recreate the table.
```

### Example 5: Import Existing Table

```bash
# Import an existing reference table
terraform import datadog_reference_table.imported_table "00000000-0000-0000-0000-000000000000"
```

```hcl
# After import, define the configuration
resource "datadog_reference_table" "imported_table" {
  table_name  = "existing_table"
  description = "Previously created table"
  source      = "S3"
  
  file_metadata {
    sync_enabled = true
    
    access_details {
      aws_detail {
        aws_account_id  = "123456789000"
        aws_bucket_name = "existing-bucket"
        file_path       = "data/existing.csv"
      }
    }
  }
  
  schema {
    primary_keys = ["id"]
    
    field {
      name = "id"
      type = "STRING"
    }
    
    field {
      name = "value"
      type = "STRING"
    }
  }
  
  tags = ["imported:true"]
}
```

## Data Source: `datadog_reference_table`

The data source supports two mutually exclusive query methods:
- **By `table_name`**: Human-readable name (recommended for most use cases)
- **By `id`**: UUID returned from the API (useful when you only have the ID)

### Example 1: Query by Table Name

```hcl
# Query using the table name
data "datadog_reference_table" "users" {
  table_name = "users_reference_table"  # The actual table name in Datadog
}

# Use the data source outputs
output "users_table_info" {
  value = {
    id          = data.datadog_reference_table.users.id
    description = data.datadog_reference_table.users.description
    source      = data.datadog_reference_table.users.source
    status      = data.datadog_reference_table.users.status
    row_count   = data.datadog_reference_table.users.row_count
    created_by  = data.datadog_reference_table.users.created_by
    updated_at  = data.datadog_reference_table.users.updated_at
  }
}

# Access schema information
output "users_primary_keys" {
  value = data.datadog_reference_table.users.schema.primary_keys
}

output "users_fields" {
  value = [for field in data.datadog_reference_table.users.schema.fields : {
    name = field.name
    type = field.type
  }]
}
```

### Example 2: Query by Table ID

```hcl
# Query using the UUID (if you only have the ID)
data "datadog_reference_table" "lookup_by_uuid" {
  id = "a1b2c3d4-1234-5678-90ab-cdef12345678"  # The UUID from Datadog API
}

output "table_name_from_id" {
  value = data.datadog_reference_table.lookup_by_uuid.table_name
}

output "table_description" {
  value = data.datadog_reference_table.lookup_by_uuid.description
}
```

### Example 3: Query Read-Only Source Types (External Integrations)

```hcl
# Query a ServiceNow table (created via ServiceNow integration, not manageable via Terraform resource)
data "datadog_reference_table" "servicenow_incidents" {
  table_name = "servicenow_incidents"
}

output "servicenow_table_info" {
  value = {
    id         = data.datadog_reference_table.servicenow_incidents.id
    source     = data.datadog_reference_table.servicenow_incidents.source  # Will be "SERVICENOW"
    row_count  = data.datadog_reference_table.servicenow_incidents.row_count
    status     = data.datadog_reference_table.servicenow_incidents.status
  }
}

# Query rows from the ServiceNow table
data "datadog_reference_table_rows" "recent_incidents" {
  table_id = data.datadog_reference_table.servicenow_incidents.id
  row_ids  = ["INC0001", "INC0002", "INC0003"]
}
```

### Example 4: Use Data Source to Reference Another Resource

```hcl
# Create a reference table
resource "datadog_reference_table" "products" {
  table_name = "product_catalog"
  source     = "S3"
  
  file_metadata {
    sync_enabled = true
    
    access_details {
      aws_detail {
        aws_account_id  = "123456789000"
        aws_bucket_name = "my-data-bucket"
        file_path       = "products.csv"
      }
    }
  }
  
  schema {
    primary_keys = ["product_id"]
    
    field {
      name = "product_id"
      type = "STRING"
    }
    
    field {
      name = "product_name"
      type = "STRING"
    }
  }
}

# Query the table via data source
data "datadog_reference_table" "products_info" {
  table_name = datadog_reference_table.products.table_name
  
  # This ensures data source refresh happens after resource updates
  depends_on = [datadog_reference_table.products]
}

# Use in a log pipeline or other resource
output "reference_table_id" {
  value       = data.datadog_reference_table.products_info.id
  description = "Use this ID in enrichment processors"
}
```

## Data Source: `datadog_reference_table_rows`

### Example 1: Query Specific Rows

```hcl
data "datadog_reference_table_rows" "specific_users" {
  table_id = datadog_reference_table.products_s3.id
  
  row_ids = [
    "user123",
    "user456",
    "user789"
  ]
}

output "user_rows" {
  value = data.datadog_reference_table_rows.specific_users.rows
}

# Example output structure:
# rows = [
#   {
#     id = "user123"
#     values = {
#       account_id = "user123"
#       user_name = "John Doe"
#       tier = "premium"
#       max_connections = "100"
#     }
#   },
#   {
#     id = "user456"
#     values = {
#       account_id = "user456"
#       user_name = "Jane Smith"
#       tier = "standard"
#       max_connections = "50"
#     }
#   }
# ]
```

### Example 2: Using Row Data in Locals

```hcl
data "datadog_reference_table" "config" {
  table_name = "service_configuration"
}

data "datadog_reference_table_rows" "service_configs" {
  table_id = data.datadog_reference_table.config.id
  
  row_ids = [
    "web-service",
    "api-service",
    "worker-service"
  ]
}

# Extract configuration values
locals {
  service_configs = {
    for row in data.datadog_reference_table_rows.service_configs.rows :
    row.values.service_name => row.values
  }
}

# Use in other resources
output "web_service_config" {
  value = local.service_configs["web-service"]
}
```
