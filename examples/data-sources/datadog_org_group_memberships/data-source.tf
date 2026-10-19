# List every org that is a member of the given org group.
data "datadog_org_group_memberships" "by_group" {
  org_group_id = "019d97e8-74d4-7060-806e-e731c8f1d005"
}

# Or look up the group an organization currently belongs to.
data "datadog_org_group_memberships" "by_org" {
  org_uuid = "ff4a8255-6931-58d1-add0-a6b3602d5421"
}
