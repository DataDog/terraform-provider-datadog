# Create new datastore resource

resource "datadog_datastore" "foo" {
    description                      = "My application datastore"
    name                             = "my-datastore"
    org_access                       = "contributor"
    primary_column_name              = "id"
    primary_key_generation_strategy  = "none"
}