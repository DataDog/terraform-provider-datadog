---
layout: "datadog"
page_title: "Datadog: datadog_integration_slack"
sidebar_current: "docs-datadog-resource-integration_slack"
description: |-
  Provides a Datadog - Slack integration resource. This can be used to create and manage the integrations.
---

# datadog_integration_slack

Provides a Datadog - Slack integration resource. This can be used to create, manage and update Datadog - Slack integration.

## Example Usage

```hcl
# Create a new Datadog - Slack integration
resource "datadog_integration_slack" "foo1" {
	account = "slack_pro_account"
	url = "http:///myslack/url"
   
   channels {
	   channel_name = "#prod-outage"
	   transfer_all_user_comments = false       
   }

   channels {
	   channel_name = "#staging-outage"
	   transfer_all_user_comments = false       
   }
}
```

## Argument Reference

The following arguments are supported:

* `account` - (Required) The account name you want to set.
* `url` - (Required) Slack webhook url.
* `channels` - (Required) Set of Slack channels.
  * `channel_name` - (Required) The Slack Channel name to use.
  * `transfer_all_user_comments` - (Optional) If you would like to be notified for every comment on a graph, set it to `True`.

### See also
* [Datadog API Reference > Integrations > Slack](https://docs.datadoghq.com/api/?lang=bash#integration-slack)
