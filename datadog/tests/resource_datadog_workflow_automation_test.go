package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

// Vars shared for data source and resource tests
var (
	testWorkflowDescription = "My description."
	testWorkflowTags        = "[\"service:foo\", \"team:bar\", \"foo:bar\"]"
	testWorkflowSpec        = `{
	"steps": [],
	"triggers": [
		{
			"startStepNames": [],
			"workflowTrigger": {}
		}
	]
}`
	testInvalidWorkflowSpec = `{
	"foo": "bar",
	"steps": [],
	"triggers": [
		{
			"startStepNames": [],
			"workflowTrigger": {}
		}
	]
}`
)

func TestAccDatadogWorkflowAutomationResource(t *testing.T) {
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	workflowName := uniqueEntityName(ctx, t)
	resourceName := "datadog_workflow_automation.my_workflow"

	// Simulates the behavior of the `jsonencode` function in Terraform which will store the json_spec with no whitespace
	r, err := regexp.Compile(`\s+`)
	if err != nil {
		t.Error("Unexpected error compiling regex")
	}
	testWorkflowEmptySpecNoWhitespace := r.ReplaceAllString(testWorkflowSpec, "")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogWorkflowDestroy(providers.frameworkProvider, resourceName),
		Steps: []resource.TestStep{
			{
				Config: testWorkflowAutomationResourceConfig(workflowName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogWorkflowExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", workflowName),
					resource.TestCheckResourceAttr(resourceName, "description", testWorkflowDescription),
					resource.TestCheckTypeSetElemAttr(resourceName, "tags.*", "service:foo"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tags.*", "team:bar"),
					resource.TestCheckTypeSetElemAttr(resourceName, "tags.*", "foo:bar"),
					resource.TestCheckResourceAttr(resourceName, "published", "false"),
					resource.TestCheckResourceAttr(resourceName, "spec_json", testWorkflowEmptySpecNoWhitespace),
				),
			},
			{
				Config:      testInvalidWorkflowAutomationResourceConfig(workflowName),
				ExpectError: regexp.MustCompile("Error running apply"),
			},
		},
	})
}

// testWorkflowAutomationResourceConfig shared for data source and resource tests
func testWorkflowAutomationResourceConfig(workflowName string) string {
	return fmt.Sprintf(`
	resource "datadog_workflow_automation" "my_workflow" {
		name        = "%s"
		description = "%s"
		tags        = %s
		published   = false

		spec_json = jsonencode(
%s
		)
	}`, workflowName, testWorkflowDescription, testWorkflowTags, testWorkflowSpec)
}

func testInvalidWorkflowAutomationResourceConfig(workflowName string) string {
	return fmt.Sprintf(`
	resource "datadog_workflow_automation" "invalid_workflow" {
		name        = "%s"
		description = "%s"
		tags        = %s
		published   = false
		spec_json = jsonencode(
%s
		)
	}`, workflowName, testWorkflowDescription, testWorkflowTags, testInvalidWorkflowSpec)
}

func testAccCheckDatadogWorkflowExists(accProvider *fwprovider.FrameworkProvider, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := datadogWorkflowExistsHelper(auth, s, apiInstances, n); err != nil {
			return err
		}
		return nil
	}
}

func datadogWorkflowExistsHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances, name string) error {
	id := s.RootModule().Resources[name].Primary.ID
	if _, _, err := apiInstances.GetWorkflowAutomationApiV2().GetWorkflow(ctx, id); err != nil {
		return fmt.Errorf("received an error retrieving workflow: %s", err)
	}
	return nil
}
