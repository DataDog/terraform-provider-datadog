---
page_title: "Dashboard resource names"
description: |-
    Understand the canonical datadog_dashboard resource and its datadog_dashboard_v2 alias.
---

### Dashboard resource names

`datadog_dashboard` is the recommended Terraform resource for creating and managing Datadog dashboards. For the full schema, see the [`datadog_dashboard` resource reference](../resources/dashboard.md).

`datadog_dashboard_v2` is retained as an alias for `datadog_dashboard`. The two resource names have identical schemas and behavior. `datadog_dashboard` is not deprecated. Use `datadog_dashboard` for new configurations.

Existing `datadog_dashboard_v2` configurations remain supported and do not need to be renamed. Changing a resource type also changes its Terraform address, so review Terraform's state-migration requirements before renaming an existing resource and verify that the resulting plan does not replace the dashboard.

### Upgrading from the previous `datadog_dashboard` implementation

Existing state loads unchanged and most configurations need no changes. The following configurations now fail at `terraform plan`. Setting `validate = false` does not skip these checks:

- **`q` combined with `query` or `formula` in the same widget request.** Use one query style per request: either `q` alone, or `query` with `formula`. Previously, `q` with `query` silently dropped the `query` blocks, and `q` with `formula` sent both.
- **More than one block where only one is supported:** `request.fill` and `request.size` on hostmap widgets, `request.x` and `request.y` on scatterplot widgets, `style` and `style.display` on toplist widgets, and `request.query` on topology map widgets. Previously, only the first block was sent to the API and the rest produced a permanent diff.
- **`powerpack_definition` or `split_graph_definition` inside a `group_definition`.** The API does not accept these widgets inside groups.

Other behavior changes:

- **Plan-time validation.** Widgets are validated against the Datadog API during `terraform plan` (`validate` defaults to `true`), so plans now need API access. Set `validate = false` to skip this API call, for example for offline plans.
- **Deprecated fields.** `q` and the other legacy request query fields (`apm_query`, `log_query`, `rum_query`, `security_query`, `process_query`) and toplist `style.display.type` are deprecated and now show a warning. They keep working; prefer `query` with `formula`, and the `flat {}` or `stacked {}` display blocks.
- **Widget IDs.** Widget IDs from state are sent on update, so widget links and tab mappings keep pointing at the same widget while widget order is unchanged. IDs follow list position: after reordering or inserting widgets, an existing ID can be reassigned to a different widget.

For `datadog_dashboard_v2` users, note widgets with an empty `content` now fail at plan, matching `datadog_dashboard`.
