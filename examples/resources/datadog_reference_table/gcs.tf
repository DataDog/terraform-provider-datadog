# Create a reference table from Google Cloud Storage
resource "datadog_reference_table" "gcs_table" {
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

    fields {
      name = "location_id"
      type = "STRING"
    }

    fields {
      name = "store_name"
      type = "STRING"
    }

    fields {
      name = "region"
      type = "STRING"
    }

    fields {
      name = "employee_count"
      type = "INT32"
    }
  }

  tags = ["source:gcs", "team:retail"]
}

