resource "datadog_observability_pipeline" "test" {
  name = "test pipeline"
  config {
    source {
      id = "source-1"

      datadog_agent {
        tls {
          crt_file = "/etc/certs/client.crt"
          key_file = "/etc/certs/client.key"
          ca_file  = "/etc/certs/ca.crt"
        }
      }
    }

    processor_group {
      id           = "processor-group-1"
      enabled      = true
      include      = "service:my-service"
      inputs       = ["source-1"]
      display_name = "processor group"

      processor {
        id           = "parser-1"
        enabled      = true
        include      = "service:my-service"
        display_name = "json parser"

        parse_json {
          field = "message"
        }
      }

      processor {
        id           = "filter-1"
        enabled      = true
        include      = "service:my-service"
        display_name = "filter"

        filter {}
      }
    }

    destination {
      id     = "destination-1"
      inputs = ["processor-group-1"]

      datadog_logs {}
    }
  }
}