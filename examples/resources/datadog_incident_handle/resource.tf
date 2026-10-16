resource "datadog_incident_type" "critical" {
  name        = "Critical"
  description = "Customer-impacting incidents"
}

# Mentioning @incident-p1 in a monitor notification declares a SEV-1 Critical incident.
resource "datadog_incident_handle" "p1" {
  name             = "@incident-p1"
  incident_type_id = datadog_incident_type.critical.id
  severity         = ["SEV-1"]
}
