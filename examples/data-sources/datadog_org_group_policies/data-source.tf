# List every policy attached to the given org group.
data "datadog_org_group_policies" "all" {
  org_group_id = "019d97e8-74d4-7060-806e-e731c8f1d005"
}

# Or look up a specific policy by name to reference its ID elsewhere.
data "datadog_org_group_policies" "widget_copy_paste" {
  org_group_id = "019d97e8-74d4-7060-806e-e731c8f1d005"
  policy_name  = "is_widget_copy_paste_enabled"
}
