package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDatadogIntegrationAWSAccountUuidDatasource(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	awsAccountId := "123456789012"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceIntegrationAWSAccountUuidConfig(awsAccountId),
				Check: resource.ComposeTestCheckFunc(
					checkDatadogIntegrationAWSAccountUuidCount(providers.frameworkProvider, awsAccountId),
				),
			},
		},
	})
}

func testAccDatasourceIntegrationAWSAccountUuidConfig(awsAccountId string) string {
	return fmt.Sprintf(`data "datadog_integration_aws_account_uuid" "foo" {
		aws_account_id = "%s"
	}`, awsAccountId)
}

func checkDatadogIntegrationAWSAccountUuidCount(accProvider *fwprovider.FrameworkProvider, awsAccountId string) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		/*
			apiInstances := accProvider.DatadogApiInstances
			auth := accProvider.Auth

			params := datadogV2.ListAWSAccountsOptionalParameters{
				AwsAccountId: &awsAccountId,
			}
			resp, _, err := apiInstances.GetAWSIntegrationApiV2().ListAWSAccounts(auth, params)
			if err != nil {
				return err
			}

			if len(resp.GetData()) != 1 {
				return fmt.Errorf("Expected exactly one account with ID %s", awsAccountId)
			}
		*/

		actualId := state.RootModule().Resources["data.datadog_integration_aws_account_uuid.foo"].Primary.ID
		//expectedId := resp.GetData()[0].GetId()
		expectedId := "be093cc6-1fe4-4c19-955e-abdcf983ece9"

		if actualId != expectedId {
			return fmt.Errorf("Expected account config uuid %s, got %s", expectedId, actualId)
		}

		return nil
	}
}
