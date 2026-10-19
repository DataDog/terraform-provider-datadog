resource "datadog_org_group" "prod" {
  name = "Production Environments"
}

# Moves the given organization into the prod org group.
resource "datadog_org_group_membership" "example" {
  org_group_id = datadog_org_group.prod.id
  org_uuid     = "ff4a8255-6931-58d1-add0-a6b3602d5421"
}
