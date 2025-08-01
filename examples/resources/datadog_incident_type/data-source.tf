# Data source examples for datadog_incident_type

# Look up incident type by name
data "datadog_incident_type" "security" {
  name = "Security Incident"
}

# Look up incident type by ID  
data "datadog_incident_type" "by_id" {
  id = "abc123-def456-ghi789"
}

# Use data source output in other resources
resource "datadog_monitor" "example" {
  name    = "Security Monitor"
  type    = "metric alert"
  message = "Security alert - incident type: ${data.datadog_incident_type.security.name}"

  query = "avg(last_5m):avg:system.cpu.user{*} > 0.9"

  # Reference the incident type ID
  tags = ["incident_type:${data.datadog_incident_type.security.id}"]
}

# Output incident type information
output "incident_type_details" {
  value = {
    id          = data.datadog_incident_type.security.id
    name        = data.datadog_incident_type.security.name
    description = data.datadog_incident_type.security.description
    is_default  = data.datadog_incident_type.security.is_default
  }
}