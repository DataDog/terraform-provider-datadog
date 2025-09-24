package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDatadogIntegrationAWSExternalIDDatasource(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceIntegrationAWSExternalIDConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.datadog_integration_aws_external_id.foo", "external_id"),
					checkDatadogIntegrationAWSExternalIDDataSource(providers.frameworkProvider),
				),
			},
		},
	})
}

func testAccDatasourceIntegrationAWSExternalIDConfig() string {
	return `

data "datadog_integration_aws_external_id" "foo" {
  aws_account_id = "123456789012"
}
`
}

func checkDatadogIntegrationAWSExternalIDDataSource(accProvider *fwprovider.FrameworkProvider) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		// Verify the external_id matches the value returned by the API for the target account
		api := accProvider.DatadogApiInstances.GetAWSIntegrationApiV2()
		auth := accProvider.Auth

		// List accounts and find the one for 123456789012
		accountsResp, _, err := api.ListAWSAccounts(auth)
		if err != nil {
			return err
		}

		expectedExternalID := ""
		for _, item := range accountsResp.Data {
			attrs := item.GetAttributes()
			if attrs.GetAwsAccountId() != "123456789012" {
				continue
			}
			authConfig, ok := attrs.GetAuthConfigOk()
			if !ok || authConfig.AWSAuthConfigRole == nil {
				continue
			}
			expectedExternalID = authConfig.AWSAuthConfigRole.GetExternalId()
			break
		}

		if expectedExternalID == "" {
			return fmt.Errorf("could not find role-based auth config with external ID for account 123456789012")
		}

		resourceAttributes := state.RootModule().Resources["data.datadog_integration_aws_external_id.foo"].Primary.Attributes
		got := resourceAttributes["external_id"]
		if got == "" {
			return fmt.Errorf("external_id should not be empty")
		}
		if got != expectedExternalID {
			return fmt.Errorf("expected external_id %q, got %q", expectedExternalID, got)
		}
		return nil
	}
}
