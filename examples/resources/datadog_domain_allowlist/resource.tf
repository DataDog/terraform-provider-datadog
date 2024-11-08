resource "datadog_ip_allowlist" "example" {
  enabled = true
  domains = ["@gmail.com"]
}