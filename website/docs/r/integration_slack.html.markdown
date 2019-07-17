---
layout: "datadog"
page_title: "Datadog: datadog_integration_slack"
sidebar_current: "docs-datadog-resource-integration_slack"
description: |-
  Provides a Datadog - Slack integration resource. This can be used to create and manage the integration.
---

# datadog_integration_slack

Provides a Datadog - Slack resource. This can be used to create and manage Datadog - Slack integration.

## Example Usage

```
# Create a new Datadog - Slack integration
resource "datadog_integration_slack" "slack" {
  service_hook {
    account = "testing_account"
    url  = "https://hooks.slack.com/services/1/1"
  }
  
  channel {
    channel_name = "#testing_foo"
    transfer_all_user_comments = true
    account  = "testing_account"
  }
  channel {
    channel_name = "#testing_bar"
    account = "testing_account"
  }
}
```

## Argument Reference

The following arguments are supported:

* `service_hook` - (Required) List of Slack service hook objects.
  * `account` - (Required) Your Slack account name.
  * `url` - (Required) Your Slack account hook URL.
* `channel` - (Required) List of Slack channel objects.
  * `channel_name` - (Required) Your Slack channel name.
  * `transfer_all_user_comments` - (Optional) To be notified for every comment on a graph, set it to true.
  * `account` - (Required) Your Service hook account associated with this channel.


### See also
* [Datadog API Reference > Integration Slack](https://docs.datadoghq.com/api/?lang=bash#integration-slack)
