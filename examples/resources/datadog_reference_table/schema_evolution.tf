# Example: Schema Evolution - Adding Fields (In-Place Update)
# 
# This example demonstrates how to add new fields to an existing reference table.
# Additive schema changes (adding new fields) are supported and will update the table in-place.
# 
# IMPORTANT: Destructive schema changes are NOT supported:
# - Removing fields
# - Changing field types
# - Changing primary_keys
# 
# For destructive changes, you must manually delete and recreate the table.

# Initial schema
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

    fields {
      name = "customer_id"
      type = "STRING"
    }

    fields {
      name = "name"
      type = "STRING"
    }

    # NEW FIELDS ADDED - This is supported and will update in-place
    # Make sure to update the CSV file in S3 with these new columns first!
    fields {
      name = "email"
      type = "STRING"
    }

    fields {
      name = "signup_year"
      type = "INT32"
    }
  }

  tags = ["team:sales"]
}

# Note: If you need to make destructive changes (remove fields, change types, or change primary_keys):
# 1. Remove the resource from Terraform state: terraform state rm datadog_reference_table.customers
# 2. Update your configuration with the new schema
# 3. Run terraform apply to recreate the table
# 
# Warning: This will cause downtime and any enrichment processors using this table will fail temporarily.

