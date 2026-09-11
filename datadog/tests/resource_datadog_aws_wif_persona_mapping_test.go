package test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

const awsWifTestAccountIdentifier = "tf-testacccurrentuserdatasource-local@example.com"

func TestAccAwsWifPersonaMapping(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	accountID := uniqueAWSAccountID(ctx, t)
	resourceName := "datadog_aws_wif_persona_mapping.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckAwsWifPersonaMappingDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWifPersonaMappingConfig(accountID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWifPersonaMappingExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "arn_pattern", fmt.Sprintf("arn:aws:sts::%s:assumed-role/terraform-runner/*", accountID)),
					resource.TestCheckResourceAttrSet(resourceName, "account_identifier"),
					resource.TestCheckResourceAttrSet(resourceName, "account_uuid"),
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

func testAccAwsWifPersonaMappingConfig(accountID string) string {
	return fmt.Sprintf(`
resource "datadog_integration_aws_account" "test" {
  aws_account_id = %[1]s
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

resource "datadog_aws_wif_persona_mapping" "test" {
  # This is the caller in the standard acceptance-test org, which guarantees
  # the API's required permission-subset relation.
  account_identifier = %[2]q
  arn_pattern         = "arn:aws:sts::${datadog_integration_aws_account.test.aws_account_id}:assumed-role/terraform-runner/*"
}
`, accountID, awsWifTestAccountIdentifier)
}

func testAccCheckAwsWifPersonaMappingExists(provider *fwprovider.FrameworkProvider, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		id := state.RootModule().Resources[resourceName].Primary.ID
		_, httpResponse, err := provider.DatadogApiInstances.GetCloudAuthenticationApiV2().GetAWSCloudAuthPersonaMapping(provider.Auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResponse, "error checking AWS WIF persona mapping existence")
		}
		return nil
	}
}

func testAccCheckAwsWifPersonaMappingDestroy(provider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, stateResource := range state.RootModule().Resources {
			if stateResource.Type != "datadog_aws_wif_persona_mapping" {
				continue
			}

			err := retry.RetryContext(provider.Auth, 30*time.Second, func() *retry.RetryError {
				_, httpResponse, err := provider.DatadogApiInstances.GetCloudAuthenticationApiV2().GetAWSCloudAuthPersonaMapping(provider.Auth, stateResource.Primary.ID)
				if err != nil {
					if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
						return nil
					}
					return retry.NonRetryableError(utils.TranslateClientError(err, httpResponse, "error checking AWS WIF persona mapping destruction"))
				}
				return retry.RetryableError(fmt.Errorf("AWS WIF persona mapping %s still exists", stateResource.Primary.ID))
			})
			if err != nil {
				return err
			}
		}
		return nil
	}
}
