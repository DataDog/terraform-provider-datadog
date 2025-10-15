package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccDeploymentRuleBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogDeploymentRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDeploymentRule(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDeploymentRuleExists(providers.frameworkProvider),

					resource.TestCheckResourceAttr(
						"datadog_deployment_rule.foo", "dry_run", "UPDATE ME"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_rule.foo", "name", "My deployment rule"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_rule.foo", "type", "faulty_deployment_detection"),
				),
			},
		},
	})
}

func testAccCheckDatadogDeploymentRule(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`resource "datadog_deployment_rule" "foo" {
    gate_id = "UPDATE ME"
    dry_run = "UPDATE ME"
    name = "My deployment rule"
    type = "faulty_deployment_detection"
}`, uniq)
}

func testAccCheckDatadogDeploymentRuleDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := DeploymentRuleDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func DeploymentRuleDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_deployment_rule" {
				continue
			}
			gateId := r.Primary.Attributes["gate_id"]
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetDeploymentGatesApiV2().GetDeploymentRule(auth, gateId, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving DeploymentRule %s", err)}
			}
			return &utils.RetryableError{Prob: "DeploymentRule still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogDeploymentRuleExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := deploymentRuleExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func deploymentRuleExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_deployment_rule" {
			continue
		}
		gateId := r.Primary.Attributes["gate_id"]
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetDeploymentGatesApiV2().GetDeploymentRule(auth, gateId, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving DeploymentRule")
		}
	}
	return nil
}
