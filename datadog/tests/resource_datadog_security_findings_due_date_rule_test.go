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

// TestAccDatadogSecurityFindingsDueDateRule covers the basic create -> update -> import lifecycle,
// exercising every attribute (name, enabled, finding_types, query, due_from, reason_description)
// and the due_days_per_severity blocks (adding entries, changing values).
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
				Config: fmt.Sprintf(`
resource "datadog_security_findings_due_date_rule" "test" {
  name    = "%s"
  enabled = true
  rule {
    finding_types = ["misconfiguration"]
    query         = "env:prod"
  }
  action {
    due_from           = "first_seen"
    reason_description = "Standard remediation SLA"
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
`, uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsDueDateRuleExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", uniq),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "rule.finding_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.finding_types.0", "misconfiguration"),
					resource.TestCheckResourceAttr(resourceName, "rule.query", "env:prod"),
					resource.TestCheckResourceAttr(resourceName, "action.due_from", "first_seen"),
					resource.TestCheckResourceAttr(resourceName, "action.reason_description", "Standard remediation SLA"),
					resource.TestCheckResourceAttr(resourceName, "action.due_days_per_severity.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "action.due_days_per_severity.0.severity", "critical"),
					resource.TestCheckResourceAttr(resourceName, "action.due_days_per_severity.0.due_in_days", "7"),
					resource.TestCheckResourceAttr(resourceName, "action.due_days_per_severity.1.severity", "high"),
					resource.TestCheckResourceAttr(resourceName, "action.due_days_per_severity.1.due_in_days", "30"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "datadog_security_findings_due_date_rule" "test" {
  name    = "%s-updated"
  enabled = false
  rule {
    finding_types = ["misconfiguration", "secret"]
    query         = "env:staging"
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
`, uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsDueDateRuleExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", uniq+"-updated"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "rule.finding_types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "action.due_from", "fix_available"),
					// reason_description was set on create and dropped here: it must clear.
					resource.TestCheckNoResourceAttr(resourceName, "action.reason_description"),
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

// TestAccDatadogSecurityFindingsDueDateRule_Minimal configures only the required fields and verifies
// the computed default for `enabled` (true) and that omitted optionals are absent from state.
func TestAccDatadogSecurityFindingsDueDateRule_Minimal(t *testing.T) {
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
				Config: testAccCheckDatadogSecurityFindingsDueDateRuleConfigMinimal(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsDueDateRuleExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", uniq),
					// `enabled` is omitted from config: the schema default must apply.
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckNoResourceAttr(resourceName, "rule.query"),
					resource.TestCheckResourceAttr(resourceName, "action.due_from", "first_seen"),
					resource.TestCheckNoResourceAttr(resourceName, "action.reason_description"),
					resource.TestCheckResourceAttr(resourceName, "action.due_days_per_severity.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.due_days_per_severity.0.severity", "critical"),
					resource.TestCheckResourceAttr(resourceName, "action.due_days_per_severity.0.due_in_days", "7"),
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

// TestAccDatadogSecurityFindingsDueDateRule_DueDaysPerSeverityChurn focuses on the
// due_days_per_severity list: it starts with three severities and a reason_description, then shrinks
// to a single severity with a changed due_in_days and no reason_description, asserting the list is
// rewritten cleanly with no perpetual diff.
func TestAccDatadogSecurityFindingsDueDateRule_DueDaysPerSeverityChurn(t *testing.T) {
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
				// Three severities plus an optional reason_description.
				Config: fmt.Sprintf(`
resource "datadog_security_findings_due_date_rule" "test" {
  name = "%s"
  rule {
    finding_types = ["misconfiguration"]
  }
  action {
    due_from           = "first_seen"
    reason_description = "context"
    due_days_per_severity {
      severity    = "critical"
      due_in_days = 7
    }
    due_days_per_severity {
      severity    = "high"
      due_in_days = 30
    }
    due_days_per_severity {
      severity    = "low"
      due_in_days = 180
    }
  }
}
`, uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsDueDateRuleExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "action.reason_description", "context"),
					resource.TestCheckResourceAttr(resourceName, "action.due_days_per_severity.#", "3"),
				),
			},
			{
				// Shrink to one severity with a changed due_in_days; reason_description is dropped.
				Config: fmt.Sprintf(`
resource "datadog_security_findings_due_date_rule" "test" {
  name = "%s"
  rule {
    finding_types = ["misconfiguration"]
  }
  action {
    due_from = "first_seen"
    due_days_per_severity {
      severity    = "critical"
      due_in_days = 14
    }
  }
}
`, uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsDueDateRuleExists(providers.frameworkProvider, resourceName),
					resource.TestCheckNoResourceAttr(resourceName, "action.reason_description"),
					resource.TestCheckResourceAttr(resourceName, "action.due_days_per_severity.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.due_days_per_severity.0.severity", "critical"),
					resource.TestCheckResourceAttr(resourceName, "action.due_days_per_severity.0.due_in_days", "14"),
				),
			},
		},
	})
}

// TestAccDatadogSecurityFindingsDueDateRule_Validation exercises the schema validators. Each step
// fails at plan time (before any API call), so these are independent of the rule's runtime behavior.
func TestAccDatadogSecurityFindingsDueDateRule_Validation(t *testing.T) {
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
resource "datadog_security_findings_due_date_rule" "test" {
  name = "%s"
  rule {
    finding_types = []
  }
  action {
    due_from = "first_seen"
    due_days_per_severity {
      severity    = "critical"
      due_in_days = 7
    }
  }
}
`, uniq),
				ExpectError: regexp.MustCompile("must contain at least 1"),
			},
			{
				// finding_types value must be a valid enum member.
				Config: fmt.Sprintf(`
resource "datadog_security_findings_due_date_rule" "test" {
  name = "%s"
  rule {
    finding_types = ["not_a_finding_type"]
  }
  action {
    due_from = "first_seen"
    due_days_per_severity {
      severity    = "critical"
      due_in_days = 7
    }
  }
}
`, uniq),
				ExpectError: regexp.MustCompile("invalid value"),
			},
			{
				// due_from value must be a valid enum member.
				Config: fmt.Sprintf(`
resource "datadog_security_findings_due_date_rule" "test" {
  name = "%s"
  rule {
    finding_types = ["misconfiguration"]
  }
  action {
    due_from = "not_a_due_from"
    due_days_per_severity {
      severity    = "critical"
      due_in_days = 7
    }
  }
}
`, uniq),
				ExpectError: regexp.MustCompile("invalid value"),
			},
			{
				// severity value must be a valid enum member.
				Config: fmt.Sprintf(`
resource "datadog_security_findings_due_date_rule" "test" {
  name = "%s"
  rule {
    finding_types = ["misconfiguration"]
  }
  action {
    due_from = "first_seen"
    due_days_per_severity {
      severity    = "not_a_severity"
      due_in_days = 7
    }
  }
}
`, uniq),
				ExpectError: regexp.MustCompile("invalid value"),
			},
			{
				// A severity may not appear more than once in due_days_per_severity.
				Config: fmt.Sprintf(`
resource "datadog_security_findings_due_date_rule" "test" {
  name = "%s"
  rule {
    finding_types = ["misconfiguration"]
  }
  action {
    due_from = "first_seen"
    due_days_per_severity {
      severity    = "critical"
      due_in_days = 7
    }
    due_days_per_severity {
      severity    = "critical"
      due_in_days = 30
    }
  }
}
`, uniq),
				ExpectError: regexp.MustCompile("used more than once"),
			},
			{
				// The rule block is required.
				Config: fmt.Sprintf(`
resource "datadog_security_findings_due_date_rule" "test" {
  name = "%s"
  action {
    due_from = "first_seen"
    due_days_per_severity {
      severity    = "critical"
      due_in_days = 7
    }
  }
}
`, uniq),
				ExpectError: regexp.MustCompile("must have a configuration value"),
			},
			{
				// The action block is required.
				Config: fmt.Sprintf(`
resource "datadog_security_findings_due_date_rule" "test" {
  name = "%s"
  rule {
    finding_types = ["misconfiguration"]
  }
}
`, uniq),
				ExpectError: regexp.MustCompile("must have a configuration value"),
			},
		},
	})
}

// TestAccDatadogSecurityFindingsDueDateRule_OutOfBandDelete deletes the rule directly through the
// API and asserts Terraform detects it as drift (a non-empty plan that would recreate it).
func TestAccDatadogSecurityFindingsDueDateRule_OutOfBandDelete(t *testing.T) {
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
				Config: testAccCheckDatadogSecurityFindingsDueDateRuleConfigMinimal(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsDueDateRuleExists(providers.frameworkProvider, resourceName),
					// Let the create settle, then delete the rule out-of-band; the follow-up plan must be non-empty.
					testAccDueDateRuleDeleteOutOfBand(providers.frameworkProvider, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccDatadogSecurityFindingsDueDateRule_OutOfBandUpdate renames the rule directly through the
// API and asserts Terraform detects the drift and restores the configured value on the next apply.
func TestAccDatadogSecurityFindingsDueDateRule_OutOfBandUpdate(t *testing.T) {
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
				Config: testAccCheckDatadogSecurityFindingsDueDateRuleConfigMinimal(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsDueDateRuleExists(providers.frameworkProvider, resourceName),
				),
			},
			{
				// Mutate the rule out-of-band; the refreshed plan must be non-empty (drift detected).
				Config: testAccCheckDatadogSecurityFindingsDueDateRuleConfigMinimal(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccDueDateRuleRenameOutOfBand(providers.frameworkProvider, resourceName, uniq+"-drifted"),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				// Re-applying the unchanged config must restore the configured name.
				Config: testAccCheckDatadogSecurityFindingsDueDateRuleConfigMinimal(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsDueDateRuleExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", uniq),
				),
			},
		},
	})
}

// testAccCheckDatadogSecurityFindingsDueDateRuleConfigMinimal sets only the required fields. It is
// the one config reused across multiple tests; single-use configs are inlined into their TestStep.
func testAccCheckDatadogSecurityFindingsDueDateRuleConfigMinimal(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_security_findings_due_date_rule" "test" {
  name = "%s"
  rule {
    finding_types = ["misconfiguration"]
  }
  action {
    due_from = "first_seen"
    due_days_per_severity {
      severity    = "critical"
      due_in_days = 7
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

// testAccDueDateRuleDeleteOutOfBand deletes the named rule directly through the API, simulating an
// external deletion that Terraform should detect as drift.
func testAccDueDateRuleDeleteOutOfBand(accProvider *fwprovider.FrameworkProvider, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		api := accProvider.DatadogApiInstances.GetSecurityMonitoringApiV2()
		auth := accProvider.Auth

		r, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		id, err := uuid.Parse(r.Primary.ID)
		if err != nil {
			return fmt.Errorf("invalid due date rule ID %s: %w", r.Primary.ID, err)
		}
		if httpResp, err := api.DeleteSecurityFindingsAutomationDueDateRule(auth, id); err != nil {
			return utils.TranslateClientError(err, httpResp, "error deleting due date rule out-of-band")
		}
		return nil
	}
}

// testAccDueDateRuleRenameOutOfBand renames the named rule directly through the API, simulating an
// external modification that Terraform should detect as drift.
func testAccDueDateRuleRenameOutOfBand(accProvider *fwprovider.FrameworkProvider, resourceName, newName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		api := accProvider.DatadogApiInstances.GetSecurityMonitoringApiV2()
		auth := accProvider.Auth

		r, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		id, err := uuid.Parse(r.Primary.ID)
		if err != nil {
			return fmt.Errorf("invalid due date rule ID %s: %w", r.Primary.ID, err)
		}

		current, httpResp, err := api.GetSecurityFindingsAutomationDueDateRule(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving due date rule for out-of-band update")
		}
		currentData := current.GetData()
		attrs := currentData.GetAttributes()

		updateAttrs := datadogV2.NewDueDateRuleAttributesCreateWithDefaults()
		updateAttrs.SetName(newName)
		updateAttrs.SetEnabled(attrs.GetEnabled())
		updateAttrs.SetRule(attrs.GetRule())
		updateAttrs.SetAction(attrs.GetAction())

		data := datadogV2.NewDueDateRuleDataCreateWithDefaults()
		data.SetType(datadogV2.DUEDATERULETYPE_DUE_DATE_RULES)
		data.SetAttributes(*updateAttrs)

		body := datadogV2.NewDueDateRuleUpdateRequestWithDefaults()
		body.SetData(*data)

		if _, httpResp, err := api.UpdateSecurityFindingsAutomationDueDateRule(auth, id, *body); err != nil {
			return utils.TranslateClientError(err, httpResp, "error updating due date rule out-of-band")
		}
		return nil
	}
}
