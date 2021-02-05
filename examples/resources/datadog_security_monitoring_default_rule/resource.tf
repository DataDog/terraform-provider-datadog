resource "datadog_security_monitoring_default_rule" "adefaultrule" {
    rule_id = "ojo-qef-3g3"
    enabled = true

    # Change the notifications for the high case
    case {
        status = "high"
        notifications = ["@me"]
    }
}
