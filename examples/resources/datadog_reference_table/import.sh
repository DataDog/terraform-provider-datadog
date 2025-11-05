#!/bin/bash
# Import an existing reference table by its UUID

terraform import datadog_reference_table.imported_table "00000000-0000-0000-0000-000000000000"

# After importing, add the resource configuration to your .tf file:
# 
# resource "datadog_reference_table" "imported_table" {
#   table_name  = "existing_table"
#   description = "Previously created table"
#   source      = "S3"  # or "GCS" or "AZURE"
#   
#   file_metadata {
#     sync_enabled = true
#     
#     access_details {
#       aws_detail {
#         aws_account_id  = "123456789000"
#         aws_bucket_name = "existing-bucket"
#         file_path       = "data/existing.csv"
#       }
#     }
#   }
#   
#   schema {
#     primary_keys = ["id"]
#     
#     fields {
#       name = "id"
#       type = "STRING"
#     }
#     
#     fields {
#       name = "value"
#       type = "STRING"
#     }
#   }
#   
#   tags = ["imported:true"]
# }
