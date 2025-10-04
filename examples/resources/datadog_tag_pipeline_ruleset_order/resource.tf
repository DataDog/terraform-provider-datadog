# Create new tag_pipeline_ruleset_order resource

resource "datadog_tag_pipeline_ruleset_order" "my_ruleset_order" {
  ruleset_ids = [
    "ruleset-id-1",
    "ruleset-id-2",
    "ruleset-id-3"
  ]
}
