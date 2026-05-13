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

// TestAccDatadogSecurityMonitoringDefaultRule_Migration_Basic verifies that
// upgrading a default rule's state from the last SDKv2 release (v4.7.0) to
// the framework version produces no plan diff.
//
// The default rule is import-only: there is no Create. The test imports an
// existing default rule under v4.7.0, applies a non-trivial config under
// v4.7.0, then re-applies the same config under the framework provider with
// ExpectEmptyPlan() — the empirical proof of HashiCorp's "no behavior change"
// migration invariant.
//
// Skipped unless RECORD=true: the v4.7.0 ExternalProvider runs in its own
// subprocess with its own HTTP client, bypassing the in-process VCR recorder.
func TestAccDatadogSecurityMonitoringDefaultRule_Migration_Basic(t *testing.T) {
	if !isRecording() {
		t.Skip("migration tests require a live Datadog API via RECORD=true (ExternalProviders subprocess bypasses VCR)")
	}

	t.Parallel()
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
			// v4.7.0: apply non-trivial config (toggle decrease_criticality + add a custom tag)
			{
				ExternalProviders: v4_7_0,
				Config:            defaultRuleMigrationProviderBlock + testAccDatadogSecurityMonitoringDefaultRuleDynamicCriticality(),
				Check:             testAccCheckDatadogSecurityMonitoringDefaultDynamicCriticality(),
			},
			// FW: same config — plan must be empty (no migration drift)
			{
				ProtoV6ProviderFactories: accProviders,
				Config:                   defaultRuleMigrationProviderBlock + testAccDatadogSecurityMonitoringDefaultRuleDynamicCriticality(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}
