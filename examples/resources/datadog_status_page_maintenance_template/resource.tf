# Create new datadog_status_page_maintenance_template resource

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

resource "datadog_status_page_maintenance_template" "acme_maintenance_template" {
  page_id                 = datadog_status_page.acme_status_page.id
  name                    = "Scheduled Acme API maintenance"
  maintenance_title       = "Acme API maintenance"
  component_ids           = [datadog_status_page_component.acme_api.id]
  scheduled_description   = "Maintenance has been scheduled for the Acme API."
  in_progress_description = "Maintenance is in progress for the Acme API."
  completed_description   = "Maintenance is complete for the Acme API."
}
