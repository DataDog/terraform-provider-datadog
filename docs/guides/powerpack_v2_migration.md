---
subcategory: ""
page_title: "Migrating from datadog_powerpack to datadog_powerpack_v2"
description: |-
    Guide for migrating Terraform configurations from datadog_powerpack to datadog_powerpack_v2.
---

### Migrating from datadog_powerpack to datadog_powerpack_v2

~> **Beta Resource** `datadog_powerpack_v2` is currently in beta and may be subject to changes. We recommend testing in non-production environments before adopting it for critical infrastructure.

The [`datadog_powerpack_v2`](../resources/powerpack_v2.md) resource is an updated version of `datadog_powerpack` that improves compliance with Datadog's powerpack API spec. It shares the same widget support as [`datadog_dashboard_v2`](../resources/dashboard_v2.md), giving it consistent widget coverage and serialization behavior. Both resources can coexist in the same Terraform configuration, so you can migrate incrementally.

## What's different

### HCL configuration is mostly identical

The resource type name changes; most attributes and widget blocks remain the same. The primary difference is in how widget queries are expressed — `datadog_powerpack_v2` uses the same query syntax as `datadog_dashboard_v2`.

### Widget query syntax

Legacy metric query strings (`q = "avg:system.cpu.user{*}"`) continue to work. For formula-based queries, use `query`/`formula` blocks:

```terraform
widget {
  timeseries_definition {
    request {
      formula { formula_expression = "query1 / query2" }
      query {
        metric_query {
          name  = "query1"
          query = "sum:requests.count{*}"
        }
      }
      query {
        metric_query {
          name  = "query2"
          query = "sum:requests.errors{*}"
        }
      }
    }
  }
}
```

### Import works the same way

Both resources use the powerpack ID for import:

```bash
# v1
terraform import datadog_powerpack.my_powerpack abc-def-ghi

# v2
terraform import datadog_powerpack_v2.my_powerpack abc-def-ghi
```

## Migration steps

### Step 1: Get the powerpack ID

```bash
terraform state show datadog_powerpack.my_powerpack | grep '^\s*id.*\"'
```

### Step 2: Update the configuration

Change the resource type in your `.tf` file:

```diff
-resource "datadog_powerpack" "my_powerpack" {
+resource "datadog_powerpack_v2" "my_powerpack" {
   name        = "My Powerpack"
   description = "Service health widgets"
   tags        = ["team:platform"]

   widget {
     timeseries_definition {
       title = "CPU Usage"
       request {
         q            = "avg:system.cpu.user{service:web}"
         display_type = "line"
       }
     }
   }
 }
```

Update any references in other resources:

```diff
-  powerpack_id = datadog_powerpack.my_powerpack.id
+  powerpack_id = datadog_powerpack_v2.my_powerpack.id
```

### Step 3: Remove v1 state and import v2

```bash
# Remove the old resource from state (does NOT delete the powerpack from Datadog)
terraform state rm datadog_powerpack.my_powerpack

# Import under the new resource type
terraform import datadog_powerpack_v2.my_powerpack <powerpack-id>
```

### Step 4: Verify

```bash
terraform plan
```

A clean plan with no changes confirms a successful migration.

## Troubleshooting

### Plan shows diffs after import

Common causes:

- **Ordering differences** — If the API returns widgets in a different order than your config, reorder them to match.
- **Default values** — Some fields have server-side defaults (e.g., `show_title`). Add them explicitly to suppress diffs.
- **Query style mismatch** — If your v1 config mixed `q` with `formula`/`query` blocks on the same request, `datadog_powerpack_v2` will flag this as a configuration error. Separate them: use `q` alone for legacy queries, or `query`/`formula` blocks for formula-based queries.

### Both v1 and v2 managing the same powerpack

Never manage the same powerpack with both `datadog_powerpack` and `datadog_powerpack_v2` simultaneously. Always remove the v1 state entry before importing as v2.
