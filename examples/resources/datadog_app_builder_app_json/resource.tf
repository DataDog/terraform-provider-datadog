# Example app JSON - loaded from a file (required for ${} syntax)
resource "datadog_app_builder_app_json" "example_app_from_file" {
  app_json = file("${path.module}/resource.json")
  override_action_query_names_to_connection_ids = {
    "listTeams0" = datadog_action_connection.example_connection.id
  }
}

# Example app JSON - inline basic with optional fields
resource "datadog_app_builder_app_json" "example_app_inline_basic_with_optional_fields" {
  name                  = "Example Terraform App - Basic"
  description           = "Created using the Datadog provider in Terraform."
  root_instance_name    = "grid0"
  tags                  = ["team:app-builder", "service:app-builder-api", "tag_key:tag_value", "terraform_app"]
  publish_status_update = "publish"

  override_action_query_names_to_connection_ids = {
    "listTeams0" = datadog_action_connection.example_connection.id
  }

  app_json = jsonencode(
    {
      "queries" : [
        {
          "id" : "22a65cea-b96f-44fa-a1e2-008885d7bd2e",
          "type" : "action",
          "name" : "listTeams0",
          "events" : [],
          "properties" : {
            "spec" : {
              "fqn" : "com.datadoghq.dd.teams.listTeams",
              "connectionId" : "11111111-2222-3333-4444-555555555555",
              "inputs" : {}
            }
          }
        }
      ],
      "id" : "11111111-2222-3333-4444-555555555555",
      "components" : [
        {
          "events" : [],
          "name" : "grid0",
          "properties" : {
            "children" : []
          },
          "type" : "grid"
        }
      ],
      "description" : "This description will be overridden",
      "favorite" : false,
      "name" : "This name will be overridden",
      "rootInstanceName" : "This rootInstanceName will be overridden",
      "selfService" : false,
      "tags" : ["these_tags_will_be:overridden", "foo:bar"],
      "connections" : [
        {
          "id" : "11111111-2222-3333-4444-555555555555",
          "name" : "A connection that will be overridden"
        }
      ]
    }
  )
}
