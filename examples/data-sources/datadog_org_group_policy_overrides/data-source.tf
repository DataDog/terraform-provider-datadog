# List every override attached to the given org group. Useful for discovering
# overrides auto-created by the server in response to membership or policy events.
data "datadog_org_group_policy_overrides" "all" {
  org_group_id = "019d97e8-74d4-7060-806e-e731c8f1d005"
}

# Scope the search to a single policy and/or organization. org_uuid is a
# client-side filter since the API does not natively support it.
data "datadog_org_group_policy_overrides" "for_org" {
  org_group_id = "019d97e8-74d4-7060-806e-e731c8f1d005"
  policy_id    = "019d97e9-f340-78c1-aa50-6a953c2bd006"
  org_uuid     = "ff4a8255-6931-58d1-add0-a6b3602d5421"
}
