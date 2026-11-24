data "datadog_integration_slack_channels" "example" {
  account_name = "my_slack_account"
}

locals {
  # Names keep the leading "#", matching both the data source output and channel_name.
  wanted_channels = toset(["#alerts-prod", "#alerts-staging"])
}

resource "datadog_integration_slack_channel" "channels" {
  for_each = local.wanted_channels

  account_name = data.datadog_integration_slack_channels.example.account_name
  channel_name = each.key

  display {
    message  = true
    notified = true
    snapshot = false
    tags     = true
  }
}

# Creating a channel already configured in Datadog (for example from the UI) fails, so adopt it instead.
# Importing an address already in state is a no-op, so the block can stay. Requires Terraform 1.7 or later.
import {
  for_each = setintersection(local.wanted_channels, toset(data.datadog_integration_slack_channels.example.channels[*].name))

  to = datadog_integration_slack_channel.channels[each.key]
  id = "${data.datadog_integration_slack_channels.example.account_name}:${each.key}"
}
