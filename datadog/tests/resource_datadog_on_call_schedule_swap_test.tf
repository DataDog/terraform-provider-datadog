resource "datadog_on_call_schedule" "swap" {
  name      = "SWAP_SCHEDULE_NAME"
  time_zone = "UTC"
  layer {
    effective_date = "2025-01-01T00:00:00Z"
    interval {
      days = 1
    }
    rotation_start = "2025-01-01T00:00:00Z"
    users          = [null]
    name           = "LAYER_A_NAME"
  }
  layer {
    effective_date = "2025-01-01T00:00:00Z"
    interval {
      days = 2
    }
    rotation_start = "2025-01-01T00:00:00Z"
    users          = [null]
    name           = "LAYER_B_NAME"
  }
}
