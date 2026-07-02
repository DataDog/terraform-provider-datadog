resource "datadog_incident_type" "example" {
  name        = "Security Incident"
  description = "Security-related incidents requiring immediate attention"
}

# A dropdown user-defined field with a fixed set of valid values.
resource "datadog_incident_user_defined_field" "example" {
  name          = "root_cause"
  display_name  = "Root Cause"
  type          = 1 # 1=dropdown, 2=multiselect, 3=textbox, 4=textarray, 5=metrictag, 6=autocomplete, 7=number, 8=datetime
  category      = "what_happened"
  default_value = "service_bug"
  incident_type = datadog_incident_type.example.id

  valid_values {
    display_name = "Service Bug"
    value        = "service_bug"
    description  = "A bug in the service code."
  }

  valid_values {
    display_name = "Human Error"
    value        = "human_error"
  }
}
