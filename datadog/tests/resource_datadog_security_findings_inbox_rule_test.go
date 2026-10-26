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

// TestAccDatadogSecurityFindingsInboxRule covers the basic create -> update -> import lifecycle,
// exercising changes to every attribute (name, enabled, finding_types, query, action.description).
func TestAccDatadogSecurityFindingsInboxRule(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	resourceName := "datadog_security_findings_inbox_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogSecurityFindingsInboxRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "datadog_security_findings_inbox_rule" "test" {
  name    = "%s"
  enabled = true
  rule = {
    finding_types = ["misconfiguration"]
    query         = "env:dev @severity:low"
  }
  action = {
    description = "Needs triage"
  }
}
`, uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsInboxRuleExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", uniq),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "rule.finding_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.finding_types.0", "misconfiguration"),
					resource.TestCheckResourceAttr(resourceName, "rule.query", "env:dev @severity:low"),
					resource.TestCheckResourceAttr(resourceName, "action.description", "Needs triage"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "datadog_security_findings_inbox_rule" "test" {
  name    = "%s-updated"
  enabled = false
  rule = {
    finding_types = ["misconfiguration", "secret"]
    query         = "env:prod"
  }
  action = {
    description = "Triage production misconfigurations"
  }
}
`, uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsInboxRuleExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", uniq+"-updated"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "rule.finding_types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "rule.query", "env:prod"),
					resource.TestCheckResourceAttr(resourceName, "action.description", "Triage production misconfigurations"),
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

// TestAccDatadogSecurityFindingsInboxRule_Minimal configures only the required fields and verifies
// the computed default for `enabled` (true) and that omitted optionals are absent from state.
func TestAccDatadogSecurityFindingsInboxRule_Minimal(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	resourceName := "datadog_security_findings_inbox_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogSecurityFindingsInboxRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSecurityFindingsInboxRuleConfigMinimal(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsInboxRuleExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", uniq),
					// `enabled` is omitted from config: the schema default must apply.
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "rule.finding_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.finding_types.0", "misconfiguration"),
					resource.TestCheckNoResourceAttr(resourceName, "rule.query"),
					// The action block is required but its description is optional and omitted.
					resource.TestCheckNoResourceAttr(resourceName, "action.description"),
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

// TestAccDatadogSecurityFindingsInboxRule_OptionalChurn sets the optional fields, then removes
// them, asserting they clear cleanly with no perpetual diff.
func TestAccDatadogSecurityFindingsInboxRule_OptionalChurn(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	resourceName := "datadog_security_findings_inbox_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogSecurityFindingsInboxRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				// All optionals present.
				Config: fmt.Sprintf(`
resource "datadog_security_findings_inbox_rule" "test" {
  name    = "%s"
  enabled = false
  rule = {
    finding_types = ["misconfiguration"]
    query         = "env:prod"
  }
  action = {
    description = "context"
  }
}
`, uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsInboxRuleExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "rule.query", "env:prod"),
					resource.TestCheckResourceAttr(resourceName, "action.description", "context"),
				),
			},
			{
				// All optionals removed: they must clear, and the plan must be empty afterwards.
				Config: testAccCheckDatadogSecurityFindingsInboxRuleConfigMinimal(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsInboxRuleExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckNoResourceAttr(resourceName, "rule.query"),
					resource.TestCheckNoResourceAttr(resourceName, "action.description"),
				),
			},
		},
	})
}

// TestAccDatadogSecurityFindingsInboxRule_Validation exercises the schema validators. Each step
// fails at plan time (before any API call), so these are independent of the rule's runtime behavior.
func TestAccDatadogSecurityFindingsInboxRule_Validation(t *testing.T) {
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
resource "datadog_security_findings_inbox_rule" "test" {
  name = "%s"
  rule = {
    finding_types = []
  }
  action = {}
}
`, uniq),
				ExpectError: regexp.MustCompile("must contain at least 1"),
			},
			{
				// finding_types value must be a valid enum member.
				Config: fmt.Sprintf(`
resource "datadog_security_findings_inbox_rule" "test" {
  name = "%s"
  rule = {
    finding_types = ["not_a_finding_type"]
  }
  action = {}
}
`, uniq),
				ExpectError: regexp.MustCompile("invalid value"),
			},
			{
				// The rule attribute is required.
				Config: fmt.Sprintf(`
resource "datadog_security_findings_inbox_rule" "test" {
  name   = "%s"
  action = {}
}
`, uniq),
				ExpectError: regexp.MustCompile("is required, but no definition was found"),
			},
			{
				// The action attribute is required: the API rejects requests without it.
				Config: fmt.Sprintf(`
resource "datadog_security_findings_inbox_rule" "test" {
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

// TestAccDatadogSecurityFindingsInboxRule_OutOfBandDelete deletes the rule directly through the API
// and asserts Terraform detects it as drift (a non-empty plan that would recreate it).
func TestAccDatadogSecurityFindingsInboxRule_OutOfBandDelete(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	resourceName := "datadog_security_findings_inbox_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogSecurityFindingsInboxRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSecurityFindingsInboxRuleConfigMinimal(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsInboxRuleExists(providers.frameworkProvider, resourceName),
					// Let the create settle, then delete the rule out-of-band; the follow-up plan must be non-empty.
					testAccInboxRuleDeleteOutOfBand(providers.frameworkProvider, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccDatadogSecurityFindingsInboxRule_OutOfBandUpdate renames the rule directly through the API
// and asserts Terraform detects the drift and restores the configured value on the next apply.
func TestAccDatadogSecurityFindingsInboxRule_OutOfBandUpdate(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	resourceName := "datadog_security_findings_inbox_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogSecurityFindingsInboxRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSecurityFindingsInboxRuleConfigMinimal(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsInboxRuleExists(providers.frameworkProvider, resourceName),
				),
			},
			{
				// Mutate the rule out-of-band; the refreshed plan must be non-empty (drift detected).
				Config: testAccCheckDatadogSecurityFindingsInboxRuleConfigMinimal(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccInboxRuleRenameOutOfBand(providers.frameworkProvider, resourceName, uniq+"-drifted"),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				// Re-applying the unchanged config must restore the configured name.
				Config: testAccCheckDatadogSecurityFindingsInboxRuleConfigMinimal(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsInboxRuleExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", uniq),
				),
			},
		},
	})
}

// testAccCheckDatadogSecurityFindingsInboxRuleConfigMinimal sets only the required fields. It is
// the one config reused across multiple tests; single-use configs are inlined into their TestStep.
func testAccCheckDatadogSecurityFindingsInboxRuleConfigMinimal(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_security_findings_inbox_rule" "test" {
  name = "%s"
  rule = {
    finding_types = ["misconfiguration"]
  }
  action = {}
}
`, uniq)
}

func testAccCheckDatadogSecurityFindingsInboxRuleExists(accProvider *fwprovider.FrameworkProvider, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		r, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		id, err := uuid.Parse(r.Primary.ID)
		if err != nil {
			return fmt.Errorf("invalid inbox rule ID %s: %w", r.Primary.ID, err)
		}
		_, httpResp, err := apiInstances.GetSecurityMonitoringApiV2().GetSecurityFindingsAutomationInboxRule(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving inbox rule")
		}
		return nil
	}
}

func testAccCheckDatadogSecurityFindingsInboxRuleDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_security_findings_inbox_rule" {
				continue
			}
			err := utils.Retry(2, 10, func() error {
				id, parseErr := uuid.Parse(r.Primary.ID)
				if parseErr != nil {
					return &utils.RetryableError{Prob: fmt.Sprintf("invalid inbox rule ID %s", r.Primary.ID)}
				}
				_, httpResp, err := apiInstances.GetSecurityMonitoringApiV2().GetSecurityFindingsAutomationInboxRule(auth, id)
				if err != nil {
					if httpResp != nil && httpResp.StatusCode == 404 {
						return nil
					}
					return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving inbox rule %s", err)}
				}
				return &utils.RetryableError{Prob: "inbox rule still exists"}
			})
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// testAccInboxRuleDeleteOutOfBand deletes the named rule directly through the API, simulating an
// external deletion that Terraform should detect as drift.
func testAccInboxRuleDeleteOutOfBand(accProvider *fwprovider.FrameworkProvider, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		api := accProvider.DatadogApiInstances.GetSecurityMonitoringApiV2()
		auth := accProvider.Auth

		r, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		id, err := uuid.Parse(r.Primary.ID)
		if err != nil {
			return fmt.Errorf("invalid inbox rule ID %s: %w", r.Primary.ID, err)
		}
		if httpResp, err := api.DeleteSecurityFindingsAutomationInboxRule(auth, id); err != nil {
			return utils.TranslateClientError(err, httpResp, "error deleting inbox rule out-of-band")
		}
		return nil
	}
}

// testAccInboxRuleRenameOutOfBand renames the named rule directly through the API, simulating an
// external modification that Terraform should detect as drift.
func testAccInboxRuleRenameOutOfBand(accProvider *fwprovider.FrameworkProvider, resourceName, newName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		api := accProvider.DatadogApiInstances.GetSecurityMonitoringApiV2()
		auth := accProvider.Auth

		r, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		id, err := uuid.Parse(r.Primary.ID)
		if err != nil {
			return fmt.Errorf("invalid inbox rule ID %s: %w", r.Primary.ID, err)
		}

		current, httpResp, err := api.GetSecurityFindingsAutomationInboxRule(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving inbox rule for out-of-band update")
		}
		currentData := current.GetData()
		attrs := currentData.GetAttributes()

		updateAttrs := datadogV2.NewInboxRuleAttributesCreateWithDefaults()
		updateAttrs.SetName(newName)
		updateAttrs.SetEnabled(attrs.GetEnabled())
		updateAttrs.SetRule(attrs.GetRule())
		updateAttrs.SetAction(attrs.GetAction())

		data := datadogV2.NewInboxRuleDataUpdateWithDefaults()
		data.SetId(id)
		data.SetType(datadogV2.INBOXRULETYPE_INBOX_RULES)
		data.SetAttributes(*updateAttrs)

		body := datadogV2.NewInboxRuleUpdateRequestWithDefaults()
		body.SetData(*data)

		if _, httpResp, err := api.UpdateSecurityFindingsAutomationInboxRule(auth, id, *body); err != nil {
			return utils.TranslateClientError(err, httpResp, "error updating inbox rule out-of-band")
		}
		return nil
	}
}
