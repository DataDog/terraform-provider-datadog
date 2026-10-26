# Create new datadog_status_page_component resource

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

resource "datadog_status_page_component" "acme_infrastructure" {
  page_id  = datadog_status_page.acme_status_page.id
  name     = "Infrastructure"
  type     = "group"
  position = 1

  components = [
    {
      name     = "US Region"
      type     = "component"
      position = 0
    },
    {
      name     = "EU Region"
      type     = "component"
      position = 1
    }
  ]
}
