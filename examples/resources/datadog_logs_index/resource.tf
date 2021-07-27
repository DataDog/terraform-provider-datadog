# A sample Datadog logs index resource definition.

resource "datadog_logs_index" "sample_index" {
  name           = "your index"
  daily_limit    = 200000
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
