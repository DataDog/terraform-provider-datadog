package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDatadogWebhookOauth2ClientCredentials_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniqueName := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogWebhookOauth2ClientCredentialsDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogWebhookOauth2ClientCredentialsBasicConfig(uniqueName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogWebhookOauth2ClientCredentialsExists(providers.frameworkProvider, "datadog_webhook_oauth2_client_credentials.foo"),
					resource.TestCheckResourceAttr("datadog_webhook_oauth2_client_credentials.foo", "name", uniqueName),
					resource.TestCheckResourceAttr("datadog_webhook_oauth2_client_credentials.foo", "access_token_url", "https://example.com/oauth2/token"),
					resource.TestCheckResourceAttr("datadog_webhook_oauth2_client_credentials.foo", "client_id", "test-client-id"),
					resource.TestCheckResourceAttr("datadog_webhook_oauth2_client_credentials.foo", "client_secret", "test-client-secret"),
					resource.TestCheckResourceAttr("datadog_webhook_oauth2_client_credentials.foo", "audience", "https://example.com/api"),
					resource.TestCheckNoResourceAttr("datadog_webhook_oauth2_client_credentials.foo", "scope"),
					resource.TestCheckResourceAttrSet("datadog_webhook_oauth2_client_credentials.foo", "id"),
				),
			},
			{
				Config: testAccCheckDatadogWebhookOauth2ClientCredentialsBasicConfigUpdated(uniqueName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogWebhookOauth2ClientCredentialsExists(providers.frameworkProvider, "datadog_webhook_oauth2_client_credentials.foo"),
					resource.TestCheckResourceAttr("datadog_webhook_oauth2_client_credentials.foo", "name", uniqueName+"UPDATED"),
					resource.TestCheckResourceAttr("datadog_webhook_oauth2_client_credentials.foo", "access_token_url", "https://example.com/oauth2/token/updated"),
					resource.TestCheckResourceAttr("datadog_webhook_oauth2_client_credentials.foo", "client_id", "test-client-id-updated"),
					resource.TestCheckResourceAttr("datadog_webhook_oauth2_client_credentials.foo", "client_secret", "test-client-secret-updated"),
					resource.TestCheckNoResourceAttr("datadog_webhook_oauth2_client_credentials.foo", "audience"),
					resource.TestCheckResourceAttr("datadog_webhook_oauth2_client_credentials.foo", "scope", "read write"),
				),
			},
		},
	})
}

func testAccCheckDatadogWebhookOauth2ClientCredentialsBasicConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_webhook_oauth2_client_credentials" "foo" {
  name             = "%[1]s"
  access_token_url = "https://example.com/oauth2/token"
  client_id        = "test-client-id"
  client_secret    = "test-client-secret"
  audience         = "https://example.com/api"
}`, uniq)
}

func testAccCheckDatadogWebhookOauth2ClientCredentialsBasicConfigUpdated(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_webhook_oauth2_client_credentials" "foo" {
  name             = "%[1]sUPDATED"
  access_token_url = "https://example.com/oauth2/token/updated"
  client_id        = "test-client-id-updated"
  client_secret    = "test-client-secret-updated"
  scope            = "read write"
}`, uniq)
}

func testAccCheckDatadogWebhookOauth2ClientCredentialsExists(accProvider *fwprovider.FrameworkProvider, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		id := s.RootModule().Resources[name].Primary.ID
		_, httpresp, err := apiInstances.GetWebhooksIntegrationApiV2().GetOAuth2ClientCredentials(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpresp, "error checking webhook OAuth2 client credentials existence")
		}
		return nil
	}
}

func testAccCheckDatadogWebhookOauth2ClientCredentialsDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_webhook_oauth2_client_credentials" {
				continue
			}

			id := r.Primary.ID

			_, httpResp, err := apiInstances.GetWebhooksIntegrationApiV2().GetOAuth2ClientCredentials(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					continue
				}
				return fmt.Errorf("received an error retrieving webhook OAuth2 client credentials: %s", err.Error())
			}
			return fmt.Errorf("webhook OAuth2 client credentials %s still exists", id)
		}

		return nil
	}
}
