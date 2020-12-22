# Default rules need to be imported using their ID before applying.
resource "datadog_security_monitoring_default_rule" "adefaultrule" {
}

terraform import datadog_security_monitoring_default_rule.adefaultrule m0o-hto-lkb

