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

func TestAccDatadogSecurityFindingsTicketCreationRule(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	resourceName := "datadog_security_findings_ticket_creation_rule.test"
	projectID := "11111111-1111-1111-1111-111111111111"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogSecurityFindingsTicketCreationRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSecurityFindingsTicketCreationRuleConfig(uniq, projectID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsTicketCreationRuleExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", uniq),
					resource.TestCheckResourceAttr(resourceName, "rule.finding_types.0", "misconfiguration"),
					resource.TestCheckResourceAttr(resourceName, "action.project_id", projectID),
					resource.TestCheckResourceAttr(resourceName, "action.target", "jira"),
					resource.TestCheckResourceAttr(resourceName, "action.max_tickets_per_day", "100"),
				),
			},
			{
				Config: testAccCheckDatadogSecurityFindingsTicketCreationRuleConfigUpdated(uniq, projectID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsTicketCreationRuleExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", uniq+"-updated"),
					resource.TestCheckResourceAttr(resourceName, "action.max_tickets_per_day", "50"),
					resource.TestCheckResourceAttr(resourceName, "action.assignee_id", "22222222-2222-2222-2222-222222222222"),
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

func testAccCheckDatadogSecurityFindingsTicketCreationRuleConfig(uniq, projectID string) string {
	return fmt.Sprintf(`
resource "datadog_security_findings_ticket_creation_rule" "test" {
  name    = "%s"
  enabled = true
  rule {
    finding_types = ["misconfiguration"]
    query         = "env:prod @severity:critical"
  }
  action {
    project_id          = "%s"
    target              = "jira"
    max_tickets_per_day = 100
  }
}
`, uniq, projectID)
}

func testAccCheckDatadogSecurityFindingsTicketCreationRuleConfigUpdated(uniq, projectID string) string {
	return fmt.Sprintf(`
resource "datadog_security_findings_ticket_creation_rule" "test" {
  name    = "%s-updated"
  enabled = true
  rule {
    finding_types = ["misconfiguration"]
    query         = "env:prod @severity:critical"
  }
  action {
    project_id          = "%s"
    target              = "jira"
    assignee_id         = "22222222-2222-2222-2222-222222222222"
    max_tickets_per_day = 50
    fields = jsonencode({
      labels = ["security"]
    })
  }
}
`, uniq, projectID)
}

func testAccCheckDatadogSecurityFindingsTicketCreationRuleExists(accProvider *fwprovider.FrameworkProvider, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		r, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		id, err := uuid.Parse(r.Primary.ID)
		if err != nil {
			return fmt.Errorf("invalid ticket creation rule ID %s: %w", r.Primary.ID, err)
		}
		_, httpResp, err := apiInstances.GetSecurityMonitoringApiV2().GetSecurityFindingsAutomationTicketCreationRule(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving ticket creation rule")
		}
		return nil
	}
}

func testAccCheckDatadogSecurityFindingsTicketCreationRuleDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_security_findings_ticket_creation_rule" {
				continue
			}
			err := utils.Retry(2, 10, func() error {
				id, parseErr := uuid.Parse(r.Primary.ID)
				if parseErr != nil {
					return &utils.RetryableError{Prob: fmt.Sprintf("invalid ticket creation rule ID %s", r.Primary.ID)}
				}
				_, httpResp, err := apiInstances.GetSecurityMonitoringApiV2().GetSecurityFindingsAutomationTicketCreationRule(auth, id)
				if err != nil {
					if httpResp != nil && httpResp.StatusCode == 404 {
						return nil
					}
					return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving ticket creation rule %s", err)}
				}
				return &utils.RetryableError{Prob: "ticket creation rule still exists"}
			})
			if err != nil {
				return err
			}
		}
		return nil
	}
}
