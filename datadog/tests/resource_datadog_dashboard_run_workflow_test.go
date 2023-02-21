package test

import (
	"testing"
)

const datadogDashboardRunWorkflowConfig = `
resource "datadog_dashboard" "run_workflow_dashboard" {
  title        = "{{uniq}}"
  description  = "Created using the Datadog provider in Terraform"
  layout_type  = "ordered"
  reflow_type  = "auto"
  is_read_only = true
  widget {
    run_workflow_definition {
      title       = "My workflow widget"
      title_size  = "13"
      title_align = "left"
      workflow_id = "2e055f16-8b6a-4cdd-b452-17a34c44b160"
      input {
        name  = "env"
        value = "$Env.value"
      }
      input {
        name  = "Foo"
        value = "$Env"
      }
    }
  }
}
`

var datadogDashboardRunWorkflowAsserts = []string{
	"title = {{uniq}}",
	"description = Created using the Datadog provider in Terraform",
	"layout_type = ordered",
	"reflow_type = auto",
	"is_read_only = true",
	"widget.0.run_workflow_definition.0.title = My workflow widget",
	"widget.0.run_workflow_definition.0.workflow_id = 2e055f16-8b6a-4cdd-b452-17a34c44b160",
	"widget.0.run_workflow_definition.0.input.0.name = env",
	"widget.0.run_workflow_definition.0.input.0.value = $Env.value",
	"widget.0.run_workflow_definition.0.input.1.name = Foo",
	"widget.0.run_workflow_definition.0.input.1.value = $Env",
}

func TestAccDatadogDashboardRunWorkflow(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardRunWorkflowConfig, "datadog_dashboard.run_workflow_dashboard", datadogDashboardRunWorkflowAsserts)
}

func TestAccDatadogDashboardRunWorkflow_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardRunWorkflowConfig, "datadog_dashboard.run_workflow_dashboard")
}
