---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "datadog_monitor_json Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog monitor JSON resource. This can be used to create and manage Datadog monitors using the JSON definition.
---

# datadog_monitor_json (Resource)

Provides a Datadog monitor JSON resource. This can be used to create and manage Datadog monitors using the JSON definition.

## Example Usage

```terraform
resource "datadog_monitor_json" "monitor_json" {
  monitor = <<-EOF
{
    "name": "Example monitor - service check",
    "type": "service check",
    "query": "\"ntp.in_sync\".by(\"*\").last(2).count_by_status()",
    "message": "Change the message triggers if any host's clock goes out of sync with the time given by NTP. The offset threshold is configured in the Agent's 'ntp.yaml' file.\n\nSee [Troubleshooting NTP Offset issues](https://docs.datadoghq.com/agent/troubleshooting/ntp for more details on cause and resolution.",
    "tags": [],
    "multi": true,
	"restricted_roles": null,
    "options": {
        "include_tags": true,
        "new_host_delay": 150,
        "notify_audit": false,
        "notify_no_data": false,
        "thresholds": {
            "warning": 1,
            "ok": 1,
            "critical": 1
        }
    },
    "priority": null,
    "classification": "custom"
}
EOF
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `monitor` (String) The JSON formatted definition of the monitor.

### Optional

- `url` (String) The URL of the monitor.

### Read-Only

- `id` (String) The ID of this resource.

## Import

Import is supported using the following syntax:

The [`terraform import` command](https://developer.hashicorp.com/terraform/cli/commands/import) can be used, for example:

```shell
terraform import datadog_monitor_json.monitor_json 123456
```
