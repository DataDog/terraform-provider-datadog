resource "datadog_incident_type" "example" {
  name        = "My Incident Type"
  description = "Incident type for critical production issues"
}

resource "datadog_incident_notification_template" "example" {
  name          = "My Notification Template"
  subject       = "SEV-1 Incident: {{incident.title}}"
  content       = <<-EOF
An incident has been declared.

Title: {{incident.title}}
Severity: {{incident.severity}}
Status: {{incident.status}}

Please join the incident channel for updates.
EOF
  category      = "alert"
  incident_type = datadog_incident_type.example.id
}

resource "datadog_incident_notification_rule" "example" {
  enabled    = true
  trigger    = "incident_created_trigger"
  visibility = "organization"

  handles = [
    "@team-email@company.com",
    "@slack-channel-alerts",
    "@pagerduty-service"
  ]

  # Trigger for SEV-1 and SEV-2 incidents
  conditions {
    field  = "severity"
    values = ["SEV-1", "SEV-2"]
  }

  # Also trigger for incidents affecting production services
  conditions {
    field  = "services"
    values = ["web-service", "api-service", "database-service"]
  }

  # Re-notify when status or severity changes
  renotify_on = ["status", "severity"]

  # Associate with incident type
  incident_type = datadog_incident_type.example.id

  # Use custom notification template
  notification_template = datadog_incident_notification_template.example.id
}