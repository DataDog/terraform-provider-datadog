---
page_title: "datadog_synthetics_global_variable Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog synthetics global variable resource. This can be used to create and manage Datadog synthetics global variables.
---

# Resource `datadog_synthetics_global_variable`

Provides a Datadog synthetics global variable resource. This can be used to create and manage Datadog synthetics global variables.



## Schema

### Required

- **name** (String, Required)
- **value** (String, Required)

### Optional

- **description** (String, Optional)
- **id** (String, Optional) The ID of this resource.
- **secure** (Boolean, Optional)
- **tags** (List of String, Optional)


