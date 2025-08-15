
# Create new dataset resource


resource "datadog_dataset" "foo" {
  name       = "HR Dataset"
  principals = ["role:00000000-0000-1111-0000-000000000000"]

  product_filters {
    product = "rum"
    filters = ["@application.id:123"]
  }
}