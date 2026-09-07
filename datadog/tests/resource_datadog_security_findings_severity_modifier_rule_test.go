package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

// TestAccDatadogSecurityFindingsSeverityModifierRule covers the basic create -> update -> import
// lifecycle, exercising changes to every attribute (name, enabled, finding_types, query, and
// switching the action between `set` and `shift`).
func TestAccDatadogSecurityFindingsSeverityModifierRule(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	resourceName := "datadog_security_findings_severity_modifier_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogSecurityFindingsSeverityModifierRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "datadog_security_findings_severity_modifier_rule" "test" {
  name    = "%s"
  enabled = true
  rule = {
    finding_types = ["secret"]
    query         = "env:prod"
  }
  action = {
    set = {
      severity    = "critical"
      description = "Secrets in production are always critical"
    }
  }
}
`, uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsSeverityModifierRuleExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", uniq),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "rule.finding_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.finding_types.0", "secret"),
					resource.TestCheckResourceAttr(resourceName, "rule.query", "env:prod"),
					resource.TestCheckResourceAttr(resourceName, "action.set.severity", "critical"),
					resource.TestCheckResourceAttr(resourceName, "action.set.description", "Secrets in production are always critical"),
					resource.TestCheckNoResourceAttr(resourceName, "action.shift"),
				),
			},
			{
				// Switch the action from `set` to `shift`.
				Config: fmt.Sprintf(`
resource "datadog_security_findings_severity_modifier_rule" "test" {
  name    = "%s-updated"
  enabled = false
  rule = {
    finding_types = ["misconfiguration", "secret"]
    query         = "env:dev"
  }
  action = {
    shift = {
      severity_delta = "down_one"
    }
  }
}
`, uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsSeverityModifierRuleExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", uniq+"-updated"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "rule.finding_types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "action.shift.severity_delta", "down_one"),
					resource.TestCheckNoResourceAttr(resourceName, "action.shift.description"),
					resource.TestCheckNoResourceAttr(resourceName, "action.set"),
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

// TestAccDatadogSecurityFindingsSeverityModifierRule_Minimal configures only the required fields
// and verifies the computed default for `enabled` (true) and that omitted optionals are absent
// from state.
func TestAccDatadogSecurityFindingsSeverityModifierRule_Minimal(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	resourceName := "datadog_security_findings_severity_modifier_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogSecurityFindingsSeverityModifierRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSecurityFindingsSeverityModifierRuleConfigMinimal(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsSeverityModifierRuleExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", uniq),
					// `enabled` is omitted from config: the schema default must apply.
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "rule.finding_types.#", "1"),
					resource.TestCheckNoResourceAttr(resourceName, "rule.query"),
					resource.TestCheckResourceAttr(resourceName, "action.set.severity", "high"),
					resource.TestCheckNoResourceAttr(resourceName, "action.set.description"),
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

// TestAccDatadogSecurityFindingsSeverityModifierRule_Validation exercises the schema validators.
// Each step fails at plan time (before any API call), so these are independent of the rule's
// runtime behavior.
func TestAccDatadogSecurityFindingsSeverityModifierRule_Validation(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				// Exactly one of `set`/`shift` must be provided: neither is an error.
				Config: fmt.Sprintf(`
resource "datadog_security_findings_severity_modifier_rule" "test" {
  name = "%s"
  rule = {
    finding_types = ["secret"]
  }
  action = {}
}
`, uniq),
				ExpectError: regexp.MustCompile("Missing Attribute Configuration|Invalid Attribute Combination"),
			},
			{
				// Exactly one of `set`/`shift` must be provided: both is an error.
				Config: fmt.Sprintf(`
resource "datadog_security_findings_severity_modifier_rule" "test" {
  name = "%s"
  rule = {
    finding_types = ["secret"]
  }
  action = {
    set = {
      severity = "high"
    }
    shift = {
      severity_delta = "down_one"
    }
  }
}
`, uniq),
				ExpectError: regexp.MustCompile("Invalid Attribute Combination"),
			},
			{
				// severity value must be a valid enum member.
				Config: fmt.Sprintf(`
resource "datadog_security_findings_severity_modifier_rule" "test" {
  name = "%s"
  rule = {
    finding_types = ["secret"]
  }
  action = {
    set = {
      severity = "not_a_severity"
    }
  }
}
`, uniq),
				ExpectError: regexp.MustCompile("invalid value"),
			},
			{
				// severity_delta value must be a valid enum member.
				Config: fmt.Sprintf(`
resource "datadog_security_findings_severity_modifier_rule" "test" {
  name = "%s"
  rule = {
    finding_types = ["secret"]
  }
  action = {
    shift = {
      severity_delta = "sideways"
    }
  }
}
`, uniq),
				ExpectError: regexp.MustCompile("invalid value"),
			},
			{
				// The rule attribute is required.
				Config: fmt.Sprintf(`
resource "datadog_security_findings_severity_modifier_rule" "test" {
  name = "%s"
  action = {
    set = {
      severity = "high"
    }
  }
}
`, uniq),
				ExpectError: regexp.MustCompile("is required, but no definition was found"),
			},
		},
	})
}

// TestAccDatadogSecurityFindingsSeverityModifierRule_OutOfBandDelete deletes the rule directly
// through the API and asserts Terraform detects it as drift (a non-empty plan that would recreate
// it).
func TestAccDatadogSecurityFindingsSeverityModifierRule_OutOfBandDelete(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)
	resourceName := "datadog_security_findings_severity_modifier_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogSecurityFindingsSeverityModifierRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogSecurityFindingsSeverityModifierRuleConfigMinimal(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsSeverityModifierRuleExists(providers.frameworkProvider, resourceName),
					// Let the create settle, then delete the rule out-of-band; the follow-up plan must be non-empty.
					testAccSeverityModifierRuleDeleteOutOfBand(providers.frameworkProvider, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// testAccCheckDatadogSecurityFindingsSeverityModifierRuleConfigMinimal sets only the required
// fields. It is the one config reused across multiple tests; single-use configs are inlined into
// their TestStep.
func testAccCheckDatadogSecurityFindingsSeverityModifierRuleConfigMinimal(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_security_findings_severity_modifier_rule" "test" {
  name = "%s"
  rule = {
    finding_types = ["secret"]
  }
  action = {
    set = {
      severity = "high"
    }
  }
}
`, uniq)
}

func testAccCheckDatadogSecurityFindingsSeverityModifierRuleExists(accProvider *fwprovider.FrameworkProvider, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		r, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		id, err := uuid.Parse(r.Primary.ID)
		if err != nil {
			return fmt.Errorf("invalid severity modifier rule ID %s: %w", r.Primary.ID, err)
		}
		_, httpResp, err := apiInstances.GetSecurityMonitoringApiV2().GetSecurityFindingsAutomationSeverityModifierRule(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving severity modifier rule")
		}
		return nil
	}
}

func testAccCheckDatadogSecurityFindingsSeverityModifierRuleDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_security_findings_severity_modifier_rule" {
				continue
			}
			err := utils.Retry(2, 10, func() error {
				id, parseErr := uuid.Parse(r.Primary.ID)
				if parseErr != nil {
					return &utils.RetryableError{Prob: fmt.Sprintf("invalid severity modifier rule ID %s", r.Primary.ID)}
				}
				_, httpResp, err := apiInstances.GetSecurityMonitoringApiV2().GetSecurityFindingsAutomationSeverityModifierRule(auth, id)
				if err != nil {
					if httpResp != nil && httpResp.StatusCode == 404 {
						return nil
					}
					return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving severity modifier rule %s", err)}
				}
				return &utils.RetryableError{Prob: "severity modifier rule still exists"}
			})
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// testAccSeverityModifierRuleDeleteOutOfBand deletes the named rule directly through the API,
// simulating an external deletion that Terraform should detect as drift.
func testAccSeverityModifierRuleDeleteOutOfBand(accProvider *fwprovider.FrameworkProvider, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		api := accProvider.DatadogApiInstances.GetSecurityMonitoringApiV2()
		auth := accProvider.Auth

		r, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		id, err := uuid.Parse(r.Primary.ID)
		if err != nil {
			return fmt.Errorf("invalid severity modifier rule ID %s: %w", r.Primary.ID, err)
		}
		if httpResp, err := api.DeleteSecurityFindingsAutomationSeverityModifierRule(auth, id); err != nil {
			return utils.TranslateClientError(err, httpResp, "error deleting severity modifier rule out-of-band")
		}
		return nil
	}
}
