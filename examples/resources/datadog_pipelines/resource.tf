# Create new pipelines resource

resource "datadog_pipelines" "test" {
  name = "test TF pipeline"
  config {

    source {
      datadog_agent {
        id = "source-1"
      }
    }

    processor {
      parse_json {
        id      = "filter-1"
        include = "service:nginx"
        field   = "message"
        inputs  = ["source-1"]
      }
    }

    processor {
      filter {
        id      = "filter-2"
        include = "service:nginx"
        inputs  = ["filter-1"]
      }
    }

    processor {
      parse_json {
        id      = "filter-3"
        include = "service:nginx"
        field   = "message"
        inputs  = ["filter-2"]
      }
    }

    destination {
      datadog_logs {
        id     = "sink-1"
        inputs = ["filter-3"]
      }
    }

  }
}
