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

func TestAccDatadogSyntheticsSuite_importBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	suiteName := uniqueEntityName(ctx, t)
	frameworkProvider := providers.frameworkProvider

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testSyntheticsSuiteIsDestroyed(frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: createSyntheticsSuiteConfig(suiteName),
			},
			{
				ResourceName:      "datadog_synthetics_suite.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDatadogSyntheticsSuite_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	frameworkProvider := providers.frameworkProvider

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testSyntheticsSuiteIsDestroyed(frameworkProvider),
		Steps: []resource.TestStep{
			createSyntheticsSuiteStep(ctx, frameworkProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsSuite_Updated(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	frameworkProvider := providers.frameworkProvider

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testSyntheticsSuiteIsDestroyed(frameworkProvider),
		Steps: []resource.TestStep{
			createSyntheticsSuiteStep(ctx, frameworkProvider, t),
			updateSyntheticsSuiteStep(ctx, frameworkProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsSuite_WithTests(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	frameworkProvider := providers.frameworkProvider

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testSyntheticsSuiteIsDestroyed(frameworkProvider),
		Steps: []resource.TestStep{
			createSyntheticsSuiteWithTestsStep(ctx, frameworkProvider, t),
		},
	})
}

func createSyntheticsSuiteStep(ctx context.Context, accProvider *fwprovider.FrameworkProvider, t *testing.T) resource.TestStep {
	suiteName := uniqueEntityName(ctx, t)
	return resource.TestStep{
		Config: createSyntheticsSuiteConfig(suiteName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsSuiteExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_suite.foo", "name", suiteName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_suite.foo", "message", "This is a test suite"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_suite.foo", "tags.#", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_suite.foo", "tags.0", "env:test"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_suite.foo", "tags.1", "team:synthetics"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_suite.foo", "options.0.alerting_threshold", "0.5"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_suite.foo", "tests.#", "0"),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_suite.foo", "public_id"),
		),
	}
}

func createSyntheticsSuiteConfig(uniqSuiteName string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_suite" "foo" {
	name    = "%s"
	message = "This is a test suite"
	tags    = ["env:test", "team:synthetics"]

	options {
		alerting_threshold = 0.5
	}
}`, uniqSuiteName)
}

func createSyntheticsSuiteConfigMinimal(uniqSuiteName string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_suite" "foo" {
	name    = "%s"
	message = "This is a test suite"
}`, uniqSuiteName)
}

func updateSyntheticsSuiteStep(ctx context.Context, accProvider *fwprovider.FrameworkProvider, t *testing.T) resource.TestStep {
	suiteName := uniqueEntityName(ctx, t) + "_updated"
	return resource.TestStep{
		Config: updateSyntheticsSuiteConfig(suiteName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsSuiteExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_suite.foo", "name", suiteName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_suite.foo", "message", "Updated test suite"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_suite.foo", "tags.#", "3"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_suite.foo", "tags.0", "env:prod"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_suite.foo", "tags.1", "team:synthetics"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_suite.foo", "tags.2", "updated:true"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_suite.foo", "options.0.alerting_threshold", "0.7"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_suite.foo", "tests.#", "0"),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_suite.foo", "public_id"),
		),
	}
}

func updateSyntheticsSuiteConfig(uniqSuiteName string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_suite" "foo" {
	name    = "%s"
	message = "Updated test suite"
	tags    = ["env:prod", "team:synthetics", "updated:true"]

	options {
		alerting_threshold = 0.7
	}
}`, uniqSuiteName)
}

func createSyntheticsSuiteWithTestsStep(ctx context.Context, accProvider *fwprovider.FrameworkProvider, t *testing.T) resource.TestStep {
	suiteName := uniqueEntityName(ctx, t)
	return resource.TestStep{
		Config: createSyntheticsSuiteWithTestsConfig(suiteName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsSuiteExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_suite.bar", "name", suiteName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_suite.bar", "message", "Suite with tests"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_suite.bar", "options.0.alerting_threshold", "0.6"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_suite.bar", "tests.#", "2"),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_suite.bar", "tests.0.public_id"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_suite.bar", "tests.0.alerting_criticality", "critical"),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_suite.bar", "tests.1.public_id"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_suite.bar", "tests.1.alerting_criticality", "ignore"),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_suite.bar", "public_id"),
		),
	}
}

func createSyntheticsSuiteWithTestsConfig(uniqSuiteName string) string {
	// Note: This config creates test resources first, then references them in the suite
	return fmt.Sprintf(`
resource "datadog_synthetics_test" "test1" {
	type    = "api"
	subtype = "http"
	request_definition {
		method = "GET"
		url    = "https://www.example.com"
	}
	assertion {
		type     = "statusCode"
		operator = "is"
		target   = "200"
	}
	locations = ["aws:us-east-2"]
	options_list {
		tick_every = 900
	}
	name    = "%[1]s-test1"
	message = "Test 1"
	tags    = ["test:1"]
	status  = "paused"
}

resource "datadog_synthetics_test" "test2" {
	type    = "api"
	subtype = "http"
	request_definition {
		method = "GET"
		url    = "https://www.example.org"
	}
	assertion {
		type     = "statusCode"
		operator = "is"
		target   = "200"
	}
	locations = ["aws:us-east-2"]
	options_list {
		tick_every = 900
	}
	name    = "%[1]s-test2"
	message = "Test 2"
	tags    = ["test:2"]
	status  = "paused"
}

resource "datadog_synthetics_suite" "bar" {
	name    = "%[1]s"
	message = "Suite with tests"
	tags    = ["env:test"]

	options {
		alerting_threshold = 0.6
	}

	tests {
		public_id            = datadog_synthetics_test.test1.id
		alerting_criticality = "critical"
	}

	tests {
		public_id            = datadog_synthetics_test.test2.id
		alerting_criticality = "ignore"
	}
}`, uniqSuiteName)
}

func testSyntheticsSuiteExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_synthetics_suite" {
				continue
			}

			if _, _, err := apiInstances.GetSyntheticsApiV2().GetSyntheticsSuite(auth, r.Primary.ID); err != nil {
				return fmt.Errorf("received an error retrieving synthetics suite %s", err)
			}
		}
		return nil
	}
}

func testSyntheticsSuiteIsDestroyed(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := SyntheticsSuiteDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func SyntheticsSuiteDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_synthetics_suite" {
			continue
		}
		err := utils.Retry(2, 10, func() error {
			_, httpresp, err := apiInstances.GetSyntheticsApiV2().GetSyntheticsSuite(auth, r.Primary.ID)
			if err != nil {
				if httpresp != nil && httpresp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("error retrieving synthetics suite: %s", err)}
			}
			return &utils.RetryableError{Prob: "synthetics suite still exists"}
		})

		if err != nil {
			return err
		}
	}
	return nil
}
