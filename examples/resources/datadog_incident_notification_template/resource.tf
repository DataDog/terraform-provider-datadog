# Create a notification template for incident alerts
resource "datadog_incident_notification_template" "security_incident" {
  name          = "Security Incident Template"
  subject       = "SEV-1 Security Incident: {{incident.title}}"
  content       = <<-EOF
🚨 SECURITY INCIDENT DECLARED 🚨

**Incident Details:**
- Title: {{incident.title}}
- Severity: {{incident.severity}}
- Status: {{incident.status}}
- Declared at: {{incident.created}}

**Affected Services:**
{{#each incident.services}}
- {{name}}
{{/each}}

**Commander:** {{incident.commander}}

**Next Steps:**
1. Join the incident Slack channel: #incident-{{incident.id}}
2. Review the incident details in Datadog
3. Await further instructions from the incident commander

For more information: {{incident.url}}
EOF
  category      = "alert"
  incident_type = datadog_incident_type.security.id
}

# Reference incident type
resource "datadog_incident_type" "security" {
  name        = "Security Incident"
  description = "Security-related incidents requiring immediate attention"
}