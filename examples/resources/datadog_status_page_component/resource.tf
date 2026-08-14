resource "datadog_status_page" "example" {
  name               = "Platform Status"
  domain_prefix      = "platform"
  type               = "internal"
  visualization_type = "bars_only"
}

resource "datadog_status_page_component" "group" {
  status_page_id = datadog_status_page.example.id
  name           = "Note Generation"
  type           = "group"
  position       = 0
}

resource "datadog_status_page_component" "api" {
  status_page_id = datadog_status_page.example.id
  name           = "API"
  type           = "component"
  position       = 0
  group_id       = datadog_status_page_component.group.id
}
