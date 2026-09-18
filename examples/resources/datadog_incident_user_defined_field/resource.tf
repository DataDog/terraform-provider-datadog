resource "datadog_incident_type" "example" {
  name        = "Security Incident"
  description = "Security-related incidents requiring immediate attention"
}

# A dropdown user-defined field with a fixed set of valid values.
resource "datadog_incident_user_defined_field" "example" {
  name          = "root_cause"
  display_name  = "Root Cause"
  type          = "dropdown" # dropdown, multiselect, textbox, textarray, metrictag, autocomplete, number, datetime
  category      = "what_happened"
  default_value = "service_bug"
  incident_type = datadog_incident_type.example.id

  valid_value {
    display_name = "Service Bug"
    value        = "service_bug"
    description  = "A bug in the service code."
  }

  valid_value {
    display_name = "Human Error"
    value        = "human_error"
  }
}
