package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDatadogIntegrationAWSExternalIDDatasource(t *testing.T) {
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	accountID := uniqueAWSAccountID(ctx, t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckDatadogIntegrationAwsAccountDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceIntegrationAWSExternalIDConfig(accountID, uniq),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.datadog_integration_aws_external_id.foo", "external_id"),
					checkDatadogIntegrationAWSExternalIDDataSource(providers.frameworkProvider, accountID),
				),
			},
		},
	})
}

func testAccDatasourceIntegrationAWSExternalIDConfig(accountID, uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_aws_account" "foo" {
  aws_account_id = %s
  account_tags   = ["tag:%s"]
  aws_partition  = "aws"
  aws_regions {}
  auth_config {
    aws_auth_config_role {
      role_name = "test"
    }
  }
  logs_config {
    lambda_forwarder {}
  }
  metrics_config {
    namespace_filters {}
  }
  resources_config {}
  traces_config {
    xray_services {}
  }
}

data "datadog_integration_aws_external_id" "foo" {
  aws_account_id = datadog_integration_aws_account.foo.aws_account_id
}
`, accountID, uniq)
}

func checkDatadogIntegrationAWSExternalIDDataSource(accProvider *fwprovider.FrameworkProvider, accountID string) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		// Verify the external_id matches the value returned by the API for the target account
		api := accProvider.DatadogApiInstances.GetAWSIntegrationApiV2()
		auth := accProvider.Auth

		optionalParams := datadogV2.NewListAWSAccountsOptionalParameters().WithAwsAccountId(accountID)
		accountsResp, _, err := api.ListAWSAccounts(auth, *optionalParams)
		if err != nil {
			return err
		}

		expectedExternalID := ""
		if len(accountsResp.Data) > 0 {
			attrs := accountsResp.Data[0].GetAttributes()
			authConfig, ok := attrs.GetAuthConfigOk()
			if ok && authConfig.AWSAuthConfigRole != nil {
				expectedExternalID = authConfig.AWSAuthConfigRole.GetExternalId()
			}
		}

		if expectedExternalID == "" {
			return fmt.Errorf("could not find role-based auth config with external ID for account %s", accountID)
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
