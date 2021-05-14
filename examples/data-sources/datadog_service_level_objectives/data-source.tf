data "datadog_service_level_objectives" "ft_foo_slos" {
  tags_query = "owner:ft-foo"
}