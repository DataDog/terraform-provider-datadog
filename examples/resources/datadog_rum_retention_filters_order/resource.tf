# Create new rum_application resource

resource "datadog_rum_application" "my_rum_application" {
  name = "my-rum-application-test"
  type = "browser"
}

# Retrieve rum_retention_filters for rum_application created above

data "datadog_rum_retention_filters" "my_retention_filters" {
  application_id = resource.datadog_rum_application.my_rum_application.id
}

# Create new rum_retention_filter resource.
# 'depends_on' is to prevent creating rum_retention_filter and retrieving rum_retention_filters from running in parallel for race condition.

resource "datadog_rum_retention_filter" "new_rum_retention_filter" {
  application_id = resource.datadog_rum_application.my_rum_application.id
  name = "testing.rum.retention_filter"
  event_type = "action"
  sample_rate = 60
  query = "@session.has_replay:true"
  enabled = true
  depends_on = [data.datadog_rum_retention_filters.my_retention_filters]
}

# Create new rum_retention_filters_order resource for reordering

resource "datadog_rum_retention_filters_order" "my_rum_retention_filters_order" {
  application_id = resource.datadog_rum_application.my_rum_application.id
  retention_filter_ids = concat([
    for rf in data.datadog_rum_retention_filters.my_retention_filters.retention_filters :
      rf.id if startswith(rf.id, "default")
    ], [datadog_rum_retention_filter.new_rum_retention_filter.id])
}
