# Record only metadata for the coding assistant's requests, whatever the organization default is.
resource "datadog_ai_gateway_data_capture_override" "coding_assistant" {
  name             = "coding-assistant-metadata-only"
  mode             = "metadata_only"
  metadata_capture = ["tool_usage", "tool_names"]
  tags = {
    ml_app = "coding-assistant"
  }
}
