---
page_title: "datadog_synthetics_global_variable"
---

# datadog_synthetics_global_variable Resource

Provides a Datadog synthetics global variable resource. This can be used to create and manage Datadog synthetics global variables.

## Example Usage

```hcl
resource "datadog_synthetics_global_variable" "test_api" {
    name = "EXAMPLE_VARIABLE"
    description = "Description of the variable"
    tags = ["foo:bar", "env:test"]
    value = "variable-value"
}
```

## Secure global variables

Secure global variables are not supported for now.

## Argument Reference

The following arguments are supported:

- `name`: (Required) Synthetics global variable name
- `description`: (Optional) Description of the global variable
- `tags`: (Required) A list of tags to associate with your synthetics global variable.
- `value`: (Required) The value of the global variable.

## Import

Synthetics global variables can be imported using their string ID, e.g.

```
$ terraform import datadog_synthetics_global_variable.fizz abcde123-fghi-456-jkl-mnopqrstuv
```
