package test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

func TestAccSecurityMonitoringCriticalAsset_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	assetName := uniqueEntityName(ctx, t)
	resourceName := "datadog_security_monitoring_critical_asset.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckSecurityMonitoringCriticalAssetDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityMonitoringCriticalAssetConfig(assetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityMonitoringCriticalAssetExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "query", "source:runtime-security-agent"),
					resource.TestCheckResourceAttr(resourceName, "rule_query", "type:(log_detection OR signal_correlation OR workload_security OR application_security) ruleId:007-d1a-1f3"),
					resource.TestCheckResourceAttr(resourceName, "severity", "increase"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "2"),
				),
			},
		},
	})
}

func TestAccSecurityMonitoringCriticalAsset_Update(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	assetName := uniqueEntityName(ctx, t)
	resourceName := "datadog_security_monitoring_critical_asset.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckSecurityMonitoringCriticalAssetDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityMonitoringCriticalAssetConfig(assetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityMonitoringCriticalAssetExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "severity", "increase"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
			{
				Config: testAccSecurityMonitoringCriticalAssetConfigUpdated(assetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityMonitoringCriticalAssetExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "query", "source:cloudtrail"),
					resource.TestCheckResourceAttr(resourceName, "severity", "high"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "1"),
				),
			},
		},
	})
}

func testAccSecurityMonitoringCriticalAssetConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_security_monitoring_critical_asset" "test" {
  enabled    = true
  query      = "source:runtime-security-agent"
  rule_query = "type:(log_detection OR signal_correlation OR workload_security OR application_security) ruleId:007-d1a-1f3"
  severity   = "increase"
  tags       = ["test:tf-%s", "team:security"]
}
`, uniq)
}

func testAccSecurityMonitoringCriticalAssetConfigUpdated(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_security_monitoring_critical_asset" "test" {
  enabled    = false
  query      = "source:cloudtrail"
  rule_query = "type:(log_detection OR signal_correlation OR workload_security OR application_security) *"
  severity   = "high"
  tags       = ["test:tf-%s"]
}
`, uniq)
}

func testAccCheckSecurityMonitoringCriticalAssetExists(accProvider *fwprovider.FrameworkProvider, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiClient := accProvider.DatadogApiInstances.GetSecurityMonitoringApiV2()
		auth := accProvider.Auth

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		_, httpResp, err := apiClient.GetSecurityMonitoringCriticalAsset(auth, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error checking security monitoring critical asset existence: %v", err)
		}

		if httpResp.StatusCode != 200 {
			return fmt.Errorf("received status code %d when checking critical asset existence", httpResp.StatusCode)
		}

		return nil
	}
}

func testAccCheckSecurityMonitoringCriticalAssetDestroy(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiClient := accProvider.DatadogApiInstances.GetSecurityMonitoringApiV2()
		auth := accProvider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_security_monitoring_critical_asset" {
				continue
			}

			_, httpResp, err := apiClient.GetSecurityMonitoringCriticalAsset(auth, r.Primary.ID)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					continue
				}
				return fmt.Errorf("error checking if security monitoring critical asset was destroyed: %v", err)
			}

			if httpResp.StatusCode != 404 {
				return errors.New("critical asset still exists")
			}
		}

		return nil
	}
}
