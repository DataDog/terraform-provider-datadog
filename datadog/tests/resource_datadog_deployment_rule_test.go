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
						"datadog_deployment_rule.foo", "dry_run", "false"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_rule.foo", "name", "My deployment rule"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_rule.foo", "type", "faulty_deployment_detection"),
					resource.TestCheckResourceAttr("datadog_deployment_rule.foo", "options.duration", "10"),
					resource.TestCheckResourceAttr("datadog_deployment_rule.foo", "options.excluded_resources.0", "resource1"),
				),
			},
			{
				Config: testAccCheckDatadogDeploymentRuleUpdated(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDeploymentRuleExists(providers.frameworkProvider),

					resource.TestCheckResourceAttr(
						"datadog_deployment_rule.foo", "dry_run", "true"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_rule.foo", "name", "Updated deployment rule"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_rule.foo", "type", "faulty_deployment_detection"),
					resource.TestCheckResourceAttr("datadog_deployment_rule.foo", "options.duration", "15"),
					resource.TestCheckResourceAttr("datadog_deployment_rule.foo", "options.excluded_resources.0", "resource2"),
					resource.TestCheckResourceAttr("datadog_deployment_rule.foo", "options.excluded_resources.1", "resource3"),
				),
			},
		},
	})
}

func testAccCheckDatadogDeploymentRule(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_deployment_gate" "test_gate" {
	service = "test-service"
	env = "prod"
	identifier = "%s"
}

resource "datadog_deployment_rule" "foo" {
    gate_id = datadog_deployment_gate.test_gate.id
    dry_run = "false"
    name = "My deployment rule"
    type = "faulty_deployment_detection"
    options {
        duration = 10
        excluded_resources = ["resource1"]
    }
}`, uniq)
}

func TestAccDeploymentRuleTypeForceNew(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogDeploymentRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDeploymentRuleTypeForceNew(uniq, "faulty_deployment_detection"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDeploymentRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_deployment_rule.foo", "type", "faulty_deployment_detection"),
				),
			},
			{
				Config: testAccCheckDatadogDeploymentRuleTypeForceNew(uniq, "monitor"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDeploymentRuleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_deployment_rule.foo", "type", "monitor"),
				),
			},
		},
	})
}

func testAccCheckDatadogDeploymentRuleTypeForceNew(uniq string, ruleType string) string {
	if ruleType == "monitor" {
		return fmt.Sprintf(`
resource "datadog_deployment_gate" "test_gate" {
	service = "test-service"
	env = "prod"
	identifier = "%s"
}

resource "datadog_deployment_rule" "foo" {
    gate_id = datadog_deployment_gate.test_gate.id
    dry_run = "false"
    name = "My deployment rule"
    type = "monitor"
    options {
        query = "service:test-service"
    }
}`, uniq)
	}
	return fmt.Sprintf(`
resource "datadog_deployment_gate" "test_gate" {
	service = "test-service"
	env = "prod"
	identifier = "%s"
}

resource "datadog_deployment_rule" "foo" {
    gate_id = datadog_deployment_gate.test_gate.id
    dry_run = "false"
    name = "My deployment rule"
    type = "faulty_deployment_detection"
    options {
        duration = 10
        excluded_resources = ["resource1"]
    }
}`, uniq)
}

func testAccCheckDatadogDeploymentRuleUpdated(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_deployment_gate" "test_gate" {
	service = "test-service"
	env = "prod"
	identifier = "%s"
}

resource "datadog_deployment_rule" "foo" {
    gate_id = datadog_deployment_gate.test_gate.id
    dry_run = "true"
    name = "Updated deployment rule"
    type = "faulty_deployment_detection"
    options {
        duration = 15
        excluded_resources = ["resource2", "resource3"]
    }
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
