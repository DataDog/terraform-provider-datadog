# Create a reference table from an S3 bucket
resource "datadog_reference_table" "s3_table" {
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

    fields {
      name = "product_id"
      type = "STRING"
    }

    fields {
      name = "product_name"
    type = "STRING"
    }

    fields {
      name = "category"
      type = "STRING"
    }

    fields {
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