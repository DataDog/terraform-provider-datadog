---
page_title: "datadog_logs_custom_pipeline Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  
---

# Resource `datadog_logs_custom_pipeline`





## Schema

### Required

- **filter** (Block List, Min: 1) (see [below for nested schema](#nestedblock--filter))
- **name** (String, Required)

### Optional

- **id** (String, Optional) The ID of this resource.
- **is_enabled** (Boolean, Optional)
- **processor** (Block List) (see [below for nested schema](#nestedblock--processor))

<a id="nestedblock--filter"></a>
### Nested Schema for `filter`

Required:

- **query** (String, Required)


<a id="nestedblock--processor"></a>
### Nested Schema for `processor`

Optional:

- **arithmetic_processor** (Block List, Max: 1) (see [below for nested schema](#nestedblock--processor--arithmetic_processor))
- **attribute_remapper** (Block List, Max: 1) (see [below for nested schema](#nestedblock--processor--attribute_remapper))
- **category_processor** (Block List, Max: 1) (see [below for nested schema](#nestedblock--processor--category_processor))
- **date_remapper** (Block List, Max: 1) (see [below for nested schema](#nestedblock--processor--date_remapper))
- **geo_ip_parser** (Block List, Max: 1) (see [below for nested schema](#nestedblock--processor--geo_ip_parser))
- **grok_parser** (Block List, Max: 1) (see [below for nested schema](#nestedblock--processor--grok_parser))
- **lookup_processor** (Block List, Max: 1) (see [below for nested schema](#nestedblock--processor--lookup_processor))
- **message_remapper** (Block List, Max: 1) (see [below for nested schema](#nestedblock--processor--message_remapper))
- **pipeline** (Block List, Max: 1) (see [below for nested schema](#nestedblock--processor--pipeline))
- **service_remapper** (Block List, Max: 1) (see [below for nested schema](#nestedblock--processor--service_remapper))
- **status_remapper** (Block List, Max: 1) (see [below for nested schema](#nestedblock--processor--status_remapper))
- **string_builder_processor** (Block List, Max: 1) (see [below for nested schema](#nestedblock--processor--string_builder_processor))
- **trace_id_remapper** (Block List, Max: 1) (see [below for nested schema](#nestedblock--processor--trace_id_remapper))
- **url_parser** (Block List, Max: 1) (see [below for nested schema](#nestedblock--processor--url_parser))
- **user_agent_parser** (Block List, Max: 1) (see [below for nested schema](#nestedblock--processor--user_agent_parser))

<a id="nestedblock--processor--arithmetic_processor"></a>
### Nested Schema for `processor.arithmetic_processor`

Required:

- **expression** (String, Required)
- **target** (String, Required)

Optional:

- **is_enabled** (Boolean, Optional)
- **is_replace_missing** (Boolean, Optional)
- **name** (String, Optional)


<a id="nestedblock--processor--attribute_remapper"></a>
### Nested Schema for `processor.attribute_remapper`

Required:

- **source_type** (String, Required)
- **sources** (List of String, Required)
- **target** (String, Required)
- **target_type** (String, Required)

Optional:

- **is_enabled** (Boolean, Optional)
- **name** (String, Optional)
- **override_on_conflict** (Boolean, Optional)
- **preserve_source** (Boolean, Optional)
- **target_format** (String, Optional)


<a id="nestedblock--processor--category_processor"></a>
### Nested Schema for `processor.category_processor`

Required:

- **category** (Block List, Min: 1) (see [below for nested schema](#nestedblock--processor--category_processor--category))
- **target** (String, Required)

Optional:

- **is_enabled** (Boolean, Optional)
- **name** (String, Optional)

<a id="nestedblock--processor--category_processor--category"></a>
### Nested Schema for `processor.category_processor.category`

Required:

- **filter** (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--processor--category_processor--category--filter))
- **name** (String, Required)

<a id="nestedblock--processor--category_processor--category--filter"></a>
### Nested Schema for `processor.category_processor.category.name`

Required:

- **query** (String, Required)




<a id="nestedblock--processor--date_remapper"></a>
### Nested Schema for `processor.date_remapper`

Required:

- **sources** (List of String, Required)

Optional:

- **is_enabled** (Boolean, Optional)
- **name** (String, Optional)


<a id="nestedblock--processor--geo_ip_parser"></a>
### Nested Schema for `processor.geo_ip_parser`

Required:

- **sources** (List of String, Required)
- **target** (String, Required)

Optional:

- **is_enabled** (Boolean, Optional)
- **name** (String, Optional)


<a id="nestedblock--processor--grok_parser"></a>
### Nested Schema for `processor.grok_parser`

Required:

- **grok** (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--processor--grok_parser--grok))
- **source** (String, Required)

Optional:

- **is_enabled** (Boolean, Optional)
- **name** (String, Optional)
- **samples** (List of String, Optional)

<a id="nestedblock--processor--grok_parser--grok"></a>
### Nested Schema for `processor.grok_parser.grok`

Required:

- **match_rules** (String, Required)
- **support_rules** (String, Required)



<a id="nestedblock--processor--lookup_processor"></a>
### Nested Schema for `processor.lookup_processor`

Required:

- **lookup_table** (List of String, Required)
- **source** (String, Required)
- **target** (String, Required)

Optional:

- **default_lookup** (String, Optional)
- **is_enabled** (Boolean, Optional)
- **name** (String, Optional)


<a id="nestedblock--processor--message_remapper"></a>
### Nested Schema for `processor.message_remapper`

Required:

- **sources** (List of String, Required)

Optional:

- **is_enabled** (Boolean, Optional)
- **name** (String, Optional)


<a id="nestedblock--processor--pipeline"></a>
### Nested Schema for `processor.pipeline`

Required:

- **filter** (Block List, Min: 1) (see [below for nested schema](#nestedblock--processor--pipeline--filter))
- **name** (String, Required)

Optional:

- **is_enabled** (Boolean, Optional)
- **processor** (Block List) (see [below for nested schema](#nestedblock--processor--pipeline--processor))

<a id="nestedblock--processor--pipeline--filter"></a>
### Nested Schema for `processor.pipeline.filter`

Required:

- **query** (String, Required)


<a id="nestedblock--processor--pipeline--processor"></a>
### Nested Schema for `processor.pipeline.processor`

Optional:

- **arithmetic_processor** (Block List, Max: 1) (see [below for nested schema](#nestedblock--processor--pipeline--processor--arithmetic_processor))
- **attribute_remapper** (Block List, Max: 1) (see [below for nested schema](#nestedblock--processor--pipeline--processor--attribute_remapper))
- **category_processor** (Block List, Max: 1) (see [below for nested schema](#nestedblock--processor--pipeline--processor--category_processor))
- **date_remapper** (Block List, Max: 1) (see [below for nested schema](#nestedblock--processor--pipeline--processor--date_remapper))
- **geo_ip_parser** (Block List, Max: 1) (see [below for nested schema](#nestedblock--processor--pipeline--processor--geo_ip_parser))
- **grok_parser** (Block List, Max: 1) (see [below for nested schema](#nestedblock--processor--pipeline--processor--grok_parser))
- **lookup_processor** (Block List, Max: 1) (see [below for nested schema](#nestedblock--processor--pipeline--processor--lookup_processor))
- **message_remapper** (Block List, Max: 1) (see [below for nested schema](#nestedblock--processor--pipeline--processor--message_remapper))
- **service_remapper** (Block List, Max: 1) (see [below for nested schema](#nestedblock--processor--pipeline--processor--service_remapper))
- **status_remapper** (Block List, Max: 1) (see [below for nested schema](#nestedblock--processor--pipeline--processor--status_remapper))
- **string_builder_processor** (Block List, Max: 1) (see [below for nested schema](#nestedblock--processor--pipeline--processor--string_builder_processor))
- **trace_id_remapper** (Block List, Max: 1) (see [below for nested schema](#nestedblock--processor--pipeline--processor--trace_id_remapper))
- **url_parser** (Block List, Max: 1) (see [below for nested schema](#nestedblock--processor--pipeline--processor--url_parser))
- **user_agent_parser** (Block List, Max: 1) (see [below for nested schema](#nestedblock--processor--pipeline--processor--user_agent_parser))

<a id="nestedblock--processor--pipeline--processor--arithmetic_processor"></a>
### Nested Schema for `processor.pipeline.processor.user_agent_parser`

Required:

- **expression** (String, Required)
- **target** (String, Required)

Optional:

- **is_enabled** (Boolean, Optional)
- **is_replace_missing** (Boolean, Optional)
- **name** (String, Optional)


<a id="nestedblock--processor--pipeline--processor--attribute_remapper"></a>
### Nested Schema for `processor.pipeline.processor.user_agent_parser`

Required:

- **source_type** (String, Required)
- **sources** (List of String, Required)
- **target** (String, Required)
- **target_type** (String, Required)

Optional:

- **is_enabled** (Boolean, Optional)
- **name** (String, Optional)
- **override_on_conflict** (Boolean, Optional)
- **preserve_source** (Boolean, Optional)
- **target_format** (String, Optional)


<a id="nestedblock--processor--pipeline--processor--category_processor"></a>
### Nested Schema for `processor.pipeline.processor.user_agent_parser`

Required:

- **category** (Block List, Min: 1) (see [below for nested schema](#nestedblock--processor--pipeline--processor--user_agent_parser--category))
- **target** (String, Required)

Optional:

- **is_enabled** (Boolean, Optional)
- **name** (String, Optional)

<a id="nestedblock--processor--pipeline--processor--user_agent_parser--category"></a>
### Nested Schema for `processor.pipeline.processor.user_agent_parser.category`

Required:

- **filter** (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--processor--pipeline--processor--user_agent_parser--category--filter))
- **name** (String, Required)

<a id="nestedblock--processor--pipeline--processor--user_agent_parser--category--filter"></a>
### Nested Schema for `processor.pipeline.processor.user_agent_parser.category.name`

Required:

- **query** (String, Required)




<a id="nestedblock--processor--pipeline--processor--date_remapper"></a>
### Nested Schema for `processor.pipeline.processor.user_agent_parser`

Required:

- **sources** (List of String, Required)

Optional:

- **is_enabled** (Boolean, Optional)
- **name** (String, Optional)


<a id="nestedblock--processor--pipeline--processor--geo_ip_parser"></a>
### Nested Schema for `processor.pipeline.processor.user_agent_parser`

Required:

- **sources** (List of String, Required)
- **target** (String, Required)

Optional:

- **is_enabled** (Boolean, Optional)
- **name** (String, Optional)


<a id="nestedblock--processor--pipeline--processor--grok_parser"></a>
### Nested Schema for `processor.pipeline.processor.user_agent_parser`

Required:

- **grok** (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--processor--pipeline--processor--user_agent_parser--grok))
- **source** (String, Required)

Optional:

- **is_enabled** (Boolean, Optional)
- **name** (String, Optional)
- **samples** (List of String, Optional)

<a id="nestedblock--processor--pipeline--processor--user_agent_parser--grok"></a>
### Nested Schema for `processor.pipeline.processor.user_agent_parser.grok`

Required:

- **match_rules** (String, Required)
- **support_rules** (String, Required)



<a id="nestedblock--processor--pipeline--processor--lookup_processor"></a>
### Nested Schema for `processor.pipeline.processor.user_agent_parser`

Required:

- **lookup_table** (List of String, Required)
- **source** (String, Required)
- **target** (String, Required)

Optional:

- **default_lookup** (String, Optional)
- **is_enabled** (Boolean, Optional)
- **name** (String, Optional)


<a id="nestedblock--processor--pipeline--processor--message_remapper"></a>
### Nested Schema for `processor.pipeline.processor.user_agent_parser`

Required:

- **sources** (List of String, Required)

Optional:

- **is_enabled** (Boolean, Optional)
- **name** (String, Optional)


<a id="nestedblock--processor--pipeline--processor--service_remapper"></a>
### Nested Schema for `processor.pipeline.processor.user_agent_parser`

Required:

- **sources** (List of String, Required)

Optional:

- **is_enabled** (Boolean, Optional)
- **name** (String, Optional)


<a id="nestedblock--processor--pipeline--processor--status_remapper"></a>
### Nested Schema for `processor.pipeline.processor.user_agent_parser`

Required:

- **sources** (List of String, Required)

Optional:

- **is_enabled** (Boolean, Optional)
- **name** (String, Optional)


<a id="nestedblock--processor--pipeline--processor--string_builder_processor"></a>
### Nested Schema for `processor.pipeline.processor.user_agent_parser`

Required:

- **target** (String, Required)
- **template** (String, Required)

Optional:

- **is_enabled** (Boolean, Optional)
- **is_replace_missing** (Boolean, Optional)
- **name** (String, Optional)


<a id="nestedblock--processor--pipeline--processor--trace_id_remapper"></a>
### Nested Schema for `processor.pipeline.processor.user_agent_parser`

Required:

- **sources** (List of String, Required)

Optional:

- **is_enabled** (Boolean, Optional)
- **name** (String, Optional)


<a id="nestedblock--processor--pipeline--processor--url_parser"></a>
### Nested Schema for `processor.pipeline.processor.user_agent_parser`

Required:

- **sources** (List of String, Required)
- **target** (String, Required)

Optional:

- **is_enabled** (Boolean, Optional)
- **name** (String, Optional)
- **normalize_ending_slashes** (Boolean, Optional)


<a id="nestedblock--processor--pipeline--processor--user_agent_parser"></a>
### Nested Schema for `processor.pipeline.processor.user_agent_parser`

Required:

- **sources** (List of String, Required)
- **target** (String, Required)

Optional:

- **is_enabled** (Boolean, Optional)
- **is_encoded** (Boolean, Optional)
- **name** (String, Optional)




<a id="nestedblock--processor--service_remapper"></a>
### Nested Schema for `processor.service_remapper`

Required:

- **sources** (List of String, Required)

Optional:

- **is_enabled** (Boolean, Optional)
- **name** (String, Optional)


<a id="nestedblock--processor--status_remapper"></a>
### Nested Schema for `processor.status_remapper`

Required:

- **sources** (List of String, Required)

Optional:

- **is_enabled** (Boolean, Optional)
- **name** (String, Optional)


<a id="nestedblock--processor--string_builder_processor"></a>
### Nested Schema for `processor.string_builder_processor`

Required:

- **target** (String, Required)
- **template** (String, Required)

Optional:

- **is_enabled** (Boolean, Optional)
- **is_replace_missing** (Boolean, Optional)
- **name** (String, Optional)


<a id="nestedblock--processor--trace_id_remapper"></a>
### Nested Schema for `processor.trace_id_remapper`

Required:

- **sources** (List of String, Required)

Optional:

- **is_enabled** (Boolean, Optional)
- **name** (String, Optional)


<a id="nestedblock--processor--url_parser"></a>
### Nested Schema for `processor.url_parser`

Required:

- **sources** (List of String, Required)
- **target** (String, Required)

Optional:

- **is_enabled** (Boolean, Optional)
- **name** (String, Optional)
- **normalize_ending_slashes** (Boolean, Optional)


<a id="nestedblock--processor--user_agent_parser"></a>
### Nested Schema for `processor.user_agent_parser`

Required:

- **sources** (List of String, Required)
- **target** (String, Required)

Optional:

- **is_enabled** (Boolean, Optional)
- **is_encoded** (Boolean, Optional)
- **name** (String, Optional)


