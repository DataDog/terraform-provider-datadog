package test

import (
	"testing"
)

const datadogPowerpackRun_WorkflowTest = `
resource "datadog_powerpack" "run_workflow_powerpack" {
	name         = "{{uniq}}"
    tags = ["tag:foo1"]
	description   = "Created using the Datadog provider in Terraform"
	template_variables {
		defaults = ["defaults"]
		name     = "datacenter"
	}
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

var datadogPowerpackRun_WorkflowTestAsserts = []string{
	// Powerpack metadata
	"name = {{uniq}}",
	"description = Created using the Datadog provider in Terraform",
	"widget.# = 1",
	"tags.# = 1",
	"tags.0 = tag:foo1",
	// Run Workflow widget
	"widget.0.run_workflow_definition.0.title = My workflow widget",
	"widget.0.run_workflow_definition.0.workflow_id = 2e055f16-8b6a-4cdd-b452-17a34c44b160",
	"widget.0.run_workflow_definition.0.input.0.name = env",
	"widget.0.run_workflow_definition.0.input.0.value = $Env.value",
	"widget.0.run_workflow_definition.0.input.1.name = Foo",
	"widget.0.run_workflow_definition.0.input.1.value = $Env",
	// Template Variables
	"template_variables.# = 1",
	"template_variables.0.name = datacenter",
	"template_variables.0.defaults.# = 1",
	"template_variables.0.defaults.0 = defaults",
}

func TestAccDatadogPowerpackRun_Workflow(t *testing.T) {
	testAccDatadogPowerpackWidgetUtil(t, datadogPowerpackRun_WorkflowTest, "datadog_powerpack.run_workflow_powerpack", datadogPowerpackRun_WorkflowTestAsserts)
}
