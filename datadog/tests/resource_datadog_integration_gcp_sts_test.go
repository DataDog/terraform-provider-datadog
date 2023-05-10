package test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestGCPSTSIntegrationBasic(t *testing.T) {
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testBasicGCPSTSResource(uniq),
				Check: resource.ComposeTestCheckFunc(
					testGCPIntegrationAccountExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp_sts.foo", "service_account_email", uniq),
					// Verify computed values
					resource.TestCheckResourceAttrSet("datadog_integration_gcp_sts.foo", "delegate_email"),
					resource.TestCheckResourceAttrSet("datadog_integration_gcp_sts.foo", "id"),
					// Verify nil values
					resource.TestCheckNoResourceAttr("datadog_integration_gcp_sts.foo", "host_filters"),
					resource.TestCheckNoResourceAttr("datadog_integration_gcp_sts.foo", "automute"),
					resource.TestCheckNoResourceAttr("datadog_integration_gcp_sts.foo", "enable_cspm"),
				),
			},
		},
		CheckDestroy: testGCPSTSDestroy(providers.frameworkProvider),
	})
}

func TestGCPSTSIntegrationResourceUpdates(t *testing.T) {
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testBasicGCPSTSResource(uniq),
				Check: resource.ComposeTestCheckFunc(
					testGCPIntegrationAccountExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp_sts.foo", "service_account_email", uniq),
					// Verify computed values
					resource.TestCheckResourceAttrSet("datadog_integration_gcp_sts.foo", "delegate_email"),
					resource.TestCheckResourceAttrSet("datadog_integration_gcp_sts.foo", "id"),
					// Verify nil values
					resource.TestCheckNoResourceAttr("datadog_integration_gcp_sts.foo", "host_filters"),
					resource.TestCheckNoResourceAttr("datadog_integration_gcp_sts.foo", "automute"),
					resource.TestCheckNoResourceAttr("datadog_integration_gcp_sts.foo", "enable_cspm"),
				),
			},
			{
				Config: testGCPIntegrationUpdate(uniq),
				Check: resource.ComposeTestCheckFunc(
					testGCPIntegrationAccountExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp_sts.foo", "service_account_email", uniq),
					// Verify computed values
					resource.TestCheckResourceAttrSet("datadog_integration_gcp_sts.foo", "delegate_email"),
					resource.TestCheckResourceAttrSet("datadog_integration_gcp_sts.foo", "id"),
					// Verify updated values
					resource.TestCheckResourceAttr("datadog_integration_gcp_sts.foo", "host_filters.0", "filter_one"),
					resource.TestCheckResourceAttr("datadog_integration_gcp_sts.foo", "host_filters.1", "filter_two"),
					resource.TestCheckResourceAttr("datadog_integration_gcp_sts.foo", "host_filters.2", "filter_three"),
					resource.TestCheckResourceAttr("datadog_integration_gcp_sts.foo", "automute", "true"),
					resource.TestCheckResourceAttr("datadog_integration_gcp_sts.foo", "enable_cspm", "true"),
				),
			},
			// Verify removing updates works
			{
				Config: testBasicGCPSTSResource(uniq),
				Check: resource.ComposeTestCheckFunc(
					testGCPIntegrationAccountExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp_sts.foo", "service_account_email", uniq),
					// Verify computed values
					resource.TestCheckResourceAttrSet("datadog_integration_gcp_sts.foo", "delegate_email"),
					resource.TestCheckResourceAttrSet("datadog_integration_gcp_sts.foo", "id"),
					// Verify nil values
					resource.TestCheckNoResourceAttr("datadog_integration_gcp_sts.foo", "host_filters"),
					resource.TestCheckNoResourceAttr("datadog_integration_gcp_sts.foo", "automute"),
					resource.TestCheckNoResourceAttr("datadog_integration_gcp_sts.foo", "enable_cspm"),
				),
			},
		},
		CheckDestroy: testGCPSTSDestroy(providers.frameworkProvider),
	})
}

// Update all configurable fields
func testGCPIntegrationUpdate(uniq string) string {
	return fmt.Sprintf(`
	resource "datadog_integration_gcp_sts" "foo" {
		service_account_email = "%s"
		host_filters = ["filter_one", "filter_two", "filter_three"]
		automute = true
		enable_cspm = true
	}`, uniq)
}

func testGCPIntegrationAccountExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := integrationGCPSTSAccountExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func integrationGCPSTSAccountExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_integration_gcp_sts" {
			continue
		}
		id := r.Primary.ID

		accounts, httpResp, err := apiInstances.GetGCPStsIntegrationApiV2().ListGCPSTSAccounts(auth)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving STS enabled GCP service account")
		}

		accountWasFound, err := findGCPAccount(id, accounts)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving STS enabled GCP service account")
		}
		if !accountWasFound {
			return errors.New("error finding your account with id:" + id)
		}
	}
	return nil
}

func testBasicGCPSTSResource(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`
resource "datadog_integration_gcp_sts" "foo" {
    service_account_email = "%s"
}`, uniq)
}

func testGCPSTSDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth
		if err := GCPSTSIntegrationDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func GCPSTSIntegrationDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_integration_gcp_sts" {
				continue
			}
			id := r.Primary.ID

			stsEnabledAccounts, _, err := apiInstances.GetGCPStsIntegrationApiV2().ListGCPSTSAccounts(auth)
			if err != nil {
				return &utils.RetryableError{Prob: fmt.Sprintf("error retrieving STS enabled accounts %s", err)}
			}

			accountWasFound, err := findGCPAccount(id, stsEnabledAccounts)
			if err != nil {
				return err
			}

			if !accountWasFound {
				return nil
			}

			return &utils.RetryableError{Prob: "GCP STS account still exists"}
		}
		return nil
	})
	return err
}

func findGCPAccount(accountID string, stsEnabledAccounts datadogV2.STSEnabledAccountData) (bool, error) {
	accounts, ok := stsEnabledAccounts.GetDataOk()
	if !ok {
		return false, errors.New("error retrieving Data from GCP accounts")
	}

	for _, account := range *accounts {
		id, ok := account.GetIdOk()
		if !ok {
			return false, errors.New("error retrieving id from account")
		}

		if *id == accountID {
			return true, nil
		}
	}

	return false, nil
}
