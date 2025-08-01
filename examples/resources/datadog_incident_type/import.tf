# Import examples for datadog_incident_type

# To import an existing incident type, use the incident type ID
# terraform import datadog_incident_type.existing <incident-type-id>

# Example: Import an existing incident type
resource "datadog_incident_type" "imported" {
  name        = "Imported Incident Type"
  description = "This incident type was imported from existing infrastructure"
  is_default  = false
}

# After creating the resource block above, run:
# terraform import datadog_incident_type.imported abc123-def456-ghi789

# You can then run terraform plan to see any differences between
# your configuration and the actual state of the resource