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
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

const aiGatewayBudgetResourceName = "datadog_ai_gateway_budget.foo"

func TestAccAiGatewayBudget(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	name := uniqueEntityName(ctx, t)
	mlApp := name
	var firstID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAiGatewayBudgetDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				// The API lowercases tag keys, and tags replace the budget, so an uppercase key would replace it on every run.
				Config:      strings.Replace(testAccAiGatewayBudgetConfig(name, mlApp, "DAILY"), "ml_app =", "ML_APP =", 1),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`must be lowercase`),
			},
			{
				Config: testAccAiGatewayBudgetConfig(name, mlApp, "DAILY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAiGatewayBudgetExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(aiGatewayBudgetResourceName, "name", name),
					resource.TestCheckResourceAttr(aiGatewayBudgetResourceName, "capacity_microcents", "100000000"),
					resource.TestCheckResourceAttr(aiGatewayBudgetResourceName, "window_policy", "DAILY"),
					resource.TestCheckResourceAttr(aiGatewayBudgetResourceName, "enforcement", "BLOCK"),
					resource.TestCheckResourceAttr(aiGatewayBudgetResourceName, "warning_threshold", "0.8"),
					resource.TestCheckResourceAttr(aiGatewayBudgetResourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(aiGatewayBudgetResourceName, "tags.ml_app", mlApp),
					resource.TestCheckNoResourceAttr(aiGatewayBudgetResourceName, "selector"),
					func(s *terraform.State) error {
						firstID = s.RootModule().Resources[aiGatewayBudgetResourceName].Primary.ID
						return nil
					},
				),
			},
			{
				// Fields the update endpoint accepts change in place, and the server must show the new values.
				Config: testAccAiGatewayBudgetConfigUpdated(name, mlApp),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(aiGatewayBudgetResourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(aiGatewayBudgetResourceName, "name", name+"-renamed"),
					resource.TestCheckResourceAttr(aiGatewayBudgetResourceName, "capacity_microcents", "250000000"),
					resource.TestCheckResourceAttr(aiGatewayBudgetResourceName, "enforcement", "NOTIFY"),
					resource.TestCheckResourceAttr(aiGatewayBudgetResourceName, "warning_threshold", "0.9"),
					resource.TestCheckResourceAttr(aiGatewayBudgetResourceName, "enabled", "false"),
					testAccCheckDatadogAiGatewayBudgetOnServer(providers.frameworkProvider, 250000000, false),
				),
			},
			{
				// window_policy can only be set on create, so changing it replaces the budget, and the old one must be gone.
				Config: testAccAiGatewayBudgetConfig(name, mlApp, "WEEKLY"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(aiGatewayBudgetResourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(aiGatewayBudgetResourceName, "window_policy", "WEEKLY"),
					func(s *terraform.State) error {
						if s.RootModule().Resources[aiGatewayBudgetResourceName].Primary.ID == firstID {
							return fmt.Errorf("budget %s was not replaced", firstID)
						}
						_, status, err := getAiGatewayBudget(providers.frameworkProvider, firstID)
						if status != 404 {
							return fmt.Errorf("replaced budget %s: status %d, err %w, want 404", firstID, status, err)
						}
						return nil
					},
				),
			},
			{
				ResourceName:      aiGatewayBudgetResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccAiGatewayBudgetSelector checks that a template budget's selector, which the API sorts and canonicalizes,
// reads back without a diff, since any difference would replace the budget.
func TestAccAiGatewayBudgetSelector(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	name := uniqueEntityName(ctx, t)
	mlApp := name

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAiGatewayBudgetDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "datadog_ai_gateway_budget" "foo" {
  name                = "%s"
  capacity_microcents = 50000000
  window_policy       = "MONTHLY"
  enforcement         = "BLOCK"
  selector = [
    { dim = "tag:ml_app", values = ["%s"] },
    { dim = "provider", values = ["openai", "anthropic"], per_value = true },
    { dim = "model", per_value = true },
  ]
}`, name, mlApp),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAiGatewayBudgetExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(aiGatewayBudgetResourceName, "selector.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(aiGatewayBudgetResourceName, "selector.*", map[string]string{
						"dim":       "model",
						"per_value": "true",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(aiGatewayBudgetResourceName, "selector.*", map[string]string{
						"dim":       "provider",
						"per_value": "true",
						"values.#":  "2",
					}),
				),
			},
			{
				ResourceName:      aiGatewayBudgetResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAiGatewayBudgetConfig(name, mlApp, windowPolicy string) string {
	return fmt.Sprintf(`
resource "datadog_ai_gateway_budget" "foo" {
  name                = "%s"
  capacity_microcents = 100000000
  window_policy       = "%s"
  enforcement         = "BLOCK"
  warning_threshold   = 0.8
  tags = {
    ml_app = "%s"
  }
}`, name, windowPolicy, mlApp)
}

func testAccAiGatewayBudgetConfigUpdated(name, mlApp string) string {
	return fmt.Sprintf(`
resource "datadog_ai_gateway_budget" "foo" {
  name                    = "%s-renamed"
  capacity_microcents     = 250000000
  apply_to_current_period = true
  window_policy           = "DAILY"
  enforcement             = "NOTIFY"
  enabled                 = false
  tags = {
    ml_app = "%s"
  }
}`, name, mlApp)
}

func getAiGatewayBudget(accProvider *fwprovider.FrameworkProvider, id string) (datadogV2.AiGatewayBudgetResponse, int, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return datadogV2.AiGatewayBudgetResponse{}, 0, err
	}
	api := datadogV2.NewAIGatewayApi(accProvider.DatadogApiInstances.HttpClient)
	resp, httpResp, err := api.GetAiGatewayBudget(accProvider.Auth, parsed)
	status := 0
	if httpResp != nil {
		status = httpResp.StatusCode
	}
	return resp, status, err
}

func testAccCheckDatadogAiGatewayBudgetDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		return utils.Retry(2, 5, func() error {
			for _, r := range s.RootModule().Resources {
				if r.Type != "datadog_ai_gateway_budget" {
					continue
				}
				_, status, err := getAiGatewayBudget(accProvider, r.Primary.ID)
				if err != nil {
					if status == 404 {
						continue
					}
					return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving AI Gateway budget: %s", err)}
				}
				return &utils.RetryableError{Prob: "AI Gateway budget still exists"}
			}
			return nil
		})
	}
}

func testAccCheckDatadogAiGatewayBudgetExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_ai_gateway_budget" {
				continue
			}
			if _, _, err := getAiGatewayBudget(accProvider, r.Primary.ID); err != nil {
				return fmt.Errorf("received an error retrieving AI Gateway budget: %w", err)
			}
		}
		return nil
	}
}

// testAccCheckDatadogAiGatewayBudgetOnServer reads the budget from the API, so it fails if the provider only
// updated state while the server kept the old values.
func testAccCheckDatadogAiGatewayBudgetOnServer(accProvider *fwprovider.FrameworkProvider, capacity int64, enabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r, ok := s.RootModule().Resources[aiGatewayBudgetResourceName]
		if !ok {
			return fmt.Errorf("%s not found in state", aiGatewayBudgetResourceName)
		}
		resp, _, err := getAiGatewayBudget(accProvider, r.Primary.ID)
		if err != nil {
			return fmt.Errorf("received an error retrieving AI Gateway budget: %w", err)
		}
		attributes := resp.Data.Attributes
		if attributes.CapacityMicrocents != capacity || attributes.Enabled != enabled {
			return fmt.Errorf("server has capacity %d and enabled %t, want %d and %t",
				attributes.CapacityMicrocents, attributes.Enabled, capacity, enabled)
		}
		return nil
	}
}
