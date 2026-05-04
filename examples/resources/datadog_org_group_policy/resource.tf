resource "datadog_org_group" "prod" {
  name = "Production Environments"
}

# Disables widget copy-paste for every member org of the prod group.
# enforcement_tier = "DEFAULT" means member orgs can still override the value;
# use "ENFORCE" to make the value immutable for members.
resource "datadog_org_group_policy" "example" {
  org_group_id     = datadog_org_group.prod.id
  policy_name      = "is_widget_copy_paste_enabled"
  content          = jsonencode({ "org_config" : false })
  enforcement_tier = "DEFAULT"
}
