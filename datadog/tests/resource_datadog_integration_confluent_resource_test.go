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

func TestAccIntegrationConfluentResourceBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIntegrationConfluentResourceDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationConfluentResource(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationConfluentResourceExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_confluent_resource.foo", "resource_type", "kafka"),
					resource.TestCheckResourceAttr(
						"datadog_integration_confluent_resource.foo", "resource_id", "12345678910"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_integration_confluent_resource.foo", "tags.*", "mytag"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_integration_confluent_resource.foo", "tags.*", "mytag2:myvalue"),
					resource.TestCheckResourceAttr(
						"datadog_integration_confluent_resource.foo", "enable_custom_metrics", "false"),
				),
			},
			{
				Config: testAccCheckDatadogIntegrationConfluentResourceUpdated(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationConfluentResourceExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_confluent_resource.foo", "resource_type", "connector"),
					resource.TestCheckResourceAttr(
						"datadog_integration_confluent_resource.foo", "resource_id", "12345678910"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_integration_confluent_resource.foo", "tags.*", "mytag"),
					resource.TestCheckResourceAttr(
						"datadog_integration_confluent_resource.foo", "enable_custom_metrics", "true"),
				),
			},
		},
	})
}

func testAccCheckDatadogIntegrationConfluentResource(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`
resource "datadog_integration_confluent_account" "foo" {
	api_key = "%s"
	api_secret = "test-api-secret-123"
}

resource "datadog_integration_confluent_resource" "foo" {
    account_id    = datadog_integration_confluent_account.foo.id
	resource_id   = "12345678910"
    resource_type = "kafka"
    tags = ["mytag", "mytag2:myvalue"]
}`, uniq)
}

func testAccCheckDatadogIntegrationConfluentResourceUpdated(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`
resource "datadog_integration_confluent_account" "foo" {
	api_key = "%s"
	api_secret = "test-api-secret-123"
}

resource "datadog_integration_confluent_resource" "foo" {
    account_id    = datadog_integration_confluent_account.foo.id
	resource_id   = "12345678910"
    resource_type = "connector"
    tags = ["mytag"]
    enable_custom_metrics = true
}`, uniq)
}

func testAccCheckDatadogIntegrationConfluentResourceDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := IntegrationConfluentResourceDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func IntegrationConfluentResourceDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_integration_confluent_resource" {
				continue
			}
			accountId := r.Primary.Attributes["account_id"]
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetConfluentCloudApiV2().GetConfluentResource(auth, accountId, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving Confluent resource %s", err)}
			}
			return &utils.RetryableError{Prob: "Confluent resource still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogIntegrationConfluentResourceExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := integrationConfluentResourceExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func integrationConfluentResourceExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_integration_confluent_resource" {
			continue
		}
		accountId := r.Primary.Attributes["account_id"]
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetConfluentCloudApiV2().GetConfluentResource(auth, accountId, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving Confluent resource")
		}
	}
	return nil
}
