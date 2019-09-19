---
layout: "datadog"
page_title: "Datadog: datadog_logs_pipeline"
sidebar_current: "docs-datadog-resource-logs-pipeline"
description: |-
  Provides a Datadog logs pipeline resource. This can be used to create and manage logs pipelines.
---

# datadog_logs_pipeline

Provides a Datadog [Logs Pipeline API](https://docs.datadoghq.com/api/?lang=python#logs-pipelines) resource. This can be used to create and manage Datadog logs pipelines.


## Example Usage: Datadog logs pipeline

```hcl
# Create a new Datadog logs pipeline
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
## Example Usage: Datadog logs pipeline order

```hcl
resource "datadog_logs_pipelineorder" "sample_pipeline_order" {
    name = "sample_pipeline_order"
    depends_on = [
        "datadog_logs_pipeline.sample_pipeline"
    ]
    pipelines = [
        "${datadog_logs_pipeline.sample_pipeline.id}"
    ]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Your pipeline name.
* `is_enabled` - (Optional, default = false) Boolean value to enable your pipeline.
* `filter` - (Required) Defines your pipeline filter. Only logs that match the filter criteria are processed by this pipeline.
  * `query` - (Required)
* `processor` - (Optional) processors or nested pipelines. See [below](logs_pipeline.html#Processors) for more detailed descriptions.

### Processors

* arithmetic_processor
  * `name` - (Optional) Name of the processor
  * `is_enabled` - (Optional, default = false) If the processor is enabled or not.
  * `expression` - (Required) Arithmetic operation between one or more log attributes.
  * `target` - (Required) Name of the attribute that contains the result of the arithmetic operation.
  * `is_replace_missing` - (Optional, default = false) If true, it replaces all missing attributes of expression by 0, false skip the operation if an attribute is missing.
* attribute_remapper
  * `name` - (Optional) Name of the processor
  * `is_enabled` - (Optional, default = false) If the processor is enabled or not.
  * `sources` - (Required) List of source attributes or tags.
  * `source_type` - (Required) Defines if the sources are from log `attribute` or `tag`. 
  * `target` - (Required) Final `attribute` or `tag` name to remap the sources to.
  * `target_type` - (Required) Defines if the target is a log `attribute` or `tag`.
  * `preserve_source` - (Optional, default = false) Remove or preserve the remapped source element.
  * `override_on_conflict` - (Optional, default = false) Override or not the target element if already set.
* category_processor
  * `name` - (Optional) Name of the processor
  * `is_enabled` - (Optional, default = false) If the processor is enabled or not.
  * `target` - (Required) Name of the target attribute which value is defined by the matching category.
  * `category` - (Required) List of filters to match or not a log and their corresponding name to assign a custom value to the log.
    * `name` - (Required) 
    * `filter`
      * `query` - (Required)
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
  * `processor` - (Optional) [Processors](logs_pipeline.html#Processors). It can not be pipelines anymore.
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
  * `normalize_ending_slashes` - (Optional, default = false) If needs to normalize the ending slashes or not.
* user_agent_parser
  * `name` - (Optional) Name of the processor
  * `is_enabled` - (Optional, default = false) If the processor is enabled or not.
  * `sources` - (Required) List of source attributes.
  * `target` - (Required) Name of the parent attribute that contains all the extracted details from the sources.
  * `is_encoded` - (Optional, default = false) If the source attribute is url encoded or not.

## Attributes Reference

* `name` - (Required) The name attribute in resource `datadog_logs_pipelineorder` needs to be unique. It's better to set to the same value as the resource `NAME`. 
There is no related field can be found in  [Logs Pipeline API](https://docs.datadoghq.com/api/?lang=python#get-pipeline-order).
* `pipelines` - (Required) The pipeline ids list. The order of pipeline ids in this attribute defines the overall pipelines order for Logs.

**Note:** The [read-only pipelines](https://docs.datadoghq.com/logs/processing/pipelines/#integration-pipelines) also need to be included in the `pipelines`
list in resource `datadog_logs_pipelineorder`. You can update the order of these pipelines, but you can't modify the **read-only** pipelines 
through this provider. 

## Import

For the pipelines that are already created (including the **read-only** ones), you can include them into terraform by `import` operation.
At the moment of writing this documentation, terraform requires to explicitly create resources that match the existing pipelines to be able to
import them. You will need to do ```terraform import <resource.name> <pipelineID>``` for each existing pipeline.

## Important Notes

Each `datadog_logs_pipeline` resource defines a complete pipeline. The order of pipelines are maintained in resource `datadog_logs_pipelineorder`.
There should be just one `datadog_logs_pipelineorder` resource. When creating a new pipeline, you need to **explicitly** add this pipeline to 
`datadog_logs_pipelineorder` resource to keep track of this pipeline. Similarly, when a pipeline needs to be destroyed, do not forget to also remove
its references from `datadog_logs_pipelineorder` resource. 
