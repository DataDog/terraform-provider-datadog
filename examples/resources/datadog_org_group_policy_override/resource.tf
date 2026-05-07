resource "datadog_org_group" "prod" {
  name = "Production Environments"
}

resource "datadog_org_group_membership" "exempt_org" {
  org_group_id = datadog_org_group.prod.id
  org_uuid     = "ff4a8255-6931-58d1-add0-a6b3602d5421"
}

resource "datadog_org_group_policy" "widget_copy_paste" {
  org_group_id     = datadog_org_group.prod.id
  policy_name      = "is_widget_copy_paste_enabled"
  content          = jsonencode({ "org_config" : false })
  enforcement_tier = "DEFAULT"
}

# Exempts the given organization from the widget_copy_paste policy.
# The org keeps its current value for is_widget_copy_paste_enabled regardless
# of the policy. The resource must target a policy whose tier is not ENFORCE.
resource "datadog_org_group_policy_override" "example" {
  org_group_id = datadog_org_group.prod.id
  policy_id    = datadog_org_group_policy.widget_copy_paste.id
  org_uuid     = "ff4a8255-6931-58d1-add0-a6b3602d5421"
  org_site     = "us1"
  depends_on   = [datadog_org_group_membership.exempt_org]
}
