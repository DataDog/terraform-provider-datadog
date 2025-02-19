# Create an Application Security exclusion filter on a path
resource "datadog_appsec_exclusion_filter" "exclude_on_path" {
  description = "Exclude false positives on a path"
  enabled     = true
  path_glob   = "/accounts/*"
  rules_target {
    tags {
      category = "attack_attempt"
      type     = "lfi"
    }
  }
  scope {
    env     = "www"
    service = "prod"
  }
}

# Create an Application Security exclusion filter for trusted IPs
resource "datadog_appsec_exclusion_filter" "trusted_ips" {
  description = "Office IP network"
  enabled     = true
  ip_list = [
    "198.10.14.53/24"
  ]
  on_match = "monitor"
}
