# Create new datadog_status_page_degradation_template resource

resource "datadog_status_page" "example" {
  name               = "Example Status Page"
  type               = "public"
  domain_prefix      = "example"
  visualization_type = "bars_and_uptime_percentage"

  components = [
    {
      name = "API"
      type = "component"
    }
  ]
}

resource "datadog_status_page_degradation_template" "example" {
  page_id           = datadog_status_page.example.id
  name              = "API degradation"
  degradation_title = "API is degraded"

  components_affected = [
    {
      id     = datadog_status_page.example.components[0].id
      status = "degraded"
    }
  ]

  updates = [
    {
      message = "We are investigating the issue."
      status  = "investigating"
    }
  ]
}
