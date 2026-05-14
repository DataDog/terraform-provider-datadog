package test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// Pin the provider source explicitly so step transitions across
// ExternalProviders / ProtoV6ProviderFactories don't default-resolve
// `datadog` to `hashicorp/datadog` and trip the lock file.
const defaultRuleMigrationProviderBlock = `
terraform {
  required_providers {
    datadog = {
      source = "DataDog/datadog"
    }
  }
}
`

// runDefaultRuleMigrationTest verifies that upgrading a default rule's state
// from the last SDKv2 release (v4.7.0) to the framework version produces no
// plan diff for the given config.
//
// The default rule is import-only. The test imports an existing default rule
// under v4.7.0, applies applyConfig under v4.7.0, then re-applies the same
// config under the framework provider with ExpectEmptyPlan() — the empirical
// proof of HashiCorp's "no behavior change" migration invariant.
//
// Skipped unless RECORD=true: the v4.7.0 ExternalProvider runs in its own
// subprocess with its own HTTP client, bypassing the in-process VCR recorder.
func runDefaultRuleMigrationTest(t *testing.T, applyConfig string, applyCheck resource.TestCheckFunc) {
	t.Helper()
	if !isRecording() {
		t.Skip("migration tests require a live Datadog API via RECORD=true (ExternalProviders subprocess bypasses VCR)")
	}

	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	v4_7_0 := map[string]resource.ExternalProvider{
		"datadog": {
			VersionConstraint: "4.7.0",
			Source:            "DataDog/datadog",
		},
	}

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// v4.7.0: discover an existing default rule via datasource
			{
				ExternalProviders: v4_7_0,
				Config:            defaultRuleMigrationProviderBlock + testAccDatadogSecurityMonitoringDefaultDatasource(),
			},
			// v4.7.0: import the rule
			{
				ExternalProviders:  v4_7_0,
				Config:             defaultRuleMigrationProviderBlock + testAccCheckDatadogSecurityMonitoringDefaultNoop(),
				ResourceName:       tfSecurityDefaultRuleName,
				ImportState:        true,
				ImportStateIdFunc:  idFromDatasource,
				ImportStatePersist: true,
			},
			// v4.7.0: apply the scenario config
			{
				ExternalProviders: v4_7_0,
				Config:            defaultRuleMigrationProviderBlock + applyConfig,
				Check:             applyCheck,
			},
			// FW: same config — plan must be empty (no migration drift)
			{
				ProtoV6ProviderFactories: accProviders,
				Config:                   applyConfig,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// FW: restore to base config so subsequent re-recording sessions
			// start from a clean rule state.
			{
				ProtoV6ProviderFactories: accProviders,
				Config:                   testAccCheckDatadogSecurityMonitoringDefaultNoop(),
			},
		},
	})
}

func TestAccDatadogSecurityMonitoringDefaultRule_Migration_Basic(t *testing.T) {
	runDefaultRuleMigrationTest(t,
		testAccDatadogSecurityMonitoringDefaultRuleDynamicCriticality(),
		testAccCheckDatadogSecurityMonitoringDefaultDynamicCriticality(),
	)
}

func TestAccDatadogSecurityMonitoringDefaultRule_Migration_AddTag(t *testing.T) {
	runDefaultRuleMigrationTest(t,
		testAccDatadogSecurityMonitoringDefaultRuleAddTag(),
		testAccCheckDatadogSecurityMonitoringDefaultRuleAddTag(),
	)
}

func TestAccDatadogSecurityMonitoringDefaultRule_Migration_CustomMessage(t *testing.T) {
	runDefaultRuleMigrationTest(t,
		testAccDatadogSecurityMonitoringDefaultRuleCustomMessage(),
		testAccCheckDatadogSecurityMonitoringDefaultCustomMessage(),
	)
}

func TestAccDatadogSecurityMonitoringDefaultRule_Migration_CustomName(t *testing.T) {
	runDefaultRuleMigrationTest(t,
		testAccDatadogSecurityMonitoringDefaultRuleCustomName(),
		testAccCheckDatadogSecurityMonitoringDefaultCustomName(),
	)
}

func TestAccDatadogSecurityMonitoringDefaultRule_Migration_Enabled(t *testing.T) {
	runDefaultRuleMigrationTest(t,
		testAccDatadogSecurityMonitoringDefaultRuleEnabled(),
		testAccCheckDatadogSecurityMonitoringDefaultEnabled(),
	)
}

func TestAccDatadogSecurityMonitoringDefaultRule_Migration_CustomStatus(t *testing.T) {
	runDefaultRuleMigrationTest(t,
		testAccDatadogSecurityMonitoringDefaultRuleCustomStatus(),
		testAccCheckDatadogSecurityMonitoringDefaultCustomStatus(),
	)
}

func TestAccDatadogSecurityMonitoringDefaultRule_Migration_CustomQueryExtension(t *testing.T) {
	runDefaultRuleMigrationTest(t,
		testAccDatadogSecurityMonitoringDefaultRuleCustomQueryExtension(),
		testAccCheckDatadogSecurityMonitoringDefaultCustomQueryExtension(),
	)
}
