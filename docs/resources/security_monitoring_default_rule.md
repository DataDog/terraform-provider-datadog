---
page_title: "datadog_security_monitoring_default_rule Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog Security Monitoring Rule API resource for default rules.
---

# Resource `datadog_security_monitoring_default_rule`

Provides a Datadog Security Monitoring Rule API resource for default rules.

## Example Usage

```terraform
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

## Schema

### Optional

- **case** (Block List, Max: 5) Cases of the rule, this is used to update notifications. (see [below for nested schema](#nestedblock--case))
- **enabled** (Boolean) Enable the rule.
- **id** (String) The ID of this resource.

<a id="nestedblock--case"></a>
### Nested Schema for `case`

Required:

- **notifications** (List of String) Notification targets for each rule case.
- **status** (String) Status of the rule case to match.

## Import

Import is supported using the following syntax:

```shell
# Default rules need to be imported using their ID before applying.
resource "datadog_security_monitoring_default_rule" "adefaultrule" {
}

terraform import datadog_security_monitoring_default_rule.adefaultrule m0o-hto-lkb
```
