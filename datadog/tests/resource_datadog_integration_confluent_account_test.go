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

func TestAccIntegrationConfluentAccountBasic(t *testing.T) {
	// t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIntegrationConfluentAccountDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationConfluentAccountBasic(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationConfluentAccountExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_confluent_account.foo", "api_key", uniq),
					resource.TestCheckResourceAttr(
						"datadog_integration_confluent_account.foo", "api_secret", "test-api-secret-123"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_integration_confluent_account.foo", "tags.*", "foo:bar"),
				),
			},
		},
	})
}

func TestAccIntegrationConfluentAccountUpdated(t *testing.T) {
	// t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIntegrationConfluentAccountDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationConfluentAccountBasic(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationConfluentAccountExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_confluent_account.foo", "api_key", uniq),
					resource.TestCheckResourceAttr(
						"datadog_integration_confluent_account.foo", "api_secret", "test-api-secret-123"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_integration_confluent_account.foo", "tags.*", "foo:bar"),
				),
			},
			{
				Config: testAccCheckDatadogIntegrationConfluentAccountBasicUpdated(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationConfluentAccountExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_confluent_account.foo", "api_key", uniq),
					resource.TestCheckResourceAttr(
						"datadog_integration_confluent_account.foo", "api_secret", "test-api-secret-123"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_integration_confluent_account.foo", "tags.*", "mytag"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_integration_confluent_account.foo", "tags.*", "mytag2:myvalue"),
				),
			},
		},
	})
}

func testAccCheckDatadogIntegrationConfluentAccountBasic(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`
resource "datadog_integration_confluent_account" "foo" {
    api_key = "%s"
    api_secret = "test-api-secret-123"
	tags = ["foo:bar"]
}`, uniq)
}

func testAccCheckDatadogIntegrationConfluentAccountBasicUpdated(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`
resource "datadog_integration_confluent_account" "foo" {
    api_key = "%s"
    api_secret = "test-api-secret-123"
    tags = ["mytag", "mytag2:myvalue"]
}`, uniq)
}

func testAccCheckDatadogIntegrationConfluentAccountDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := IntegrationConfluentAccountDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func IntegrationConfluentAccountDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_integration_confluent_account" {
				continue
			}
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetConfluentCloudApiV2().GetConfluentAccount(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving Confluent Account %s", err)}
			}
			return &utils.RetryableError{Prob: "Monitor still Confluent Account"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogIntegrationConfluentAccountExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := integrationConfluentAccountExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func integrationConfluentAccountExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_integration_confluent_account" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetConfluentCloudApiV2().GetConfluentAccount(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving Confluent Account")
		}
	}
	return nil
}
