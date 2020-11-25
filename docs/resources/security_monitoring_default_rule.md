page_title: "datadog_security_monitoring_default_rule"
---

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
- `rule_id`: (Required) The id of a default rule.
- `enabled`: (Optional) Enable the default rule.
- `disabled`: (Optional) Disable the default rule.
- `case`: (Optional) Change the notifications of a case.
    - `status`: Severity of the case.
    - `notifications`: Notification targets for the case.

## Configuring many rules

It is possible to configure many default rules using the `datadog_security_monitoring_rules` datasource.

```hcl
# List all default rules with tag "security:attack"
datasource "datadog_security_rules" "attack_default_rules" {
    default_only_filter = true
    tags_filter = ["security:attack"]
}

# Create a resource for each default rule id and enable the rule
resource "datadog_security_monitoring_default_rule" "attack_rule" {
    for_each = toset(data.datadog_security_monitoring_rules.rules.rule_ids)
    rule_id = each.key
    enabled = true
}
```
