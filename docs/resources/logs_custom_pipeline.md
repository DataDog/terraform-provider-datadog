---
page_title: "datadog_logs_custom_pipeline Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog Logs Pipeline API https://docs.datadoghq.com/api/v1/logs-pipelines/ resource, which is used to create and manage Datadog logs custom pipelines. Each datadog_logs_custom_pipeline resource defines a complete pipeline. The order of the pipelines is maintained in a different resource: datadog_logs_pipeline_order. When creating a new pipeline, you need to explicitly add this pipeline to the datadog_logs_pipeline_order resource to track the pipeline. Similarly, when a pipeline needs to be destroyed, remove its references from the datadog_logs_pipeline_order resource.
---

# Resource `datadog_logs_custom_pipeline`

Provides a Datadog [Logs Pipeline API](https://docs.datadoghq.com/api/v1/logs-pipelines/) resource, which is used to create and manage Datadog logs custom pipelines. Each `datadog_logs_custom_pipeline` resource defines a complete pipeline. The order of the pipelines is maintained in a different resource: `datadog_logs_pipeline_order`. When creating a new pipeline, you need to **explicitly** add this pipeline to the `datadog_logs_pipeline_order` resource to track the pipeline. Similarly, when a pipeline needs to be destroyed, remove its references from the `datadog_logs_pipeline_order` resource.

## Example Usage

```terraform
resource "datadog_logs_custom_pipeline" "sample_pipeline" {
  filter {
    query = "source:foo"
  }
  name       = "sample pipeline"
  is_enabled = true
  processor {
    arithmetic_processor {
      expression         = "(time1 - time2)*1000"
      target             = "my_arithmetic"
      is_replace_missing = true
      name               = "sample arithmetic processor"
      is_enabled         = true
    }
  }
  processor {
    attribute_remapper {
      sources              = ["db.instance"]
      source_type          = "tag"
      target               = "db"
      target_type          = "attribute"
      target_format        = "string"
      preserve_source      = true
      override_on_conflict = false
      name                 = "sample attribute processor"
      is_enabled           = true
    }
  }
  processor {
    category_processor {
      target = "foo.severity"
      category {
        name = "debug"
        filter {
          query = "@severity: \".\""
        }
      }
      category {
        name = "verbose"
        filter {
          query = "@severity: \"-\""
        }
      }
      name       = "sample category processor"
      is_enabled = true
    }
  }
  processor {
    date_remapper {
      sources    = ["_timestamp", "published_date"]
      name       = "sample date remapper"
      is_enabled = true
    }
  }
  processor {
    geo_ip_parser {
      sources    = ["network.client.ip"]
      target     = "network.client.geoip"
      name       = "sample geo ip parser"
      is_enabled = true
    }
  }
  processor {
    grok_parser {
      samples = ["sample log 1"]
      source  = "message"
      grok {
        support_rules = ""
        match_rules   = "Rule %%{word:my_word2} %%{number:my_float2}"
      }
      name       = "sample grok parser"
      is_enabled = true
    }
  }
  processor {
    lookup_processor {
      source         = "service_id"
      target         = "service_name"
      lookup_table   = ["1,my service"]
      default_lookup = "unknown service"
      name           = "sample lookup processor"
      is_enabled     = true
    }
  }
  processor {
    message_remapper {
      sources    = ["msg"]
      name       = "sample message remapper"
      is_enabled = true
    }
  }
  processor {
    pipeline {
      filter {
        query = "source:foo"
      }
      processor {
        url_parser {
          name                     = "sample url parser"
          sources                  = ["url", "extra"]
          target                   = "http_url"
          normalize_ending_slashes = true
        }
      }
      name       = "nested pipeline"
      is_enabled = true
    }
  }
  processor {
    service_remapper {
      sources    = ["service"]
      name       = "sample service remapper"
      is_enabled = true
    }
  }
  processor {
    status_remapper {
      sources    = ["info", "trace"]
      name       = "sample status remapper"
      is_enabled = true
    }
  }
  processor {
    string_builder_processor {
      target             = "user_activity"
      template           = "%%{user.name} logged in at %%{timestamp}"
      name               = "sample string builder processor"
      is_enabled         = true
      is_replace_missing = false
    }
  }
  processor {
    trace_id_remapper {
      sources    = ["dd.trace_id"]
      name       = "sample trace id remapper"
      is_enabled = true
    }
  }
  processor {
    user_agent_parser {
      sources    = ["user", "agent"]
      target     = "http_agent"
      is_encoded = false
      name       = "sample user agent parser"
      is_enabled = true
    }
  }
}
```

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

- **query** (String, Required) Filter criteria of the category.


<a id="nestedblock--processor"></a>
### Nested Schema for `processor`

Optional:

- **arithmetic_processor** (Block List, Max: 1) Arithmetic Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#arithmetic-processor) (see [below for nested schema](#nestedblock--processor--arithmetic_processor))
- **attribute_remapper** (Block List, Max: 1) Attribute Remapper Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#remapper) (see [below for nested schema](#nestedblock--processor--attribute_remapper))
- **category_processor** (Block List, Max: 1) Category Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#category-processor) (see [below for nested schema](#nestedblock--processor--category_processor))
- **date_remapper** (Block List, Max: 1) Date Remapper Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#log-date-remapper) (see [below for nested schema](#nestedblock--processor--date_remapper))
- **geo_ip_parser** (Block List, Max: 1) Date GeoIP Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#geoip-parser) (see [below for nested schema](#nestedblock--processor--geo_ip_parser))
- **grok_parser** (Block List, Max: 1) Grok Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#grok-parser) (see [below for nested schema](#nestedblock--processor--grok_parser))
- **lookup_processor** (Block List, Max: 1) Lookup Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#lookup-processor) (see [below for nested schema](#nestedblock--processor--lookup_processor))
- **message_remapper** (Block List, Max: 1) Message Remapper Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#log-message-remapper) (see [below for nested schema](#nestedblock--processor--message_remapper))
- **pipeline** (Block List, Max: 1) (see [below for nested schema](#nestedblock--processor--pipeline))
- **service_remapper** (Block List, Max: 1) Service Remapper Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#service-remapper) (see [below for nested schema](#nestedblock--processor--service_remapper))
- **status_remapper** (Block List, Max: 1) Status Remapper Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#log-status-remapper) (see [below for nested schema](#nestedblock--processor--status_remapper))
- **string_builder_processor** (Block List, Max: 1) String Builder Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#string-builder-processor) (see [below for nested schema](#nestedblock--processor--string_builder_processor))
- **trace_id_remapper** (Block List, Max: 1) Trace ID Remapper Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#trace-remapper) (see [below for nested schema](#nestedblock--processor--trace_id_remapper))
- **url_parser** (Block List, Max: 1) URL Parser Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#url-parser) (see [below for nested schema](#nestedblock--processor--url_parser))
- **user_agent_parser** (Block List, Max: 1) User-Agent Parser Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#user-agent-parser) (see [below for nested schema](#nestedblock--processor--user_agent_parser))

<a id="nestedblock--processor--arithmetic_processor"></a>
### Nested Schema for `processor.arithmetic_processor`

Required:

- **expression** (String, Required) Arithmetic operation between one or more log attributes.
- **target** (String, Required) Name of the attribute that contains the result of the arithmetic operation.

Optional:

- **is_enabled** (Boolean, Optional) Boolean value to enable your pipeline.
- **is_replace_missing** (Boolean, Optional) If true, it replaces all missing attributes of expression by 0, false skips the operation if an attribute is missing.
- **name** (String, Optional) Your pipeline name.


<a id="nestedblock--processor--attribute_remapper"></a>
### Nested Schema for `processor.attribute_remapper`

Required:

- **source_type** (String, Required) Defines where the sources are from (log attribute or tag).
- **sources** (List of String, Required) List of source attributes or tags.
- **target** (String, Required) Final attribute or tag name to remap the sources.
- **target_type** (String, Required) Defines if the target is a log attribute or tag.

Optional:

- **is_enabled** (Boolean, Optional) If the processor is enabled or not.
- **name** (String, Optional) Name of the processor
- **override_on_conflict** (Boolean, Optional) Override the target element if already set.
- **preserve_source** (Boolean, Optional) Remove or preserve the remapped source element.
- **target_format** (String, Optional) If the `target_type` of the remapper is `attribute`, try to cast the value to a new specific type. If the cast is not possible, the original type is kept. `string`, `integer`, or `double` are the possible types. If the `target_type` is `tag`, this parameter may not be specified.


<a id="nestedblock--processor--category_processor"></a>
### Nested Schema for `processor.category_processor`

Required:

- **category** (Block List, Min: 1) List of filters to match or exclude a log with their corresponding name to assign a custom value to the log. (see [below for nested schema](#nestedblock--processor--category_processor--category))
- **target** (String, Required) Name of the target attribute whose value is defined by the matching category.

Optional:

- **is_enabled** (Boolean, Optional) If the processor is enabled or not.
- **name** (String, Optional) Name of the category

<a id="nestedblock--processor--category_processor--category"></a>
### Nested Schema for `processor.category_processor.category`

Required:

- **filter** (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--processor--category_processor--category--filter))
- **name** (String, Required)

<a id="nestedblock--processor--category_processor--category--filter"></a>
### Nested Schema for `processor.category_processor.category.name`

Required:

- **query** (String, Required) Filter criteria of the category.




<a id="nestedblock--processor--date_remapper"></a>
### Nested Schema for `processor.date_remapper`

Required:

- **sources** (List of String, Required) List of source attributes.

Optional:

- **is_enabled** (Boolean, Optional) If the processor is enabled or not.
- **name** (String, Optional) Name of the processor.


<a id="nestedblock--processor--geo_ip_parser"></a>
### Nested Schema for `processor.geo_ip_parser`

Required:

- **sources** (List of String, Required) List of source attributes.
- **target** (String, Required) Name of the parent attribute that contains all the extracted details from the sources.

Optional:

- **is_enabled** (Boolean, Optional) If the processor is enabled or not.
- **name** (String, Optional) Name of the processor.


<a id="nestedblock--processor--grok_parser"></a>
### Nested Schema for `processor.grok_parser`

Required:

- **grok** (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--processor--grok_parser--grok))
- **source** (String, Required) Name of the log attribute to parse.

Optional:

- **is_enabled** (Boolean, Optional) If the processor is enabled or not.
- **name** (String, Optional) Name of the processor
- **samples** (List of String, Optional) List of sample logs for this parser. It can save up to 5 samples. Each sample takes up to 5000 characters.

<a id="nestedblock--processor--grok_parser--grok"></a>
### Nested Schema for `processor.grok_parser.grok`

Required:

- **match_rules** (String, Required) Match rules for your grok parser.
- **support_rules** (String, Required) Support rules for your grok parser.



<a id="nestedblock--processor--lookup_processor"></a>
### Nested Schema for `processor.lookup_processor`

Required:

- **lookup_table** (List of String, Required) List of entries of the lookup table using `key,value` format.
- **source** (String, Required) Name of the source attribute used to do the lookup.
- **target** (String, Required) Name of the attribute that contains the result of the lookup.

Optional:

- **default_lookup** (String, Optional) Default lookup value to use if there is no entry in the lookup table for the value of the source attribute.
- **is_enabled** (Boolean, Optional) If the processor is enabled or not.
- **name** (String, Optional) Name of the processor


<a id="nestedblock--processor--message_remapper"></a>
### Nested Schema for `processor.message_remapper`

Required:

- **sources** (List of String, Required) List of source attributes.

Optional:

- **is_enabled** (Boolean, Optional) If the processor is enabled or not.
- **name** (String, Optional) Name of the processor.


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

- **query** (String, Required) Filter criteria of the category.


<a id="nestedblock--processor--pipeline--processor"></a>
### Nested Schema for `processor.pipeline.processor`

Optional:

- **arithmetic_processor** (Block List, Max: 1) Arithmetic Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#arithmetic-processor) (see [below for nested schema](#nestedblock--processor--pipeline--processor--arithmetic_processor))
- **attribute_remapper** (Block List, Max: 1) Attribute Remapper Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#remapper) (see [below for nested schema](#nestedblock--processor--pipeline--processor--attribute_remapper))
- **category_processor** (Block List, Max: 1) Category Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#category-processor) (see [below for nested schema](#nestedblock--processor--pipeline--processor--category_processor))
- **date_remapper** (Block List, Max: 1) Date Remapper Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#log-date-remapper) (see [below for nested schema](#nestedblock--processor--pipeline--processor--date_remapper))
- **geo_ip_parser** (Block List, Max: 1) Date GeoIP Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#geoip-parser) (see [below for nested schema](#nestedblock--processor--pipeline--processor--geo_ip_parser))
- **grok_parser** (Block List, Max: 1) Grok Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#grok-parser) (see [below for nested schema](#nestedblock--processor--pipeline--processor--grok_parser))
- **lookup_processor** (Block List, Max: 1) Lookup Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#lookup-processor) (see [below for nested schema](#nestedblock--processor--pipeline--processor--lookup_processor))
- **message_remapper** (Block List, Max: 1) Message Remapper Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#log-message-remapper) (see [below for nested schema](#nestedblock--processor--pipeline--processor--message_remapper))
- **service_remapper** (Block List, Max: 1) Service Remapper Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#service-remapper) (see [below for nested schema](#nestedblock--processor--pipeline--processor--service_remapper))
- **status_remapper** (Block List, Max: 1) Status Remapper Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#log-status-remapper) (see [below for nested schema](#nestedblock--processor--pipeline--processor--status_remapper))
- **string_builder_processor** (Block List, Max: 1) String Builder Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#string-builder-processor) (see [below for nested schema](#nestedblock--processor--pipeline--processor--string_builder_processor))
- **trace_id_remapper** (Block List, Max: 1) Trace ID Remapper Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#trace-remapper) (see [below for nested schema](#nestedblock--processor--pipeline--processor--trace_id_remapper))
- **url_parser** (Block List, Max: 1) URL Parser Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#url-parser) (see [below for nested schema](#nestedblock--processor--pipeline--processor--url_parser))
- **user_agent_parser** (Block List, Max: 1) User-Agent Parser Processor. More information can be found in the [official docs](https://docs.datadoghq.com/logs/processing/processors/?tab=ui#user-agent-parser) (see [below for nested schema](#nestedblock--processor--pipeline--processor--user_agent_parser))

<a id="nestedblock--processor--pipeline--processor--arithmetic_processor"></a>
### Nested Schema for `processor.pipeline.processor.user_agent_parser`

Required:

- **expression** (String, Required) Arithmetic operation between one or more log attributes.
- **target** (String, Required) Name of the attribute that contains the result of the arithmetic operation.

Optional:

- **is_enabled** (Boolean, Optional) Boolean value to enable your pipeline.
- **is_replace_missing** (Boolean, Optional) If true, it replaces all missing attributes of expression by 0, false skips the operation if an attribute is missing.
- **name** (String, Optional) Your pipeline name.


<a id="nestedblock--processor--pipeline--processor--attribute_remapper"></a>
### Nested Schema for `processor.pipeline.processor.user_agent_parser`

Required:

- **source_type** (String, Required) Defines where the sources are from (log attribute or tag).
- **sources** (List of String, Required) List of source attributes or tags.
- **target** (String, Required) Final attribute or tag name to remap the sources.
- **target_type** (String, Required) Defines if the target is a log attribute or tag.

Optional:

- **is_enabled** (Boolean, Optional) If the processor is enabled or not.
- **name** (String, Optional) Name of the processor
- **override_on_conflict** (Boolean, Optional) Override the target element if already set.
- **preserve_source** (Boolean, Optional) Remove or preserve the remapped source element.
- **target_format** (String, Optional) If the `target_type` of the remapper is `attribute`, try to cast the value to a new specific type. If the cast is not possible, the original type is kept. `string`, `integer`, or `double` are the possible types. If the `target_type` is `tag`, this parameter may not be specified.


<a id="nestedblock--processor--pipeline--processor--category_processor"></a>
### Nested Schema for `processor.pipeline.processor.user_agent_parser`

Required:

- **category** (Block List, Min: 1) List of filters to match or exclude a log with their corresponding name to assign a custom value to the log. (see [below for nested schema](#nestedblock--processor--pipeline--processor--user_agent_parser--category))
- **target** (String, Required) Name of the target attribute whose value is defined by the matching category.

Optional:

- **is_enabled** (Boolean, Optional) If the processor is enabled or not.
- **name** (String, Optional) Name of the category

<a id="nestedblock--processor--pipeline--processor--user_agent_parser--category"></a>
### Nested Schema for `processor.pipeline.processor.user_agent_parser.category`

Required:

- **filter** (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--processor--pipeline--processor--user_agent_parser--category--filter))
- **name** (String, Required)

<a id="nestedblock--processor--pipeline--processor--user_agent_parser--category--filter"></a>
### Nested Schema for `processor.pipeline.processor.user_agent_parser.category.name`

Required:

- **query** (String, Required) Filter criteria of the category.




<a id="nestedblock--processor--pipeline--processor--date_remapper"></a>
### Nested Schema for `processor.pipeline.processor.user_agent_parser`

Required:

- **sources** (List of String, Required) List of source attributes.

Optional:

- **is_enabled** (Boolean, Optional) If the processor is enabled or not.
- **name** (String, Optional) Name of the processor.


<a id="nestedblock--processor--pipeline--processor--geo_ip_parser"></a>
### Nested Schema for `processor.pipeline.processor.user_agent_parser`

Required:

- **sources** (List of String, Required) List of source attributes.
- **target** (String, Required) Name of the parent attribute that contains all the extracted details from the sources.

Optional:

- **is_enabled** (Boolean, Optional) If the processor is enabled or not.
- **name** (String, Optional) Name of the processor.


<a id="nestedblock--processor--pipeline--processor--grok_parser"></a>
### Nested Schema for `processor.pipeline.processor.user_agent_parser`

Required:

- **grok** (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--processor--pipeline--processor--user_agent_parser--grok))
- **source** (String, Required) Name of the log attribute to parse.

Optional:

- **is_enabled** (Boolean, Optional) If the processor is enabled or not.
- **name** (String, Optional) Name of the processor
- **samples** (List of String, Optional) List of sample logs for this parser. It can save up to 5 samples. Each sample takes up to 5000 characters.

<a id="nestedblock--processor--pipeline--processor--user_agent_parser--grok"></a>
### Nested Schema for `processor.pipeline.processor.user_agent_parser.grok`

Required:

- **match_rules** (String, Required) Match rules for your grok parser.
- **support_rules** (String, Required) Support rules for your grok parser.



<a id="nestedblock--processor--pipeline--processor--lookup_processor"></a>
### Nested Schema for `processor.pipeline.processor.user_agent_parser`

Required:

- **lookup_table** (List of String, Required) List of entries of the lookup table using `key,value` format.
- **source** (String, Required) Name of the source attribute used to do the lookup.
- **target** (String, Required) Name of the attribute that contains the result of the lookup.

Optional:

- **default_lookup** (String, Optional) Default lookup value to use if there is no entry in the lookup table for the value of the source attribute.
- **is_enabled** (Boolean, Optional) If the processor is enabled or not.
- **name** (String, Optional) Name of the processor


<a id="nestedblock--processor--pipeline--processor--message_remapper"></a>
### Nested Schema for `processor.pipeline.processor.user_agent_parser`

Required:

- **sources** (List of String, Required) List of source attributes.

Optional:

- **is_enabled** (Boolean, Optional) If the processor is enabled or not.
- **name** (String, Optional) Name of the processor.


<a id="nestedblock--processor--pipeline--processor--service_remapper"></a>
### Nested Schema for `processor.pipeline.processor.user_agent_parser`

Required:

- **sources** (List of String, Required) List of source attributes.

Optional:

- **is_enabled** (Boolean, Optional) If the processor is enabled or not.
- **name** (String, Optional) Name of the processor.


<a id="nestedblock--processor--pipeline--processor--status_remapper"></a>
### Nested Schema for `processor.pipeline.processor.user_agent_parser`

Required:

- **sources** (List of String, Required) List of source attributes.

Optional:

- **is_enabled** (Boolean, Optional) If the processor is enabled or not.
- **name** (String, Optional) Name of the processor.


<a id="nestedblock--processor--pipeline--processor--string_builder_processor"></a>
### Nested Schema for `processor.pipeline.processor.user_agent_parser`

Required:

- **target** (String, Required) The name of the attribute that contains the result of the template.
- **template** (String, Required) The formula with one or more attributes and raw text.

Optional:

- **is_enabled** (Boolean, Optional) If the processor is enabled or not.
- **is_replace_missing** (Boolean, Optional) If it replaces all missing attributes of template by an empty string.
- **name** (String, Optional) The name of the processor.


<a id="nestedblock--processor--pipeline--processor--trace_id_remapper"></a>
### Nested Schema for `processor.pipeline.processor.user_agent_parser`

Required:

- **sources** (List of String, Required) List of source attributes.

Optional:

- **is_enabled** (Boolean, Optional) If the processor is enabled or not.
- **name** (String, Optional) Name of the processor.


<a id="nestedblock--processor--pipeline--processor--url_parser"></a>
### Nested Schema for `processor.pipeline.processor.user_agent_parser`

Required:

- **sources** (List of String, Required) List of source attributes.
- **target** (String, Required) Name of the parent attribute that contains all the extracted details from the sources.

Optional:

- **is_enabled** (Boolean, Optional) If the processor is enabled or not.
- **name** (String, Optional) Name of the processor
- **normalize_ending_slashes** (Boolean, Optional) Normalize the ending slashes or not.


<a id="nestedblock--processor--pipeline--processor--user_agent_parser"></a>
### Nested Schema for `processor.pipeline.processor.user_agent_parser`

Required:

- **sources** (List of String, Required) List of source attributes.
- **target** (String, Required) Name of the parent attribute that contains all the extracted details from the sources.

Optional:

- **is_enabled** (Boolean, Optional) If the processor is enabled or not.
- **is_encoded** (Boolean, Optional) If the source attribute is URL encoded or not.
- **name** (String, Optional) Name of the processor




<a id="nestedblock--processor--service_remapper"></a>
### Nested Schema for `processor.service_remapper`

Required:

- **sources** (List of String, Required) List of source attributes.

Optional:

- **is_enabled** (Boolean, Optional) If the processor is enabled or not.
- **name** (String, Optional) Name of the processor.


<a id="nestedblock--processor--status_remapper"></a>
### Nested Schema for `processor.status_remapper`

Required:

- **sources** (List of String, Required) List of source attributes.

Optional:

- **is_enabled** (Boolean, Optional) If the processor is enabled or not.
- **name** (String, Optional) Name of the processor.


<a id="nestedblock--processor--string_builder_processor"></a>
### Nested Schema for `processor.string_builder_processor`

Required:

- **target** (String, Required) The name of the attribute that contains the result of the template.
- **template** (String, Required) The formula with one or more attributes and raw text.

Optional:

- **is_enabled** (Boolean, Optional) If the processor is enabled or not.
- **is_replace_missing** (Boolean, Optional) If it replaces all missing attributes of template by an empty string.
- **name** (String, Optional) The name of the processor.


<a id="nestedblock--processor--trace_id_remapper"></a>
### Nested Schema for `processor.trace_id_remapper`

Required:

- **sources** (List of String, Required) List of source attributes.

Optional:

- **is_enabled** (Boolean, Optional) If the processor is enabled or not.
- **name** (String, Optional) Name of the processor.


<a id="nestedblock--processor--url_parser"></a>
### Nested Schema for `processor.url_parser`

Required:

- **sources** (List of String, Required) List of source attributes.
- **target** (String, Required) Name of the parent attribute that contains all the extracted details from the sources.

Optional:

- **is_enabled** (Boolean, Optional) If the processor is enabled or not.
- **name** (String, Optional) Name of the processor
- **normalize_ending_slashes** (Boolean, Optional) Normalize the ending slashes or not.


<a id="nestedblock--processor--user_agent_parser"></a>
### Nested Schema for `processor.user_agent_parser`

Required:

- **sources** (List of String, Required) List of source attributes.
- **target** (String, Required) Name of the parent attribute that contains all the extracted details from the sources.

Optional:

- **is_enabled** (Boolean, Optional) If the processor is enabled or not.
- **is_encoded** (Boolean, Optional) If the source attribute is URL encoded or not.
- **name** (String, Optional) Name of the processor

## Import

Import is supported using the following syntax:

```shell
# For the previously created custom pipelines, you can include them in Terraform with the import operation. Currently, Terraform requires you to explicitly create resources that match the existing pipelines to import them.
terraform import <resource.name> <pipelineID>
```
