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
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
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
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
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

func TestAccDeploymentGateWithRules(t *testing.T) {
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogDeploymentGateDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDeploymentGateWithRules(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDeploymentGateExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "service", "my-service"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "env", "production"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "rule.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "rule.0.name", "fdd_rule"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "rule.0.type", "faulty_deployment_detection"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "rule.0.dry_run", "false"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "rule.0.options.duration", "1300"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "rule.1.name", "monitor_rule"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "rule.1.type", "monitor"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "rule.1.dry_run", "true"),
					resource.TestCheckResourceAttrSet(
						"datadog_deployment_gate.foo", "rule.0.id"),
					resource.TestCheckResourceAttrSet(
						"datadog_deployment_gate.foo", "rule.1.id"),
				),
			},
		},
	})
}

func testAccCheckDatadogDeploymentGateWithRules(uniq string) string {
	return fmt.Sprintf(`resource "datadog_deployment_gate" "foo" {
    service = "my-service"
    env = "production"
    identifier = "%s"
    dry_run = false

    rule {
        name = "fdd_rule"
        type = "faulty_deployment_detection"
        dry_run = false
        options {
            duration = 1300
        }
    }

    rule {
        name = "monitor_rule"
        type = "monitor"
        dry_run = true
        options {
            duration = 300
            query = "test_query"
        }
    }
}`, uniq)
}

func TestAccDeploymentGateRuleWithExcludedResources(t *testing.T) {
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogDeploymentGateDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDeploymentGateWithExcludedResources(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDeploymentGateExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "rule.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "rule.0.name", "fdd_with_exclusions"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "rule.0.type", "faulty_deployment_detection"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "rule.0.options.excluded_resources.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "rule.0.options.excluded_resources.0", "GET /api/v1/health"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "rule.0.options.excluded_resources.1", "GET /api/v1/status"),
				),
			},
		},
	})
}

func testAccCheckDatadogDeploymentGateWithExcludedResources(uniq string) string {
	return fmt.Sprintf(`resource "datadog_deployment_gate" "foo" {
    service = "my-service"
    env = "production"
    identifier = "%s"
    dry_run = false

    rule {
        name = "fdd_with_exclusions"
        type = "faulty_deployment_detection"
        dry_run = false
        options {
            duration = 1300
            excluded_resources = ["GET /api/v1/health", "GET /api/v1/status"]
        }
    }
}`, uniq)
}

func TestAccDeploymentGateModifyRuleDryRun(t *testing.T) {
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogDeploymentGateDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			// Create gate with rule where dry_run = false
			{
				Config: testAccCheckDatadogDeploymentGateModifyRuleDryRunInitial(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDeploymentGateExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "rule.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "rule.0.dry_run", "false"),
					resource.TestCheckResourceAttrSet(
						"datadog_deployment_gate.foo", "rule.0.id"),
				),
			},
			// Update dry_run to true (should update, not recreate)
			{
				Config: testAccCheckDatadogDeploymentGateModifyRuleDryRunUpdated(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDeploymentGateExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "rule.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "rule.0.dry_run", "true"),
					resource.TestCheckResourceAttrSet(
						"datadog_deployment_gate.foo", "rule.0.id"),
				),
			},
		},
	})
}

func testAccCheckDatadogDeploymentGateModifyRuleDryRunInitial(uniq string) string {
	return fmt.Sprintf(`resource "datadog_deployment_gate" "foo" {
    service = "my-service"
    env = "production"
    identifier = "%s"
    dry_run = false

    rule {
        name = "test_rule"
        type = "faulty_deployment_detection"
        dry_run = false
        options {
            duration = 1300
        }
    }
}`, uniq)
}

func testAccCheckDatadogDeploymentGateModifyRuleDryRunUpdated(uniq string) string {
	return fmt.Sprintf(`resource "datadog_deployment_gate" "foo" {
    service = "my-service"
    env = "production"
    identifier = "%s"
    dry_run = false

    rule {
        name = "test_rule"
        type = "faulty_deployment_detection"
        dry_run = true
        options {
            duration = 1300
        }
    }
}`, uniq)
}

func TestAccDeploymentGateModifyRuleOptions(t *testing.T) {
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogDeploymentGateDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			// Create gate with FDD rule with basic options
			{
				Config: testAccCheckDatadogDeploymentGateModifyRuleOptionsInitial(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDeploymentGateExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "rule.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "rule.0.options.duration", "1300"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "rule.0.options.excluded_resources.#", "0"),
					resource.TestCheckResourceAttrSet(
						"datadog_deployment_gate.foo", "rule.0.id"),
				),
			},
			// Update options: change duration and add excluded_resources
			{
				Config: testAccCheckDatadogDeploymentGateModifyRuleOptionsUpdated(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDeploymentGateExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "rule.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "rule.0.options.duration", "1800"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "rule.0.options.excluded_resources.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "rule.0.options.excluded_resources.0", "GET /health"),
					resource.TestCheckResourceAttrSet(
						"datadog_deployment_gate.foo", "rule.0.id"),
				),
			},
		},
	})
}

func testAccCheckDatadogDeploymentGateModifyRuleOptionsInitial(uniq string) string {
	return fmt.Sprintf(`resource "datadog_deployment_gate" "foo" {
    service = "my-service"
    env = "production"
    identifier = "%s"
    dry_run = false

    rule {
        name = "fdd_rule"
        type = "faulty_deployment_detection"
        dry_run = false
        options {
            duration = 1300
        }
    }
}`, uniq)
}

func testAccCheckDatadogDeploymentGateModifyRuleOptionsUpdated(uniq string) string {
	return fmt.Sprintf(`resource "datadog_deployment_gate" "foo" {
    service = "my-service"
    env = "production"
    identifier = "%s"
    dry_run = false

    rule {
        name = "fdd_rule"
        type = "faulty_deployment_detection"
        dry_run = false
        options {
            duration = 1800
            excluded_resources = ["GET /health"]
        }
    }
}`, uniq)
}

func TestAccDeploymentGateChangeRuleType(t *testing.T) {
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogDeploymentGateDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			// Create gate with FDD rule
			{
				Config: testAccCheckDatadogDeploymentGateChangeRuleTypeInitial(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDeploymentGateExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "rule.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "rule.0.name", "test_rule"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "rule.0.type", "faulty_deployment_detection"),
					resource.TestCheckResourceAttrSet(
						"datadog_deployment_gate.foo", "rule.0.id"),
				),
			},
			// Change rule type from FDD to monitor (should recreate rule)
			{
				Config: testAccCheckDatadogDeploymentGateChangeRuleTypeUpdated(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDeploymentGateExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "rule.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "rule.0.name", "test_rule"),
					resource.TestCheckResourceAttr(
						"datadog_deployment_gate.foo", "rule.0.type", "monitor"),
					resource.TestCheckResourceAttrSet(
						"datadog_deployment_gate.foo", "rule.0.id"),
				),
			},
		},
	})
}

func testAccCheckDatadogDeploymentGateChangeRuleTypeInitial(uniq string) string {
	return fmt.Sprintf(`resource "datadog_deployment_gate" "foo" {
    service = "my-service"
    env = "production"
    identifier = "%s"
    dry_run = false

    rule {
        name = "test_rule"
        type = "faulty_deployment_detection"
        dry_run = false
        options {
            duration = 1300
        }
    }
}`, uniq)
}

func testAccCheckDatadogDeploymentGateChangeRuleTypeUpdated(uniq string) string {
	return fmt.Sprintf(`resource "datadog_deployment_gate" "foo" {
    service = "my-service"
    env = "production"
    identifier = "%s"
    dry_run = false

    rule {
        name = "test_rule"
        type = "monitor"
        dry_run = false
        options {
            duration = 300
            query = "service:test"
        }
    }
}`, uniq)
}
