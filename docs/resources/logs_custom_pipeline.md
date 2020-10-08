---
page_title: "datadog_logs_custom_pipeline"
---

# datadog_logs_custom_pipeline Resource

Provides a Datadog [Logs Pipeline API](https://docs.datadoghq.com/api/v1/logs-pipelines/) resource, which is used to create and manage Datadog logs custom pipelines.

## Example Usage

Create a Datadog logs pipeline:

```hcl
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
            target_format = "String"
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

## Argument Reference

The following arguments are supported:

- `filter`: (Required) Defines your pipeline filter. Only logs that match the filter criteria are processed by this pipeline.
  - `query`: (Required) Defines the filter criteria.
- `name`: (Required) Your pipeline name.
- `is_enabled`: (Optional, default = false) Boolean value to enable your pipeline.
- `processor`: (Optional) Processors or nested pipelines. See [below](logs_custom_pipeline.html#Processors) for more detailed descriptions.

**Note** A pipeline or its processors are disabled by default if `is_enabled` is not explicitly set to `true`.

### Processors

- arithmetic_processor
  - `expression`: (Required) Arithmetic operation between one or more log attributes.
  - `target`: (Required) Name of the attribute that contains the result of the arithmetic operation.
  - `is_replace_missing`: (Optional, default = false) If true, it replaces all missing attributes of expression by 0, false skips the operation if an attribute is missing.
  - `name`: (Optional) Name of the processor
  - `is_enabled`: (Optional, default = false) If the processor is enabled or not.
- attribute_remapper
  - `sources`: (Required) List of source attributes or tags.
  - `source_type`: (Required) Defines where the sources are from (log `attribute` or `tag`).
  - `target`: (Required) Final `attribute` or `tag` name to remap the sources.
  - `target_type`: (Required) Defines if the target is a log `attribute` or `tag`.
  - `target_format`: (Optional) If the `target_type` of the remapper is `attribute`, try to cast the value to a new specific type. If the cast is not possible, the original type is kept. `string`, `integer`, or `double` are the possible types. If the `target_type` is `tag`, it should not be specified, otherwise it will raise an issue.
  - `preserve_source`: (Optional, default = false) Remove or preserve the remapped source element.
  - `override_on_conflict`: (Optional, default = false) Override the target element if already set.
  - `name`: (Optional) Name of the processor
  - `is_enabled`: (Optional, default = false) If the processor is enabled or not.
- category_processor
  - `target`: (Required) Name of the target attribute whose value is defined by the matching category.
  - `category`: (Required) List of filters to match or exclude a log with their corresponding name to assign a custom value to the log.
    - `name`: (Required) Name of the cateory.
    - `filter`
      - `query`: (Required) Filter criteria of the category.
  - `name`: (Optional) Name of the processor
  - `is_enabled`: (Optional, default = false) If the processor is enabled or not.
- date_remapper
  - `sources`: (Required) List of source attributes.
  - `name`: (Optional) Name of the processor.
  - `is_enabled`: (Optional, default = false) If the processor is enabled or not.
- geo_ip_parser
  - `sources`: (Required) List of source attributes.
  - `target`: (Required) Name of the parent attribute that contains all the extracted details from the `sources`.
  - `name`: (Optional) Name of the processor.
  - `is_enabled`: (Optional, default = false) If the processor is enabled or not.
- grok_parser
  - `samples`: (Optional) List of sample logs for this parser. It can save up to 5 samples. Each sample takes up to 5000 characters.
  - `source`: (Required) Name of the log attribute to parse.
  - `grok`
    - `support_rules`: (Required) Support rules for your grok parser.
    - `match_rules`: (Required) Match rules for your grok parser.
  - `name`: (Optional) Name of the processor.
  - `is_enabled`: (Optional, default = false) If the processor is enabled or not.
- lookup_processor
  - `source`: (Required) Name of the source attribute used to do the lookup.
  - `target`: (Required) Name of the attribute that contains the result of the lookup.
  - `lookup_table`: (Required) List of entries of the lookup table using `"key,value"` format.
  - `default_lookup`: (Optional) Default lookup value to use if there is no entry in the lookup table for the value of the source attribute.
  - `name`: (Optional) Name of the processor.
  - `is_enabled`: (Optional, default = false) If the processor is enabled or not.
- message_remapper
  - `sources`: (Required) List of source attributes.
  - `name`: (Optional) Name of the processor.
  - `is_enabled`: (Optional, default = false) If the processor is enabled or not.
- pipeline
  - `filter`: (Required) Defines the nested pipeline filter. Only logs that match the filter criteria are processed by this pipeline.
    - `query`: (Required)
  - `processor`: (Optional) [Processors](logs_custom_pipeline.html#Processors). Nested pipeline can't take any other nested pipeline as its processor.
  - `name`: (Optional) Name of the nested pipeline.
  - `is_enabled`: (Optional, default = false) If the processor is enabled or not.
- service_remapper
  - `sources`: (Required) List of source attributes.
  - `name`: (Optional) Name of the processor
  - `is_enabled`: (Optional, default = false) If the processor is enabled or not.
- status_remapper
  - `sources`: (Required) List of source attributes.
  - `name`: (Optional) Name of the processor
  - `is_enabled`: (Optional, default = false) If the processor is enabled or not.
- string_builder_processor
  - `target`: (Required) The name of the attribute that contains the result of the template.
  - `template`: (Required) The formula with one or more attributes and raw text.
  - `name`: (Optional) The name of the processor.
  - `is_enabled`: (Optional, default = false) If the processor is enabled or not.
  - `is_replace_missing`: (Optional, default = false) If it replaces all missing attributes of `template` by an empty string.
- trace_id_remapper
  - `sources`: (Required) List of source attributes.
  - `name`: (Optional) Name of the processor
  - `is_enabled`: (Optional, default = false) If the processor is enabled or not.
- url_parser
  - `sources`: (Required) List of source attributes.
  - `target`: (Required) Name of the parent attribute that contains all the extracted details from the `sources`.
  - `normalize_ending_slashes`: (Optional, default = false) Normalize the ending slashes or not.
  - `name`: (Optional) Name of the processor
  - `is_enabled`: (Optional, default = false) If the processor is enabled or not.
- user_agent_parser
  - `sources`: (Required) List of source attributes.
  - `target`: (Required) Name of the parent attribute that contains all the extracted details from the sources.
  - `is_encoded`: (Optional, default = false) If the source attribute is URL encoded or not.
  - `name`: (Optional) Name of the processor
  - `is_enabled`: (Optional, default = false) If the processor is enabled or not.

Even though some arguments are optional, we still recommend you to state **all** arguments in the resource to avoid unnecessary confusion.

## Import

For the previously created custom pipelines, you can include them in Terraform with the `import` operation. Currently, Terraform requires you to explicitly create resources that match the existing pipelines to import them. Use `terraform import <resource.name> <pipelineID>` for each existing pipeline.

## Important Notes

Each `datadog_logs_custom_pipeline` resource defines a complete pipeline. The order of the pipelines is maintained in a different resource [datadog_logs_pipeline_order](logs_pipeline_order.html#datadog_logs_pipeline_order). When creating a new pipeline, you need to **explicitly** add this pipeline to the `datadog_logs_pipeline_order` resource to track the pipeline. Similarly, when a pipeline needs to be destroyed, remove its references from the `datadog_logs_pipeline_order` resource.
