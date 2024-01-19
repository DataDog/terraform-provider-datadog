# Create new openapi_api resource

resource "datadog_openapi_api" "my-api" {
  spec = file("./path/my-api.yaml")
}
