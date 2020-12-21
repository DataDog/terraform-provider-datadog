---
page_title: "datadog_security_monitoring_default_rule Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog Security Monitoring Rule API resource for default rules.
---

# Resource `datadog_security_monitoring_default_rule`

Provides a Datadog Security Monitoring Rule API resource for default rules.



## Schema

### Optional

- **case** (Block List, Max: 5) Cases of the rule, this is used to update notifications. (see [below for nested schema](#nestedblock--case))
- **enabled** (Boolean, Optional) Enable the rule.
- **id** (String, Optional) The ID of this resource.

<a id="nestedblock--case"></a>
### Nested Schema for `case`

Required:

- **notifications** (List of String, Required) Notification targets for each rule case.
- **status** (String, Required) Status of the rule case to match.


