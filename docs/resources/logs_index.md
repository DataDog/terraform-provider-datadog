---
page_title: "datadog_logs_index"
---

# datadog_logs_index Resource

Provides a Datadog [Logs Index API](https://docs.datadoghq.com/api/v1/logs-indexes/) resource. This can be used to create and manage Datadog logs indexes.

## Example Usage

A sample Datadog logs index resource definition. Note that at this point, it is not possible to create new logs indexes through Terraform, so the `name` field must match a name of an already existing index. If you want to keep the current state of the index, we suggest importing it (see below).

```hcl
resource "datadog_logs_index" "sample_index" {
    name = "your index"
    filter {
        query = "*"
    }
    exclusion_filter {
        name = "Filter coredns logs"
        is_enabled = true
        filter {
            query = "app:coredns"
            sample_rate = 0.97
        }
    }
    exclusion_filter {
        name = "Kubernetes apiserver"
        is_enabled = true
        filter {
            query = "service:kube_apiserver"
            sample_rate = 1.0
        }
    }
}
```

## Argument Reference

The following arguments are supported:

-   `name`: (Required) The name of the index.
-   `filter`: (Required) Logs filter.
    -   `query`: (Required) Logs filter criteria. Only logs matching this filter criteria are considered for this index.
-   `exclusion_filter`: (Optional) List of exclusion filters.
    -   `filter`: (Optional)
        -   `query`: (Optional) Only logs matching the filter criteria and the query of the parent index will be considered for this exclusion filter.
        -   `sample_rate`: (Optional, default = 0.0) The fraction of logs excluded by the exclusion filter, when active.
    -   `name`: (Optional) The name of the exclusion filter.
    -   `is_enabled`: (Optional, default = false) A boolean stating if the exclusion is active or not.

## Import

The current Datadog Terraform provider version does not support the creation and deletion of indexes. To manage the existing indexes, do `terraform import <datadog_logs_index.name> <indexName>` to import them to Terraform. If you create a resource which does not match the name of any existing index, `terraform apply` will throw `Not Found` error code.

## Important Notes

The order of indexes is maintained in the separated resource [datadog_logs_index_order](logs_index_order.html#datadog_logs_index_order).
