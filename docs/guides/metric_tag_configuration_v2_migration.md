---
page_title: "Metric tag configuration v2 migration"
description: |-
    Manage metric tags independently of distribution percentiles and migrate without recreating tag configurations.
---

# Metric tag configuration v2 migration

`datadog_metric_tag_configuration_v2` manages only a metric's tag configuration. It does not enable, disable, or manage distribution percentile aggregations. The metric must already exist; creating this Terraform resource does not submit metric data.

The legacy `datadog_metric_tag_configuration` resource remains supported with unchanged behavior. In particular, it sends `include_percentiles` for distribution metrics even when omitted from HCL. Omitting that argument or adding `ignore_changes` is **not** a safe way to relinquish percentile ownership.

## Choose one owner for each setting

| Configuration | Guidance |
| --- | --- |
| Tags-only v2 | Manages tags and exclude-tags mode; leaves percentiles unchanged. |
| Legacy tag configuration | Retains its existing combined tag/percentile behavior. |
| Legacy and v2 managing the same metric's tag configuration | Unsupported. Remove legacy ownership without destroying the remote configuration before transferring it. |
| Two resources or Terraform states managing the same setting | Unsupported. Terraform does not enforce exclusive ownership of a remote metric. |

A `_v2` resource name does not provide conflict detection. Documentation and migration procedures help users avoid overlapping management; they do not enforce exclusivity. Different metrics can continue to use different resource types while you migrate incrementally.

Standalone percentile and Historical Metrics Ingestion resources are separate follow-up work, not part of this prerequisite change. Do not add references to unavailable resource types. If you need Terraform to continue managing percentiles on a distribution, stay on the legacy resource until the standalone percentile resource is available. Migrating only the tag configuration deliberately leaves percentile enablement unmanaged, though the existing remote value is preserved.

Historical Metrics Ingestion and an explicit tag configuration are incompatible according to the API's eligibility rules. Do not configure them together on the same metric. Independent Terraform resource boundaries do not override product constraints.

## Migrate an existing tag configuration

This example uses a gauge. The same tag ownership handoff applies to a distribution, subject to the percentile-management caveat above. Use your actual existing metric name, metric type, tags, and exclude-tags mode; do not copy the sample settings over a different existing configuration.

1. Back up Terraform state securely and capture the current metric tag configuration and, for a distribution, its percentile setting. State backups may contain secrets; do not commit them.
2. Replace the legacy resource block with a tags-only v2 block preserving all owned tag settings.
3. Add a `removed` block with `destroy = false` for the old address and an `import` block for the new address. Terraform 1.7+ is needed for this declarative handoff; use a Terraform version supported by the provider.
4. Review the plan. It should forget the old address without destroying the remote object and import the same metric under the new address. Stop if it proposes remote tag updates, deletes, replacements, or unexpected percentile changes.
5. Apply the reviewed ownership-only plan once, then run `terraform plan` again. Verify no changes and confirm the tag settings and percentile status are unchanged. Remove the transitional `removed`/`import` blocks afterward if desired.

Before:

```terraform
resource "datadog_metric_tag_configuration" "queue" {
  metric_name       = "app.queue.depth"
  metric_type       = "gauge"
  tags              = ["env", "service"]
  exclude_tags_mode = false
}
```

After:

```terraform
removed {
  from = datadog_metric_tag_configuration.queue

  lifecycle {
    destroy = false
  }
}

resource "datadog_metric_tag_configuration_v2" "queue" {
  metric_name       = "app.queue.depth"
  metric_type       = "gauge"
  tags              = ["env", "service"]
  exclude_tags_mode = false
}

import {
  to = datadog_metric_tag_configuration_v2.queue
  id = "app.queue.depth"
}
```

Do not apply an intermediate configuration that simply removes the old block: Terraform would destroy the remote tag configuration. A bare resource-type rename is also insufficient. This resource does not implement automatic cross-type `moved` support. Operators on versions without declarative removal/import can perform an equivalent backed-up `terraform state rm` followed immediately by `terraform import`, without applying between those operations.

When standalone percentile management becomes available, import its resource only for metrics whose percentiles are already enabled and should remain managed. For disabled metrics, do not declare a presence-means-enabled percentile resource: creating one would enable the feature. Verify a no-op plan after the complete handoff.

## Read-after-write behavior

Metric configuration reads can be eventually consistent. Creates and updates retain acknowledged state and wait up to one minute for tag settings to become visible. Deletes wait for the tag configuration to become absent. The provider does not use a fixed sleep or a PostgreSQL fallback. A verification timeout does not necessarily mean the acknowledged remote write failed; inspect the remote configuration and rerun Terraform to reconcile. Ordinary refresh reports external tag drift instead of waiting for the desired state to reappear.

Destroying a distribution's tag configuration restores its default tag behavior; it does not delete the metric or disable percentiles. Confirm the target site's API supports preserving omitted percentile settings before rollout, and validate this contract in a sandbox before production adoption.

## Scope and compatibility

The v2 resource intentionally omits the legacy `include_percentiles` and deprecated `aggregations` arguments. Use `tags = []` with `exclude_tags_mode = false` to configure no queryable tag keys. Exclude-tags mode requires at least one tag. Changing `metric_name` or `metric_type` replaces the Terraform resource; changing tags or exclude-tags mode updates it in place.
