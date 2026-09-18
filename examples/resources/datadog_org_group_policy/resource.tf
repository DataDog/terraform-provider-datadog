resource "datadog_org_group" "prod" {
  name = "Production Environments"
}

# Disables widget copy-paste for every member org of the prod group.
# enforcement_tier = "OVERRIDE_ALLOWED" means member orgs can still override the value;
# use "GROUP_MANAGED" to make the value immutable for members.
resource "datadog_org_group_policy" "example" {
  org_group_id     = datadog_org_group.prod.id
  policy_name      = "is_widget_copy_paste_enabled"
  content          = jsonencode({ "org_config" : false })
  enforcement_tier = "OVERRIDE_ALLOWED"
}

# Provisions a shared role with the given permissions into every member org.
# role policies only support GROUP_MANAGED/DELEGATE; DELEGATE disables the
# shared role and cannot be reversed.
resource "datadog_org_group_policy" "role_example" {
  org_group_id     = datadog_org_group.prod.id
  policy_name      = "finance_read_only"
  policy_type      = "role"
  content          = jsonencode({ "permissions" : ["<permission-uuid>"] })
  enforcement_tier = "GROUP_MANAGED"
}
