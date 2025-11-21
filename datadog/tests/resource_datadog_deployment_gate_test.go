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

func TestAccDeploymentGateBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogDeploymentGateDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDeploymentGate(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDeploymentGateExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "dry_run", "false"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "env", "production"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "identifier", uniq),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "service", "my-service"),
				),
			},
			{
				Config: testAccCheckDatadogDeploymentGateUpdated(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDeploymentGateExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "dry_run", "true"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "env", "production"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "identifier", uniq),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "service", "my-service"),
				),
			},
		},
	})
}

func testAccCheckDatadogDeploymentGate(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`resource "datadog_deployment_gate" "foo" {
    dry_run = "false"
    env = "production"
    identifier = "%s"
    service = "my-service"
}`, uniq)
}

func testAccCheckDatadogDeploymentGateUpdated(uniq string) string {
	return fmt.Sprintf(`resource "datadog_deployment_gate" "foo" {
    dry_run = "true"
    env = "production"
    identifier = "%s"
    service = "my-service"
}`, uniq)
}

func TestAccDeploymentGateForceNew(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogDeploymentGateDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDeploymentGateForceNew(uniq, "production", "my-service"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDeploymentGateExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "env", "production"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "service", "my-service"),
				),
			},
			{
				Config: testAccCheckDatadogDeploymentGateForceNew(uniq, "staging", "my-service"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDeploymentGateExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "env", "staging"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "service", "my-service"),
				),
			},
			{
				Config: testAccCheckDatadogDeploymentGateForceNew(uniq, "staging", "updated-service"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDeploymentGateExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "env", "staging"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "service", "updated-service"),
				),
			},
		},
	})
}

func testAccCheckDatadogDeploymentGateForceNew(uniq string, env string, service string) string {
	return fmt.Sprintf(`resource "datadog_deployment_gate" "foo" {
    dry_run = "false"
    env = "%s"
    identifier = "%s"
    service = "%s"
}`, env, uniq, service)
}

func testAccCheckDatadogDeploymentGateDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := DeploymentGateDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func DeploymentGateDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_deployment_gate" {
				continue
			}
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetDeploymentGatesApiV2().GetDeploymentGate(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving DeploymentGate %s", err)}
			}
			return &utils.RetryableError{Prob: "DeploymentGate still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogDeploymentGateExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := deploymentGateExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func deploymentGateExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_deployment_gate" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetDeploymentGatesApiV2().GetDeploymentGate(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving DeploymentGate")
		}
	}
	return nil
}
