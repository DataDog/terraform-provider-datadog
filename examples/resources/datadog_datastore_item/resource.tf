# Create a datastore and add items to it

resource "datadog_datastore" "example" {
  name                            = "users-datastore"
  description                     = "Datastore for user data"
  primary_column_name             = "id"
  primary_key_generation_strategy = "none"
}

# Create a datastore item with the primary key specified in the value map

resource "datadog_datastore_item" "user1" {
  datastore_id = datadog_datastore.example.id
  item_key     = "user-123"
  value = {
    id       = "user-123"
    username = "john_doe"
    email    = "john@example.com"
    status   = "active"
  }
}

# Create another datastore item

resource "datadog_datastore_item" "user2" {
  datastore_id = datadog_datastore.example.id
  item_key     = "user-456"
  value = {
    id       = "user-456"
    username = "jane_doe"
    email    = "jane@example.com"
    status   = "active"
  }
}
