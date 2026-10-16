---
page_title: "Dashboard resource names"
description: |-
    Understand the canonical datadog_dashboard resource and its datadog_dashboard_v2 alias.
---

### Dashboard resource names

`datadog_dashboard` is the canonical Terraform resource for creating and managing Datadog dashboards. It uses the current API-aligned dashboard implementation documented in the [`datadog_dashboard` resource reference](../resources/dashboard.md).

`datadog_dashboard_v2` is retained as an alias for `datadog_dashboard`. The two resource names have identical schemas and behavior. `datadog_dashboard` is not deprecated; prefer it for new configurations.

Existing `datadog_dashboard_v2` configurations remain supported and do not need to be renamed. Changing a resource type also changes its Terraform address, so review Terraform's state-migration requirements before renaming an existing resource and verify that the resulting plan does not replace the dashboard.
