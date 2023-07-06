package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
)

func TestAccDatadogSyntheticsPrivateLocation_importBasic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	privateLocationName := uniqueEntityName(ctx, t)
	roleName := uniqueEntityName(ctx, t) + "_role"
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsPrivateLocationIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			{
				Config: createSyntheticsPrivateLocationConfig(privateLocationName, roleName),
			},
			{
				ResourceName:            "datadog_synthetics_private_location.foo",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"config"},
			},
		},
	})
}

func TestAccDatadogSyntheticsPrivateLocation_Basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsPrivateLocationIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsPrivateLocationStep(ctx, accProvider, t),
		},
	})
}

func TestAccDatadogSyntheticsPrivateLocation_Updated(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testSyntheticsPrivateLocationIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsPrivateLocationStep(ctx, accProvider, t),
			updateSyntheticsPrivateLocationStep(ctx, accProvider, t),
		},
	})
}

func createSyntheticsPrivateLocationStep(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T) resource.TestStep {
	privateLocationName := uniqueEntityName(ctx, t)
	roleName := uniqueEntityName(ctx, t) + "_role"
	return resource.TestStep{
		Config: createSyntheticsPrivateLocationConfig(privateLocationName, roleName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsPrivateLocationExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_private_location.foo", "name", privateLocationName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_private_location.foo", "description", "a private location"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_private_location.foo", "tags.#", "2"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_private_location.foo", "tags.0", "foo:bar"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_private_location.foo", "tags.1", "baz"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_private_location.foo", "metadata.0.restricted_roles.#", "1"),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_private_location.foo", "config"),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_private_location.foo", "id"),
		),
	}
}

func createSyntheticsPrivateLocationConfig(uniqPrivateLocation string, uniqRole string) string {
	return fmt.Sprintf(`
resource "datadog_role" "rbac_role" {
	name = "%s"
}

resource "datadog_synthetics_private_location" "foo" {
	name = "%s"
	description = "a private location"
	tags = ["foo:bar", "baz"]
	metadata {
		restricted_roles = ["${datadog_role.rbac_role.id}"]
	}
}`, uniqRole, uniqPrivateLocation)
}

func updateSyntheticsPrivateLocationStep(ctx context.Context, accProvider func() (*schema.Provider, error), t *testing.T) resource.TestStep {
	privateLocationName := uniqueEntityName(ctx, t) + "_updated"
	roleName := uniqueEntityName(ctx, t) + "_role_updated"
	return resource.TestStep{
		Config: updateSyntheticsPrivateLocationConfig(privateLocationName, roleName),
		Check: resource.ComposeTestCheckFunc(
			testSyntheticsPrivateLocationExists(accProvider),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_private_location.foo", "name", privateLocationName),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_private_location.foo", "description", "an updated private location"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_private_location.foo", "tags.#", "3"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_private_location.foo", "tags.0", "foo:bar"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_private_location.foo", "tags.1", "baz"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_private_location.foo", "tags.2", "env:test"),
			resource.TestCheckResourceAttr(
				"datadog_synthetics_private_location.foo", "metadata.0.restricted_roles.#", "1"),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_private_location.foo", "config"),
			resource.TestCheckResourceAttrSet(
				"datadog_synthetics_private_location.foo", "id"),
		),
	}
}

func updateSyntheticsPrivateLocationConfig(uniqPrivateLocation string, uniqRole string) string {
	return fmt.Sprintf(`
resource "datadog_role" "rbac_role" {
	name = "%s"
}

resource "datadog_synthetics_private_location" "foo" {
	name = "%s"
	description = "an updated private location"
	tags = ["foo:bar", "baz", "env:test"]
	metadata {
		restricted_roles = ["${datadog_role.rbac_role.id}"]
	}
}`, uniqRole, uniqPrivateLocation)
}

func testSyntheticsPrivateLocationExists(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type == "datadog_role" {
				continue
			}

			if _, _, err := apiInstances.GetSyntheticsApiV1().GetPrivateLocation(auth, r.Primary.ID); err != nil {
				return fmt.Errorf("received an error retrieving synthetics private location %s", err)
			}
		}
		return nil
	}
}

func testSyntheticsPrivateLocationIsDestroyed(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		for _, r := range s.RootModule().Resources {
			if r.Type == "datadog_role" {
				continue
			}

			if _, httpresp, err := apiInstances.GetSyntheticsApiV1().GetPrivateLocation(auth, r.Primary.ID); err != nil {
				if httpresp != nil && httpresp.StatusCode == 404 {
					continue
				}
				return fmt.Errorf("received an error retrieving synthetics private location %s", err)
			}
			return fmt.Errorf("synthetics private location still exists")
		}
		return nil
	}
}
