## page_title: "datadog_security_monitoring_default_rule"

# datadog_security_monitoring_default_rule Resource

Provides a Datadog [Security Monitoring Rule API](https://docs.datadoghq.com/api/v2/security-monitoring/) resource for default rules.

## Example Usage

Enable a default rule and configure it's notifications.

```hcl
resource "datadog_security_monitoring_default_rule" "adefaultrule" {
    rule_id = "ojo-qef-3g3"
    enabled = true

    # Change the notifications for the high case
    case {
        status = "high"
        notifications = ["@me"]
    }
}
```

## Argument Reference

The following arguments are supported:

-   `enabled`: (Optional, default = True) Whether the default rule is enabled.
-   `case`: (Optional) Change the notifications of a case.
    -   `status`: Severity of the case.
    -   `notifications`: Notification targets for the case.

## Importing

Default rules need to be imported using their ID before applying.

```hcl
resource "datadog_security_monitoring_default_rule" "adefaultrule" {
}
```

```
terraform import datadog_security_monitoring_default_rule.adefaultrule
```
