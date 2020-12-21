---
page_title: "datadog_monitor Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog monitor resource. This can be used to create and manage Datadog monitors.
---

# Resource `datadog_monitor`

Provides a Datadog monitor resource. This can be used to create and manage Datadog monitors.



## Schema

### Required

- **message** (String, Required)
- **name** (String, Required)
- **query** (String, Required)
- **type** (String, Required)

### Optional

- **enable_logs_sample** (Boolean, Optional)
- **escalation_message** (String, Optional)
- **evaluation_delay** (Number, Optional)
- **force_delete** (Boolean, Optional)
- **id** (String, Optional) The ID of this resource.
- **include_tags** (Boolean, Optional)
- **locked** (Boolean, Optional)
- **new_host_delay** (Number, Optional)
- **no_data_timeframe** (Number, Optional)
- **notify_audit** (Boolean, Optional)
- **notify_no_data** (Boolean, Optional)
- **priority** (Number, Optional)
- **renotify_interval** (Number, Optional)
- **require_full_window** (Boolean, Optional)
- **silenced** (Map of Number, Optional, Deprecated)
- **tags** (Set of String, Optional)
- **threshold_windows** (Map of String, Optional)
- **thresholds** (Map of String, Optional)
- **timeout_h** (Number, Optional)
- **validate** (Boolean, Optional)


