resource "datadog_monitor" "watchdog_monitor" {
  name               = "Watchdog Monitor TF"
  type               = "event-v2 alert"
  message            = "Check out this Watchdog Service Monitor"
  escalation_message = "Escalation message"

  query = "events(\"priority:all sources:watchdog tags:(story_type:service env:test_env service:testservice_aggregate)\").rollup(\"count\").by(\"service,resource_name\").last(\"30m\") > 0"

  include_tags = true

  tags = ["foo:bar", "baz"]
}