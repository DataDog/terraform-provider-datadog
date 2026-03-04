package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

func TestAccDatadogWorkflowAutomationDatasource(t *testing.T) {
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	workflowName := uniqueEntityName(ctx, t)

	// Simulates the behavior of the `jsonencode` function in Terraform which will store the json_spec with no whitespace
	r, err := regexp.Compile(`\s+`)
	if err != nil {
		t.Error("Unexpected error compiling regex")
	}
	testWorkflowEmptySpecNoWhitespace := r.ReplaceAllString(testWorkflowSpec, "")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogWorkflowDestroy(providers.frameworkProvider, "datadog_workflow_automation.my_workflow"),
		Steps: []resource.TestStep{
			{
				Config: testWorkflowAutomationDataSourceConfig(workflowName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.datadog_workflow_automation.my_workflow", "id"),
					resource.TestCheckResourceAttr("data.datadog_workflow_automation.my_workflow", "name", workflowName),
					resource.TestCheckResourceAttr("data.datadog_workflow_automation.my_workflow", "description", testWorkflowDescription),
					resource.TestCheckTypeSetElemAttr("data.datadog_workflow_automation.my_workflow", "tags.*", "service:foo"),
					resource.TestCheckTypeSetElemAttr("data.datadog_workflow_automation.my_workflow", "tags.*", "team:bar"),
					resource.TestCheckTypeSetElemAttr("data.datadog_workflow_automation.my_workflow", "tags.*", "foo:bar"),
					resource.TestCheckResourceAttr("data.datadog_workflow_automation.my_workflow", "published", "false"),
					resource.TestCheckResourceAttr("data.datadog_workflow_automation.my_workflow", "spec_json", testWorkflowEmptySpecNoWhitespace),
				),
			},
		},
	})
}

func testWorkflowAutomationDataSourceConfig(name string) string {
	return fmt.Sprintf(`
	%s
	data "datadog_workflow_automation" "my_workflow" {
		id = datadog_workflow_automation.my_workflow.id
		depends_on = [datadog_workflow_automation.my_workflow]
	}`, testWorkflowAutomationResourceConfig(name))
}

func testAccCheckDatadogWorkflowDestroy(accProvider *fwprovider.FrameworkProvider, resourceName string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		resource := s.RootModule().Resources[resourceName]
		_, httpRes, err := apiInstances.GetWorkflowAutomationApiV2().GetWorkflow(auth, resource.Primary.ID)
		if err != nil {
			if httpRes.StatusCode == 404 {
				return nil
			}
			return err
		}

		return fmt.Errorf("workflow destroy check failed")
	}
}
