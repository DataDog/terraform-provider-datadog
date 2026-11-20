package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

func TestAccDatadogFlakyTestsManagementPolicies_Basic(t *testing.T) {
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	repo := uniqueEntityName(ctx, t)
	resourceName := "datadog_flaky_tests_management_policies.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogFlakyTestsManagementPoliciesDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogFlakyTestsManagementPoliciesConfig(repo),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogFlakyTestsManagementPoliciesExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "repository_id", repo),
					resource.TestCheckResourceAttr(resourceName, "attempt_to_fix.retries", "5"),
					resource.TestCheckResourceAttr(resourceName, "quarantined.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "quarantined.auto_quarantine_rule.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "quarantined.auto_quarantine_rule.window_seconds", "1209600"),
					resource.TestCheckResourceAttr(resourceName, "quarantined.branch_rule.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "quarantined.branch_rule.branches.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "quarantined.failure_rate_rule.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "quarantined.failure_rate_rule.min_runs", "10"),
					resource.TestCheckResourceAttr(resourceName, "quarantined.failure_rate_rule.threshold", "0.3"),
					resource.TestCheckResourceAttr(resourceName, "disabled.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "disabled.auto_disable_rule.status", "quarantined"),
					resource.TestCheckResourceAttr(resourceName, "disabled.auto_disable_rule.window_seconds", "2592000"),
					resource.TestCheckResourceAttr(resourceName, "disabled.failure_rate_rule.status", "quarantined"),
					resource.TestCheckResourceAttr(resourceName, "disabled.failure_rate_rule.threshold", "0.5"),
				),
			},
			{
				Config: testAccCheckDatadogFlakyTestsManagementPoliciesConfigUpdated(repo),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogFlakyTestsManagementPoliciesExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "attempt_to_fix.retries", "10"),
					resource.TestCheckResourceAttr(resourceName, "quarantined.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "quarantined.failure_rate_rule.threshold", "0.4"),
					resource.TestCheckResourceAttr(resourceName, "disabled.enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDatadogFlakyTestsManagementPoliciesConfig(repo string) string {
	return fmt.Sprintf(`
resource "datadog_flaky_tests_management_policies" "foo" {
  repository_id = "%s"

  attempt_to_fix {
    retries = 5
  }

  quarantined {
    enabled = true

    auto_quarantine_rule {
      enabled        = true
      window_seconds = 1209600
    }

    branch_rule {
      enabled  = true
      branches = ["main"]
    }

    failure_rate_rule {
      enabled   = true
      branches  = ["main"]
      min_runs  = 10
      threshold = 0.3
    }
  }

  disabled {
    enabled = true

    auto_disable_rule {
      enabled        = true
      status         = "quarantined"
      window_seconds = 2592000
    }

    branch_rule {
      enabled  = true
      branches = ["main"]
    }

    failure_rate_rule {
      enabled   = true
      branches  = ["main"]
      min_runs  = 10
      status    = "quarantined"
      threshold = 0.5
    }
  }
}`, repo)
}

func testAccCheckDatadogFlakyTestsManagementPoliciesConfigUpdated(repo string) string {
	return fmt.Sprintf(`
resource "datadog_flaky_tests_management_policies" "foo" {
  repository_id = "%s"

  attempt_to_fix {
    retries = 10
  }

  quarantined {
    enabled = false

    failure_rate_rule {
      enabled   = true
      branches  = ["main"]
      min_runs  = 10
      threshold = 0.4
    }
  }

  disabled {
    enabled = false
  }
}`, repo)
}

func testAccCheckDatadogFlakyTestsManagementPoliciesExists(accProvider *fwprovider.FrameworkProvider, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource '%s' not found in the state", resourceName)
		}
		if res.Type != "datadog_flaky_tests_management_policies" {
			return fmt.Errorf("resource %s is not of type datadog_flaky_tests_management_policies, found %s instead", resourceName, res.Type)
		}

		getRequest := datadogV2.NewTestOptimizationFlakyTestsManagementPoliciesGetRequest(
			*datadogV2.NewTestOptimizationFlakyTestsManagementPoliciesGetRequestData(
				*datadogV2.NewTestOptimizationFlakyTestsManagementPoliciesGetRequestAttributes(res.Primary.ID),
				datadogV2.TESTOPTIMIZATIONGETFLAKYTESTSMANAGEMENTPOLICIESREQUESTDATATYPE_TEST_OPTIMIZATION_GET_FLAKY_TESTS_MANAGEMENT_POLICIES_REQUEST,
			),
		)
		_, _, err := accProvider.DatadogApiInstances.GetTestOptimizationApiV2().GetFlakyTestsManagementPolicies(accProvider.Auth, *getRequest)
		if err != nil {
			return fmt.Errorf("received an error retrieving flaky tests management policies: %s", err)
		}
		return nil
	}
}

func testAccCheckDatadogFlakyTestsManagementPoliciesDestroy(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// The API has no delete endpoint; policies persist after the resource is
		// removed from state, so there is nothing to check here.
		return nil
	}
}
