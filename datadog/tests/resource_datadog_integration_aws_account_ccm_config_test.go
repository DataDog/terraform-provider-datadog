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

func TestAccIntegrationAwsAccountCcmConfigBasic(t *testing.T) {
	t.Parallel()
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIntegrationAwsAccountCcmConfigDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationAwsAccountCcmConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationAwsAccountCcmConfigExists(providers.frameworkProvider),
				),
			},
		},
	})
}

func testAccCheckDatadogIntegrationAwsAccountCcmConfig() string {
	return `resource "datadog_integration_aws_account_ccm_config" "foo" {
  aws_account_config_id = "b2087a32-4d4f-45b1-9321-1a0a48e9d7cf"

  ccm_config {
    data_export_configs {
      report_name   = "cost-and-usage-report"
      report_prefix = "reports"
      report_type   = "CUR2.0"
      bucket_name   = "billing"
      bucket_region = "us-east-1"
    }
  }
}`
}

func testAccCheckDatadogIntegrationAwsAccountCcmConfigDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := IntegrationAwsAccountCcmConfigDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func IntegrationAwsAccountCcmConfigDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_integration_aws_account_ccm_config" {
				continue
			}
			awsAccountConfigId := r.Primary.Attributes["aws_account_config_id"]

			_, httpResp, err := apiInstances.GetAWSIntegrationApiV2().GetAWSAccountCCMConfig(auth, awsAccountConfigId)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving IntegrationAwsAccountCcmConfig %s", err)}
			}
			return &utils.RetryableError{Prob: "IntegrationAwsAccountCcmConfig still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogIntegrationAwsAccountCcmConfigExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := integrationAwsAccountCcmConfigExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func integrationAwsAccountCcmConfigExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_integration_aws_account_ccm_config" {
			continue
		}
		awsAccountConfigId := r.Primary.Attributes["aws_account_config_id"]

		_, httpResp, err := apiInstances.GetAWSIntegrationApiV2().GetAWSAccountCCMConfig(auth, awsAccountConfigId)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving IntegrationAwsAccountCcmConfig")
		}
	}
	return nil
}
