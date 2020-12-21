---
page_title: "datadog_integration_pagerduty Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog - PagerDuty resource. This can be used to create and manage Datadog - PagerDuty integration. This resource is deprecated and should only be used for legacy purposes.
---

# Resource `datadog_integration_pagerduty`

Provides a Datadog - PagerDuty resource. This can be used to create and manage Datadog - PagerDuty integration. This resource is deprecated and should only be used for legacy purposes.



## Schema

### Required

- **subdomain** (String, Required)

### Optional

- **api_token** (String, Optional)
- **id** (String, Optional) The ID of this resource.
- **individual_services** (Boolean, Optional)
- **schedules** (List of String, Optional)
- **services** (Block List, Deprecated) A list of service names and service keys. (see [below for nested schema](#nestedblock--services))

<a id="nestedblock--services"></a>
### Nested Schema for `services`

Required:

- **service_key** (String, Required)
- **service_name** (String, Required)


