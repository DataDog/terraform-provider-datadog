# Create a new Datadog timeboard
resource "datadog_timeboard" "redis" {
  title       = "Redis Timeboard (created via Terraform)"
  description = "created using the Datadog provider in Terraform"
  read_only   = true

  graph {
    title = "Redis latency (ms)"
    viz   = "timeseries"

    request {
      q    = "avg:redis.info.latency_ms{$host}"
      type = "bars"

      # NOTE: this will only work with TF >= 0.12; see metadata_json
      # documentation below for example on usage with TF < 0.12
      metadata_json = jsonencode({
        "avg:redis.info.latency_ms{$host}": {
          "alias": "Redis latency"
        }
      })
    }

    request {
        log_query {
          index = "mcnulty"
          compute {
            aggregation = "avg"
            facet = "@duration"
            interval = 5000
          }
          search {
            query = "status:info"
          }
          group_by {
            facet = "host"
            limit = 10
            sort {
              aggregation = "avg"
              order = "desc"
              facet = "@duration"
            }
          }
        }
        type = "area"
      }

      request {
        apm_query {
          index = "apm-search"
          compute {
            aggregation = "avg"
            facet = "@duration"
            interval = 5000
          }
          search {
            query = "type:web"
          }
          group_by {
            facet = "resource_name"
            limit = 50
            sort {
              aggregation = "avg"
              order = "desc"
              facet = "@string_query.interval"
            }
          }
        }
        type = "bars"
      }

      request {
        process_query {
          metric = "process.stat.cpu.total_pct"
          search_by = "error"
          filter_by = ["active"]
          limit = 50
        }
        type = "area"
      }
  }

  graph {
    title = "Redis memory usage"
    viz   = "timeseries"

    request {
      q       = "avg:redis.mem.used{$host} - avg:redis.mem.lua{$host}, avg:redis.mem.lua{$host}"
      stacked = true
    }

    request {
      q = "avg:redis.mem.rss{$host}"

      style = {
        palette = "warm"
      }
    }
  }

  graph {
    title = "Top System CPU by Docker container"
    viz   = "toplist"

    request {
      q = "top(avg:docker.cpu.system{*} by {container_name}, 10, 'mean', 'desc')"
    }
  }

  template_variable {
    name   = "host"
    prefix = "host"
  }
}
