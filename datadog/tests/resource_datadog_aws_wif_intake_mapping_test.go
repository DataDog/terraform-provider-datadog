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

func TestAccAwsWifIntakeMapping(t *testing.T) {
	skipIfNoCassette(t)
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	accountID := uniqueAWSAccountID(ctx, t)
	resourceName := "datadog_aws_wif_intake_mapping.test"
	var originalID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckAwsWifIntakeMappingDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWifIntakeMappingConfig(accountID, "DatadogAgentRole"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWifIntakeMappingExists(providers.frameworkProvider, resourceName),
					func(state *terraform.State) error {
						originalID = state.RootModule().Resources[resourceName].Primary.ID
						return nil
					},
					resource.TestCheckResourceAttr(resourceName, "arn_pattern", fmt.Sprintf("arn:aws:sts::%s:assumed-role/DatadogAgentRole/*", accountID)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsWifIntakeMappingConfig(accountID, "DatadogAgentRoleUpdated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWifIntakeMappingExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "arn_pattern", fmt.Sprintf("arn:aws:sts::%s:assumed-role/DatadogAgentRoleUpdated/*", accountID)),
					func(state *terraform.State) error {
						if state.RootModule().Resources[resourceName].Primary.ID == originalID {
							return fmt.Errorf("changing arn_pattern did not replace the mapping")
						}
						return nil
					},
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

func testAccAwsWifIntakeMappingConfig(accountID, roleName string) string {
	return fmt.Sprintf(`
resource "datadog_integration_aws_account" "test" {
  aws_account_id = %[1]q
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
resource "datadog_aws_wif_intake_mapping" "test" {
  arn_pattern = "arn:aws:sts::${datadog_integration_aws_account.test.aws_account_id}:assumed-role/%[2]s/*"
}
`, accountID, roleName)
}

func testAccCheckAwsWifIntakeMappingExists(provider *fwprovider.FrameworkProvider, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		stateResource, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		id := stateResource.Primary.ID
		return retry.RetryContext(provider.Auth, 30*time.Second, func() *retry.RetryError {
			_, httpResponse, err := provider.DatadogApiInstances.GetCloudAuthenticationApiV2().GetAWSCloudAuthIntakeMapping(provider.Auth, id)
			if err == nil {
				return nil
			}
			translatedError := utils.TranslateClientError(err, httpResponse, "error checking AWS WIF intake mapping existence")
			if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
				return retry.RetryableError(translatedError)
			}
			return retry.NonRetryableError(translatedError)
		})
	}
}

func testAccCheckAwsWifIntakeMappingDestroy(provider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, stateResource := range state.RootModule().Resources {
			if stateResource.Type != "datadog_aws_wif_intake_mapping" {
				continue
			}

			err := retry.RetryContext(provider.Auth, 30*time.Second, func() *retry.RetryError {
				_, httpResponse, err := provider.DatadogApiInstances.GetCloudAuthenticationApiV2().GetAWSCloudAuthIntakeMapping(provider.Auth, stateResource.Primary.ID)
				if err != nil {
					if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
						return nil
					}
					return retry.NonRetryableError(utils.TranslateClientError(err, httpResponse, "error checking AWS WIF intake mapping destruction"))
				}
				return retry.RetryableError(fmt.Errorf("AWS WIF intake mapping %s still exists", stateResource.Primary.ID))
			})
			if err != nil {
				return err
			}
		}
		return nil
	}
}
