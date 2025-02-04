# Example app json - loaded from a file (necessary if includes ${} syntax)
resource "datadog_app" "example_app_from_file" {
  app_json = file("${path.module}/resource.json")
}

# Example app json - inline basic
resource "datadog_app" "example_app_inline_basic" {
  app_json = jsonencode(
    {
      "queries" : [],
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
      "description" : "Created using the Datadog provider in Terraform.",
      "name" : "Example Terraform App - Simple",
      "rootInstanceName" : "grid0",
      "tags" : [],
    }
  )
}
