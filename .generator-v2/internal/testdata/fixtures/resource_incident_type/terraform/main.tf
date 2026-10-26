# Terraform configuration for the resource_incident_type fixture.
#
# One stanza and one variable: `name` is what the update step changes, so it is
# the only value a step needs to supply (terraform-plugin-testing's
# ConfigVariables, or -var on the CLI). Everything else is a literal — a step
# that needs to vary one of them should promote that field to a variable then,
# rather than the file guessing in advance which fields a future cassette will
# touch.
#
# The nested `configuration` is written as a block, not as an assigned object:
# the generator emits schema.SingleNestedBlock for a nested object, so
# `configuration { ... }` is the syntax the generated resource accepts. The
# hand-written datadog_incident_type models the same field as a nested
# attribute (`configuration = { ... }`) — see this fixture's README.

variable "name" {
  type        = string
  description = "Name of the incident type. Required by the API, and the field the update step changes."
  default     = "tfgen-fixture-incident-type"
}

resource "datadog_incident_type" "fixture" {
  name        = var.name
  description = "Incident type created by the tfgen resource_incident_type fixture."
  prefix      = "TFGEN"
  is_default  = false

  configuration {
    private_incidents            = false
    private_incidents_by_default = false
    allow_workflows              = true
    allow_incident_deletion      = false
    editable_timestamps          = false
    test_incidents               = true
    create_message               = "Declare an incident of this type."
    slug_source                  = "default"
  }
}
