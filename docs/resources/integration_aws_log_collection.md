---
page_title: "datadog_integration_aws_log_collection"
---

# datadog_integration_aws_log_collection Resource

Provides a Datadog - Amazon Web Services integration log collection resource. This can be used to manage which
AWS services logs are collected from for an account.

## Example Usage

```hcl
# Create a new Datadog - Amazon Web Services integration lambda arn
resource "datadog_integration_aws_log_collection" "main" {
  account_id = "1234567890"
  services = ["lambda"]
}
```

## Argument Reference

The following arguments are supported:

* `account_id` - (Required) Your AWS Account ID without dashes.
* `services` - (Required) A list of services to collect logs from. See the
[api docs](https://docs.datadoghq.com/api/v1/aws-logs-integration/#get-list-of-aws-log-ready-services) for more details on which
services are supported.

### See also
* [Datadog API Reference > Integrations > AWS](https://docs.datadoghq.com/api/v1/aws-integration/)

## Attributes Reference

All provided arguments are exported.

## Import

Amazon Web Services log collection integrations can be imported using the `account ID`.

```
$ terraform import datadog_integration_aws_log_collection.test 1234567890
```
