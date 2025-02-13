---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "datadog_domain_allowlist Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides the Datadog Email Domain Allowlist resource. This can be used to manage the Datadog Email Domain Allowlist.
---

# datadog_domain_allowlist (Resource)

Provides the Datadog Email Domain Allowlist resource. This can be used to manage the Datadog Email Domain Allowlist.

## Example Usage

```terraform
resource "datadog_domain_allowlist" "example" {
  enabled = true
  domains = ["@gmail.com"]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `domains` (List of String) The domains within the domain allowlist.
- `enabled` (Boolean) Whether the Email Domain Allowlist is enabled.

### Read-Only

- `id` (String) The ID of this resource.
