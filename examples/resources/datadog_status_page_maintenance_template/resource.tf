# Create new datadog_status_page_maintenance_template resource

resource "datadog_status_page" "example" {
  name               = "Example Status Page"
  type               = "public"
  domain_prefix      = "example"
  visualization_type = "bars_and_uptime_percentage"

  components {
    name = "API"
    type = "component"
  }
}

resource "datadog_status_page_maintenance_template" "example" {
  page_id                 = datadog_status_page.example.id
  name                    = "Scheduled API maintenance"
  maintenance_title       = "API maintenance"
  component_ids           = [datadog_status_page.example.components[0].id]
  scheduled_description   = "Maintenance is scheduled."
  in_progress_description = "Maintenance is in progress."
  completed_description   = "Maintenance is complete."
}
