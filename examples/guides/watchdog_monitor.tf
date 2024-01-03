resource "datadog_monitor" "watchdog_monitor" {
  name    = "Watchdog detected an anomaly: {{event.title}}"
  type    = "event-v2 alert"
  message = "Watchdog monitor created from Terraform"

  query = "events(\"source:watchdog story_category:apm env:test_env\").rollup(\"count\").by(\"story_key,service,resource_name\").last(\"30m\") > 0"

  tags = ["terraform:true"]
}
