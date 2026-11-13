package test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

const aiGatewayRateLimitResourceName = "datadog_ai_gateway_rate_limit.foo"

func TestAccAiGatewayRateLimit(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	name := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAiGatewayRateLimitDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				// The API lowercases tag keys, so an uppercase key would show as a change on every plan.
				Config:      strings.Replace(testAccAiGatewayRateLimitConfig(name), "provider =", "Provider =", 1),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`must be lowercase`),
			},
			{
				Config: testAccAiGatewayRateLimitConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAiGatewayRateLimitExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(aiGatewayRateLimitResourceName, "name", name),
					resource.TestCheckResourceAttr(aiGatewayRateLimitResourceName, "mode", "ALERT"),
					resource.TestCheckResourceAttr(aiGatewayRateLimitResourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(aiGatewayRateLimitResourceName, "requests_per_minute", "100"),
					resource.TestCheckResourceAttr(aiGatewayRateLimitResourceName, "input_tokens_per_minute", "50000"),
					resource.TestCheckResourceAttr(aiGatewayRateLimitResourceName, "tags.provider", "openai"),
				),
			},
			{
				// Removing a limit and the tags must clear them on the server, not leave them in place.
				Config: testAccAiGatewayRateLimitConfigUpdated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAiGatewayRateLimitExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(aiGatewayRateLimitResourceName, "mode", "ENFORCE"),
					resource.TestCheckResourceAttr(aiGatewayRateLimitResourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(aiGatewayRateLimitResourceName, "requests_per_minute", "200"),
					resource.TestCheckNoResourceAttr(aiGatewayRateLimitResourceName, "input_tokens_per_minute"),
					resource.TestCheckNoResourceAttr(aiGatewayRateLimitResourceName, "tags.%"),
					testAccCheckDatadogAiGatewayRateLimitCleared(providers.frameworkProvider),
				),
			},
			{
				ResourceName:      aiGatewayRateLimitResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAiGatewayRateLimitConfig(name string) string {
	return fmt.Sprintf(`
resource "datadog_ai_gateway_rate_limit" "foo" {
  name                    = "%s"
  mode                    = "ALERT"
  requests_per_minute     = 100
  input_tokens_per_minute = 50000
  tags = {
    provider = "openai"
  }
}`, name)
}

func testAccAiGatewayRateLimitConfigUpdated(name string) string {
	return fmt.Sprintf(`
resource "datadog_ai_gateway_rate_limit" "foo" {
  name                = "%s"
  mode                = "ENFORCE"
  enabled             = false
  requests_per_minute = 200
}`, name)
}

func getAiGatewayRateLimit(accProvider *fwprovider.FrameworkProvider, id string) (datadogV2.AiGatewayRateLimitResponse, int, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return datadogV2.AiGatewayRateLimitResponse{}, 0, err
	}
	api := datadogV2.NewAIGatewayApi(accProvider.DatadogApiInstances.HttpClient)
	resp, httpResp, err := api.GetAiGatewayRateLimit(accProvider.Auth, parsed)
	status := 0
	if httpResp != nil {
		status = httpResp.StatusCode
	}
	return resp, status, err
}

func testAccCheckDatadogAiGatewayRateLimitDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		return utils.Retry(2, 5, func() error {
			for _, r := range s.RootModule().Resources {
				if r.Type != "datadog_ai_gateway_rate_limit" {
					continue
				}
				_, status, err := getAiGatewayRateLimit(accProvider, r.Primary.ID)
				if err != nil {
					if status == 404 {
						continue
					}
					return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving AI Gateway rate limit: %s", err)}
				}
				return &utils.RetryableError{Prob: "AI Gateway rate limit still exists"}
			}
			return nil
		})
	}
}

func testAccCheckDatadogAiGatewayRateLimitExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_ai_gateway_rate_limit" {
				continue
			}
			if _, _, err := getAiGatewayRateLimit(accProvider, r.Primary.ID); err != nil {
				return fmt.Errorf("received an error retrieving AI Gateway rate limit: %w", err)
			}
		}
		return nil
	}
}

// testAccCheckDatadogAiGatewayRateLimitCleared reads the rate limit from the API, so it fails if the provider
// only dropped the removed fields from state while the server kept them.
func testAccCheckDatadogAiGatewayRateLimitCleared(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r, ok := s.RootModule().Resources[aiGatewayRateLimitResourceName]
		if !ok {
			return fmt.Errorf("%s not found in state", aiGatewayRateLimitResourceName)
		}
		resp, _, err := getAiGatewayRateLimit(accProvider, r.Primary.ID)
		if err != nil {
			return fmt.Errorf("received an error retrieving AI Gateway rate limit: %w", err)
		}
		attributes := resp.Data.Attributes
		if v := attributes.InputTokensPerMinute.Get(); v != nil && *v > 0 {
			return fmt.Errorf("input_tokens_per_minute is still %d on the server", *v)
		}
		if len(attributes.Tags) > 0 {
			return fmt.Errorf("tags are still %v on the server", attributes.Tags)
		}
		return nil
	}
}
