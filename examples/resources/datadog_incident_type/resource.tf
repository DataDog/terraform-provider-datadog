# Basic incident type
resource "datadog_incident_type" "example" {
  name        = "Security Incident"
  description = "Security-related incidents requiring immediate attention"
  is_default  = false
}

# Incident type with behavior configuration
resource "datadog_incident_type" "with_configuration" {
  name        = "Customer Impacting"
  description = "Incidents that impact customers"

  configuration = {
    private_incidents            = true
    private_incidents_by_default = false
    allow_workflows              = true
    allow_incident_deletion      = false
    editable_timestamps          = true
    test_incidents               = false
    create_message               = "Follow the SEV runbook before declaring."
    slug_source                  = "servicenow"
  }
}
