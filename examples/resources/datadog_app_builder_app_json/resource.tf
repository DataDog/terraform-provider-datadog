variable "connection_id" {
  type    = string
  default = "11111111-2222-3333-4444-555555555555"
}

# Example App Builder App resource - loaded from a file (required for ${} syntax, else use $${} inline for escaping)
resource "datadog_app_builder_app" "example_app_from_file" {
  app_json = file("${path.module}/resource.json")
  override_action_query_names_to_connection_ids = {
    "listTeams0" = var.connection_id
  }
}

# Example App Builder App resource - inline with optional fields
resource "datadog_app_builder_app" "example_app_inline_optional" {
  name               = "Example Terraform App - Inline - Optional"
  description        = "Created using the Datadog provider in Terraform."
  root_instance_name = "grid0"
  published          = false

  override_action_query_names_to_connection_ids = {
    "listTeams0" = var.connection_id
  }

  app_json = jsonencode(
    {
      "queries" : [
        {
          "id" : "11111111-1111-1111-1111-111111111111",
          "type" : "action",
          "name" : "listTeams0",
          "events" : [],
          "properties" : {
            "spec" : {
              "fqn" : "com.datadoghq.dd.teams.listTeams",
              "connectionId" : "11111111-1111-1111-1111-111111111111",
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
      "tags" : [],
      "connections" : [
        {
          "id" : "11111111-1111-1111-1111-111111111111",
          "name" : "A connection that will be overridden"
        }
      ],
      "deployment" : {
        "id" : "11111111-1111-1111-1111-111111111111",
        "app_version_id" : "00000000-0000-0000-0000-000000000000"
      }
    }
  )
}

variable "name" {
  type    = string
  default = "Example Terraform App - Inline - Escaped - Interpolated"
}

variable "description" {
  type    = string
  default = "Created using the Datadog provider in Terraform. Variable interpolation."
}

variable "root_instance_name" {
  type    = string
  default = "grid0"
}

# Example App Builder App resource - inline with escaped $${} syntax for Javascript and ${} for Terraform variables
resource "datadog_app_builder_app" "example_app_inline_escaped_interpolated" {
  app_json = jsonencode(
    {
      "queries" : [
        {
          "id" : "11111111-1111-1111-1111-111111111111",
          "name" : "listTeams0",
          "type" : "action",
          "properties" : {
            "onlyTriggerManually" : false,
            "outputs" : "$${((outputs) => {// Use 'outputs' to reference the query's unformatted output.\n\n// TODO: Apply transformations to the raw query output\n\nreturn outputs.data.map(item => item.attributes.name);})(self.rawOutputs)}",
            "spec" : {
              "fqn" : "com.datadoghq.dd.teams.listTeams",
              "inputs" : {},
              "connectionId" : "${var.connection_id}"
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
      "description" : "${var.description}",
      "favorite" : false,
      "name" : "${var.name}",
      "rootInstanceName" : "${var.root_instance_name}",
      "selfService" : false,
      "tags" : [],
      "connections" : [
        {
          "id" : "${var.connection_id}",
          "name" : "A connection that will be overridden"
        }
      ],
      "deployment" : {
        "id" : "11111111-1111-1111-1111-111111111111",
        "app_version_id" : "00000000-0000-0000-0000-000000000000"
      }
    }
  )
}
