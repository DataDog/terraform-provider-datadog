resource "datadog_notebook" "foo" {
  cells {
    type = "notebook_cells"
    attributes {
      markdown_cell {
        definition {
          text = "## markdown text"
        }
      }
    }
  }

  cells {
    type = "notebook_cells"
    attributes {
      timeseries_cell {
        definition {
          request {
            q            = "avg:system.cpu.user{app:general} by {env}"
            display_type = "line"
            style {
              palette    = "warm"
              line_type  = "dashed"
              line_width = "thin"
            }
            metadata {
              expression = "avg:system.cpu.user{app:general} by {env}"
              alias_name = "Alpha"
            }
          }
        }
        split_by {
          keys = ["test"]
          tags = ["test:true"]
        }
        time {
          notebook_absolute_time {
            start = "2021-02-24T19:18:28+08:00"
            end   = "2021-02-24T19:20:28+08:00"
          }
        }
        graph_size = "m"
      }
    }
  }


  cells {
    type = "notebook_cells"
    attributes {
      toplist_cell {
        definition {
          request {
            q = "avg:system.cpu.user{app:general} by {env}"
            conditional_formats {
              comparator = "<"
              value      = "2"
              palette    = "white_on_green"
            }
            conditional_formats {
              comparator = ">"
              value      = "2.2"
              palette    = "white_on_red"
            }
          }
          title = "Widget Title"

        }
        split_by {
          keys = ["test"]
          tags = ["test:true"]
        }
        time {
          notebook_absolute_time {
            start = "2021-02-24T19:18:28+08:00"
            end   = "2021-02-24T19:20:28+08:00"
          }
        }
        graph_size = "m"
      }
    }
  }

  cells {
    type = "notebook_cells"
    attributes {
      heatmap_cell {
        definition {
          request {
            q = "avg:system.load.1{env:staging} by {account}"
            style {
              palette = "warm"
            }
          }
        }
        split_by {
          keys = ["test"]
          tags = ["test:true"]
        }
        time {
          notebook_absolute_time {
            start = "2021-02-24T19:18:28+08:00"
            end   = "2021-02-24T19:20:28+08:00"
          }
        }
        graph_size = "m"
      }
    }
  }

  cells {
    type = "notebook_cells"
    attributes {
      distribution_cell {
        definition {
          request {
            q = "avg:system.load.1{env:staging} by {account}"
            style {
              palette = "warm"
            }
          }
        }
        split_by {
          keys = ["test"]
          tags = ["test:true"]
        }
        time {
          notebook_absolute_time {
            start = "2021-02-24T19:18:28+08:00"
            end   = "2021-02-24T19:20:28+08:00"
          }
        }
        graph_size = "m"
      }
    }
  }

  cells {
    type = "notebook_cells"
    attributes {
      log_stream_cell {
        definition {
          indexes             = ["main"]
          query               = "error"
          columns             = ["core_host", "core_service", "tag_source"]
          show_date_column    = true
          show_message_column = true
          message_display     = "expanded-md"
          sort {
            column = "time"
            order  = "desc"
          }

        }
        time {
          notebook_relative_time {
            live_span = "5m"
          }
        }
        graph_size = "m"
      }
    }
  }

  time {
    notebook_absolute_time {
      start = "2021-02-24T19:18:28+08:00"
      end   = "2021-02-24T19:20:28+08:00"
    }
  }

  name = "random notebook test test"

  status = "published"

  metadata {
    is_template    = false
    take_snapshots = true
    type           = "postmortem"
  }
}
