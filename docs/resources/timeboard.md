---
page_title: "datadog_timeboard Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog timeboard resource. This can be used to create and manage Datadog timeboards.
---

# Resource `datadog_timeboard`

Provides a Datadog timeboard resource. This can be used to create and manage Datadog timeboards.

## Example Usage

```terraform
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

## Schema

### Required

- **description** (String, Required) A description of the dashboard's content.
- **graph** (Block List, Min: 1) A list of graph definitions. (see [below for nested schema](#nestedblock--graph))
- **title** (String, Required) The name of the dashboard.

### Optional

- **id** (String, Optional) The ID of this resource.
- **read_only** (Boolean, Optional)
- **template_variable** (Block List) A list of template variables for using Dashboard templating. (see [below for nested schema](#nestedblock--template_variable))

<a id="nestedblock--graph"></a>
### Nested Schema for `graph`

Required:

- **request** (Block List, Min: 1) (see [below for nested schema](#nestedblock--graph--request))
- **title** (String, Required) The name of the graph.
- **viz** (String, Required)

Optional:

- **autoscale** (Boolean, Optional) Automatically scale graphs
- **custom_unit** (String, Optional) Use a custom unit (like 'users')
- **events** (List of String, Optional) Filter for events to be overlayed on the graph.
- **group** (List of String, Optional) A list of groupings for hostmap type graphs.
- **include_no_metric_hosts** (Boolean, Optional) Include hosts without metrics in hostmap graphs
- **include_ungrouped_hosts** (Boolean, Optional) Include ungrouped hosts in hostmap graphs
- **marker** (Block List) (see [below for nested schema](#nestedblock--graph--marker))
- **node_type** (String, Optional) Type of nodes to show in hostmap graphs (either 'host' or 'container').
- **precision** (String, Optional) How many digits to show
- **scope** (List of String, Optional) A list of scope filters for hostmap type graphs.
- **style** (Map of String, Optional)
- **text_align** (String, Optional) How to align text
- **yaxis** (Map of String, Optional)

<a id="nestedblock--graph--request"></a>
### Nested Schema for `graph.request`

Optional:

- **aggregator** (String, Optional)
- **apm_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--graph--request--apm_query))
- **change_type** (String, Optional) Type of change for change graphs.
- **compare_to** (String, Optional) The time period to compare change against in change graphs.
- **conditional_format** (Block List) A list of conditional formatting rules. (see [below for nested schema](#nestedblock--graph--request--conditional_format))
- **extra_col** (String, Optional) If set to 'present', this will include the present values in change graphs.
- **increase_good** (Boolean, Optional) Decides whether to represent increases as good or bad in change graphs.
- **log_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--graph--request--log_query))
- **metadata_json** (String, Optional)
- **order_by** (String, Optional) The field a change graph will be ordered by.
- **order_direction** (String, Optional) Sort change graph in ascending or descending order.
- **process_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--graph--request--process_query))
- **q** (String, Optional)
- **stacked** (Boolean, Optional)
- **style** (Map of String, Optional)
- **type** (String, Optional)

<a id="nestedblock--graph--request--apm_query"></a>
### Nested Schema for `graph.request.apm_query`

Required:

- **compute** (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--graph--request--apm_query--compute))
- **index** (String, Required)

Optional:

- **group_by** (Block List) (see [below for nested schema](#nestedblock--graph--request--apm_query--group_by))
- **search** (Block List, Max: 1) (see [below for nested schema](#nestedblock--graph--request--apm_query--search))

<a id="nestedblock--graph--request--apm_query--compute"></a>
### Nested Schema for `graph.request.apm_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)


<a id="nestedblock--graph--request--apm_query--group_by"></a>
### Nested Schema for `graph.request.apm_query.search`

Required:

- **facet** (String, Required)

Optional:

- **limit** (Number, Optional)
- **sort** (Block List, Max: 1) (see [below for nested schema](#nestedblock--graph--request--apm_query--search--sort))

<a id="nestedblock--graph--request--apm_query--search--sort"></a>
### Nested Schema for `graph.request.apm_query.search.sort`

Required:

- **aggregation** (String, Required)
- **order** (String, Required)

Optional:

- **facet** (String, Optional)



<a id="nestedblock--graph--request--apm_query--search"></a>
### Nested Schema for `graph.request.apm_query.search`

Required:

- **query** (String, Required)



<a id="nestedblock--graph--request--conditional_format"></a>
### Nested Schema for `graph.request.conditional_format`

Required:

- **comparator** (String, Required) Comparator (<, >, etc)

Optional:

- **custom_bg_color** (String, Optional) Custom background color (e.g., #205081)
- **custom_fg_color** (String, Optional) Custom foreground color (e.g., #59afe1)
- **palette** (String, Optional) The palette to use if this condition is met.
- **value** (String, Optional) Value that is threshold for conditional format


<a id="nestedblock--graph--request--log_query"></a>
### Nested Schema for `graph.request.log_query`

Required:

- **compute** (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--graph--request--log_query--compute))
- **index** (String, Required)

Optional:

- **group_by** (Block List) (see [below for nested schema](#nestedblock--graph--request--log_query--group_by))
- **search** (Block List, Max: 1) (see [below for nested schema](#nestedblock--graph--request--log_query--search))

<a id="nestedblock--graph--request--log_query--compute"></a>
### Nested Schema for `graph.request.log_query.search`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (Number, Optional)


<a id="nestedblock--graph--request--log_query--group_by"></a>
### Nested Schema for `graph.request.log_query.search`

Required:

- **facet** (String, Required)

Optional:

- **limit** (Number, Optional)
- **sort** (Block List, Max: 1) (see [below for nested schema](#nestedblock--graph--request--log_query--search--sort))

<a id="nestedblock--graph--request--log_query--search--sort"></a>
### Nested Schema for `graph.request.log_query.search.sort`

Required:

- **aggregation** (String, Required)
- **order** (String, Required)

Optional:

- **facet** (String, Optional)



<a id="nestedblock--graph--request--log_query--search"></a>
### Nested Schema for `graph.request.log_query.search`

Required:

- **query** (String, Required)



<a id="nestedblock--graph--request--process_query"></a>
### Nested Schema for `graph.request.process_query`

Required:

- **metric** (String, Required)

Optional:

- **filter_by** (List of String, Optional)
- **limit** (Number, Optional)
- **search_by** (String, Optional)



<a id="nestedblock--graph--marker"></a>
### Nested Schema for `graph.marker`

Required:

- **type** (String, Required)
- **value** (String, Required)

Optional:

- **label** (String, Optional)



<a id="nestedblock--template_variable"></a>
### Nested Schema for `template_variable`

Required:

- **name** (String, Required) The name of the variable.

Optional:

- **default** (String, Optional) The default value for the template variable on dashboard load.
- **prefix** (String, Optional) The tag prefix associated with the variable. Only tags with this prefix will appear in the variable dropdown.

## Import

Import is supported using the following syntax:

```shell
# Timeboards can be imported using their numeric ID, e.g.
terraform import datadog_timeboard.my_service_timeboard 2081
```
