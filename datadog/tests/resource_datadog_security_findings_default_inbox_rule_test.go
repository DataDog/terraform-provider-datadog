package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

// Default inbox rules always exist server-side, so the acceptance tests operate on the real
// `secret_default_rule` rather than creating resources. They mutate the org's enabled flag for
// that rule and always restore it to `true` (the server-side default) at the end, so they must not
// run in parallel with each other.

// TestAccDatadogSecurityFindingsDefaultInboxRule_Import imports an existing default rule with no
// attributes in config (adopting its current state), then manages `enabled`: declaring
// `enabled = false` disables the rule through the toggle endpoint, and `enabled = true` restores it.
func TestAccDatadogSecurityFindingsDefaultInboxRule_Import(t *testing.T) {
	// No t.Parallel(): the test mutates the org-wide enabled flag of a persistent default rule.
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	resourceName := "datadog_security_findings_default_inbox_rule.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				// Import the existing rule; the config declares the resource with no attributes, so
				// `enabled` adopts the imported (server-side) value and the following plan is clean.
				Config: `
resource "datadog_security_findings_default_inbox_rule" "test" {
}
`,
				ResourceName:       resourceName,
				ImportState:        true,
				ImportStateId:      "secret_default_rule",
				ImportStatePersist: true,
				ImportStateVerify:  true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsDefaultInboxRuleExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "id", "secret_default_rule"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "rule.finding_types.#"),
					resource.TestCheckResourceAttrSet(resourceName, "action.description"),
				),
			},
			{
				// Declare `enabled = false`: the update must disable the rule server-side.
				Config: `
resource "datadog_security_findings_default_inbox_rule" "test" {
  enabled = false
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsDefaultInboxRuleExists(providers.frameworkProvider, resourceName),
					testAccCheckDatadogSecurityFindingsDefaultInboxRuleEnabled(providers.frameworkProvider, resourceName, false),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
				),
			},
			{
				// Declare `enabled = true`: the rule is restored to the server-side default state.
				Config: `
resource "datadog_security_findings_default_inbox_rule" "test" {
  enabled = true
}
`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogSecurityFindingsDefaultInboxRuleExists(providers.frameworkProvider, resourceName),
					testAccCheckDatadogSecurityFindingsDefaultInboxRuleEnabled(providers.frameworkProvider, resourceName, true),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
		},
	})
}

// TestAccDatadogSecurityFindingsDefaultInboxRule_CreateFails asserts that applying the resource
// without importing it first fails with the import instructions: default rules cannot be created.
func TestAccDatadogSecurityFindingsDefaultInboxRule_CreateFails(t *testing.T) {
	// No t.Parallel(): consistent with the other default inbox rule tests.
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: `
resource "datadog_security_findings_default_inbox_rule" "test" {
  enabled = false
}
`,
				ExpectError: regexp.MustCompile("cannot be created"),
			},
		},
	})
}

func testAccCheckDatadogSecurityFindingsDefaultInboxRuleExists(accProvider *fwprovider.FrameworkProvider, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		r, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		// The default rule ID is a fixed string (e.g. "secret_default_rule"), not a UUID.
		_, httpResp, err := apiInstances.GetSecurityMonitoringApiV2().GetSecurityFindingsAutomationDefaultInboxRule(auth, r.Primary.ID)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving default inbox rule")
		}
		return nil
	}
}

// testAccCheckDatadogSecurityFindingsDefaultInboxRuleEnabled verifies the rule's live enabled flag
// directly through the API, guarding that the toggle endpoints were actually called.
func testAccCheckDatadogSecurityFindingsDefaultInboxRuleEnabled(accProvider *fwprovider.FrameworkProvider, resourceName string, wantEnabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		r, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		resp, httpResp, err := apiInstances.GetSecurityMonitoringApiV2().GetSecurityFindingsAutomationDefaultInboxRule(auth, r.Primary.ID)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving default inbox rule")
		}
		data := resp.GetData()
		attrs := data.GetAttributes()
		if got := attrs.GetEnabled(); got != wantEnabled {
			return fmt.Errorf("default inbox rule %s: expected enabled=%t, got %t", r.Primary.ID, wantEnabled, got)
		}
		return nil
	}
}
