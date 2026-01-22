# Create a new datastore resource

resource "datadog_datastore" "foo" {
  name                            = "my-datastore"
  description                     = "My application datastore"
  primary_column_name             = "id"
  primary_key_generation_strategy = "none"
  org_access                      = "contributor"
}

# Create a datastore with auto-generated UUIDs for primary keys

resource "datadog_datastore" "auto_uuid" {
  name                            = "my-uuid-datastore"
  description                     = "Datastore with auto-generated primary keys"
  primary_column_name             = "uuid"
  primary_key_generation_strategy = "uuid"
}
