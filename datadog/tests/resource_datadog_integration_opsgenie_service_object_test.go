package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDatadogIntegrationOpsgenieServiceObject_Basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)
	reg, _ := regexp.Compile("[^a-zA-Z0-9\u00a0-\uffef.+\\-_]")
	serviceName := reg.ReplaceAllString(uniqueEntityName(ctx, t), "")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogIntegrationOpsgenieServiceDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationOpsgenieServiceObjectConfigCreate(serviceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationOpsgenieServiceObjectExists(accProvider, "datadog_integration_opsgenie_service_object.testing_foo"),
					testAccCheckDatadogIntegrationOpsgenieServiceObjectExists(accProvider, "datadog_integration_opsgenie_service_object.testing_bar"),
					resource.TestCheckResourceAttr(
						"datadog_integration_opsgenie_service_object.testing_foo", "name", serviceName+"_foo"),
					resource.TestCheckResourceAttr(
						"datadog_integration_opsgenie_service_object.testing_foo", "opsgenie_api_key", "00000000-0000-0000-0000-000000000000"),
					resource.TestCheckResourceAttr(
						"datadog_integration_opsgenie_service_object.testing_foo", "region", "us"),
					resource.TestCheckResourceAttr(
						"datadog_integration_opsgenie_service_object.testing_bar", "name", serviceName+"_bar"),
					resource.TestCheckResourceAttr(
						"datadog_integration_opsgenie_service_object.testing_bar", "opsgenie_api_key", "11111111-1111-1111-1111-111111111111"),
					resource.TestCheckResourceAttr(
						"datadog_integration_opsgenie_service_object.testing_bar", "region", "custom"),
					resource.TestCheckResourceAttr(
						"datadog_integration_opsgenie_service_object.testing_bar", "custom_url", "https://example.com/custom"),
				),
			},
			{
				Config: testAccCheckDatadogIntegrationOpsgenieServiceObjectConfigUpdate(serviceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationOpsgenieServiceObjectExists(accProvider, "datadog_integration_opsgenie_service_object.testing_foo"),
					testAccCheckDatadogIntegrationOpsgenieServiceObjectExists(accProvider, "datadog_integration_opsgenie_service_object.testing_bar"),
					resource.TestCheckResourceAttr(
						"datadog_integration_opsgenie_service_object.testing_foo", "name", serviceName+"_foo_updated"),
					resource.TestCheckResourceAttr(
						"datadog_integration_opsgenie_service_object.testing_foo", "opsgenie_api_key", "11111111-1111-1111-1111-111111111111"),
					resource.TestCheckResourceAttr(
						"datadog_integration_opsgenie_service_object.testing_foo", "region", "eu"),
					resource.TestCheckResourceAttr(
						"datadog_integration_opsgenie_service_object.testing_bar", "name", serviceName+"_bar_updated"),
					resource.TestCheckResourceAttr(
						"datadog_integration_opsgenie_service_object.testing_bar", "opsgenie_api_key", "00000000-0000-0000-0000-000000000000"),
					resource.TestCheckResourceAttr(
						"datadog_integration_opsgenie_service_object.testing_bar", "region", "custom"),
					resource.TestCheckResourceAttr(
						"datadog_integration_opsgenie_service_object.testing_bar", "custom_url", "https://example.com/custom/updated"),
				),
			},
		},
	})
}

func testAccCheckDatadogIntegrationOpsgenieServiceDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		if err := datadogIntegrationOpsgenieServiceObjectDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func datadogIntegrationOpsgenieServiceObjectDestroyHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_integration_opsgenie_service_object" {
			continue
		}
		id := r.Primary.ID
		_, httpResponse, err := apiInstances.GetOpsgenieIntegrationApiV2().GetOpsgenieService(ctx, id)

		if err != nil {
			if httpResponse.StatusCode == 404 {
				continue
			}
			return fmt.Errorf("received an error retrieving Opsgenie service %s", err)
		}
		return fmt.Errorf("opsgenie service still exists")
	}
	return nil
}

func testAccCheckDatadogIntegrationOpsgenieServiceObjectExists(accProvider func() (*schema.Provider, error), resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		if err := datadogIntegrationOpsgenieServiceObjectExistsHelper(auth, s, apiInstances, resourceName); err != nil {
			return err
		}
		return nil
	}
}

func datadogIntegrationOpsgenieServiceObjectExistsHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances, resourceName string) error {
	id := s.RootModule().Resources[resourceName].Primary.ID
	if _, _, err := apiInstances.GetOpsgenieIntegrationApiV2().GetOpsgenieService(ctx, id); err != nil {
		return fmt.Errorf("received an error retrieving Opsgenie service %s", err)
	}
	return nil
}

func testAccCheckDatadogIntegrationOpsgenieServiceObjectConfigCreate(uniq string) string {
	return fmt.Sprintf(`
		resource "datadog_integration_opsgenie_service_object" "testing_foo" {
			name = "%s_foo"
			opsgenie_api_key  = "00000000-0000-0000-0000-000000000000"
			region = "us"
		}
		
		resource "datadog_integration_opsgenie_service_object" "testing_bar" {
			name = "%s_bar"
			opsgenie_api_key  = "11111111-1111-1111-1111-111111111111"
			region = "custom"
			custom_url = "https://example.com/custom"
		}
	`, uniq, uniq)

}

func testAccCheckDatadogIntegrationOpsgenieServiceObjectConfigUpdate(uniq string) string {
	return fmt.Sprintf(`
		resource "datadog_integration_opsgenie_service_object" "testing_foo" {
			name = "%s_foo_updated"
			opsgenie_api_key  = "11111111-1111-1111-1111-111111111111"
			region = "eu"
		}
		
		resource "datadog_integration_opsgenie_service_object" "testing_bar" {
			name = "%s_bar_updated"
			opsgenie_api_key  = "00000000-0000-0000-0000-000000000000"
			region = "custom"
			custom_url = "https://example.com/custom/updated"
		}
	`, uniq, uniq)
}
