package test

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

func TestAccDatadogAgentlessScanningAzureScanOptions_Basic(t *testing.T) {
	t.Parallel()
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	subscriptionID := "12345678-1234-1234-1234-123456789012" // Test Azure subscription ID

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAgentlessScanningAzureScanOptionsDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogAgentlessScanningAzureScanOptionsConfig(subscriptionID, true, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAgentlessScanningAzureScanOptionsExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_agentless_scanning_azure_scan_options.test", "azure_subscription_id", subscriptionID),
					resource.TestCheckResourceAttr("datadog_agentless_scanning_azure_scan_options.test", "compliance_host", "true"),
					resource.TestCheckResourceAttr("datadog_agentless_scanning_azure_scan_options.test", "vuln_containers_os", "true"),
					resource.TestCheckResourceAttr("datadog_agentless_scanning_azure_scan_options.test", "vuln_host_os", "true"),
				),
			},
		},
	})
}

func TestAccDatadogAgentlessScanningAzureScanOptions_InvalidSubscriptionID(t *testing.T) {
	t.Parallel()
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	subscriptionID := "not-a-uuid" // Invalid: not a UUID

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAgentlessScanningAzureScanOptionsDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckDatadogAgentlessScanningAzureScanOptionsConfig(subscriptionID, true, true, true),
				ExpectError: regexp.MustCompile("must be a valid Azure subscription ID"),
			},
		},
	})
}

func TestAccDatadogAgentlessScanningAzureScanOptions_Update(t *testing.T) {
	t.Parallel()
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	subscriptionID := "12345678-1234-1234-1234-123456789012" // Test Azure subscription ID

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAgentlessScanningAzureScanOptionsDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogAgentlessScanningAzureScanOptionsConfig(subscriptionID, true, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAgentlessScanningAzureScanOptionsExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_agentless_scanning_azure_scan_options.test", "compliance_host", "true"),
					resource.TestCheckResourceAttr("datadog_agentless_scanning_azure_scan_options.test", "vuln_containers_os", "true"),
					resource.TestCheckResourceAttr("datadog_agentless_scanning_azure_scan_options.test", "vuln_host_os", "true"),
				),
			},
			{
				Config: testAccCheckDatadogAgentlessScanningAzureScanOptionsConfig(subscriptionID, false, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAgentlessScanningAzureScanOptionsExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_agentless_scanning_azure_scan_options.test", "compliance_host", "false"),
					resource.TestCheckResourceAttr("datadog_agentless_scanning_azure_scan_options.test", "vuln_containers_os", "false"),
					resource.TestCheckResourceAttr("datadog_agentless_scanning_azure_scan_options.test", "vuln_host_os", "false"),
				),
			},
		},
	})
}

func TestAccDatadogAgentlessScanningAzureScanOptions_Import(t *testing.T) {
	t.Parallel()
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	subscriptionID := "12345678-1234-1234-1234-123456789012" // Test Azure subscription ID

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAgentlessScanningAzureScanOptionsDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogAgentlessScanningAzureScanOptionsConfig(subscriptionID, true, true, true),
			},
			{
				ResourceName:      "datadog_agentless_scanning_azure_scan_options.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDatadogAgentlessScanningAzureScanOptionsConfig(subscriptionID string, complianceHost, vulnContainers, vulnHost bool) string {
	return fmt.Sprintf(`
resource "datadog_agentless_scanning_azure_scan_options" "test" {
  azure_subscription_id = "%s"
  compliance_host       = %s
  vuln_containers_os    = %s
  vuln_host_os          = %s
}`, subscriptionID, strconv.FormatBool(complianceHost), strconv.FormatBool(vulnContainers), strconv.FormatBool(vulnHost))
}

func testAccCheckDatadogAgentlessScanningAzureScanOptionsExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_agentless_scanning_azure_scan_options" {
				continue
			}

			_, _, err := apiInstances.GetAgentlessScanningApiV2().GetAzureScanOptions(auth, r.Primary.ID)
			if err != nil {
				return fmt.Errorf("received an error retrieving agentless scanning azure scan options: %s", err)
			}
		}
		return nil
	}
}

func testAccCheckDatadogAgentlessScanningAzureScanOptionsDestroy(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_agentless_scanning_azure_scan_options" {
				continue
			}

			subscriptionID := r.Primary.ID

			_, _, err := apiInstances.GetAgentlessScanningApiV2().GetAzureScanOptions(auth, r.Primary.ID)
			if err != nil {
				// If we get an error, assume the resource is gone
				continue
			}

			return fmt.Errorf("agentless scanning azure scan options %s still exists", subscriptionID)
		}
		return nil
	}
}
