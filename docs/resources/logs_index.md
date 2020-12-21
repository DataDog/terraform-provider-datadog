---
page_title: "datadog_logs_index Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog Logs Index API resource. This can be used to create and manage Datadog logs indexes.
---

# Resource `datadog_logs_index`

Provides a Datadog Logs Index API resource. This can be used to create and manage Datadog logs indexes.



## Schema

### Required

- **filter** (Block List, Min: 1) (see [below for nested schema](#nestedblock--filter))
- **name** (String, Required)

### Optional

- **exclusion_filter** (Block List) (see [below for nested schema](#nestedblock--exclusion_filter))
- **id** (String, Optional) The ID of this resource.

<a id="nestedblock--filter"></a>
### Nested Schema for `filter`

Required:

- **query** (String, Required)


<a id="nestedblock--exclusion_filter"></a>
### Nested Schema for `exclusion_filter`

Optional:

- **filter** (Block List) (see [below for nested schema](#nestedblock--exclusion_filter--filter))
- **is_enabled** (Boolean, Optional)
- **name** (String, Optional)

<a id="nestedblock--exclusion_filter--filter"></a>
### Nested Schema for `exclusion_filter.filter`

Optional:

- **query** (String, Optional)
- **sample_rate** (Number, Optional)


