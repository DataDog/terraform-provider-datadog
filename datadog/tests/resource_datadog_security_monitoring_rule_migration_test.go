package test

import (
	"context"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// validateAttrLine matches a `validate = true|false` line (any indentation /
// spacing) so it can be stripped from configs used in migration tests.
var validateAttrLine = regexp.MustCompile(`(?m)^\s*validate\s*=\s*(true|false)\s*\n`)

// stripValidateAttr removes `validate = ...` from a resource config. `validate`
// is a write-only dry-run flag that the API never persists; SDKv2 v4.5.0 stored
// it in state as null (via DiffSuppressFunc), so keeping it in migration
// configs produces a spurious null→true diff on the FW step. Real users don't
// leave `validate = true` across applies either.
func stripValidateAttr(config string) string {
	return validateAttrLine.ReplaceAllString(config, "")
}

// Migration tests verify that upgrading from the last SDKv2 release (v4.5.0) to this
// plugin-framework version produces no unexpected plan diff.
//
// These tests use ExternalProviders to spin up v4.5.0 from the Terraform registry as a
// real subprocess. That subprocess has its own HTTP client and bypasses the in-process VCR
// recorder, so these tests are only runnable in RECORD mode (real Datadog API credentials
// required). They are skipped automatically in replay/CI.
func skipUnlessMigrationRecording(t *testing.T) {
	t.Helper()
	if !isRecording() {
		t.Skip("migration tests require a live Datadog API via RECORD=true (ExternalProviders subprocess bypasses VCR)")
	}
}

// runSecurityMonitoringRuleMigrationTest proves that upgrading from v4.5.0 to the current
// plugin-framework version produces no plan diff.
//
//   - Step 1: v4.5.0 applies the config (ExternalProvider subprocess); check asserts the
//     state the old provider wrote, so a regression in v4.5.0 output is also caught.
//   - Step 2: FW provider, same config — plan must be empty (no unexpected migration diff).
func runSecurityMonitoringRuleMigrationTest(t *testing.T, config string, check resource.TestCheckFunc) {
	t.Helper()
	skipUnlessMigrationRecording(t)
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	config = stripValidateAttr(config)

	resource.Test(t, resource.TestCase{
		CheckDestroy: testAccCheckDatadogSecurityMonitoringRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"datadog": {
						VersionConstraint: "4.5.0",
						Source:            "DataDog/datadog",
					},
				},
				Config: config,
				Check:  check,
			},
			{
				ProtoV6ProviderFactories: accProviders,
				Config:                   config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccDatadogSecurityMonitoringRule_Migration_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, _ := testAccFrameworkMuxProviders(context.Background(), t)
	ruleName := uniqueEntityName(ctx, t)
	runSecurityMonitoringRuleMigrationTest(t,
		testAccCheckDatadogSecurityMonitoringCreatedConfig(ruleName),
		testAccCheckDatadogSecurityMonitorCreatedCheck(providers.frameworkProvider, ruleName),
	)
}

func TestAccDatadogSecurityMonitoringRule_Migration_SignalCorrelation(t *testing.T) {
	t.Parallel()
	ctx, providers, _ := testAccFrameworkMuxProviders(context.Background(), t)
	ruleName := uniqueEntityName(ctx, t)
	runSecurityMonitoringRuleMigrationTest(t,
		testAccCheckDatadogSecurityMonitoringCreatedSignalCorrelationConfig(ruleName),
		testAccCheckDatadogSecurityMonitorCreatedSignalCorrelationCheck(providers.frameworkProvider, ruleName),
	)
}

func TestAccDatadogSecurityMonitoringRule_Migration_CWS(t *testing.T) {
	t.Parallel()
	ctx, providers, _ := testAccFrameworkMuxProviders(context.Background(), t)
	ruleName := uniqueEntityName(ctx, t)
	runSecurityMonitoringRuleMigrationTest(t,
		testAccCheckDatadogSecurityMonitoringCreatedConfigCwsRule(ruleName),
		testAccCheckDatadogSecurityMonitoringCreatedCheckCwsRule(providers.frameworkProvider, ruleName),
	)
}

func TestAccDatadogSecurityMonitoringRule_Migration_ThirdParty(t *testing.T) {
	t.Parallel()
	ctx, providers, _ := testAccFrameworkMuxProviders(context.Background(), t)
	ruleName := uniqueEntityName(ctx, t)
	runSecurityMonitoringRuleMigrationTest(t,
		testAccCheckDatadogSecurityMonitoringCreatedThirdPartyConfig(ruleName),
		testAccCheckDatadogSecurityMonitoringCreatedThirdPartyCheck(providers.frameworkProvider, ruleName),
	)
}

func TestAccDatadogSecurityMonitoringRule_Migration_NewValue(t *testing.T) {
	t.Parallel()
	ctx, providers, _ := testAccFrameworkMuxProviders(context.Background(), t)
	ruleName := uniqueEntityName(ctx, t)
	runSecurityMonitoringRuleMigrationTest(t,
		testAccCheckDatadogSecurityMonitoringCreatedConfigNewValueRule(ruleName),
		testAccCheckDatadogSecurityMonitorCreatedCheckNewValueRule(providers.frameworkProvider, ruleName),
	)
}

func TestAccDatadogSecurityMonitoringRule_Migration_ImpossibleTravel(t *testing.T) {
	t.Parallel()
	ctx, providers, _ := testAccFrameworkMuxProviders(context.Background(), t)
	ruleName := uniqueEntityName(ctx, t)
	runSecurityMonitoringRuleMigrationTest(t,
		testAccCheckDatadogSecurityMonitoringCreatedConfigImpossibleTravelRule(ruleName),
		testAccCheckDatadogSecurityMonitorCreatedCheckImpossibleTravelRule(providers.frameworkProvider, ruleName),
	)
}

// TestAccDatadogSecurityMonitoringRule_Migration_DefaultTags verifies the documented one-time
// upgrade diff: when a user upgrades and the provider has default_tags configured, the first
// plan after upgrade will show effective_tags changing to include the default tags. After one
// apply the plan stabilizes to empty.
func TestAccDatadogSecurityMonitoringRule_Migration_DefaultTags(t *testing.T) {
	t.Parallel()
	skipUnlessMigrationRecording(t)

	ctx, providers, _ := testAccFrameworkMuxProviders(context.Background(), t)
	ruleName := uniqueEntityName(ctx, t)
	config := stripValidateAttr(testAccCheckDatadogSecurityMonitoringDefaultTags(ruleName))

	defaultTags := map[string]string{"default_key": "default_value"}
	fwWithDefaults := func() map[string]func() (tfprotov6.ProviderServer, error) {
		return map[string]func() (tfprotov6.ProviderServer, error){
			"datadog": withDefaultTagsFw(ctx, providers, defaultTags),
		}
	}

	resource.Test(t, resource.TestCase{
		CheckDestroy: testAccCheckDatadogSecurityMonitoringRuleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				// v4.5.0 creates the resource; no default_tags yet, tags is user-only
				ExternalProviders: map[string]resource.ExternalProvider{
					"datadog": {
						VersionConstraint: "4.6.0",
						Source:            "DataDog/datadog",
					},
				},
				Config: config,
			},
			{
				// FW provider with default_tags — expected one-time diff on effective_tags.
				// Apply it so the state absorbs the default tags, then verify stabilization.
				ProtoV6ProviderFactories: fwWithDefaults(),
				Config:                   config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_security_monitoring_rule.acceptance_test", "effective_tags.#", "3"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_security_monitoring_rule.acceptance_test", "effective_tags.*", "foo:bar"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_security_monitoring_rule.acceptance_test", "effective_tags.*", "baz"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_security_monitoring_rule.acceptance_test", "effective_tags.*", "default_key:default_value"),
					resource.TestCheckResourceAttr(
						"datadog_security_monitoring_rule.acceptance_test", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_security_monitoring_rule.acceptance_test", "tags.*", "foo:bar"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_security_monitoring_rule.acceptance_test", "tags.*", "baz"),
				),
			},
			{
				// After the apply above, plan should be empty (state stabilized)
				ProtoV6ProviderFactories: fwWithDefaults(),
				Config:                   config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}
