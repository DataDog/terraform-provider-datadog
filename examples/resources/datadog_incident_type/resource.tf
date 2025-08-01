# Basic incident type
resource "datadog_incident_type" "example" {
  name        = "Security Incident"
  description = "Security-related incidents requiring immediate attention"
  is_default  = false
}

# Minimal configuration (description is optional)
resource "datadog_incident_type" "minimal" {
  name = "Simple Incident"
}

# Default incident type (only one can be default)
resource "datadog_incident_type" "default" {
  name        = "General Incident"
  description = "Default incident type for unspecified incidents"
  is_default  = true
}

# Incident type with longer description
resource "datadog_incident_type" "detailed" {
  name        = "Application Performance"
  description = "Performance-related incidents affecting application response times, throughput, or availability that require immediate investigation and resolution"
  is_default  = false
}