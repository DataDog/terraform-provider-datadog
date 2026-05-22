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

const databricksTestIntegration = "databricks"

func TestAccIntegrationDatabricksAccountOAuth(t *testing.T) {
	t.Parallel()
	// Skip in live mode (RECORD=none). AMS validates the OAuth credentials
	// end-to-end against the real Databricks workspace at create time, so a
	// recorded cassette is required for this test to run in CI.
	if !isReplaying() && !isRecording() {
		t.Skip("Requires RECORD=true (record) or RECORD=false (replay)")
	}
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	resourceName := "datadog_integration_databricks_account.oauth"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIntegrationDatabricksAccountDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationDatabricksAccountOAuth(uniq, true, "0 * * * *"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationDatabricksAccountExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", uniq),
					resource.TestCheckResourceAttr(resourceName, "workspace_url", "https://dbc-097db9cd-e3d9.cloud.databricks.com/"),
					resource.TestCheckResourceAttr(resourceName, "djm_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "ccm_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "do_crawlers_cron", "0 * * * *"),
					resource.TestCheckResourceAttr(resourceName, "auth_config.oauth.client_id", "f635a6e0-60f8-448d-9087-b9359eafa329"),
					resource.TestCheckResourceAttr(resourceName, "auth_config.oauth.databricks_account_id", "f482fe01-7704-4760-af92-72ed263ac8f3"),
				),
			},
			{
				// Toggle djm_enabled off and change the cron to verify update + drift.
				Config: testAccCheckDatadogIntegrationDatabricksAccountOAuth(uniq, false, "0 0 * * *"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationDatabricksAccountExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "djm_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "do_crawlers_cron", "0 0 * * *"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"auth_config"},
			},
		},
	})
}

func testAccCheckDatadogIntegrationDatabricksAccountOAuth(uniq string, djmEnabled bool, cron string) string {
	return fmt.Sprintf(`
resource "datadog_integration_databricks_account" "oauth" {
    name          = "%s"
    workspace_url = "https://dbc-097db9cd-e3d9.cloud.databricks.com/"

    auth_config {
        oauth {
            client_id             = "f635a6e0-60f8-448d-9087-b9359eafa329"
            client_secret         = "oauth-test-client-secret"
            databricks_account_id = "f482fe01-7704-4760-af92-72ed263ac8f3"
        }
    }

    djm_enabled             = %t
    do_crawlers_cron        = "%s"
    serverless_jobs_enabled = false
}`, uniq, djmEnabled, cron)
}

func TestAccIntegrationDatabricksAccountPat(t *testing.T) {
	t.Parallel()
	// Skip in live mode (RECORD=none). The AMS API validates credentials
	// end-to-end against the real Databricks workspace at create time, so a
	// recorded cassette is required for this test to run in CI.
	if !isReplaying() && !isRecording() {
		t.Skip("Requires RECORD=true (record) or RECORD=false (replay)")
	}
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	resourceName := "datadog_integration_databricks_account.pat"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIntegrationDatabricksAccountDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationDatabricksAccountPat(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationDatabricksAccountExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(resourceName, "name", uniq),
					resource.TestCheckResourceAttr(resourceName, "workspace_url", "https://dbc-097db9cd-e3d9.cloud.databricks.com/"),
					resource.TestCheckResourceAttr(resourceName, "auth_config.pat.token", "dapi-test-token"),
					resource.TestCheckResourceAttr(resourceName, "djm_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "serverless_jobs_enabled", "false"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"auth_config"},
			},
		},
	})
}

func testAccCheckDatadogIntegrationDatabricksAccountPat(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_databricks_account" "pat" {
    name          = "%s"
    workspace_url = "https://dbc-097db9cd-e3d9.cloud.databricks.com/"

    auth_config {
        pat {
            token = "dapi-test-token"
        }
    }

    serverless_jobs_enabled = false
}`, uniq)
}

func testAccCheckDatadogIntegrationDatabricksAccountDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		return integrationDatabricksAccountDestroyHelper(auth, s, apiInstances)
	}
}

func integrationDatabricksAccountDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	return utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_integration_databricks_account" {
				continue
			}
			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetWebIntegrationsApiV2().GetWebIntegrationAccount(auth, databricksTestIntegration, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving databricks account: %s", err)}
			}
			return &utils.RetryableError{Prob: "databricks account still exists"}
		}
		return nil
	})
}

func testAccCheckDatadogIntegrationDatabricksAccountExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		return integrationDatabricksAccountExistsHelper(auth, s, apiInstances)
	}
}

func integrationDatabricksAccountExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_integration_databricks_account" {
			continue
		}
		id := r.Primary.ID

		_, httpResp, err := apiInstances.GetWebIntegrationsApiV2().GetWebIntegrationAccount(auth, databricksTestIntegration, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving databricks account")
		}
	}
	return nil
}
