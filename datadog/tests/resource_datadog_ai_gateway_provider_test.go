package test

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

const aiGatewayProviderResourceName = "datadog_ai_gateway_provider.foo"

// aiGatewayProviderTestIntegrationAccountID returns the integration account that holds the vendor key. It is created
// in the Datadog console, so recording the cassette requires DD_TEST_AI_GATEWAY_INTEGRATION_ACCOUNT_ID to name a
// real OpenAI integration account in the test org. The default is the tf-acceptance-fixture account in the test org,
// which the cassette was recorded with and CI's live runs use, so it must not be deleted.
func aiGatewayProviderTestIntegrationAccountID() string {
	if id := os.Getenv("DD_TEST_AI_GATEWAY_INTEGRATION_ACCOUNT_ID"); id != "" {
		return id
	}
	return "210b5556-01fe-4cd6-8490-9ae40f853329"
}

func TestAccAiGatewayProvider(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	name := uniqueEntityName(ctx, t)
	accountID := aiGatewayProviderTestIntegrationAccountID()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAiGatewayProviderDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccAiGatewayProviderConfig(name, name, accountID, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAiGatewayProviderExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(aiGatewayProviderResourceName, "name", name),
					resource.TestCheckResourceAttr(aiGatewayProviderResourceName, "slug", name),
					resource.TestCheckResourceAttr(aiGatewayProviderResourceName, "known_provider", "openai"),
					resource.TestCheckResourceAttr(aiGatewayProviderResourceName, "integration_account_id", accountID),
					resource.TestCheckResourceAttr(aiGatewayProviderResourceName, "cost_multiplier", "1"),
				),
			},
			{
				Config: testAccAiGatewayProviderConfig(name+"-renamed", name, accountID, "cost_multiplier = 1.5"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAiGatewayProviderExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(aiGatewayProviderResourceName, "name", name+"-renamed"),
					resource.TestCheckResourceAttr(aiGatewayProviderResourceName, "slug", name),
					resource.TestCheckResourceAttr(aiGatewayProviderResourceName, "cost_multiplier", "1.5"),
				),
			},
			{
				ResourceName:      aiGatewayProviderResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAiGatewayProviderConfig(name, slug, accountID, extra string) string {
	return fmt.Sprintf(`
resource "datadog_ai_gateway_provider" "foo" {
  name                   = "%s"
  slug                   = "%s"
  known_provider         = "openai"
  integration_account_id = "%s"
  %s
}`, name, slug, accountID, extra)
}

func testAccCheckDatadogAiGatewayProviderDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		api := datadogV2.NewAIGatewayApi(accProvider.DatadogApiInstances.HttpClient)
		return utils.Retry(2, 5, func() error {
			for _, r := range s.RootModule().Resources {
				if r.Type != "datadog_ai_gateway_provider" {
					continue
				}
				id, err := strconv.ParseInt(r.Primary.ID, 10, 64)
				if err != nil {
					return err
				}
				_, httpResp, err := api.GetAiGatewayProvider(accProvider.Auth, id)
				if err != nil {
					if httpResp != nil && httpResp.StatusCode == 404 {
						continue
					}
					return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving AI Gateway provider: %s", err)}
				}
				return &utils.RetryableError{Prob: "AI Gateway provider still exists"}
			}
			return nil
		})
	}
}

func testAccCheckDatadogAiGatewayProviderExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		api := datadogV2.NewAIGatewayApi(accProvider.DatadogApiInstances.HttpClient)
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_ai_gateway_provider" {
				continue
			}
			id, err := strconv.ParseInt(r.Primary.ID, 10, 64)
			if err != nil {
				return err
			}
			_, httpResp, err := api.GetAiGatewayProvider(accProvider.Auth, id)
			if err != nil {
				return utils.TranslateClientError(err, httpResp, "error retrieving AI Gateway provider")
			}
		}
		return nil
	}
}
