package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDatadogRUMApplication_Basic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	appName := uniqueEntityName(ctx, t)
	appNameUpdated := fmt.Sprintf("%s-updated", appName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogRUMApplicationDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogRUMApplicationConfig(appName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRUMApplicationExists(providers.frameworkProvider, "datadog_rum_application.some_app"),
					resource.TestCheckResourceAttr(
						"datadog_rum_application.some_app", "name", appName),
					resource.TestCheckResourceAttr(
						"datadog_rum_application.some_app", "type", "browser"),
				),
			},
			{
				Config: testAccCheckDatadogRUMApplicationConfig(appNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogRUMApplicationExists(providers.frameworkProvider, "datadog_rum_application.some_app"),
					resource.TestCheckResourceAttr(
						"datadog_rum_application.some_app", "name", appNameUpdated),
				),
			},
		},
	})
}

func testAccCheckDatadogRUMApplicationConfig(uniq string) string {
	return fmt.Sprintf(`
  resource "datadog_rum_application" "some_app" {
    name = "%s"
    type = "browser"
  }
    `, uniq)
}

func testAccCheckDatadogRUMApplicationExists(accProvider *fwprovider.FrameworkProvider, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		for _, r := range s.RootModule().Resources {
			id := r.Primary.ID
			if _, httpresp, err := apiInstances.GetRumApiV2().GetRUMApplication(auth, id); err != nil {
				return utils.TranslateClientError(err, httpresp, "error checking rum application existence")
			}
		}
		return nil
	}
}

func testAccCheckDatadogRUMApplicationDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		apps, _, err := apiInstances.GetRumApiV2().GetRUMApplications(auth)
		if err != nil {
			return fmt.Errorf("failed to get rum applications")
		}

		for _, r := range s.RootModule().Resources {
			id := r.Primary.ID

			// It is still possible to retrieve a rum application using the application id after it is "deleted",
			// so we just check that it is removed from the list of all rum applications
			for _, app := range apps.Data {
				if app.Attributes.ApplicationId == id {
					return fmt.Errorf("rum application with id %s still exists", id)
				}
			}
		}

		return nil
	}
}
