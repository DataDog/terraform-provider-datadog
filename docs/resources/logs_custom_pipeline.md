---
page_title: "datadog_logs_custom_pipeline Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  
---

# Resource `datadog_logs_custom_pipeline`



## Example Usage

```terraform
resource "datadog_logs_custom_pipeline" "sample_pipeline" {
    filter {
        query = "source:foo"
    }
    name = "sample pipeline"
    is_enabled = true
    processor {
        arithmetic_processor {
            expression = "(time1 - time2)*1000"
            target = "my_arithmetic"
            is_replace_missing = true
            name = "sample arithmetic processor"
            is_enabled = true
        }
    }
    processor {
        attribute_remapper {
            sources = ["db.instance"]
            source_type = "tag"
            target = "db"
            target_type = "attribute"
            target_format = "string"
            preserve_source = true
            override_on_conflict = false
            name = "sample attribute processor"
            is_enabled = true
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
            name = "sample category processor"
            is_enabled = true
        }
    }
    processor {
        date_remapper {
            sources = ["_timestamp", "published_date"]
            name = "sample date remapper"
            is_enabled = true
        }
    }
    processor {
        geo_ip_parser {
            sources = ["network.client.ip"]
            target = "network.client.geoip"
            name = "sample geo ip parser"
            is_enabled = true
        }
    }
    processor {
        grok_parser {
            samples = ["sample log 1"]
            source = "message"
            grok {
                support_rules = ""
                match_rules = "Rule %%{word:my_word2} %%{number:my_float2}"
            }
            name = "sample grok parser"
            is_enabled = true
        }
    }
    processor {
        lookup_processor {
            source = "service_id"
            target = "service_name"
            lookup_table = ["1,my service"]
            default_lookup = "unknown service"
            name = "sample lookup processor"
            is_enabled = true
        }
    }
    processor {
        message_remapper {
            sources = ["msg"]
            name = "sample message remapper"
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
                    name = "sample url parser"
                    sources = ["url", "extra"]
                    target = "http_url"
                    normalize_ending_slashes = true
                }
            }
            name = "nested pipeline"
            is_enabled = true
        }
    }
    processor {
        service_remapper {
            sources = ["service"]
            name = "sample service remapper"
            is_enabled = true
        }
    }
    processor {
        status_remapper {
            sources = ["info", "trace"]
            name = "sample status remapper"
            is_enabled = true
        }
    }
    processor {
        string_builder_processor {
            target = "user_activity"
            template = "%%{user.name} logged in at %%{timestamp}"
            name = "sample string builder processor"
            is_enabled = true
            is_replace_missing = false
        }
    }
    processor {
        trace_id_remapper {
            sources = ["dd.trace_id"]
            name = "sample trace id remapper"
            is_enabled = true
        }
    }
    processor {
        user_agent_parser {
            sources = ["user", "agent"]
            target = "http_agent"
            is_encoded = false
            name = "sample user agent parser"
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

## Import

Import is supported using the following syntax:

```shell
# For the previously created custom pipelines, you can include them in Terraform with the import operation. Currently, Terraform requires you to explicitly create resources that match the existing pipelines to import them.
terraform import <resource.name> <pipelineID>
```
