# Create new pipelines resource

resource "datadog_observability_pipeline" "test" {
  name = "test TF pipeline"
  config {

    sources {
      datadog_agent {
        id = "source-1"
      }
    }

    processors {
      parse_json {
        id      = "filter-1"
        include = "service:nginx"
        field   = "message"
        inputs  = ["source-1"]
      }

      filter {
        id      = "filter-2"
        include = "service:nginx"
        inputs  = ["filter-1"]
      }

      parse_json {
        id      = "filter-3"
        include = "service:nginx"
        field   = "message"
        inputs  = ["filter-2"]
      }
    }

    destinations {
      datadog_logs {
        id     = "sink-1"
        inputs = ["filter-3"]
      }
    }

  }
}
