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

func TestAccDatadogAgentlessScanningGcpScanOptions_Basic(t *testing.T) {
	t.Parallel()
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	projectID := "test-project" // Test GCP project ID

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAgentlessScanningGcpScanOptionsDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogAgentlessScanningGcpScanOptionsConfig(projectID, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAgentlessScanningGcpScanOptionsExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_agentless_scanning_gcp_scan_options.test", "gcp_project_id", projectID),
					resource.TestCheckResourceAttr("datadog_agentless_scanning_gcp_scan_options.test", "vuln_containers_os", "true"),
					resource.TestCheckResourceAttr("datadog_agentless_scanning_gcp_scan_options.test", "vuln_host_os", "true"),
				),
			},
		},
	})
}

func TestAccDatadogAgentlessScanningGcpScanOptions_InvalidProjectID(t *testing.T) {
	t.Parallel()
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	projectID := "1invalid-project" // Invalid: starts with a digit

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAgentlessScanningGcpScanOptionsDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckDatadogAgentlessScanningGcpScanOptionsConfig(projectID, true, true),
				ExpectError: regexp.MustCompile("must be a valid GCP project ID"),
			},
		},
	})
}

func TestAccDatadogAgentlessScanningGcpScanOptions_Update(t *testing.T) {
	if !isReplaying() {
		t.Skip("This test only supports replaying - requires GCP project integrated with agentless scanning")
	}
	t.Parallel()
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	projectID := "test-project" // Test GCP project ID

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAgentlessScanningGcpScanOptionsDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogAgentlessScanningGcpScanOptionsConfig(projectID, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAgentlessScanningGcpScanOptionsExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_agentless_scanning_gcp_scan_options.test", "vuln_containers_os", "true"),
					resource.TestCheckResourceAttr("datadog_agentless_scanning_gcp_scan_options.test", "vuln_host_os", "true"),
				),
			},
			{
				Config: testAccCheckDatadogAgentlessScanningGcpScanOptionsConfig(projectID, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAgentlessScanningGcpScanOptionsExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_agentless_scanning_gcp_scan_options.test", "vuln_containers_os", "false"),
					resource.TestCheckResourceAttr("datadog_agentless_scanning_gcp_scan_options.test", "vuln_host_os", "false"),
				),
			},
		},
	})
}

func TestAccDatadogAgentlessScanningGcpScanOptions_Import(t *testing.T) {
	t.Parallel()
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	projectID := "test-project" // Test GCP project ID

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAgentlessScanningGcpScanOptionsDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogAgentlessScanningGcpScanOptionsConfig(projectID, true, true),
			},
			{
				ResourceName:      "datadog_agentless_scanning_gcp_scan_options.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDatadogAgentlessScanningGcpScanOptionsConfig(projectID string, vulnContainers, vulnHost bool) string {
	return fmt.Sprintf(`
resource "datadog_agentless_scanning_gcp_scan_options" "test" {
  gcp_project_id     = "%s"
  vuln_containers_os = %s
  vuln_host_os       = %s
}`, projectID, strconv.FormatBool(vulnContainers), strconv.FormatBool(vulnHost))
}

func testAccCheckDatadogAgentlessScanningGcpScanOptionsExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_agentless_scanning_gcp_scan_options" {
				continue
			}

			_, _, err := apiInstances.GetAgentlessScanningApiV2().GetGcpScanOptions(auth, r.Primary.ID)
			if err != nil {
				return fmt.Errorf("received an error retrieving agentless scanning gcp scan options: %s", err)
			}
		}
		return nil
	}
}

func testAccCheckDatadogAgentlessScanningGcpScanOptionsDestroy(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_agentless_scanning_gcp_scan_options" {
				continue
			}

			projectID := r.Primary.ID

			_, _, err := apiInstances.GetAgentlessScanningApiV2().GetGcpScanOptions(auth, r.Primary.ID)
			if err != nil {
				// If we get an error, assume the resource is gone
				continue
			}

			return fmt.Errorf("agentless scanning gcp scan options %s still exists", projectID)
		}
		return nil
	}
}
