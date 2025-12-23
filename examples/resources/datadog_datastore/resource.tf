# Create new datastore resource

resource "datadog_datastore" "foo" {
    description = "UPDATE ME"
    name = "datastore-name"
    org_access = "UPDATE ME"
    primary_column_name = "UPDATE ME"
    primary_key_generation_strategy = "UPDATE ME"
}