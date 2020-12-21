---
page_title: "datadog_screenboard Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog screenboard resource. This can be used to create and manage Datadog screenboards.
---

# Resource `datadog_screenboard`

Provides a Datadog screenboard resource. This can be used to create and manage Datadog screenboards.



## Schema

### Required

- **title** (String, Required) Name of the screenboard
- **widget** (Block List, Min: 1) A list of widget definitions. (see [below for nested schema](#nestedblock--widget))

### Optional

- **height** (String, Optional) Height of the screenboard
- **id** (String, Optional) The ID of this resource.
- **read_only** (Boolean, Optional)
- **shared** (Boolean, Optional) Whether the screenboard is shared or not
- **template_variable** (Block List) A list of template variables for using Dashboard templating. (see [below for nested schema](#nestedblock--template_variable))
- **width** (String, Optional) Width of the screenboard

<a id="nestedblock--widget"></a>
### Nested Schema for `widget`

Required:

- **type** (String, Required) The type of the widget. One of [ 'free_text', 'timeseries', 'query_value', 'toplist', 'change', 'event_timeline', 'event_stream', 'image', 'note', 'alert_graph', 'alert_value', 'iframe', 'check_status', 'trace_service', 'hostmap', 'manage_status', 'log_stream', 'uptime', 'process']
- **x** (Number, Required) The position of the widget on the x axis.
- **y** (Number, Required) The position of the widget on the y axis.

Optional:

- **alert_id** (Number, Optional)
- **auto_refresh** (Boolean, Optional)
- **bgcolor** (String, Optional)
- **check** (String, Optional)
- **color** (String, Optional)
- **color_preference** (String, Optional) One of: ['background', 'text']
- **columns** (String, Optional)
- **display_format** (String, Optional) One of: ['counts', 'list', 'countsAndList']
- **env** (String, Optional)
- **event_size** (String, Optional)
- **font_size** (String, Optional)
- **group** (String, Optional)
- **group_by** (List of String, Optional)
- **grouping** (String, Optional) One of: ['cluster', 'check']
- **height** (Number, Optional) The height of the widget.
- **hide_zero_counts** (Boolean, Optional)
- **html** (String, Optional)
- **layout_version** (String, Optional)
- **legend** (Boolean, Optional)
- **legend_size** (String, Optional)
- **logset** (String, Optional)
- **manage_status_show_title** (Boolean, Optional)
- **manage_status_title_align** (String, Optional)
- **manage_status_title_size** (String, Optional)
- **manage_status_title_text** (String, Optional)
- **margin** (String, Optional) One of: ['small', 'large']
- **monitor** (Map of String, Optional)
- **must_show_breakdown** (Boolean, Optional)
- **must_show_distribution** (Boolean, Optional)
- **must_show_errors** (Boolean, Optional)
- **must_show_hits** (Boolean, Optional)
- **must_show_latency** (Boolean, Optional)
- **must_show_resource_list** (Boolean, Optional)
- **params** (Map of String, Optional)
- **precision** (String, Optional)
- **query** (String, Optional)
- **rule** (Block List) (see [below for nested schema](#nestedblock--widget--rule))
- **service_name** (String, Optional)
- **service_service** (String, Optional)
- **show_last_triggered** (Boolean, Optional)
- **size_version** (String, Optional)
- **sizing** (String, Optional) One of: ['center', 'zoom', 'fit']
- **summary_type** (String, Optional) One of: ['monitors', 'groups', 'combined']
- **tags** (List of String, Optional)
- **text** (String, Optional) For widgets of type 'free_text', the text to use.
- **text_align** (String, Optional)
- **text_size** (String, Optional)
- **tick** (Boolean, Optional)
- **tick_edge** (String, Optional)
- **tick_pos** (String, Optional)
- **tile_def** (Block List) (see [below for nested schema](#nestedblock--widget--tile_def))
- **time** (Map of String, Optional)
- **timeframes** (List of String, Optional)
- **title** (String, Optional) The name of the widget.
- **title_align** (String, Optional) The alignment of the widget's title.
- **title_size** (Number, Optional) The size of the widget's title.
- **unit** (String, Optional)
- **url** (String, Optional)
- **viz_type** (String, Optional) One of: ['timeseries', 'toplist']
- **width** (Number, Optional) The width of the widget.

<a id="nestedblock--widget--rule"></a>
### Nested Schema for `widget.rule`

Optional:

- **color** (String, Optional)
- **threshold** (Number, Optional)
- **timeframe** (String, Optional)


<a id="nestedblock--widget--tile_def"></a>
### Nested Schema for `widget.tile_def`

Required:

- **request** (Block List, Min: 1) (see [below for nested schema](#nestedblock--widget--tile_def--request))
- **viz** (String, Required)

Optional:

- **autoscale** (Boolean, Optional)
- **custom_unit** (String, Optional)
- **event** (Block List) (see [below for nested schema](#nestedblock--widget--tile_def--event))
- **group** (List of String, Optional)
- **marker** (Block List) (see [below for nested schema](#nestedblock--widget--tile_def--marker))
- **no_group_hosts** (Boolean, Optional)
- **no_metric_hosts** (Boolean, Optional)
- **node_type** (String, Optional) One of: ['host', 'container']
- **precision** (String, Optional)
- **scope** (List of String, Optional)
- **style** (Map of String, Optional)
- **text_align** (String, Optional)

<a id="nestedblock--widget--tile_def--request"></a>
### Nested Schema for `widget.tile_def.request`

Optional:

- **aggregator** (String, Optional)
- **apm_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--tile_def--request--apm_query))
- **change_type** (String, Optional)
- **compare_to** (String, Optional)
- **conditional_format** (Block List) A list of conditional formatting rules. (see [below for nested schema](#nestedblock--widget--tile_def--request--conditional_format))
- **extra_col** (String, Optional)
- **increase_good** (Boolean, Optional)
- **limit** (Number, Optional)
- **log_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--tile_def--request--log_query))
- **metadata_json** (String, Optional)
- **metric** (String, Optional)
- **order_by** (String, Optional)
- **order_dir** (String, Optional)
- **process_query** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--tile_def--request--process_query))
- **q** (String, Optional)
- **query_type** (String, Optional)
- **style** (Map of String, Optional)
- **tag_filters** (List of String, Optional)
- **text_filter** (String, Optional)
- **type** (String, Optional)

<a id="nestedblock--widget--tile_def--request--apm_query"></a>
### Nested Schema for `widget.tile_def.request.type`

Required:

- **compute** (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--widget--tile_def--request--type--compute))
- **index** (String, Required)

Optional:

- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--tile_def--request--type--group_by))
- **search** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--tile_def--request--type--search))

<a id="nestedblock--widget--tile_def--request--type--compute"></a>
### Nested Schema for `widget.tile_def.request.type.compute`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (String, Optional)


<a id="nestedblock--widget--tile_def--request--type--group_by"></a>
### Nested Schema for `widget.tile_def.request.type.group_by`

Required:

- **facet** (String, Required)

Optional:

- **limit** (Number, Optional)
- **sort** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--tile_def--request--type--group_by--sort))

<a id="nestedblock--widget--tile_def--request--type--group_by--sort"></a>
### Nested Schema for `widget.tile_def.request.type.group_by.sort`

Required:

- **aggregation** (String, Required)
- **order** (String, Required)

Optional:

- **facet** (String, Optional)



<a id="nestedblock--widget--tile_def--request--type--search"></a>
### Nested Schema for `widget.tile_def.request.type.search`

Required:

- **query** (String, Required)



<a id="nestedblock--widget--tile_def--request--conditional_format"></a>
### Nested Schema for `widget.tile_def.request.type`

Required:

- **comparator** (String, Required) Comparator (<, >, etc)

Optional:

- **color** (String, Optional) Custom color (e.g., #205081)
- **custom_bg_color** (String, Optional) Custom  background color (e.g., #205081)
- **invert** (Boolean, Optional)
- **palette** (String, Optional) The palette to use if this condition is met.
- **value** (String, Optional) Value that is threshold for conditional format


<a id="nestedblock--widget--tile_def--request--log_query"></a>
### Nested Schema for `widget.tile_def.request.type`

Required:

- **compute** (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--widget--tile_def--request--type--compute))
- **index** (String, Required)

Optional:

- **group_by** (Block List) (see [below for nested schema](#nestedblock--widget--tile_def--request--type--group_by))
- **search** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--tile_def--request--type--search))

<a id="nestedblock--widget--tile_def--request--type--compute"></a>
### Nested Schema for `widget.tile_def.request.type.compute`

Required:

- **aggregation** (String, Required)

Optional:

- **facet** (String, Optional)
- **interval** (String, Optional)


<a id="nestedblock--widget--tile_def--request--type--group_by"></a>
### Nested Schema for `widget.tile_def.request.type.group_by`

Required:

- **facet** (String, Required)

Optional:

- **limit** (Number, Optional)
- **sort** (Block List, Max: 1) (see [below for nested schema](#nestedblock--widget--tile_def--request--type--group_by--sort))

<a id="nestedblock--widget--tile_def--request--type--group_by--sort"></a>
### Nested Schema for `widget.tile_def.request.type.group_by.sort`

Required:

- **aggregation** (String, Required)
- **order** (String, Required)

Optional:

- **facet** (String, Optional)



<a id="nestedblock--widget--tile_def--request--type--search"></a>
### Nested Schema for `widget.tile_def.request.type.search`

Required:

- **query** (String, Required)



<a id="nestedblock--widget--tile_def--request--process_query"></a>
### Nested Schema for `widget.tile_def.request.type`

Required:

- **metric** (String, Required)

Optional:

- **filter_by** (List of String, Optional)
- **limit** (Number, Optional)
- **search_by** (String, Optional)



<a id="nestedblock--widget--tile_def--event"></a>
### Nested Schema for `widget.tile_def.event`

Required:

- **q** (String, Required)


<a id="nestedblock--widget--tile_def--marker"></a>
### Nested Schema for `widget.tile_def.marker`

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


