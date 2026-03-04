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

func TestAccDatadogAgentlessScanningAwsScanOptions_Basic(t *testing.T) {
	t.Parallel()
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	accountID := "123456789012" // Test AWS account ID

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAgentlessScanningAwsScanOptionsDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogAgentlessScanningAwsScanOptionsConfig(accountID, true, false, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAgentlessScanningAwsScanOptionsExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_agentless_scanning_aws_scan_options.test", "aws_account_id", accountID),
					resource.TestCheckResourceAttr("datadog_agentless_scanning_aws_scan_options.test", "lambda", "true"),
					resource.TestCheckResourceAttr("datadog_agentless_scanning_aws_scan_options.test", "sensitive_data", "false"),
					resource.TestCheckResourceAttr("datadog_agentless_scanning_aws_scan_options.test", "vuln_containers_os", "true"),
					resource.TestCheckResourceAttr("datadog_agentless_scanning_aws_scan_options.test", "vuln_host_os", "true"),
				),
			},
		},
	})
}

func TestAccDatadogAgentlessScanningAwsScanOptions_InvalidAccountID(t *testing.T) {
	t.Parallel()
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	accountID := "1nvalidaccid"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAgentlessScanningAwsScanOptionsDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckDatadogAgentlessScanningAwsScanOptionsConfig(accountID, true, false, true, true),
				ExpectError: regexp.MustCompile("must be a valid AWS account ID"),
			},
		},
	})
}

func TestAccDatadogAgentlessScanningAwsScanOptions_Update(t *testing.T) {
	t.Parallel()
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	accountID := "123456789012" // Test AWS account ID

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAgentlessScanningAwsScanOptionsDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogAgentlessScanningAwsScanOptionsConfig(accountID, true, false, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAgentlessScanningAwsScanOptionsExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_agentless_scanning_aws_scan_options.test", "lambda", "true"),
					resource.TestCheckResourceAttr("datadog_agentless_scanning_aws_scan_options.test", "sensitive_data", "false"),
				),
			},
			{
				Config: testAccCheckDatadogAgentlessScanningAwsScanOptionsConfig(accountID, false, true, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAgentlessScanningAwsScanOptionsExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_agentless_scanning_aws_scan_options.test", "lambda", "false"),
					resource.TestCheckResourceAttr("datadog_agentless_scanning_aws_scan_options.test", "sensitive_data", "true"),
					resource.TestCheckResourceAttr("datadog_agentless_scanning_aws_scan_options.test", "vuln_containers_os", "false"),
					resource.TestCheckResourceAttr("datadog_agentless_scanning_aws_scan_options.test", "vuln_host_os", "false"),
				),
			},
		},
	})
}

func TestAccDatadogAgentlessScanningAwsScanOptions_Import(t *testing.T) {
	t.Parallel()
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	accountID := "123456789012" // Test AWS account ID

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAgentlessScanningAwsScanOptionsDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogAgentlessScanningAwsScanOptionsConfig(accountID, true, false, true, true),
			},
			{
				ResourceName:      "datadog_agentless_scanning_aws_scan_options.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDatadogAgentlessScanningAwsScanOptionsConfig(accountID string, lambda, sensitiveData, vulnContainers, vulnHost bool) string {
	return fmt.Sprintf(`
resource "datadog_agentless_scanning_aws_scan_options" "test" {
  aws_account_id     = "%s"
  lambda             = %s
  sensitive_data     = %s
  vuln_containers_os = %s
  vuln_host_os       = %s
}`, accountID, strconv.FormatBool(lambda), strconv.FormatBool(sensitiveData), strconv.FormatBool(vulnContainers), strconv.FormatBool(vulnHost))
}

func testAccCheckDatadogAgentlessScanningAwsScanOptionsExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_agentless_scanning_aws_scan_options" {
				continue
			}

			accountID := r.Primary.ID

			// Check if the resource exists by listing all scan options and finding this one
			awsScanOptionsListResponse, _, err := apiInstances.GetAgentlessScanningApiV2().ListAwsScanOptions(auth)
			if err != nil {
				return fmt.Errorf("received an error retrieving agentless scanning aws scan options: %s", err)
			}

			found := false
			for _, scanOption := range awsScanOptionsListResponse.GetData() {
				if scanOption.GetId() == accountID {
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("agentless scanning aws scan options %s not found", accountID)
			}
		}
		return nil
	}
}

func testAccCheckDatadogAgentlessScanningAwsScanOptionsDestroy(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type != "datadog_agentless_scanning_aws_scan_options" {
				continue
			}

			accountID := r.Primary.ID

			// Check if the resource still exists by listing all scan options
			awsScanOptionsListResponse, _, err := apiInstances.GetAgentlessScanningApiV2().ListAwsScanOptions(auth)
			if err != nil {
				// If we get an error, assume the resource is gone
				continue
			}

			for _, scanOption := range awsScanOptionsListResponse.GetData() {
				if scanOption.GetId() == accountID {
					return fmt.Errorf("agentless scanning aws scan options %s still exists", accountID)
				}
			}
		}
		return nil
	}
}
