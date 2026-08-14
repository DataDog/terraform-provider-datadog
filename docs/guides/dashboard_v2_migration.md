---
subcategory: ""
page_title: "Migrating from datadog_dashboard to datadog_dashboard_v2"
description: |-
    Guide for migrating Terraform configurations from datadog_dashboard to datadog_dashboard_v2.
---

### Migrating from datadog_dashboard to datadog_dashboard_v2

~> **Beta Resource** `datadog_dashboard_v2` is currently in beta and may be subject to changes. We recommend testing in non-production environments before adopting it for critical infrastructure.

The [`datadog_dashboard_v2`](../resources/dashboard_v2.md) resource is an updated version of `datadog_dashboard` that improves compliance with Datadog's dashboard API spec. Both resources can coexist in the same Terraform configuration, so you can migrate dashboards incrementally.

## What's different

### HCL configuration is identical

The `datadog_dashboard_v2` resource uses the exact same HCL syntax as `datadog_dashboard`. All attributes, widget types, and nested blocks are named identically. The only change in your `.tf` files is the resource type name:

```terraform
# Before
resource "datadog_dashboard" "my_dashboard" {
  title       = "My Service Dashboard"
  layout_type = "ordered"
  # ...
}

# After
resource "datadog_dashboard_v2" "my_dashboard" {
  title       = "My Service Dashboard"
  layout_type = "ordered"
  # ...
}
```

### Import works the same way

Both resources use the dashboard ID for import:

```bash
# v1
terraform import datadog_dashboard.my_dashboard abc-def-ghi

# v2
terraform import datadog_dashboard_v2.my_dashboard abc-def-ghi
```

## Why migrate

### Forward-compatible with new API fields

`datadog_dashboard` lagged behind the Datadog API spec, causing import failures and plan errors when dashboards used fields the provider didn't yet support. `datadog_dashboard_v2` is the preferred resource going forward — it silently ignores unknown fields so imports always succeed, and new API features will be prioritized here.

### Flexible widget time spans

In addition to the existing `live_span` enum (e.g. `live_span = "1h"`), `datadog_dashboard_v2` supports two new widget-level [`time`](../resources/dashboard_v2.md#nestedblock--widget--timeseries_definition--time) configurations:

**Arbitrary live span** — any duration, not limited to the fixed enum values:
```terraform
time {
  live { value = 17; unit = "minute" }
}
```

**Fixed time range** — explicit start and end timestamps:
```terraform
time {
  fixed { from = 1712080128; to = 1712083128 }
}
```

The existing `live_span` field continues to work. `live_span` and `time` are mutually exclusive.

### Funnel widget support

The [`funnel_definition`](../resources/dashboard_v2.md#nestedblock--widget--funnel_definition) widget is available in `datadog_dashboard_v2` but was never implemented in `datadog_dashboard`.

### Toplist sort control

`datadog_dashboard_v2` adds a [`sort`](../resources/dashboard_v2.md#nestedblock--widget--toplist_definition--request--sort) block to toplist widget requests, allowing you to control the sort direction, limit, and whether to sort by formula result or group tag value.

### Formula and query support on more widgets

In `datadog_dashboard`, formula-based queries (`query`/`formula` blocks) were supported on some widgets but not others. `datadog_dashboard_v2` applies formula and query support consistently — most notably adding it to the [**distribution**](../resources/dashboard_v2.md#nestedblock--widget--distribution_definition) widget.

### Correct `number_format` serialization

The [`number_format`](../resources/dashboard_v2.md#nestedblock--widget--change_definition--request--formula--number_format) block on widget formulas is correctly serialized in `datadog_dashboard_v2`. The v1 resource has known issues with `number_format.unit.canonical` and `number_format.unit.custom` blocks on non-query-table widgets.

### Consistent legacy query types across widgets

Legacy query types like `audit_query` and `network_query` were only available on a subset of widgets in v1. In v2, they are available on all standard request widgets (change, distribution, heatmap, query_value, toplist, and sunburst).

## Migration steps

### Step 1: Identify dashboards to migrate

List your existing `datadog_dashboard` resources:

```bash
terraform state list | grep '^datadog_dashboard\.'
```

### Step 2: Get the dashboard ID

For each dashboard you want to migrate, note its ID:

```bash
terraform state show datadog_dashboard.my_dashboard | grep '^\s*id.*\"'
```

### Step 3: Update the configuration

Copy or rename the resource type in your `.tf` file. The only required change is the resource type — all attributes remain the same:

```diff
-resource "datadog_dashboard" "my_dashboard" {
+resource "datadog_dashboard_v2" "my_dashboard" {
   title       = "My Service Dashboard"
   layout_type = "ordered"
   description = "Overview of service health"

   widget {
     timeseries_definition {
       title = "CPU Usage"
       request {
         q = "avg:system.cpu.user{service:web}"
         display_type = "line"
       }
     }
   }
 }
```

If other resources reference this dashboard (e.g., `datadog_dashboard.my_dashboard.id`), update those references too:

```diff
-  dashboard_id = datadog_dashboard.my_dashboard.id
+  dashboard_id = datadog_dashboard_v2.my_dashboard.id
```

### Step 4: Remove v1 state and import v2

```bash
# Remove the old resource from state (does NOT delete the dashboard from Datadog)
terraform state rm datadog_dashboard.my_dashboard

# Import under the new resource type
terraform import datadog_dashboard_v2.my_dashboard <dashboard-id>
```

### Step 5: Verify

Run `terraform plan` to confirm there are no unexpected changes:

```bash
terraform plan
```

A clean plan with no changes confirms a successful migration. If you see diffs, see the troubleshooting section below.

## Troubleshooting

### Plan shows diffs after import

If `terraform plan` shows changes after importing, the most common causes are:

- **Ordering differences in lists** — Terraform may report diffs on list-type attributes (e.g., `custom_link`, `request`) if the API returns items in a different order than the config. Reorder items in your config to match.
- **Default values** — Some fields have server-side defaults that weren't explicitly set in your v1 config. Add them explicitly to suppress the diff (e.g., `show_legend = false`, `legend_layout = "auto"`).
- **Deprecated fields** — If your config uses `is_read_only`, consider migrating to [`restricted_roles`](../resources/dashboard_v2.md#schema). If it uses `default` on a template variable, switch to `defaults` (see [`template_variable`](../resources/dashboard_v2.md#nestedblock--template_variable)).

### Both v1 and v2 managing the same dashboard

Never manage the same dashboard with both `datadog_dashboard` and `datadog_dashboard_v2` simultaneously — they will conflict. Always remove the v1 state entry before importing as v2.

