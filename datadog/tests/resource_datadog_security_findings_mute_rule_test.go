package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccDatadogSecurityFindingsMuteRule(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	resourceName := "datadog_security_findings_mute_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogSecurityFindingsMuteRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSecurityFindingsMuteRuleConfig(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsMuteRuleExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", uniq),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "rule.finding_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.finding_types.0", "misconfiguration"),
					resource.TestCheckResourceAttr(resourceName, "rule.query", "env:dev @severity:low"),
					resource.TestCheckResourceAttr(resourceName, "action.reason", "risk_accepted"),
					resource.TestCheckResourceAttr(resourceName, "action.reason_description", "Accepted for dev environments only"),
				),
			},
			{
				Config: testAccCheckDatadogSecurityFindingsMuteRuleConfigUpdated(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsMuteRuleExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", uniq+"-updated"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "rule.finding_types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "action.reason", "false_positive"),
					resource.TestCheckResourceAttr(resourceName, "action.expire_at", "4070908800000"),
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

func testAccCheckDatadogSecurityFindingsMuteRuleConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_security_findings_mute_rule" "test" {
  name    = "%s"
  enabled = true
  rule {
    finding_types = ["misconfiguration"]
    query         = "env:dev @severity:low"
  }
  action {
    reason             = "risk_accepted"
    reason_description = "Accepted for dev environments only"
  }
}
`, uniq)
}

func testAccCheckDatadogSecurityFindingsMuteRuleConfigUpdated(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_security_findings_mute_rule" "test" {
  name    = "%s-updated"
  enabled = false
  rule {
    finding_types = ["misconfiguration", "secret"]
    query         = "env:prod"
  }
  action {
    reason    = "false_positive"
    expire_at = 4070908800000
  }
}
`, uniq)
}

func testAccCheckDatadogSecurityFindingsMuteRuleExists(accProvider *fwprovider.FrameworkProvider, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		r, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		id, err := uuid.Parse(r.Primary.ID)
		if err != nil {
			return fmt.Errorf("invalid mute rule ID %s: %w", r.Primary.ID, err)
		}
		_, httpResp, err := apiInstances.GetSecurityMonitoringApiV2().GetSecurityFindingsAutomationMuteRule(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving mute rule")
		}
		return nil
	}
}

func testAccCheckDatadogSecurityFindingsMuteRuleDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_security_findings_mute_rule" {
				continue
			}
			err := utils.Retry(2, 10, func() error {
				id, parseErr := uuid.Parse(r.Primary.ID)
				if parseErr != nil {
					return &utils.RetryableError{Prob: fmt.Sprintf("invalid mute rule ID %s", r.Primary.ID)}
				}
				_, httpResp, err := apiInstances.GetSecurityMonitoringApiV2().GetSecurityFindingsAutomationMuteRule(auth, id)
				if err != nil {
					if httpResp != nil && httpResp.StatusCode == 404 {
						return nil
					}
					return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving mute rule %s", err)}
				}
				return &utils.RetryableError{Prob: "mute rule still exists"}
			})
			if err != nil {
				return err
			}
		}
		return nil
	}
}
