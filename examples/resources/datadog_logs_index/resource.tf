# A sample Datadog logs index resource definition. Note that at this point, it is not possible to create new logs indexes through Terraform, so the name field must match a name of an already existing index. If you want to keep the current state of the index, we suggest importing it (see below).

resource "datadog_logs_index" "sample_index" {
  name = "your index"
  daily_limit = 200000
  retention_days = 7
  filter {
    query = "*"
  }
  exclusion_filter {
    name       = "Filter coredns logs"
    is_enabled = true
    filter {
      query       = "app:coredns"
      sample_rate = 0.97
    }
  }
  exclusion_filter {
    name       = "Kubernetes apiserver"
    is_enabled = true
    filter {
      query       = "service:kube_apiserver"
      sample_rate = 1.0
    }
  }
}
