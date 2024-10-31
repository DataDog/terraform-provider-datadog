resource "datadog_ip_allowlist" "example" {
  enabled = false
  domains = ["@test.com", "@datadoghq.com"]
}