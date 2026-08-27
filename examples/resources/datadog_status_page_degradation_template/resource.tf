# Create new datadog_status_page_degradation_template resource

resource "datadog_status_page" "acme_status_page" {
  name               = "Acme Status"
  type               = "public"
  domain_prefix      = "acme-status"
  visualization_type = "bars_and_uptime_percentage"
}

resource "datadog_status_page_component" "acme_api" {
  page_id  = datadog_status_page.acme_status_page.id
  name     = "API"
  type     = "component"
  position = 0
}

resource "datadog_status_page_degradation_template" "acme_degradation_template" {
  page_id           = datadog_status_page.acme_status_page.id
  name              = "Acme API Degradation"
  degradation_title = "Acme API is degraded"

  components_affected = [
    {
      id     = datadog_status_page_component.acme_api.id
      status = "degraded"
    }
  ]

  updates = [
    {
      message = "We are investigating issues with the Acme API."
      status  = "investigating"
    },
    {
      message = "We have resolved the Acme API issues."
      status  = "resolved"
    }
  ]
}
