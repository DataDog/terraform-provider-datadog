---
subcategory: ""
page_title: "Migrating from datadog_dashboard to datadog_dashboard_v2"
description: |-
    Guide for migrating Terraform configurations from datadog_dashboard to datadog_dashboard_v2.
---

### Migrating from datadog_dashboard to datadog_dashboard_v2

The `datadog_dashboard_v2` resource is the successor to `datadog_dashboard`. It provides the same dashboard management capabilities with a more maintainable implementation built on the Terraform Plugin Framework. Both resources can coexist in the same Terraform configuration, so you can migrate dashboards incrementally.

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

### All widget types are supported

Both resources support the same set of widget types: timeseries, query_value, toplist, heatmap, hostmap, change, distribution, geomap, scatterplot, sunburst, treemap, query_table, list_stream, SLO, SLO list, split graph, group, alert graph, alert value, free text, iframe, image, note, event stream, event timeline, check status, log stream, manage status, run workflow, service map, topology map, trace service, and powerpack.

### All dashboard-level attributes are supported

Top-level attributes like `title`, `layout_type`, `description`, `reflow_type`, `template_variable`, `template_variable_preset`, `notify_list`, `restricted_roles`, `dashboard_lists`, and `tags` all work identically.

### Import works the same way

Both resources use the dashboard ID for import:

```bash
# v1
terraform import datadog_dashboard.my_dashboard abc-def-ghi

# v2
terraform import datadog_dashboard_v2.my_dashboard abc-def-ghi
```

## Why migrate

### Formula and query support on more widgets

In `datadog_dashboard`, formula-based queries (`query`/`formula` blocks) were supported on some widgets but not others. In `datadog_dashboard_v2`, formula and query support is applied consistently. The most notable addition is the **distribution** widget, which now supports formula queries:

```terraform
# This works in datadog_dashboard_v2 but NOT in datadog_dashboard
resource "datadog_dashboard_v2" "example" {
  title       = "Distribution with Formulas"
  layout_type = "ordered"

  widget {
    distribution_definition {
      title = "Request Latency Distribution"
      request {
        query {
          metric_query {
            name       = "query1"
            query      = "avg:trace.http.request.duration{service:web}"
            data_source = "metrics"
            aggregator  = "avg"
          }
        }
        formula {
          formula_expression = "query1"
        }
      }
    }
  }
}
```

### Consistent legacy query types across widgets

The v1 resource had inconsistent support for legacy query types depending on the widget. For example, `audit_query` was only available on query_value, sunburst, timeseries, and toplist — but not on change or heatmap. The v2 resource standardizes these through shared field groups:

| Legacy query type | v1 widget coverage | v2 widget coverage |
|---|---|---|
| `audit_query` | query_value, sunburst, timeseries, toplist | All standard request widgets |
| `network_query` | sunburst, timeseries | All standard request widgets |
| `profile_metrics_query` | None | timeseries |

"All standard request widgets" includes: change, distribution, heatmap, query_value, toplist, and sunburst.

### Forward-compatible with new API fields

A common pain point with `datadog_dashboard` is the "object contains unparsed element" error that occurs when importing dashboards that contain fields the provider doesn't yet recognize. This happens because v1 uses typed Go structs to deserialize API responses — any unknown field causes an error.

`datadog_dashboard_v2` uses a map-based JSON engine that only reads the fields it knows about and silently ignores the rest. This means:
- Dashboards created or modified in the Datadog UI always import cleanly, even if they use features the provider hasn't added yet
- Provider upgrades never break existing imports
- You won't encounter errors like `object contains unparsed element: map[hide_incomplete_cost_data:true]` or similar

This addresses issues reported in [#2925](https://github.com/DataDog/terraform-provider-datadog/issues/2925), [#2827](https://github.com/DataDog/terraform-provider-datadog/issues/2827), and [#3148](https://github.com/DataDog/terraform-provider-datadog/issues/3148).

### Correct `number_format` serialization

The `number_format` block on widget formulas — which controls unit display (e.g., showing a metric in bytes, seconds, or a custom label) — is correctly serialized in `datadog_dashboard_v2`. The v1 resource has known issues with the `number_format.unit.canonical` and `number_format.unit.custom` blocks on non-query-table widgets, where the required `type` discriminator is not injected correctly.

### Maintained and actively developed

New widget types and fields from the Datadog OpenAPI spec are added to `datadog_dashboard_v2`. The `datadog_dashboard` resource is in maintenance mode — it will continue to work but will not receive new field additions.

### Built on the Terraform Plugin Framework

`datadog_dashboard_v2` uses the [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework), which is HashiCorp's recommended foundation for new providers. This provides:

- **Better plan-time validation** — Invalid values are caught during `terraform plan` rather than at apply time.
- **Clearer error messages** — Field-level descriptions and validators produce more informative diagnostics.
- **Consistent behavior** — Optional fields use `UseStateForUnknown` plan modifiers, reducing spurious diffs on computed fields.

### Schema driven by OpenAPI spec

The v2 resource's schema is generated from declarative `FieldSpec` definitions that mirror the Datadog OpenAPI spec. This means the Terraform schema stays closely aligned with the API, and new fields can be added with a single declaration rather than paired build/flatten functions.

## Migration steps

### Step 1: Identify dashboards to migrate

List your existing `datadog_dashboard` resources:

```bash
terraform state list | grep '^datadog_dashboard\.'
```

### Step 2: Get the dashboard ID

For each dashboard you want to migrate, note its ID:

```bash
terraform state show datadog_dashboard.my_dashboard | grep '^\s*id'
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

   template_variable {
     name    = "env"
     prefix  = "env"
     default = "production"
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

## Migrating multiple dashboards

For bulk migrations, you can script the state operations:

```bash
#!/bin/bash
# For each v1 dashboard, extract its ID, remove from state, and import as v2
for resource in $(terraform state list | grep '^datadog_dashboard\.'); do
  local_name="${resource#datadog_dashboard.}"
  dash_id=$(terraform state show "$resource" | grep '^\s*id' | awk -F'"' '{print $2}')

  echo "Migrating $resource (ID: $dash_id) -> datadog_dashboard_v2.$local_name"
  terraform state rm "$resource"
  terraform import "datadog_dashboard_v2.$local_name" "$dash_id"
done
```

Run `terraform plan` after the script completes to verify all dashboards imported cleanly.

## Troubleshooting

### Plan shows diffs after import

If `terraform plan` shows changes after importing, the most common causes are:

- **Ordering differences in lists** — Terraform may report diffs on list-type attributes (e.g., `custom_link`, `request`) if the API returns items in a different order than the config. Reorder items in your config to match.
- **Default values** — Some fields have server-side defaults that weren't explicitly set in your v1 config. Add them explicitly to suppress the diff (e.g., `show_legend = false`, `legend_layout = "auto"`).
- **Deprecated fields** — If your config uses `is_read_only`, consider migrating to `restricted_roles`. If it uses `default` on a template variable, switch to `defaults`.

### Both v1 and v2 managing the same dashboard

Never manage the same dashboard with both `datadog_dashboard` and `datadog_dashboard_v2` simultaneously — they will conflict. Always remove the v1 state entry before importing as v2.

### Computed fields causing spurious diffs

If a computed field (like `url`) shows up as a diff, run `terraform apply` once to sync the state. Subsequent plans should be clean.
