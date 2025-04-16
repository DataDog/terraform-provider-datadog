package test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

func TestCustomFramework_bgasic(t *testing.T) {
	t.Parallel()
	handle := fmt.Sprintf("handle-%d", rand.Intn(100000))
	version := "1.0"

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	path := "datadog_custom_framework.sample_rules"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogFrameworkDestroy(ctx, providers.frameworkProvider, path, version, handle),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCreateFramework(handle, version),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "handle", handle),
					resource.TestCheckResourceAttr(path, "version", version),
					resource.TestCheckResourceAttr(path, "name", "test-framework"),
					resource.TestCheckResourceAttr(path, "description", "test description"),
					resource.TestCheckResourceAttr(path, "icon_url", "test url"),
					resource.TestCheckResourceAttr(path, "requirements.#", "1"),
					resource.TestCheckResourceAttr(path, "requirements.0.name", "requirement1"),
					resource.TestCheckResourceAttr(path, "requirements.0.controls.#", "1"),
					resource.TestCheckResourceAttr(path, "requirements.0.controls.0.name", "control1"),
					resource.TestCheckResourceAttr(path, "requirements.0.controls.0.rules_id.#", "1"),
					resource.TestCheckResourceAttr(path, "requirements.0.controls.0.rules_id.0", "def-000-be9"),
					resource.TestCheckResourceAttr(path, "requirements.0.controls.0.rules_id.#", "1"),
				),
			},
		},
	})
}

func testAccCheckDatadogCreateFramework(version string, handle string) string {
	return fmt.Sprintf(`
		resource "datadog_custom_framework" "sample_rules" {
			version       = "%s"
			handle        = "%s"
			name          = "new-framework-terraform"
			description   = "test description"
			icon_url      = "test url"
			requirements  {
				name = "requirement1"
				controls {
					name = "control1"
					rules_id = ["def-000-be9"]
				}
			}
		}
	`, version, handle)
}

func testAccCheckDatadogFrameworkDestroy(ctx context.Context, accProvider *fwprovider.FrameworkProvider, resourceName string, version string, handle string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		resource := s.RootModule().Resources[resourceName]
		handle := resource.Primary.Attributes["handle"]
		version := resource.Primary.Attributes["version"]
		_, httpRes, err := apiInstances.GetSecurityMonitoringApiV2().RetrieveCustomFramework(ctx, handle, version)
		if err != nil {
			if httpRes.StatusCode == 400 {
				return nil
			}
			return err
		}

		return fmt.Errorf("framework destroy check failed")
	}
}
