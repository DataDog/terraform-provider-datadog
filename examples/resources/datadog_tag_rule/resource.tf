# Create a surfacing tag rule, which highlights non-compliant telemetry
# without rejecting it.
resource "datadog_tag_rule" "env" {
  name               = "env must be a known environment"
  source             = "logs"
  scope              = "org"
  tag_key            = "env"
  tag_value_patterns = ["prod", "staging", "dev"]
  rule_type          = "surfacing"
  required           = true
}

# Create a blocking tag rule, which rejects telemetry that violates it.
# The Datadog API only accepts `surfacing` at creation time, so the provider
# creates the rule and then immediately updates it to `blocking`. Because that
# makes the create non-atomic, it has to be opted into explicitly.
resource "datadog_tag_rule" "service" {
  name               = "service must be lowercase and hyphenated"
  source             = "spans"
  scope              = "org"
  tag_key            = "service"
  tag_value_patterns = ["^[a-z0-9-]+$"]
  rule_type          = "blocking"

  force_blocking_on_create = true

  # Keep the rule recoverable and preserve its historical compliance score
  # data when this resource is destroyed.
  hard_delete = false
}
