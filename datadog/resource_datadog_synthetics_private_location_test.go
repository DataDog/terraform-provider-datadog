package datadog

import (
	"fmt"
	"github.com/jonboulle/clockwork"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDatadogSyntheticsPrivateLocation_importBasic(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	privateLocationName := uniqueEntityName(clock, t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsPrivateLocationIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			{
				Config: createSyntheticsPrivateLocationConfig(privateLocationName),
			},
			{
				ResourceName:      "datadog_synthetics_private_location.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDatadogSyntheticsPrivateLocation_Basic(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsPrivateLocationIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsPrivateLocationStep(accProvider, clock, t),
		},
	})
}

func TestAccDatadogSyntheticsPrivateLocation_Updated(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testSyntheticsPrivateLocationIsDestroyed(accProvider),
		Steps: []resource.TestStep{
			createSyntheticsPrivateLocationStep(accProvider, clock, t),
			updateSyntheticsPrivateLocationStep(accProvider, clock, t),
		},
	})
}

func createSyntheticsPrivateLocationStep(accProvider *schema.Provider, clock clockwork.FakeClock, t *testing.T) resource.TestStep {
	privateLocationName := uniqueEntityName(clock, t)
	return resource.TestStep{
		Config: createSyntheticsPrivateLocationConfig(privateLocationName),
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
		),
	}
}

func createSyntheticsPrivateLocationConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_private_location" "foo" {
	name = "%s"
	description = "a private location"
	tags = ["foo:bar", "baz"]
}`, uniq)
}

func updateSyntheticsPrivateLocationStep(accProvider *schema.Provider, clock clockwork.FakeClock, t *testing.T) resource.TestStep {
	privateLocationName := uniqueEntityName(clock, t) + "_updated"
	return resource.TestStep{
		Config: updateSyntheticsPrivateLocationConfig(privateLocationName),
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
		),
	}
}

func updateSyntheticsPrivateLocationConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_synthetics_private_location" "foo" {
	name = "%s"
	description = "an updated private location"
	tags = ["foo:bar", "baz", "env:test"]
}`, uniq)
}

func testSyntheticsPrivateLocationExists(accProvider *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		for _, r := range s.RootModule().Resources {
			if _, _, err := datadogClientV1.SyntheticsApi.GetPrivateLocation(authV1, r.Primary.ID).Execute(); err != nil {
				return fmt.Errorf("received an error retrieving synthetics private location %s", err)
			}
		}
		return nil
	}
}

func testSyntheticsPrivateLocationIsDestroyed(accProvider *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		for _, r := range s.RootModule().Resources {
			if _, httpresp, err := datadogClientV1.SyntheticsApi.GetPrivateLocation(authV1, r.Primary.ID).Execute(); err != nil {
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
