# Create new reference_table resource

resource "datadog_reference_table" "foo" {
    description = "This is my reference table description"
    schema {
    fields {
    name = "id"
    type = "STRING"
    }
    primary_keys = ["id"]
    }
    file_metadata {
        upload_id = "1234567890"
    }
    source = "LOCAL_FILE"
    table_name = "my_reference_table"
    tags = ["tag1", "tag2"]
}