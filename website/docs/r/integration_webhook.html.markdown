---
layout: "datadog"
page_title: "Datadog: datadog_integration_webhook"
sidebar_current: "docs-datadog-resource-integration_webhook"
description: |-
  Provides a Datadog - Webhook integration resource. This can be used to create and manage the integrations.
---

# datadog_integration_webhook

Provides a Datadog - Webhook integration resource. This can be used to create and manage Datadog - Webhook integration.

Update operations are currently not supported with datadog API so any change forces a new resource.

## Example Usage

```hcl
# Create a new Datadog - Webhook integration
resource "datadog_integration_webhook" "sandbox" {
    hook {
        name: "sample hook"
        url: "http://example.com
        use_custom_payload: true
        custom_payload: "TITLE: $EVENT_TITLE"
    }

    hook {
        name: "another sample hook"
        url: "http://example2.com
        use_custom_payload: true
        custom_payload: "TITLE: $EVENT_TITLE"
        encode_as_form: true
        headers {
            "X-DataDog-Webhook": "$ID"
        }
    }
}
```

## Argument Reference

The following arguments are supported:

* `hook` - (Required) Nested block describing the webhook. The structure of this block is described [below](integration_webhook.html#nested-hook-blocks)

### Nested `hook` blocks

* `name` - (Required) Name of the webhook
* `url` - (Required) URL to send the message to
* `use_custom_payload` - (Optional) Flag a custom payload should be sent
* `custom_payload` - (Optional) The payload to be sent
* `encode_as_form` - (Optional) Flag if the payload should be form encoded
* `headers` - (Optional) Additional headers to be sent

### See also
* [Datadog API Reference > Integrations > Webhook](https://docs.datadoghq.com/api/?lang=bash#integration-webhooks)

## Import

Webhook integrations can be imported without any parameters being passed

```
terraform import datadog_integration_webhook.test 
```
