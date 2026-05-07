package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// runSecurityMonitoringFilterMigrationTest proves that upgrading from v4.7.0 to the
// current plugin-framework version produces no plan diff.
//
//   - Step 1: v4.7.0 applies the config (ExternalProvider subprocess); check asserts the
//     state the old SDKv2 provider wrote.
//   - Step 2: FW provider, same config — plan must be empty.
//
// The ExternalProvider subprocess has its own HTTP client and bypasses VCR, so this test
// is RECORD-only.
func runSecurityMonitoringFilterMigrationTest(t *testing.T, config string, check resource.TestCheckFunc) {
	t.Helper()
	skipUnlessMigrationRecording(t)
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		CheckDestroy: testAccCheckDatadogSecurityMonitoringFilterDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"datadog": {
						VersionConstraint: "4.7.0",
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

func TestAccDatadogSecurityMonitoringFilter_Migration_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, _ := testAccFrameworkMuxProviders(context.Background(), t)
	filterName := uniqueEntityName(ctx, t)
	runSecurityMonitoringFilterMigrationTest(t,
		testAccCheckDatadogSecurityMonitoringFilterCreated(filterName),
		testAccCheckDatadogSecurityMonitorFilterCreatedCheck(providers.frameworkProvider, filterName),
	)
}

// TestAccDatadogSecurityMonitoringFilter_Migration_NoExclusionFilters covers the
// no-`exclusion_filter`-block path. SDKv2 always sends an empty slice on PUT (resource
// line 250); the FW reader must produce identical state so the FW step plans empty.
func TestAccDatadogSecurityMonitoringFilter_Migration_NoExclusionFilters(t *testing.T) {
	t.Parallel()
	ctx, providers, _ := testAccFrameworkMuxProviders(context.Background(), t)
	filterName := uniqueEntityName(ctx, t)
	config := testAccCheckDatadogSecurityMonitoringFilterNoExclusion(filterName)
	check := resource.ComposeTestCheckFunc(
		testAccCheckDatadogSecurityMonitoringFilterExists(providers.frameworkProvider),
		resource.TestCheckResourceAttr(tfSecurityFilterName, "name", filterName),
		resource.TestCheckResourceAttr(tfSecurityFilterName, "is_enabled", "true"),
		resource.TestCheckResourceAttr(tfSecurityFilterName, "exclusion_filter.#", "0"),
	)
	runSecurityMonitoringFilterMigrationTest(t, config, check)
}

// TestAccDatadogSecurityMonitoringFilter_Migration_ExplicitFilteredDataType covers
// `filtered_data_type = "logs"` set explicitly — the user-set-equals-default path that
// commonly drifts on Optional+Computed migrations.
func TestAccDatadogSecurityMonitoringFilter_Migration_ExplicitFilteredDataType(t *testing.T) {
	t.Parallel()
	ctx, providers, _ := testAccFrameworkMuxProviders(context.Background(), t)
	filterName := uniqueEntityName(ctx, t)
	config := testAccCheckDatadogSecurityMonitoringFilterExplicitDataType(filterName)
	check := resource.ComposeTestCheckFunc(
		testAccCheckDatadogSecurityMonitoringFilterExists(providers.frameworkProvider),
		resource.TestCheckResourceAttr(tfSecurityFilterName, "filtered_data_type", "logs"),
	)
	runSecurityMonitoringFilterMigrationTest(t, config, check)
}

// TestAccDatadogSecurityMonitoringFilter_Migration_DefaultFilteredDataType covers
// `filtered_data_type` omitted from config — the FW Default must materialize as `"logs"`
// so post-upgrade plan stays empty against the v4.7.0-recorded state.
func TestAccDatadogSecurityMonitoringFilter_Migration_DefaultFilteredDataType(t *testing.T) {
	t.Parallel()
	ctx, providers, _ := testAccFrameworkMuxProviders(context.Background(), t)
	filterName := uniqueEntityName(ctx, t)
	config := testAccCheckDatadogSecurityMonitoringFilterNoExclusion(filterName)
	check := resource.ComposeTestCheckFunc(
		testAccCheckDatadogSecurityMonitoringFilterExists(providers.frameworkProvider),
		resource.TestCheckResourceAttr(tfSecurityFilterName, "filtered_data_type", "logs"),
	)
	runSecurityMonitoringFilterMigrationTest(t, config, check)
}

func testAccCheckDatadogSecurityMonitoringFilterNoExclusion(name string) string {
	return fmt.Sprintf(`
resource "datadog_security_monitoring_filter" "acceptance_test" {
    name = "%s"
    query = "no exclusion - %[1]s"
    is_enabled = true
}
`, name)
}

func testAccCheckDatadogSecurityMonitoringFilterExplicitDataType(name string) string {
	return fmt.Sprintf(`
resource "datadog_security_monitoring_filter" "acceptance_test" {
    name = "%s"
    query = "explicit data type - %[1]s"
    is_enabled = true
    filtered_data_type = "logs"

    exclusion_filter {
        name = "only"
        query = "noisy logs"
    }
}
`, name)
}
