---
layout: "datadog"
page_title: "Datadog: datadog_logs_pipeline"
sidebar_current: "docs-datadog-resource-logs-pipeline"
description: |-
  Provides a Datadog logs pipeline resource, which is used to create and manage logs pipelines.
---

# datadog_logs_pipeline

Provides a Datadog [Logs Pipeline API](https://docs.datadoghq.com/api/?lang=python#logs-pipelines) resource, which is used to create and manage Datadog logs pipelines.


## Example Usage

Create a Datadog logs pipeline:

```hcl
resource "datadog_logs_pipeline" "sample_pipeline" {
    name = "sample pipeline"
    is_enabled = true
    filter {
        query = "source:foo"
    }
    processor {
        arithmetic_processor {
            name = "sample arithmetic processor"
            is_enabled = true
            expression = "(time1 - time2)*1000"
            target = "my_arithmetic"
            is_replace_missing = true
        }
    }
    processor {
        attribute_remapper {
            name = "sample attribute processor"
            is_enabled = true
            sources = ["db.instance"]
            source_type = "tag"
            target = "db"
            target_type = "tag"
            preserve_source = true
            override_on_conflict = false
        }
    }
    processor {
        category_processor {
            name = "sample category processor"
            is_enabled = true
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
        }
    }
    processor {
        date_remapper {
            name = "sample date remapper"
            is_enabled = true
            sources = ["_timestamp", "published_date"]
        }
    }
    processor {
        message_remapper {
            name = "sample message remapper"
            is_enabled = true
            sources = ["msg"]
        }
    }
    processor {
        status_remapper {
            name = "sample status remapper"
            is_enabled = true
            sources = ["info", "trace"]
        }
    }
    processor {
        trace_id_remapper {
            name = "sample trace id remapper"
            is_enabled = true
            sources = ["dd.trace_id"]
        }
    }
    processor {
        service_remapper {
            name = "sample service remapper"
            is_enabled = true
            sources = ["service"]
        }
    }
    processor {
        grok_parser {
            name = "sample grok parser"
            is_enabled = true
            source = "message"
            grok {
                support_rules = ""
                match_rules = "Rule %%{word:my_word2} %%{number:my_float2}"
            }
        }
    }
    processor {
        pipeline {
            name = "nested pipeline"
            is_enabled = true
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
        }
    }
    processor {
        user_agent_parser {
            name = "sample user agent parser"
            is_enabled = true
            sources = ["user", "agent"]
            target = "http_agent"
            is_encoded = false
        }
    }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Your pipeline name.
* `is_enabled` - (Optional, default = false) Boolean value to enable your pipeline.
* `filter` - (Required) Defines your pipeline filter. Only logs that match the filter criteria are processed by this pipeline.
  * `query` - (Required) Defines the filter criteria.
* `processor` - (Optional) Processors or nested pipelines. See [below](logs_pipeline.html#Processors) for more detailed descriptions.

**Note** A pipeline or its processors are disabled by default if `is_enabled` is not explicitly set to `true`.

### Processors

* arithmetic_processor
  * `name` - (Optional) Name of the processor
  * `is_enabled` - (Optional, default = false) If the processor is enabled or not.
  * `expression` - (Required) Arithmetic operation between one or more log attributes.
  * `target` - (Required) Name of the attribute that contains the result of the arithmetic operation.
  * `is_replace_missing` - (Optional, default = false) If true, it replaces all missing attributes of expression by 0, false skips the operation if an attribute is missing.
* attribute_remapper
  * `name` - (Optional) Name of the processor
  * `is_enabled` - (Optional, default = false) If the processor is enabled or not.
  * `sources` - (Required) List of source attributes or tags.
  * `source_type` - (Required) Defines where the sources are from (log `attribute` or `tag`). 
  * `target` - (Required) Final `attribute` or `tag` name to remap the sources.
  * `target_type` - (Required) Defines if the target is a log `attribute` or `tag`.
  * `preserve_source` - (Optional, default = false) Remove or preserve the remapped source element.
  * `override_on_conflict` - (Optional, default = false) Override the target element if already set.
* category_processor
  * `name` - (Optional) Name of the processor
  * `is_enabled` - (Optional, default = false) If the processor is enabled or not.
  * `target` - (Required) Name of the target attribute whose value is defined by the matching category.
  * `category` - (Required) List of filters to match or exclude a log with their corresponding name to assign a custom value to the log.
    * `name` - (Required) Name of the cateory.
    * `filter`
      * `query` - (Required) Filter criteria of the category.
* date_remapper
  * `name` - (Optional) Name of the processor.
  * `is_enabled` - (Optional, default = false) If the processor is enabled or not.
  * `sources` - (Required) List of source attributes.
* grok_parser
  * `name` - (Optional) Name of the processor.
  * `is_enabled` - (Optional, default = false) If the processor is enabled or not.
  * `source` - (Required) Name of the log attribute to parse.
  * `grok`
    * `support_rules` - (Required) Support rules for your grok parser.
    * `match_rules` - (Required) Match rules for your grok parser.
* message_remapper
  * `name` - (Optional) Name of the processor.
  * `is_enabled` - (Optional, default = false) If the processor is enabled or not.
  * `sources` - (Required) List of source attributes.
* pipeline
  * `name` - (Optional) Name of the nested pipeline.
  * `is_enabled` - (Optional, default = false) If the processor is enabled or not.
  * `filter` - (Required) Defines the nested pipeline filter. Only logs that match the filter criteria are processed by this pipeline.
    * `query` - (Required)
  * `processor` - (Optional) [Processors](logs_pipeline.html#Processors). Nested pipeline can't take any other nested pipeline as its processor.
* service_remapper
  * `name` - (Optional) Name of the processor
  * `is_enabled` - (Optional, default = false) If the processor is enabled or not.
  * `sources` - (Required) List of source attributes.
* status_remapper
  * `name` - (Optional) Name of the processor
  * `is_enabled` - (Optional, default = false) If the processor is enabled or not.
  * `sources` - (Required) List of source attributes.
* trace_id_remapper
  * `name` - (Optional) Name of the processor
  * `is_enabled` - (Optional, default = false) If the processor is enabled or not.
  * `sources` - (Required) List of source attributes.
* url_parser
  * `name` - (Optional) Name of the processor
  * `is_enabled` - (Optional, default = false) If the processor is enabled or not.
  * `sources` - (Required) List of source attributes. 
  * `target` - (Required) Name of the parent attribute that contains all the extracted details from the `sources`.
  * `normalize_ending_slashes` - (Optional, default = false) Normalize the ending slashes or not.
* user_agent_parser
  * `name` - (Optional) Name of the processor
  * `is_enabled` - (Optional, default = false) If the processor is enabled or not.
  * `sources` - (Required) List of source attributes.
  * `target` - (Required) Name of the parent attribute that contains all the extracted details from the sources.
  * `is_encoded` - (Optional, default = false) If the source attribute is URL encoded or not.

## Import

For the previously created pipelines (including the **read-only** ones), you can include them in Terraform with the `import` operation.
Currently, Terraform requires you to explicitly create resources that match the existing pipelines to 
import them. Use ```terraform import <resource.name> <pipelineID>``` for each existing pipeline.

## Important Notes

Each `datadog_logs_pipeline` resource defines a complete pipeline. The order of pipelines are maintained in resource 
[datadog_logs_pipeline_order](logs_pipeline_order.html#datadog_logs_pipeline_order). When creating a new pipeline, you
need to **explicitly** add this pipeline to the `datadog_logs_pipeline_order` resource to track the pipeline.
Similarly, when a pipeline needs to be destroyed, remove its references from the `datadog_logs_pipeline_order` resource.
