resource "datadog_incident_type" "example" {
  name        = "Security Incident"
  description = "Security-related incidents requiring immediate attention"
}

resource "datadog_incident_user_defined_role" "tech_lead" {
  name          = "Tech Lead"
  description   = "The technical lead for the incident."
  incident_type = datadog_incident_type.example.id

  policy = {
    is_single = true
  }
}
