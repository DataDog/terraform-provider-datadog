package test

import "testing"

const datadogDashboardV2EmbeddedAppConfig = `
resource "datadog_dashboard_v2" "embedded_app_dashboard" {
  title       = "{{uniq}}"
  layout_type = "ordered"

  widget {
    embedded_app_definition {
      app_id = "7e7745f9-4343-4927-b038-80934a355915"
      title  = "Environment manager"

      input {
        name  = "environment"
        value = jsonencode("production")
      }

      input {
        name = "settings"
        value = jsonencode({
          dry_run  = false
          replicas = 3
        })
      }
    }
  }
}
`

var datadogDashboardV2EmbeddedAppAsserts = []string{
	"title = {{uniq}}",
	"widget.0.embedded_app_definition.0.app_id = 7e7745f9-4343-4927-b038-80934a355915",
	"widget.0.embedded_app_definition.0.title = Environment manager",
	"widget.0.embedded_app_definition.0.input.0.name = environment",
	`widget.0.embedded_app_definition.0.input.0.value = "production"`,
	"widget.0.embedded_app_definition.0.input.1.name = settings",
	`widget.0.embedded_app_definition.0.input.1.value = {"dry_run":false,"replicas":3}`,
}

func TestAccDatadogDashboardV2EmbeddedApp(t *testing.T) {
	config, name := datadogDashboardV2EmbeddedAppConfig, "datadog_dashboard_v2.embedded_app_dashboard"
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardV2EmbeddedApp", config, name, datadogDashboardV2EmbeddedAppAsserts)
}

func TestAccDatadogDashboardV2EmbeddedApp_import(t *testing.T) {
	config, name := datadogDashboardV2EmbeddedAppConfig, "datadog_dashboard_v2.embedded_app_dashboard"
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardV2EmbeddedApp_import", config, name)
}
