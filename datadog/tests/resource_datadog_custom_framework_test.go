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

func TestCustomFramework_create(t *testing.T) {
	t.Parallel()
	handle := fmt.Sprintf("handle-%d", rand.Intn(100000))
	version := fmt.Sprintf("version-%d", rand.Intn(100000))

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	path := "datadog_custom_framework.sample_rules"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogFrameworkDestroy(ctx, providers.frameworkProvider, path, version, handle),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCreateFramework(version, handle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "handle", handle),
					resource.TestCheckResourceAttr(path, "version", version),
					resource.TestCheckResourceAttr(path, "name", "new-framework-terraform"),
					resource.TestCheckResourceAttr(path, "description", "test description"),
					resource.TestCheckResourceAttr(path, "icon_url", "test url"),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*", map[string]string{
						"name": "requirement1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*.controls.*", map[string]string{
						"name": "control1",
					}),
					resource.TestCheckTypeSetElemAttr(path, "requirements.*.controls.*.rules_id.*", "def-000-be9"),
				),
			},
			{
				Config: testAccCheckDatadogCreateFrameworkWithMultipleRequirements(version, handle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "handle", handle),
					resource.TestCheckResourceAttr(path, "version", version),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*", map[string]string{
						"name": "compliance-requirement",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*", map[string]string{
						"name": "security-requirement",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*.controls.*", map[string]string{
						"name": "compliance-control",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*.controls.*", map[string]string{
						"name": "security-control",
					}),
					resource.TestCheckTypeSetElemAttr(path, "requirements.*.controls.*.rules_id.*", "def-000-be9"),
					resource.TestCheckTypeSetElemAttr(path, "requirements.*.controls.*.rules_id.*", "def-000-cea"),
				),
			},
		},
	})
}

func TestCustomFramework_delete(t *testing.T) {
	t.Parallel()
	handle := fmt.Sprintf("handle-%d", rand.Intn(100000))
	version := fmt.Sprintf("version-%d", rand.Intn(100000))

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	path := "datadog_custom_framework.sample_rules"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogFrameworkDestroy(ctx, providers.frameworkProvider, path, version, handle),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCreateFramework(version, handle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "handle", handle),
					resource.TestCheckResourceAttr(path, "version", version),
				),
			},
			{
				ResourceName:      path,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestCustomFramework_updateMultipleRequirements(t *testing.T) {
	t.Parallel()
	handle := fmt.Sprintf("handle-%d", rand.Intn(100000))
	version := fmt.Sprintf("version-%d", rand.Intn(100000))

	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	path := "datadog_custom_framework.sample_rules"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogFrameworkDestroy(ctx, providers.frameworkProvider, path, version, handle),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogCreateFrameworkWithMultipleRequirements(version, handle),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(path, "handle", handle),
					resource.TestCheckResourceAttr(path, "version", version),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*", map[string]string{
						"name": "compliance-requirement",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*", map[string]string{
						"name": "security-requirement",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*.controls.*", map[string]string{
						"name": "compliance-control",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(path, "requirements.*.controls.*", map[string]string{
						"name": "security-control",
					}),
					resource.TestCheckTypeSetElemAttr(path, "requirements.*.controls.*.rules_id.*", "def-000-be9"),
					resource.TestCheckTypeSetElemAttr(path, "requirements.*.controls.*.rules_id.*", "def-000-cea"),
				),
			},
		},
	})
}

func testAccCheckDatadogCreateFrameworkWithMultipleRequirements(version string, handle string) string {
	return fmt.Sprintf(`
		resource "datadog_custom_framework" "sample_rules" {
			version       = "%s"
			handle        = "%s"
			name          = "new-name"
			description   = "test description"
			icon_url      = "test url"
			requirements  {
				name = "security-requirement"
				controls {
					name = "security-control"
					rules_id = ["def-000-cea", "def-000-be9"]
				}
			}
			requirements  {
				name = "compliance-requirement"
				controls {
					name = "compliance-control"
					rules_id = ["def-000-be9", "def-000-cea"]
				}
			}
		}
	`, version, handle)
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
