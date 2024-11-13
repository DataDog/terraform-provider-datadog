resource "datadog_domain_allowlist" "example" {
  enabled = true
  domains = ["@gmail.com"]
}