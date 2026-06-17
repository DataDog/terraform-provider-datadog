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

func TestAccDatadogSecurityFindingsDueDateRule(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	resourceName := "datadog_security_findings_due_date_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogSecurityFindingsDueDateRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSecurityFindingsDueDateRuleConfig(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsDueDateRuleExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", uniq),
					resource.TestCheckResourceAttr(resourceName, "rule.finding_types.0", "misconfiguration"),
					resource.TestCheckResourceAttr(resourceName, "action.due_from", "first_seen"),
					resource.TestCheckResourceAttr(resourceName, "action.due_days_per_severity.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "action.due_days_per_severity.0.severity", "critical"),
					resource.TestCheckResourceAttr(resourceName, "action.due_days_per_severity.0.due_in_days", "7"),
				),
			},
			{
				Config: testAccCheckDatadogSecurityFindingsDueDateRuleConfigUpdated(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsDueDateRuleExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", uniq+"-updated"),
					resource.TestCheckResourceAttr(resourceName, "action.due_from", "fix_available"),
					resource.TestCheckResourceAttr(resourceName, "action.due_days_per_severity.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "action.due_days_per_severity.2.severity", "medium"),
					resource.TestCheckResourceAttr(resourceName, "action.due_days_per_severity.2.due_in_days", "90"),
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

func testAccCheckDatadogSecurityFindingsDueDateRuleConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_security_findings_due_date_rule" "test" {
  name    = "%s"
  enabled = true
  rule {
    finding_types = ["misconfiguration"]
    query         = "env:prod"
  }
  action {
    due_from = "first_seen"
    due_days_per_severity {
      severity    = "critical"
      due_in_days = 7
    }
    due_days_per_severity {
      severity    = "high"
      due_in_days = 30
    }
  }
}
`, uniq)
}

func testAccCheckDatadogSecurityFindingsDueDateRuleConfigUpdated(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_security_findings_due_date_rule" "test" {
  name    = "%s-updated"
  enabled = true
  rule {
    finding_types = ["misconfiguration"]
    query         = "env:prod"
  }
  action {
    due_from = "fix_available"
    due_days_per_severity {
      severity    = "critical"
      due_in_days = 7
    }
    due_days_per_severity {
      severity    = "high"
      due_in_days = 30
    }
    due_days_per_severity {
      severity    = "medium"
      due_in_days = 90
    }
  }
}
`, uniq)
}

func testAccCheckDatadogSecurityFindingsDueDateRuleExists(accProvider *fwprovider.FrameworkProvider, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		r, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		id, err := uuid.Parse(r.Primary.ID)
		if err != nil {
			return fmt.Errorf("invalid due date rule ID %s: %w", r.Primary.ID, err)
		}
		_, httpResp, err := apiInstances.GetSecurityMonitoringApiV2().GetSecurityFindingsAutomationDueDateRule(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving due date rule")
		}
		return nil
	}
}

func testAccCheckDatadogSecurityFindingsDueDateRuleDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_security_findings_due_date_rule" {
				continue
			}
			err := utils.Retry(2, 10, func() error {
				id, parseErr := uuid.Parse(r.Primary.ID)
				if parseErr != nil {
					return &utils.RetryableError{Prob: fmt.Sprintf("invalid due date rule ID %s", r.Primary.ID)}
				}
				_, httpResp, err := apiInstances.GetSecurityMonitoringApiV2().GetSecurityFindingsAutomationDueDateRule(auth, id)
				if err != nil {
					if httpResp != nil && httpResp.StatusCode == 404 {
						return nil
					}
					return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving due date rule %s", err)}
				}
				return &utils.RetryableError{Prob: "due date rule still exists"}
			})
			if err != nil {
				return err
			}
		}
		return nil
	}
}
