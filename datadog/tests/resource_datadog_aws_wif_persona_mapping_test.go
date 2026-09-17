package test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

// awsWifReplayAccountIdentifier is the caller recorded in the cassette. It is only
// used when replaying: the API requires the caller's permissions to be a superset of
// the mapped account's, which is guaranteed by mapping the caller to itself, so the
// live and record paths resolve the caller through datadog_current_user instead. A
// data source cannot be used when replaying because the cassette predates it and holds
// no current_user interaction, and this literal cannot be used live because the
// recorded user does not exist in other orgs. Keep it in sync with the cassette.
const awsWifReplayAccountIdentifier = "tf-testacccurrentuserdatasource-local@example.com"

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

func TestAccAwsWifPersonaMappingServiceAccount(t *testing.T) {
	skipIfNoCassette(t)
	// Run before parallel tests: a mapping for the caller itself can prevent
	// the same caller from creating another mapping, even for a service account.
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	if !isReplaying() {
		// Temporary cleanup of the exact artifact leaked by PR #4234's failed
		// acceptance run 35259842227. Remove after the cleanup run succeeds.
		const leakedID = "bb94e162-b36b-4ec1-b3c5-1844b8396f20"
		const leakedARN = "arn:aws:sts::519956565653:assumed-role/terraform-runner/*"
		api := providers.frameworkProvider.DatadogApiInstances.GetCloudAuthenticationApiV2()
		var lastHTTP *http.Response
		err := retry.RetryContext(ctx, 30*time.Second, func() *retry.RetryError {
			mapping, httpResponse, err := api.GetAWSCloudAuthPersonaMapping(ctx, leakedID)
			lastHTTP = httpResponse
			if err != nil {
				if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
					return retry.RetryableError(err)
				}
				return retry.NonRetryableError(err)
			}
			data := mapping.GetData()
			attributes := data.GetAttributes()
			if data.GetId() != leakedID || attributes.GetArnPattern() != leakedARN {
				return retry.NonRetryableError(fmt.Errorf("refusing cleanup: mapping does not match this PR's test artifact"))
			}
			_, err = api.DeleteAWSCloudAuthPersonaMapping(ctx, leakedID)
			if err != nil {
				return retry.NonRetryableError(err)
			}
			t.Logf("removed test artifact %s", leakedID)
			return nil
		})
		if err != nil && (lastHTTP == nil || lastHTTP.StatusCode != http.StatusNotFound) {
			t.Fatalf("cleanup of PR test artifact: %v", err)
		}
	}

	accountID := uniqueAWSAccountID(ctx, t)
	resourceName := "datadog_aws_wif_persona_mapping.test"
	serviceAccountName := "datadog_service_account.test"
	serviceAccountConfig := fmt.Sprintf(`
data "datadog_role" "read_only" {
  filter = "Datadog Read Only Role"
}

resource "datadog_service_account" "test" {
  email = %q
  name  = "AWS WIF acceptance test"
  # The API requires a nonempty permission set contained in the caller's.
  roles = [data.datadog_role.read_only.id]
}
`, strings.ToLower(uniqueEntityName(ctx, t))+"@example.com")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckAwsWifPersonaMappingDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWifPersonaMappingConfigWithIdentity(accountID, serviceAccountName+".id", serviceAccountConfig),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWifPersonaMappingExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "account_identifier", serviceAccountName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "account_uuid", serviceAccountName, "id"),
					resource.TestCheckResourceAttr(resourceName, "arn_pattern", fmt.Sprintf("arn:aws:sts::%s:assumed-role/terraform-runner/*", accountID)),
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
	currentUserConfig := `data "datadog_current_user" "test" {}`
	accountIdentifier := "data.datadog_current_user.test.handle"
	if isReplaying() {
		currentUserConfig = ""
		accountIdentifier = fmt.Sprintf("%q", awsWifReplayAccountIdentifier)
	}

	return testAccAwsWifPersonaMappingConfigWithIdentity(accountID, accountIdentifier, currentUserConfig)
}

func testAccAwsWifPersonaMappingConfigWithIdentity(accountID, accountIdentifier, identityConfig string) string {
	return fmt.Sprintf(`
%[2]s

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
  account_identifier = %[3]s
  arn_pattern        = "arn:aws:sts::${datadog_integration_aws_account.test.aws_account_id}:assumed-role/terraform-runner/*"
}
`, accountID, identityConfig, accountIdentifier)
}

func testAccCheckAwsWifPersonaMappingExists(provider *fwprovider.FrameworkProvider, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		stateResource, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		id := stateResource.Primary.ID
		return retry.RetryContext(provider.Auth, 30*time.Second, func() *retry.RetryError {
			_, httpResponse, err := provider.DatadogApiInstances.GetCloudAuthenticationApiV2().GetAWSCloudAuthPersonaMapping(provider.Auth, id)
			if err == nil {
				return nil
			}
			translatedError := utils.TranslateClientError(err, httpResponse, "error checking AWS WIF persona mapping existence")
			if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
				return retry.RetryableError(translatedError)
			}
			return retry.NonRetryableError(translatedError)
		})
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
