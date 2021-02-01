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

- **name** (String, Required) Synthetics global variable name.
- **value** (String, Required) The value of the global variable.

### Optional

- **description** (String, Optional) Description of the global variable.
- **id** (String, Optional) The ID of this resource.
- **parse_test_id** (String, Optional) Id of the Synthetics test to use for a variable from test.
- **parse_test_options** (Block List, Max: 1) ID of the Synthetics test to use a source of the global variable value. (see [below for nested schema](#nestedblock--parse_test_options))
- **secure** (Boolean, Optional) Sets the variable as secure. Defaults to `false`.
- **tags** (List of String, Optional) A list of tags to associate with your synthetics global variable.

<a id="nestedblock--parse_test_options"></a>
### Nested Schema for `parse_test_options`

Required:

- **parser** (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--parse_test_options--parser))
- **type** (String, Required) Defines the source to use to extract the value. Allowed enum values: `http_body`, `http_header`.

Optional:

- **field** (String, Optional) Required when type = `http_header`. Defines the header to use to extract the value

<a id="nestedblock--parse_test_options--parser"></a>
### Nested Schema for `parse_test_options.parser`

Required:

- **type** (String, Required) Type of parser to extract the value. Allowed enum values: `raw`, `json_path`, `regex`

Optional:

- **value** (String, Optional) Value for the parser to use, required for type `json_path` or `regex`.

## Import

Import is supported using the following syntax:

```shell
# Synthetics global variables can be imported using their string ID, e.g.
terraform import datadog_synthetics_global_variable.fizz abcde123-fghi-456-jkl-mnopqrstuv
```
