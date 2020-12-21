---
page_title: "datadog_integration_aws Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  
---

# Resource `datadog_integration_aws`





## Schema

### Required

- **account_id** (String, Required)
- **role_name** (String, Required)

### Optional

- **account_specific_namespace_rules** (Map of Boolean, Optional)
- **excluded_regions** (List of String, Optional)
- **filter_tags** (List of String, Optional)
- **host_tags** (List of String, Optional)
- **id** (String, Optional) The ID of this resource.

### Read-only

- **external_id** (String, Read-only)


