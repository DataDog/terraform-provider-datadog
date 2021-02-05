data "datadog_security_monitoring_rules" "test" {
  name_filter         = "attack"
  tags_filter         = ["foo:bar"]
  default_only_filter = true
}
