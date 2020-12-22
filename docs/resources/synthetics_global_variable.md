---
page_title: "datadog_synthetics_global_variable Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog synthetics global variable resource. This can be used to create and manage Datadog synthetics global variables.
---

# Resource `datadog_synthetics_global_variable`

Provides a Datadog synthetics global variable resource. This can be used to create and manage Datadog synthetics global variables.

## Example Usage

```terraform
resource "datadog_synthetics_global_variable" "test_variable" {
    name = "EXAMPLE_VARIABLE"
    description = "Description of the variable"
    tags = ["foo:bar", "env:test"]
    value = "variable-value"
}
```

## Schema

### Required

- **name** (String, Required)
- **value** (String, Required)

### Optional

- **description** (String, Optional)
- **id** (String, Optional) The ID of this resource.
- **secure** (Boolean, Optional)
- **tags** (List of String, Optional)

## Import

Import is supported using the following syntax:

```shell
# Synthetics global variables can be imported using their string ID, e.g.
terraform import datadog_synthetics_global_variable.fizz abcde123-fghi-456-jkl-mnopqrstuv
```
