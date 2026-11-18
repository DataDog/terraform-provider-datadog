# Record request metadata and tool usage, but no prompt or completion content, for requests no override matches.
resource "datadog_ai_gateway_governance" "this" {
  data_capture_mode = "metadata_only"
  metadata_capture  = ["tool_usage"]
}
