package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

const (
	ticketCreationTestProjectID  = "11111111-1111-1111-1111-111111111111"
	ticketCreationTestAssigneeID = "22222222-2222-2222-2222-222222222222"
)

// TestAccDatadogSecurityFindingsTicketCreationRule covers the basic create -> update -> import
// lifecycle, exercising every attribute (name, enabled, finding_types, query, project_id, target,
// assignee_id, fields, max_tickets_per_day). It also asserts the read-only, server-managed
// auto_disabled_reason is absent for a healthy rule.
func TestAccDatadogSecurityFindingsTicketCreationRule(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	resourceName := "datadog_security_findings_ticket_creation_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogSecurityFindingsTicketCreationRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "datadog_security_findings_ticket_creation_rule" "test" {
  name    = "%s"
  enabled = true
  rule = {
    finding_types = ["misconfiguration"]
    query         = "env:prod @severity:critical"
  }
  action = {
    project_id          = "%s"
    target              = "jira"
    max_tickets_per_day = 100
  }
}
`, uniq, ticketCreationTestProjectID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsTicketCreationRuleExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", uniq),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "rule.finding_types.0", "misconfiguration"),
					resource.TestCheckResourceAttr(resourceName, "rule.query", "env:prod @severity:critical"),
					resource.TestCheckResourceAttr(resourceName, "action.project_id", ticketCreationTestProjectID),
					resource.TestCheckResourceAttr(resourceName, "action.target", "jira"),
					resource.TestCheckResourceAttr(resourceName, "action.max_tickets_per_day", "100"),
					// Optionals omitted on create.
					resource.TestCheckNoResourceAttr(resourceName, "action.assignee_id"),
					resource.TestCheckNoResourceAttr(resourceName, "action.fields"),
					// auto_disabled_reason is read-only and only set by the server on a ticketing
					// integration error; a healthy rule must report it as absent.
					resource.TestCheckNoResourceAttr(resourceName, "action.auto_disabled_reason"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "datadog_security_findings_ticket_creation_rule" "test" {
  name    = "%s-updated"
  enabled = false
  rule = {
    finding_types = ["misconfiguration", "secret"]
    query         = "env:prod @severity:critical"
  }
  action = {
    project_id          = "%s"
    target              = "jira"
    assignee_id         = "%s"
    max_tickets_per_day = 50
    fields = jsonencode({
      labels = ["security"]
    })
  }
}
`, uniq, ticketCreationTestProjectID, ticketCreationTestAssigneeID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsTicketCreationRuleExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", uniq+"-updated"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "rule.finding_types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "action.max_tickets_per_day", "50"),
					resource.TestCheckResourceAttr(resourceName, "action.assignee_id", ticketCreationTestAssigneeID),
					resource.TestCheckResourceAttr(resourceName, "action.fields", `{"labels":["security"]}`),
					resource.TestCheckNoResourceAttr(resourceName, "action.auto_disabled_reason"),
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

// TestAccDatadogSecurityFindingsTicketCreationRule_Minimal configures only the required fields and
// verifies the computed default for `enabled` (true) and that omitted optionals (including the
// read-only auto_disabled_reason) are absent from state.
func TestAccDatadogSecurityFindingsTicketCreationRule_Minimal(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	resourceName := "datadog_security_findings_ticket_creation_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogSecurityFindingsTicketCreationRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSecurityFindingsTicketCreationRuleConfigMinimal(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsTicketCreationRuleExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", uniq),
					// `enabled` is omitted from config: the schema default must apply.
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckNoResourceAttr(resourceName, "rule.query"),
					resource.TestCheckResourceAttr(resourceName, "action.target", "jira"),
					resource.TestCheckResourceAttr(resourceName, "action.max_tickets_per_day", "100"),
					resource.TestCheckNoResourceAttr(resourceName, "action.assignee_id"),
					resource.TestCheckNoResourceAttr(resourceName, "action.fields"),
					resource.TestCheckNoResourceAttr(resourceName, "action.auto_disabled_reason"),
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

// TestAccDatadogSecurityFindingsTicketCreationRule_OptionalChurn sets every optional field, then
// removes them all, asserting they clear cleanly with no perpetual diff. The read-only
// auto_disabled_reason must remain absent throughout.
func TestAccDatadogSecurityFindingsTicketCreationRule_OptionalChurn(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	resourceName := "datadog_security_findings_ticket_creation_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogSecurityFindingsTicketCreationRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				// All optionals present.
				Config: fmt.Sprintf(`
resource "datadog_security_findings_ticket_creation_rule" "test" {
  name    = "%s"
  enabled = false
  rule = {
    finding_types = ["misconfiguration"]
    query         = "env:prod"
  }
  action = {
    project_id          = "%s"
    target              = "jira"
    assignee_id         = "%s"
    max_tickets_per_day = 25
    fields = jsonencode({
      labels = ["security"]
    })
  }
}
`, uniq, ticketCreationTestProjectID, ticketCreationTestAssigneeID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsTicketCreationRuleExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "rule.query", "env:prod"),
					resource.TestCheckResourceAttr(resourceName, "action.assignee_id", ticketCreationTestAssigneeID),
					resource.TestCheckResourceAttr(resourceName, "action.fields", `{"labels":["security"]}`),
					resource.TestCheckNoResourceAttr(resourceName, "action.auto_disabled_reason"),
				),
			},
			{
				// All optionals removed: they must clear, and the plan must be empty afterwards.
				Config: testAccCheckDatadogSecurityFindingsTicketCreationRuleConfigMinimal(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsTicketCreationRuleExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckNoResourceAttr(resourceName, "rule.query"),
					resource.TestCheckNoResourceAttr(resourceName, "action.assignee_id"),
					resource.TestCheckNoResourceAttr(resourceName, "action.fields"),
					resource.TestCheckNoResourceAttr(resourceName, "action.auto_disabled_reason"),
				),
			},
		},
	})
}

// TestAccDatadogSecurityFindingsTicketCreationRule_AutoDisabledReasonReadOnly asserts that
// auto_disabled_reason cannot be set from configuration: it is a Computed, server-managed field.
func TestAccDatadogSecurityFindingsTicketCreationRule_AutoDisabledReasonReadOnly(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "datadog_security_findings_ticket_creation_rule" "test" {
  name = "%s"
  rule = {
    finding_types = ["misconfiguration"]
  }
  action = {
    project_id           = "%s"
    target               = "jira"
    max_tickets_per_day  = 100
    auto_disabled_reason = "should not be settable"
  }
}
`, uniq, ticketCreationTestProjectID),
				ExpectError: regexp.MustCompile("read-only"),
			},
		},
	})
}

// TestAccDatadogSecurityFindingsTicketCreationRule_Validation exercises the schema validators and
// the apply-time project_id parse. Each step fails before any rule is created.
func TestAccDatadogSecurityFindingsTicketCreationRule_Validation(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				// finding_types must contain at least one element.
				Config: fmt.Sprintf(`
resource "datadog_security_findings_ticket_creation_rule" "test" {
  name = "%s"
  rule = {
    finding_types = []
  }
  action = {
    project_id          = "%s"
    target              = "jira"
    max_tickets_per_day = 100
  }
}
`, uniq, ticketCreationTestProjectID),
				ExpectError: regexp.MustCompile("must contain at least 1"),
			},
			{
				// finding_types value must be a valid enum member.
				Config: fmt.Sprintf(`
resource "datadog_security_findings_ticket_creation_rule" "test" {
  name = "%s"
  rule = {
    finding_types = ["not_a_finding_type"]
  }
  action = {
    project_id          = "%s"
    target              = "jira"
    max_tickets_per_day = 100
  }
}
`, uniq, ticketCreationTestProjectID),
				ExpectError: regexp.MustCompile("invalid value"),
			},
			{
				// target value must be a valid enum member.
				Config: fmt.Sprintf(`
resource "datadog_security_findings_ticket_creation_rule" "test" {
  name = "%s"
  rule = {
    finding_types = ["misconfiguration"]
  }
  action = {
    project_id          = "%s"
    target              = "not_a_target"
    max_tickets_per_day = 100
  }
}
`, uniq, ticketCreationTestProjectID),
				ExpectError: regexp.MustCompile("invalid value"),
			},
			{
				// project_id must be a valid UUID (enforced by the schema validator at plan time).
				Config: fmt.Sprintf(`
resource "datadog_security_findings_ticket_creation_rule" "test" {
  name = "%s"
  rule = {
    finding_types = ["misconfiguration"]
  }
  action = {
    project_id          = "not-a-uuid"
    target              = "jira"
    max_tickets_per_day = 100
  }
}
`, uniq),
				ExpectError: regexp.MustCompile("must be a valid UUID"),
			},
			{
				// The rule attribute is required.
				Config: fmt.Sprintf(`
resource "datadog_security_findings_ticket_creation_rule" "test" {
  name = "%s"
  action = {
    project_id          = "%s"
    target              = "jira"
    max_tickets_per_day = 100
  }
}
`, uniq, ticketCreationTestProjectID),
				ExpectError: regexp.MustCompile("is required, but no definition was found"),
			},
			{
				// The action attribute is required.
				Config: fmt.Sprintf(`
resource "datadog_security_findings_ticket_creation_rule" "test" {
  name = "%s"
  rule = {
    finding_types = ["misconfiguration"]
  }
}
`, uniq),
				ExpectError: regexp.MustCompile("is required, but no definition was found"),
			},
		},
	})
}

// TestAccDatadogSecurityFindingsTicketCreationRule_OutOfBandDelete deletes the rule directly through
// the API and asserts Terraform detects it as drift (a non-empty plan that would recreate it).
func TestAccDatadogSecurityFindingsTicketCreationRule_OutOfBandDelete(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	resourceName := "datadog_security_findings_ticket_creation_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogSecurityFindingsTicketCreationRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSecurityFindingsTicketCreationRuleConfigMinimal(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsTicketCreationRuleExists(providers.frameworkProvider, resourceName),
					// Let the create settle, then delete the rule out-of-band; the follow-up plan must be non-empty.
					testAccTicketCreationRuleDeleteOutOfBand(providers.frameworkProvider, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccDatadogSecurityFindingsTicketCreationRule_OutOfBandUpdate renames the rule directly through
// the API and asserts Terraform detects the drift and restores the configured value on the next apply.
func TestAccDatadogSecurityFindingsTicketCreationRule_OutOfBandUpdate(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	resourceName := "datadog_security_findings_ticket_creation_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogSecurityFindingsTicketCreationRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSecurityFindingsTicketCreationRuleConfigMinimal(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsTicketCreationRuleExists(providers.frameworkProvider, resourceName),
				),
			},
			{
				// Mutate the rule out-of-band; the refreshed plan must be non-empty (drift detected).
				Config: testAccCheckDatadogSecurityFindingsTicketCreationRuleConfigMinimal(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccTicketCreationRuleRenameOutOfBand(providers.frameworkProvider, resourceName, uniq+"-drifted"),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				// Re-applying the unchanged config must restore the configured name.
				Config: testAccCheckDatadogSecurityFindingsTicketCreationRuleConfigMinimal(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsTicketCreationRuleExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", uniq),
				),
			},
		},
	})
}

// testAccCheckDatadogSecurityFindingsTicketCreationRuleConfigMinimal sets only the required fields.
// It is the one config reused across multiple tests; single-use configs are inlined into their TestStep.
func testAccCheckDatadogSecurityFindingsTicketCreationRuleConfigMinimal(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_security_findings_ticket_creation_rule" "test" {
  name = "%s"
  rule = {
    finding_types = ["misconfiguration"]
  }
  action = {
    project_id          = "%s"
    target              = "jira"
    max_tickets_per_day = 100
  }
}
`, uniq, ticketCreationTestProjectID)
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

// testAccTicketCreationRuleDeleteOutOfBand deletes the named rule directly through the API,
// simulating an external deletion that Terraform should detect as drift.
func testAccTicketCreationRuleDeleteOutOfBand(accProvider *fwprovider.FrameworkProvider, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		api := accProvider.DatadogApiInstances.GetSecurityMonitoringApiV2()
		auth := accProvider.Auth

		r, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		id, err := uuid.Parse(r.Primary.ID)
		if err != nil {
			return fmt.Errorf("invalid ticket creation rule ID %s: %w", r.Primary.ID, err)
		}
		if httpResp, err := api.DeleteSecurityFindingsAutomationTicketCreationRule(auth, id); err != nil {
			return utils.TranslateClientError(err, httpResp, "error deleting ticket creation rule out-of-band")
		}
		return nil
	}
}

// testAccTicketCreationRuleRenameOutOfBand renames the named rule directly through the API,
// simulating an external modification that Terraform should detect as drift.
func testAccTicketCreationRuleRenameOutOfBand(accProvider *fwprovider.FrameworkProvider, resourceName, newName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		api := accProvider.DatadogApiInstances.GetSecurityMonitoringApiV2()
		auth := accProvider.Auth

		r, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		id, err := uuid.Parse(r.Primary.ID)
		if err != nil {
			return fmt.Errorf("invalid ticket creation rule ID %s: %w", r.Primary.ID, err)
		}

		current, httpResp, err := api.GetSecurityFindingsAutomationTicketCreationRule(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving ticket creation rule for out-of-band update")
		}
		currentData := current.GetData()
		attrs := currentData.GetAttributes()

		// The response action type (TicketCreationRuleActionResponse) carries the read-only
		// auto_disabled_reason and is distinct from the request type (TicketCreationRuleAction), so
		// rebuild a request action from the fields we are allowed to send back.
		respAction := attrs.GetAction()
		action := datadogV2.NewTicketCreationRuleActionWithDefaults()
		action.SetProjectId(respAction.GetProjectId())
		action.SetTarget(respAction.GetTarget())
		action.SetMaxTicketsPerDay(respAction.GetMaxTicketsPerDay())
		if respAction.HasAssigneeId() {
			action.SetAssigneeId(respAction.GetAssigneeId())
		}
		if respAction.HasFields() {
			action.SetFields(respAction.GetFields())
		}

		updateAttrs := datadogV2.NewTicketCreationRuleAttributesCreateWithDefaults()
		updateAttrs.SetName(newName)
		updateAttrs.SetEnabled(attrs.GetEnabled())
		updateAttrs.SetRule(attrs.GetRule())
		updateAttrs.SetAction(*action)

		data := datadogV2.NewTicketCreationRuleDataCreateWithDefaults()
		data.SetType(datadogV2.TICKETCREATIONRULETYPE_TICKET_CREATION_RULES)
		data.SetAttributes(*updateAttrs)

		body := datadogV2.NewTicketCreationRuleUpdateRequestWithDefaults()
		body.SetData(*data)

		if _, httpResp, err := api.UpdateSecurityFindingsAutomationTicketCreationRule(auth, id, *body); err != nil {
			return utils.TranslateClientError(err, httpResp, "error updating ticket creation rule out-of-band")
		}
		return nil
	}
}
