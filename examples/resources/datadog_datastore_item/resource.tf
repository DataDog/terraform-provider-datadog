# Create new datastore item resource

resource "datadog_datastore" "example" {
  name                            = "my-datastore"
  description                     = "Example datastore for items"
  primary_column_name             = "id"
  primary_key_generation_strategy = "none"
}

resource "datadog_datastore_item" "foo" {
  datastore_id = datadog_datastore.example.id
  item_key     = "user-123"
  value = {
    id       = "user-123"
    username = "john_doe"
    email    = "john@example.com"
    status   = "active"
  }
}
