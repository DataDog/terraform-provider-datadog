resource "datadog_csm_threats_policy" "my_policy" {
  name            = "my_policy"
  description     = "My policy"
  enabled         = true
  host_tags_lists = [["env:prod", "team:backend"], ["env:prod", "team:frontend"]]
}
