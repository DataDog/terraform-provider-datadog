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

// TestAccDatadogSecurityFindingsMuteRule covers the basic create -> update -> import lifecycle,
// exercising changes to every attribute (name, enabled, finding_types, query, reason, expire_at).
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
				Config: fmt.Sprintf(`
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
`, uniq),
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
				Config: fmt.Sprintf(`
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
`, uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsMuteRuleExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", uniq+"-updated"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "rule.finding_types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "action.reason", "false_positive"),
					resource.TestCheckResourceAttr(resourceName, "action.expire_at", "4070908800000"),
					// reason_description was set on create and dropped here: it must clear.
					resource.TestCheckNoResourceAttr(resourceName, "action.reason_description"),
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

// TestAccDatadogSecurityFindingsMuteRule_Minimal configures only the required fields and verifies
// the computed default for `enabled` (true) and that omitted optionals are absent from state.
func TestAccDatadogSecurityFindingsMuteRule_Minimal(t *testing.T) {
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
				Config: testAccCheckDatadogSecurityFindingsMuteRuleConfigMinimal(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsMuteRuleExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", uniq),
					// `enabled` is omitted from config: the schema default must apply.
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "rule.finding_types.#", "1"),
					resource.TestCheckNoResourceAttr(resourceName, "rule.query"),
					resource.TestCheckResourceAttr(resourceName, "action.reason", "no_fix"),
					resource.TestCheckNoResourceAttr(resourceName, "action.reason_description"),
					resource.TestCheckNoResourceAttr(resourceName, "action.expire_at"),
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

// TestAccDatadogSecurityFindingsMuteRule_OptionalChurn sets every optional field, then removes
// them all, asserting they clear cleanly with no perpetual diff.
func TestAccDatadogSecurityFindingsMuteRule_OptionalChurn(t *testing.T) {
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
				// All optionals present.
				Config: fmt.Sprintf(`
resource "datadog_security_findings_mute_rule" "test" {
  name    = "%s"
  enabled = false
  rule {
    finding_types = ["misconfiguration"]
    query         = "env:prod"
  }
  action {
    reason             = "no_fix"
    reason_description = "context"
    expire_at          = 4070908800000
  }
}
`, uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsMuteRuleExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "rule.query", "env:prod"),
					resource.TestCheckResourceAttr(resourceName, "action.reason_description", "context"),
					resource.TestCheckResourceAttr(resourceName, "action.expire_at", "4070908800000"),
				),
			},
			{
				// All optionals removed: they must clear, and the plan must be empty afterwards.
				Config: testAccCheckDatadogSecurityFindingsMuteRuleConfigMinimal(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsMuteRuleExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckNoResourceAttr(resourceName, "rule.query"),
					resource.TestCheckNoResourceAttr(resourceName, "action.reason_description"),
					resource.TestCheckNoResourceAttr(resourceName, "action.expire_at"),
				),
			},
		},
	})
}

// TestAccDatadogSecurityFindingsMuteRule_Validation exercises the schema validators. Each step
// fails at plan time (before any API call), so these are independent of the rule's runtime behavior.
func TestAccDatadogSecurityFindingsMuteRule_Validation(t *testing.T) {
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
resource "datadog_security_findings_mute_rule" "test" {
  name = "%s"
  rule {
    finding_types = []
  }
  action {
    reason = "no_fix"
  }
}
`, uniq),
				ExpectError: regexp.MustCompile("must contain at least 1"),
			},
			{
				// finding_types value must be a valid enum member.
				Config: fmt.Sprintf(`
resource "datadog_security_findings_mute_rule" "test" {
  name = "%s"
  rule {
    finding_types = ["not_a_finding_type"]
  }
  action {
    reason = "no_fix"
  }
}
`, uniq),
				ExpectError: regexp.MustCompile("invalid value"),
			},
			{
				// reason value must be a valid enum member.
				Config: fmt.Sprintf(`
resource "datadog_security_findings_mute_rule" "test" {
  name = "%s"
  rule {
    finding_types = ["misconfiguration"]
  }
  action {
    reason = "not_a_reason"
  }
}
`, uniq),
				ExpectError: regexp.MustCompile("invalid value"),
			},
			{
				// The rule block is required.
				Config: fmt.Sprintf(`
resource "datadog_security_findings_mute_rule" "test" {
  name = "%s"
  action {
    reason = "no_fix"
  }
}
`, uniq),
				ExpectError: regexp.MustCompile("must have a configuration value"),
			},
			{
				// The action block is required.
				Config: fmt.Sprintf(`
resource "datadog_security_findings_mute_rule" "test" {
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

// TestAccDatadogSecurityFindingsMuteRule_OutOfBandDelete deletes the rule directly through the API
// and asserts Terraform detects it as drift (a non-empty plan that would recreate it).
func TestAccDatadogSecurityFindingsMuteRule_OutOfBandDelete(t *testing.T) {
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
				Config: testAccCheckDatadogSecurityFindingsMuteRuleConfigMinimal(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsMuteRuleExists(providers.frameworkProvider, resourceName),
					// Let the create settle, then delete the rule out-of-band; the follow-up plan must be non-empty.
					testAccMuteRuleDeleteOutOfBand(providers.frameworkProvider, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccDatadogSecurityFindingsMuteRule_OutOfBandUpdate renames the rule directly through the API
// and asserts Terraform detects the drift and restores the configured value on the next apply.
func TestAccDatadogSecurityFindingsMuteRule_OutOfBandUpdate(t *testing.T) {
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
				Config: testAccCheckDatadogSecurityFindingsMuteRuleConfigMinimal(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsMuteRuleExists(providers.frameworkProvider, resourceName),
				),
			},
			{
				// Mutate the rule out-of-band; the refreshed plan must be non-empty (drift detected).
				Config: testAccCheckDatadogSecurityFindingsMuteRuleConfigMinimal(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccMuteRuleRenameOutOfBand(providers.frameworkProvider, resourceName, uniq+"-drifted"),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				// Re-applying the unchanged config must restore the configured name.
				Config: testAccCheckDatadogSecurityFindingsMuteRuleConfigMinimal(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsMuteRuleExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", uniq),
				),
			},
		},
	})
}

// testAccCheckDatadogSecurityFindingsMuteRuleConfigMinimal sets only the required fields. It is
// the one config reused across multiple tests; single-use configs are inlined into their TestStep.
func testAccCheckDatadogSecurityFindingsMuteRuleConfigMinimal(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_security_findings_mute_rule" "test" {
  name = "%s"
  rule {
    finding_types = ["misconfiguration"]
  }
  action {
    reason = "no_fix"
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

// testAccMuteRuleDeleteOutOfBand deletes the named rule directly through the API, simulating an
// external deletion that Terraform should detect as drift.
func testAccMuteRuleDeleteOutOfBand(accProvider *fwprovider.FrameworkProvider, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		api := accProvider.DatadogApiInstances.GetSecurityMonitoringApiV2()
		auth := accProvider.Auth

		r, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		id, err := uuid.Parse(r.Primary.ID)
		if err != nil {
			return fmt.Errorf("invalid mute rule ID %s: %w", r.Primary.ID, err)
		}
		if httpResp, err := api.DeleteSecurityFindingsAutomationMuteRule(auth, id); err != nil {
			return utils.TranslateClientError(err, httpResp, "error deleting mute rule out-of-band")
		}
		return nil
	}
}

// testAccMuteRuleRenameOutOfBand renames the named rule directly through the API, simulating an
// external modification that Terraform should detect as drift.
func testAccMuteRuleRenameOutOfBand(accProvider *fwprovider.FrameworkProvider, resourceName, newName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		api := accProvider.DatadogApiInstances.GetSecurityMonitoringApiV2()
		auth := accProvider.Auth

		r, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		id, err := uuid.Parse(r.Primary.ID)
		if err != nil {
			return fmt.Errorf("invalid mute rule ID %s: %w", r.Primary.ID, err)
		}

		current, httpResp, err := api.GetSecurityFindingsAutomationMuteRule(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving mute rule for out-of-band update")
		}
		currentData := current.GetData()
		attrs := currentData.GetAttributes()

		updateAttrs := datadogV2.NewMuteRuleAttributesCreateWithDefaults()
		updateAttrs.SetName(newName)
		updateAttrs.SetEnabled(attrs.GetEnabled())
		updateAttrs.SetRule(attrs.GetRule())
		updateAttrs.SetAction(attrs.GetAction())

		data := datadogV2.NewMuteRuleDataCreateWithDefaults()
		data.SetType(datadogV2.MUTERULETYPE_MUTE_RULES)
		data.SetAttributes(*updateAttrs)

		body := datadogV2.NewMuteRuleUpdateRequestWithDefaults()
		body.SetData(*data)

		if _, httpResp, err := api.UpdateSecurityFindingsAutomationMuteRule(auth, id, *body); err != nil {
			return utils.TranslateClientError(err, httpResp, "error updating mute rule out-of-band")
		}
		return nil
	}
}
