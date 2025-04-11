resource "datadog_observability_pipeline" "test" {
  name = "test pipeline"
  config {

    sources {
      kafka {
        id       = "source-1"
        group_id = "my-consumer-group"
        topics   = ["my-topic-1", "my-topic-2"]

        tls {
          crt_file = "/etc/certs/client.crt"
          key_file = "/etc/certs/client.key"
          ca_file  = "/etc/certs/ca.crt"
        }

        sasl {
          mechanism = "SCRAM-SHA-512"
        }

        librdkafka_option {
          name  = "fetch.message.max.bytes"
          value = "1048576"
        }

        librdkafka_option {
          name  = "socket.timeout.ms"
          value = "500"
        }
      }
    }

    processors {
      parse_json {
        id      = "filter-1"
        include = "service:nginx"
        field   = "message2"
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
