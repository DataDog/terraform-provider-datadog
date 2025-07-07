package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var testAppKeyRegistrationID = "2"

func TestAccDatadogAppKeyRegistrationResource(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resourceName := "datadog_app_key_registration.my_registration"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogAppKeyRegistrationDestroy(providers.frameworkProvider, resourceName),
		Steps: []resource.TestStep{
			{
				Config: testAppKeyRegistrationResourceConfig(testAppKeyRegistrationID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogAppKeyRegistrationExists(providers.frameworkProvider, resourceName),
					resource.TestCheckResourceAttr(resourceName, "id", testAppKeyRegistrationID),
				),
			},
		},
	})
}

func testAppKeyRegistrationResourceConfig(appKeyRegistrationID string) string {
	return fmt.Sprintf(`
	resource "datadog_app_key_registration" "my_registration" {
		id = "%s"
	}`, appKeyRegistrationID)
}

func testAccCheckDatadogAppKeyRegistrationExists(accProvider *fwprovider.FrameworkProvider, id string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := datadogAppKeyRegistrationExistsHelper(auth, s, apiInstances, id); err != nil {
			return err
		}
		return nil
	}
}

func datadogAppKeyRegistrationExistsHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances, id string) error {
	if _, _, err := apiInstances.GetActionConnectionApiV2().GetAppKeyRegistration(ctx, id); err != nil {
		return fmt.Errorf("received an error retrieving app key registration: %s", err)
	}
	return nil
}

func testAccCheckDatadogAppKeyRegistrationDestroy(accProvider *fwprovider.FrameworkProvider, resourceName string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		resource := s.RootModule().Resources[resourceName]
		_, httpRes, err := apiInstances.GetActionConnectionApiV2().GetAppKeyRegistration(auth, resource.Primary.ID)
		if err != nil {
			if httpRes.StatusCode == 404 {
				return nil
			}
			return err
		}

		return fmt.Errorf("app key registration destroy check failed")
	}
}
