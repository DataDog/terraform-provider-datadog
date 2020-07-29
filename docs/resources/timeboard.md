---
page_title: "datadog_timeboard"
---

# datadog_timeboard Resource

Provides a Datadog timeboard resource. This can be used to create and manage Datadog timeboards.

~> **Note:**This resource is outdated. Use the new [`datadog_dashboard`](dashboard.html) resource instead.

## Example Usage

```hcl
# Create a new Datadog timeboard
resource "datadog_timeboard" "redis" {
  title       = "Redis Timeboard (created via Terraform)"
  description = "created using the Datadog provider in Terraform"
  read_only   = true

  graph {
    title = "Redis latency (ms)"
    viz   = "timeseries"

    request {
      q    = "avg:redis.info.latency_ms{$host}"
      type = "bars"

      # NOTE: this will only work with TF >= 0.12; see metadata_json
      # documentation below for example on usage with TF < 0.12
      metadata_json = jsonencode({
        "avg:redis.info.latency_ms{$host}": {
          "alias": "Redis latency"
        }
      })
    }

    request {
        log_query {
          index = "mcnulty"
          compute {
            aggregation = "avg"
            facet = "@duration"
            interval = 5000
          }
          search {
            query = "status:info"
          }
          group_by {
            facet = "host"
            limit = 10
            sort {
              aggregation = "avg"
              order = "desc"
              facet = "@duration"
            }
          }
        }
        type = "area"
      }

      request {
        apm_query {
          index = "apm-search"
          compute {
            aggregation = "avg"
            facet = "@duration"
            interval = 5000
          }
          search {
            query = "type:web"
          }
          group_by {
            facet = "resource_name"
            limit = 50
            sort {
              aggregation = "avg"
              order = "desc"
              facet = "@string_query.interval"
            }
          }
        }
        type = "bars"
      }

      request {
        process_query {
          metric = "process.stat.cpu.total_pct"
          search_by = "error"
          filter_by = ["active"]
          limit = 50
        }
        type = "area"
      }
  }

  graph {
    title = "Redis memory usage"
    viz   = "timeseries"

    request {
      q       = "avg:redis.mem.used{$host} - avg:redis.mem.lua{$host}, avg:redis.mem.lua{$host}"
      stacked = true
    }

    request {
      q = "avg:redis.mem.rss{$host}"

      style = {
        palette = "warm"
      }
    }
  }

  graph {
    title = "Top System CPU by Docker container"
    viz   = "toplist"

    request {
      q = "top(avg:docker.cpu.system{*} by {container_name}, 10, 'mean', 'desc')"
    }
  }

  template_variable {
    name   = "host"
    prefix = "host"
  }
}
```

## Argument Reference

The following arguments are supported:

- `title` - (Required) The name of the dashboard.
- `description` - (Required) A description of the dashboard's content.
- `read_only` - (Optional) The read-only status of the timeboard. Default is false.
- `graph` - (Required) Nested block describing a graph definition. The structure of this block is described below. Multiple graph blocks are allowed within a datadog_timeboard resource.
- `template_variable` - (Optional) Nested block describing a template variable. The structure of this block is described below. Multiple template_variable blocks are allowed within a datadog_timeboard resource.

### Nested `graph` blocks

Nested `graph` blocks have the following structure:

- `title` - (Required) The name of the graph.
- `viz` - (Required) The type of visualization to use for the graph. Valid choices are "change", "distribution", "heatmap", "hostmap", "query_value", timeseries", and "toplist".
- `request` - Nested block describing a graph definition request (a metric query to plot on the graph). The structure of this block is described below. Multiple request blocks are allowed within a graph block.
- `events` - (Optional) A list of event filter strings. Note that, while supported by the Datadog API, the Datadog UI does not (currently) support multiple event filters very well, so use at your own risk.
- `autoscale` - (Optional) Boolean that determines whether to autoscale graphs.
- `precision` - (Optional) Number of digits displayed, use `*` for full precision.
- `custom_unit` - (Optional) Display a custom unit on the graph (such as 'hertz')
- `text_align` - (Optional) How to align text in the graph, can be one of 'left', 'center', or 'right'.
- `style` - (Optional) Nested block describing hostmaps. The structure of this block is described below.
- `group` - (Optional) List of groups for hostmaps (shown as 'group by' in the UI).
- `include_no_metric_hosts` - (Optional) If set to true, will display hosts on hostmap that have no reported metrics.
- `include_ungrouped_hosts` - (Optional) If set to true, will display hosts without groups on hostmaps.
- `node_type` - (Optional) What nodes to display in a hostmap. Can be one of 'host' (default) or 'container'.
- `scope` - (Optional) List of scopes for hostmaps (shown as 'filter by' in the UI).
- `yaxis` - (Optional) Nested block describing modifications to the yaxis rendering. The structure of this block is described below.
- `marker` - (Optional) Nested block describing lines / ranges added to graph for formatting. The structure of this block is described below. Multiple marker blocks are allowed within a graph block.

### Nested `graph` `marker` blocks

Nested `graph` `marker` blocks have the following structure:

- `type` - (Required) How the marker lines will look. Possible values are {"error", "warning", "info", "ok"} {"dashed", "solid", "bold"}. Example: "error dashed".
- `value` - (Required) Mathematical expression describing the marker. Examples: "y > 1", "-5 < y < 0", "y = 19".
- `label` - (Optional) A label for the line or range. **Warning:** when a label is enabled but left empty through the UI, the Datadog API returns a boolean value, not a string. This makes `terraform plan` fail with a JSON decoding error.

### Nested `graph` `yaxis` block

- `min` - (Optional) Minimum bound for the graph's yaxis, a string.
- `max` - (Optional) Maximum bound for the graph's yaxis, a string.
- `scale` - (Optional) How to scale the yaxis. Possible values are: "linear", "log", "sqrt", "pow##" (eg. pow2, pow0.5, 2 is used if only "pow" was provided). Default: "linear".

### Nested `graph` `request` blocks

Nested `graph` `request` blocks have the following structure (exactly only one of `q`, `apm_query`, `log_query` or `process_query` is required within the request block):

- `q` - (Optional) The query of the request. Pro tip: Use the JSON tab inside the Datadog UI to help build you query strings.
- `apm_query` - (Optional) The APM query to use in the widget. The structure of this block is described [below](timeboard.html#nested-graph-request-apm_query-and-log_query-blocks).
- `log_query` - (Optional) The log query to use in the widget. The structure of this block is described [below](timeboard.html#nested-graph-request-apm_query-and-log_query-blocks).
- `process_query` - (Optional) The process query to use in the widget. The structure of this block is described [below](timeboard.html#nested-graph-request-process_query-blocks).
- `aggregator` - (Optional) The aggregation method used when the number of data points outnumbers the max that can be shown.
- `stacked` - (Optional) Boolean value to determine if this is this a stacked area graph. Default: false (line chart).
- `type` - (Optional) Choose how to draw the graph. For example: "line", "bars" or "area". Default: "line".
- `style` - (Optional) Nested block to customize the graph style.
- `conditional_format` - (Optional) Nested block to customize the graph style if certain conditions are met. Currently only applies to `Query Value` and `Top List` type graphs.
- `extra_col` - (Optional, only for graphs of visualization "change") If set to "present", displays current value. Can be left empty otherwise.
- `metadata_json` - (Optional) A JSON blob (preferrably created using [jsonencode](https://www.terraform.io/docs/configuration/functions/jsonencode.html)) representing mapping of query expressions to alias names. Note that the query expressions in `metadata_json` will be ignored if they're not present in the query. For example, this is how you define `metadata_json` with Terraform >= 0.12:

  ```
  metadata_json = jsonencode({
    "avg:redis.info.latency_ms{$host}": {
      "alias": "Redis latency"
    }
  })
  ```

  And here's how you define `metadata_json` with Terraform < 0.12:

  ```
  variable "my_metadata" {
    default = {
      "avg:redis.info.latency_ms{$host}" = {
        "alias": "Redis latency"
      }
    }
  }

  resource "datadog_timeboard" "SomeTimeboard" {
    ...
        metadata_json = "${jsonencode(var.my_metadata)}"
  }
  ```

  Note that this has to be a JSON blob because of [limitations](https://github.com/hashicorp/terraform/issues/6215) of Terraform's handling complex nested structures. This is also why the key is called `metadata_json` even though it sets `metadata` attribute on the API call.

### Nested `graph` `style` block

The nested `style` block is used specifically for styling `hostmap` graphs, and has the following structure:

- `fill_max` - (Optional) Maximum value for the hostmap fill query.
- `fill_min` - (Optional) Minimum value for the hostmap fill query.
- `palette` - (Optional) Spectrum of colors to use when styling a hostmap. For example: "green_to_orange", "yellow_to_green", "YlOrRd", or "hostmap_blues". Default: "green_to_orange".
- `palette_flip` - (Optional) Flip how the hostmap is rendered. For example, with the default palette, low values are represented as green, with high values as orange. If palette_flip is "true", then low values will be orange, and high values will be green.

### Nested `graph` `request` `style` block

The nested `style` blocks has the following structure:

- `palette` - (Optional) Color of the line drawn. For example: "classic", "cool", "warm", "purple", "orange" or "gray". Default: "classic".
- `width` - (Optional) Line width. Possible values: "thin", "normal", "thick". Default: "normal".
- `type` - (Optional) Type of line drawn. Possible values: "dashed", "solid", "dotted". Default: "solid".

### Nested `graph` `request` `apm_query` and `log_query` blocks

Nested `apm_query` and `log_query` blocks have the following structure (Visit the [ Graph Primer](https://docs.datadoghq.com/graphing/) for more information about these values):

- `index` - (Required)
- `compute` - (Required). Exactly one nested block is required with the following structure:
  - `aggregation` - (Required)
  - `facet` - (Optional)
  - `interval` - (Optional)
- `search` - (Optional). One nested block is allowed with the following structure:
  - `query` - (Optional)
- `group_by` - (Optional). Multiple nested blocks are allowed with the following structure:
  - `facet` - (Optional)
  - `limit` - (Optional)
  - `sort` - (Optional). One nested block is allowed with the following structure:
    - `aggregation` - (Optional)
    - `order` - (Optional)
    - `facet` - (Optional)

### Nested `graph` `request` `process_query` blocks

Nested `process_query` blocks have the following structure (Visit the [ Graph Primer](https://docs.datadoghq.com/graphing/) for more information about these values):

- `metric` - (Required)
- `search_by` - (Required)
- `filter_by` - (Required)
- `limit` - (Required)

### Nested `graph` `request` `conditional_format` block

The nested `conditional_format` blocks has the following structure:

- `palette` - (Optional) Color scheme to be used if the condition is met. For example: "red_on_white", "white_on_red", "yellow_on_white", "white_on_yellow", "green_on_white", "white_on_green", "gray_on_white", "white_on_gray", "custom_text", "custom_bg", "custom_image".
- `comparator` - (Required) Comparison operator. Example: ">", "<".
- `value` - (Optional) Value that is the threshold for the conditional format.
- `custom_fg_color` - (Optional) Used when `palette` is set to `custom_text`. Set the color of the text to a custom web color, such as "#205081".
- `custom_bg_color` - (Optional) Used when `palette` is set to `custom_bg`. Set the color of the background to a custom web color, such as "#205081".

### Nested `template_variable` blocks

Nested `template_variable` blocks have the following structure:

- `name` - (Required) The variable name. Can be referenced as \$name in `graph` `request` `q` query strings.
- `prefix` - (Optional) The tag group. Default: no tag group.
- `default` - (Optional) The default tag. Default: "\*" (match all).

## Attributes Reference

The following attributes are exported:

- `id` - The unique ID of this timeboard in your Datadog account. The web interface URL to this timeboard can be generated by appending this ID to `https://app.datadoghq.com/dash/`

## Import

Timeboards can be imported using their numeric ID, e.g.

```
$ terraform import datadog_timeboard.my_service_timeboard 2081
```

## Dynamic Timeboards

Since Terraform 0.12, it's possible to create timeboard graphs dynamically based on contents of a list/map variable. This can be achieved by using the [dynamic blocks](https://www.terraform.io/docs/configuration/expressions.html#dynamic-blocks) feature. For example:

```
variable "my_list" {
  default = ["First", "Second", "Third"]
}

variable "my_map" {
  default = {
    "First" = "value1"
    "Second" = "value2"
  }
}

# Create a timeboard with "First", "Second" and "Third" timeseries graphs
resource "datadog_timeboard" "my_timeboard" {
  title       = "My Timeboard"
  description = "My Description"
  read_only   = true

  dynamic "graph" {
    for_each = var.my_list
    content {
      title = "${graph.value}"
      viz = "timeseries"
      request {
        q = "anomalies(sum:mycount{adapter:${graph.value}}.as_count().rollup(sum, 3600), 'robust', 4, direction='below')"
      }
    }
  }
}

# Create a timeboard with "First" and "Second" timeseries graphs, use map keys as titles and map values as adapter names
resource "datadog_timeboard" "my_timeboard_map" {
  title       = "My Timeboard From Map"
  description = "My Description"
  read_only   = true

  dynamic "graph" {
    for_each = var.my_map
    content {
      title = "${graph.key}"
      viz = "timeseries"
      request {
        q = "anomalies(sum:mycount{adapter:${graph.value}}.as_count().rollup(sum, 3600), 'robust', 4, direction='below')"
      }
    }
  }
}
```
