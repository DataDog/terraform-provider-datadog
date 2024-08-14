# Uploads an OpenAPI file from the given local path to Datadog's API catalog

resource "datadog_openapi_api" "my-api" {
  spec = file("./path/my-api.yaml")
}
