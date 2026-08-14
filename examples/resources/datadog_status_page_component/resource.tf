resource "datadog_status_page" "example" {
  name               = "Platform Status"
  domain_prefix      = "platform"
  type               = "internal"
  visualization_type = "bars_only"
}

# A group bundles child components (declared inline; groups cannot be empty).
resource "datadog_status_page_component" "note_generation" {
  status_page_id = datadog_status_page.example.id
  name           = "Note Generation"
  type           = "group"
  position       = 0

  components {
    name     = "Inpatient"
    position = 0
  }
  components {
    name     = "Ambulatory"
    position = 1
  }
}

# A standalone top-level component (no group).
resource "datadog_status_page_component" "api" {
  status_page_id = datadog_status_page.example.id
  name           = "API"
  type           = "component"
  position       = 1
}
