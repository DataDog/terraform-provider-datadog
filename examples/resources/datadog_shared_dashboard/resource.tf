
# Shared dashboard open to all users
resource "datadog_shared_dashboard" "open_dashboard" {
  dashboard_id   = "123-abc-456"
  dashboard_type = "custom_timeboard"
  global_time {
    live_span = "1h"
  }
  global_time_selectable_enabled = true
  selectable_template_vars {
    default_value = "value1"
    name          = "datacenter"
    prefix        = "datacenter"
    visible_tags  = ["value1", "value2", "value3"]
  }
  share_type = "open"
}

# Shared dashboard accessible only to a limited list of users
resource "datadog_shared_dashboard" "invite_only" {
  dashboard_id   = "123-abc-456"
  dashboard_type = "custom_timeboard"
  global_time {
    live_span = "1h"
  }
  global_time_selectable_enabled = true
  selectable_template_vars {
    default_value = "value1"
    name          = "datacenter"
    prefix        = "datacenter"
    visible_tags  = ["value1", "value2", "value3"]
  }
  share_list = ["account1@org.com", "account2@org.com"]
  share_type = "invite"
}