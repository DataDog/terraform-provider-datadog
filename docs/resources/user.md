---
page_title: "datadog_user Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog user resource. This can be used to create and manage Datadog users.
---

# Resource `datadog_user`

Provides a Datadog user resource. This can be used to create and manage Datadog users.



## Schema

### Required

- **email** (String, Required)

### Optional

- **access_role** (String, Optional)
- **disabled** (Boolean, Optional)
- **handle** (String, Optional, Deprecated)
- **id** (String, Optional) The ID of this resource.
- **is_admin** (Boolean, Optional, Deprecated)
- **name** (String, Optional)
- **role** (String, Optional, Deprecated)
- **roles** (Set of String, Optional)

### Read-only

- **verified** (Boolean, Read-only)


