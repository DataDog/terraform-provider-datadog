---
page_title: "datadog_logs_archive Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog Logs Archive API resource, which is used to create and manage Datadog logs archives.
---

# Resource `datadog_logs_archive`

Provides a Datadog Logs Archive API resource, which is used to create and manage Datadog logs archives.



## Schema

### Required

- **name** (String, Required)
- **query** (String, Required)

### Optional

- **azure** (Map of String, Optional)
- **gcs** (Map of String, Optional)
- **id** (String, Optional) The ID of this resource.
- **include_tags** (Boolean, Optional)
- **rehydration_tags** (List of String, Optional)
- **s3** (Map of String, Optional)


