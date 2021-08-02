resource "datadog_security_monitoring_filter" "my_filter" {
  name = "My filter"

  query      = "The filter is filtering."
  is_enabled = true

  exclusion_filter {
    name  = "first"
    query = "exclude some logs"
  }

  exclusion_filter {
    name  = "second"
    query = "exclude some other logs"
  }
}
