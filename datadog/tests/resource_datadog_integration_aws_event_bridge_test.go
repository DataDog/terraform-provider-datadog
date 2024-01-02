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

func TestAccIntegrationAwsEventBridgeBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	accountID := uniqueAWSAccountID(ctx, t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIntegrationAwsEventBridgeDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationAwsEventBridge(accountID, uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationAwsEventBridgeExists(providers.frameworkProvider),
				),
			},
		},
	})
}

func testAccCheckDatadogIntegrationAwsEventBridge(accountID string, uniq string) string {
	return fmt.Sprintf(`
	resource "datadog_integration_aws" "account" {
		account_id                       = "%s"
		role_name                        = "testacc-datadog-integration-role"
	  }

	resource "datadog_integration_aws_event_bridge" "foo" {
		account_id = "%s"
		event_generator_name = "%s"
		create_event_bus = false
		region = "us-east-1"
		depends_on = [datadog_integration_aws.account]
	}`, accountID, accountID, uniq)
}

func testAccCheckDatadogIntegrationAwsEventBridgeDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := IntegrationAwsEventBridgeDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func IntegrationAwsEventBridgeDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_integration_aws_event_bridge" {
				continue
			}

			_, httpResp, err := apiInstances.GetAWSIntegrationApiV1().ListAWSEventBridgeSources(auth)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving IntegrationAwsEventBridge %s", err)}
			}
			return &utils.RetryableError{Prob: "IntegrationAwsEventBridge still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogIntegrationAwsEventBridgeExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := integrationAwsEventBridgeExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func integrationAwsEventBridgeExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_integration_aws_event_bridge" {
			continue
		}

		_, httpResp, err := apiInstances.GetAWSIntegrationApiV1().ListAWSEventBridgeSources(auth)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving IntegrationAwsEventBridge")
		}
	}
	return nil
}
