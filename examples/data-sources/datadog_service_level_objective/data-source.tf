data "datadog_service_level_objective" "test" {
  name_query = "My test SLO"
  tags_query = "foo:bar"
}

data "datadog_service_level_objective" "api_slo" {
  id = data.terraform_remote_state.api.outputs.slo
}