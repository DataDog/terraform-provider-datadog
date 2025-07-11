---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "datadog_monitor_notification_rule Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog MonitorNotificationRule resource.
---

# datadog_monitor_notification_rule (Resource)

Provides a Datadog MonitorNotificationRule resource.

## Example Usage

```terraform
# Create new monitor_notification_rule resource

resource "datadog_monitor_notification_rule" "foo" {
  name       = "A notification rule name"
  recipients = ["slack-test-channel", "jira-test"]
  filter {
    tags = ["env:foo"]
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of the monitor notification rule.
- `recipients` (Set of String) List of recipients to notify.

### Optional

- `filter` (Block, Optional) (see [below for nested schema](#nestedblock--filter))

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--filter"></a>
### Nested Schema for `filter`

Required:

- `tags` (Set of String) All tags that target monitors must match.

## Import

Import is supported using the following syntax:

The [`terraform import` command](https://developer.hashicorp.com/terraform/cli/commands/import) can be used, for example:

```shell
terraform import datadog_monitor_notification_rule.new_list "00e000000-0000-1234-0000-000000000000"
```
