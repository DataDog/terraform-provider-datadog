resource "datadog_monitor" "watchdog_monitor" {
  name               = "Watchdog Monitor TF"
  type               = "event alert"
  message            = "Check out this Watchdog Service Monitor"
  escalation_message = "Escalation message"

  query = "events('priority:all sources:watchdog tags:story_type:service,env:test_env,service:test_service:_aggregate').by('service,resource_name').rollup('count').last('30m') > 0"

  include_tags = true

  tags = ["foo:bar", "baz"]
}
