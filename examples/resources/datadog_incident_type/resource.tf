# Basic incident type
resource "datadog_incident_type" "example" {
  name        = "Security Incident"
  description = "Security-related incidents requiring immediate attention"
  is_default  = false
}

# Incident type with the full configuration block shown at its default values.
# Every field is optional; omitted fields fall back to these same defaults.
resource "datadog_incident_type" "with_configuration" {
  name        = "Customer Impacting"
  description = "Incidents that impact customers"

  configuration = {
    private_incidents            = false
    private_incidents_by_default = false
    allow_workflows              = true
    allow_incident_deletion      = false
    editable_timestamps          = false
    test_incidents               = true
    create_message               = ""
    slug_source                  = "default"
  }
}
