# Create new datadog_status_page_component resource

resource "datadog_status_page" "example" {
  name               = "Example Status Page"
  type               = "public"
  domain_prefix      = "example"
  visualization_type = "bars_and_uptime_percentage"
}

resource "datadog_status_page_component" "api" {
  page_id  = datadog_status_page.example.id
  name     = "API"
  type     = "component"
  position = 0
}

resource "datadog_status_page_component" "infrastructure" {
  page_id  = datadog_status_page.example.id
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
