---
page_title: "datadog_synthetics_global_variable"
---

# datadog_synthetics_global_variable Resource

Provides a Datadog synthetics global variable resource. This can be used to create and manage Datadog synthetics global variables.

## Example Usage

```hcl
resource "datadog_synthetics_global_variable" "test_variable" {
    name = "EXAMPLE_VARIABLE"
    description = "Description of the variable"
    tags = ["foo:bar", "env:test"]
    value = "variable-value"
}
```

## Argument Reference

The following arguments are supported:

-   `name`: (Required) Synthetics global variable name.
-   `description`: (Optional) Description of the global variable.
-   `tags`: (Required) A list of tags to associate with your synthetics global variable.
-   `value`: (Required) The value of the global variable.
-   `secure`: (Optional) Sets the variable as secure, true or false.
-   `parse_test_id`: (Optional) ID of the Synthetics test to use a source of the global variable value.
-   `parse_test_options`: (Optional) Options to extract a value from a Synthetics test.
    -   `type`: (Required) Defines the source to use to extract the value. Allowed enum values: "http_body", "http_header".
    -   `parser`: (Required) Options to define how the value is extracted.
        -   `type`: (Required) Type of parser to extract the value. Allowed enum values: "raw", "json_path", "regex".
        -   `value`: (Optional) Value for the parser to use, required for type "json_path" or "regex".


## Import

Synthetics global variables can be imported using their string ID, e.g.

```
$ terraform import datadog_synthetics_global_variable.fizz abcde123-fghi-456-jkl-mnopqrstuv
```
